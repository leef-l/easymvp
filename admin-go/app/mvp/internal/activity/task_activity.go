package activity

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

const mvpTaskActiveWindow = 6 * time.Minute

type TaskSnapshot struct {
	LastActiveAt      *gtime.Time
	IsActuallyWorking bool
	Stalled           bool
}

type ProjectSummary struct {
	LastActiveAt       *gtime.Time
	IsActuallyWorking  bool
	ActiveRunningTasks int
	StalledTaskCount   int
}

type taskRecord struct {
	ID             int64       `orm:"id"`
	Status         string      `orm:"status"`
	ConversationID int64       `orm:"conversation_id"`
	StartedAt      *gtime.Time `orm:"started_at"`
	UpdatedAt      *gtime.Time `orm:"updated_at"`
}

type conversationChunkTime struct {
	ConversationID int64       `orm:"conversation_id"`
	LastChunkAt    *gtime.Time `orm:"last_chunk_at"`
}

type conversationMessageTime struct {
	ConversationID int64       `orm:"conversation_id"`
	LastMessageAt  *gtime.Time `orm:"last_message_at"`
}

func TouchTaskActivity(ctx context.Context, taskID int64) {
	if taskID <= 0 {
		return
	}
	_, _ = g.DB().Ctx(ctx).Model("mvp_task").
		Where("id", taskID).
		Where("deleted_at IS NULL").
		Data("updated_at", gtime.Now()).
		Update()
}

func TouchConversationActivity(ctx context.Context, conversationID int64) {
	if conversationID <= 0 {
		return
	}
	_, _ = g.DB().Ctx(ctx).Model("mvp_conversation").
		Where("id", conversationID).
		Where("deleted_at IS NULL").
		Data("updated_at", gtime.Now()).
		Update()
}

func TouchMessageActivity(ctx context.Context, messageID int64) {
	if messageID <= 0 {
		return
	}
	_, _ = g.DB().Ctx(ctx).Model("mvp_message").
		Where("id", messageID).
		Where("deleted_at IS NULL").
		Data("updated_at", gtime.Now()).
		Update()
}

func LoadTaskSnapshots(ctx context.Context, taskIDs []int64) (map[int64]TaskSnapshot, error) {
	result := make(map[int64]TaskSnapshot, len(taskIDs))
	if len(taskIDs) == 0 {
		return result, nil
	}

	var tasks []*taskRecord
	err := g.DB().Ctx(ctx).Model("mvp_task").
		Fields("id, status, conversation_id, started_at, updated_at").
		WhereIn("id", taskIDs).
		Where("deleted_at IS NULL").
		Scan(&tasks)
	if err != nil {
		return nil, err
	}

	conversationIDs := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		if task != nil && task.ConversationID > 0 {
			conversationIDs = append(conversationIDs, task.ConversationID)
		}
	}

	chunkMap, err := loadConversationChunkTimes(ctx, conversationIDs)
	if err != nil {
		return nil, err
	}
	messageMap, err := loadConversationMessageTimes(ctx, conversationIDs)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task == nil {
			continue
		}
		lastActiveAt := maxGTime(
			task.UpdatedAt,
			task.StartedAt,
			chunkMap[task.ConversationID],
			messageMap[task.ConversationID],
		)
		isActuallyWorking := task.Status == "running" && isRecent(lastActiveAt, mvpTaskActiveWindow)
		stalled := task.Status == "running" && !isActuallyWorking
		result[task.ID] = TaskSnapshot{
			LastActiveAt:      lastActiveAt,
			IsActuallyWorking: isActuallyWorking,
			Stalled:           stalled,
		}
	}

	return result, nil
}

func LoadProjectSummary(ctx context.Context, projectID int64) (*ProjectSummary, error) {
	var rows []*struct {
		ID int64 `orm:"id"`
	}
	if err := g.DB().Ctx(ctx).Model("mvp_task").
		Fields("id").
		Where("project_id", projectID).
		Where("status", "running").
		Where("deleted_at IS NULL").
		Scan(&rows); err != nil {
		return nil, err
	}

	taskIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.ID <= 0 {
			continue
		}
		taskIDs = append(taskIDs, row.ID)
	}

	snapshots, err := LoadTaskSnapshots(ctx, taskIDs)
	if err != nil {
		return nil, err
	}

	summary := &ProjectSummary{}
	for _, snapshot := range snapshots {
		summary.LastActiveAt = maxGTime(summary.LastActiveAt, snapshot.LastActiveAt)
		if snapshot.IsActuallyWorking {
			summary.ActiveRunningTasks++
		}
		if snapshot.Stalled {
			summary.StalledTaskCount++
		}
	}
	summary.IsActuallyWorking = summary.ActiveRunningTasks > 0
	return summary, nil
}

func loadConversationChunkTimes(ctx context.Context, conversationIDs []int64) (map[int64]*gtime.Time, error) {
	result := make(map[int64]*gtime.Time, len(conversationIDs))
	if len(conversationIDs) == 0 {
		return result, nil
	}

	var rows []*conversationChunkTime
	err := g.DB().Ctx(ctx).Model("mvp_message_chunk mc").
		LeftJoin("mvp_message m", "m.id = mc.message_id").
		Fields("m.conversation_id, MAX(mc.created_at) AS last_chunk_at").
		WhereIn("m.conversation_id", conversationIDs).
		Where("m.deleted_at IS NULL").
		Group("m.conversation_id").
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		result[row.ConversationID] = row.LastChunkAt
	}
	return result, nil
}

func loadConversationMessageTimes(ctx context.Context, conversationIDs []int64) (map[int64]*gtime.Time, error) {
	result := make(map[int64]*gtime.Time, len(conversationIDs))
	if len(conversationIDs) == 0 {
		return result, nil
	}

	var rows []*conversationMessageTime
	err := g.DB().Ctx(ctx).Model("mvp_message").
		Fields("conversation_id, MAX(updated_at) AS last_message_at").
		WhereIn("conversation_id", conversationIDs).
		Where("deleted_at IS NULL").
		Group("conversation_id").
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		result[row.ConversationID] = row.LastMessageAt
	}
	return result, nil
}

func maxGTime(items ...*gtime.Time) *gtime.Time {
	var latest *gtime.Time
	for _, item := range items {
		if item == nil {
			continue
		}
		if latest == nil || item.TimestampMilli() > latest.TimestampMilli() {
			latest = item
		}
	}
	return latest
}

func isRecent(value *gtime.Time, window time.Duration) bool {
	if value == nil {
		return false
	}
	age := gtime.Now().TimestampMilli() - value.TimestampMilli()
	return age >= 0 && time.Duration(age)*time.Millisecond <= window
}
