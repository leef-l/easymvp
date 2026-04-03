package task

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/ai/internal/model"
)

const (
	aiTaskActiveWindow   = 3 * time.Minute
	aiTaskTouchInterval  = 2 * time.Second
)

type aiTaskLogTime struct {
	TaskID    int64       `orm:"task_id"`
	LastLogAt *gtime.Time `orm:"last_log_at"`
}

type activityBufferWriter struct {
	taskID    int64
	lastTouch time.Time
	buf       bytes.Buffer
	mu        sync.Mutex
}

func newActivityBufferWriter(taskID int64) *activityBufferWriter {
	return &activityBufferWriter{taskID: taskID}
}

func (w *activityBufferWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err := w.buf.Write(p)
	if n > 0 {
		now := time.Now()
		if w.lastTouch.IsZero() || now.Sub(w.lastTouch) >= aiTaskTouchInterval {
			w.lastTouch = now
			go touchAITaskActivity(context.Background(), w.taskID)
		}
	}
	return n, err
}

func (w *activityBufferWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

func touchAITaskActivity(ctx context.Context, taskID int64) {
	if taskID <= 0 {
		return
	}
	_, _ = g.DB().Ctx(ctx).Model("ai_task").
		Where("id", taskID).
		Where("deleted_at IS NULL").
		Data("updated_at", gtime.Now()).
		Update()
}

func buildAITaskActivity(status string, startedAt, updatedAt, lastLogAt *gtime.Time) (*gtime.Time, bool, bool) {
	lastActiveAt := latestTime(lastLogAt, updatedAt, startedAt)
	if status != "running" {
		return lastActiveAt, false, false
	}
	if !isRecentTime(lastActiveAt, aiTaskActiveWindow) {
		return lastActiveAt, false, true
	}
	return lastActiveAt, true, false
}

func latestTime(times ...*gtime.Time) *gtime.Time {
	var latest *gtime.Time
	for _, item := range times {
		if item == nil {
			continue
		}
		if latest == nil || item.TimestampMilli() > latest.TimestampMilli() {
			latest = item
		}
	}
	return latest
}

func isRecentTime(value *gtime.Time, window time.Duration) bool {
	if value == nil {
		return false
	}
	age := gtime.Now().TimestampMilli() - value.TimestampMilli()
	return age >= 0 && time.Duration(age)*time.Millisecond <= window
}

func (s *sTask) enrichTaskDetailActivity(ctx context.Context, out *model.TaskDetailOutput) error {
	if out == nil || out.ID == 0 {
		return nil
	}
	logTimeMap, err := s.fetchTaskLogTimes(ctx, []int64{int64(out.ID)})
	if err != nil {
		return err
	}
	lastActiveAt, isActuallyWorking, stalled := buildAITaskActivity(
		out.Status,
		out.StartedAt,
		out.UpdatedAt,
		logTimeMap[int64(out.ID)],
	)
	out.LastActiveAt = lastActiveAt
	out.IsActuallyWorking = isActuallyWorking
	out.Stalled = stalled
	return nil
}

func (s *sTask) enrichTaskListActivity(ctx context.Context, list []*model.TaskListOutput) error {
	if len(list) == 0 {
		return nil
	}
	taskIDs := make([]int64, 0, len(list))
	for _, item := range list {
		if item == nil || item.ID == 0 {
			continue
		}
		taskIDs = append(taskIDs, int64(item.ID))
	}
	logTimeMap, err := s.fetchTaskLogTimes(ctx, taskIDs)
	if err != nil {
		return err
	}
	for _, item := range list {
		if item == nil || item.ID == 0 {
			continue
		}
		lastActiveAt, isActuallyWorking, stalled := buildAITaskActivity(
			item.Status,
			item.StartedAt,
			item.UpdatedAt,
			logTimeMap[int64(item.ID)],
		)
		item.LastActiveAt = lastActiveAt
		item.IsActuallyWorking = isActuallyWorking
		item.Stalled = stalled
	}
	return nil
}

func (s *sTask) fetchTaskLogTimes(ctx context.Context, taskIDs []int64) (map[int64]*gtime.Time, error) {
	result := make(map[int64]*gtime.Time, len(taskIDs))
	if len(taskIDs) == 0 {
		return result, nil
	}

	var rows []*aiTaskLogTime
	err := g.DB().Ctx(ctx).Model("ai_task_log").
		Fields("task_id, MAX(created_at) AS last_log_at").
		WhereIn("task_id", taskIDs).
		Group("task_id").
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		result[row.TaskID] = row.LastLogAt
	}
	return result, nil
}
