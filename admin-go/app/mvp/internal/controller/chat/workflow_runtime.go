package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/workflow/autonomy"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

const projectRuntimeSnapshotFreshWindow = 2 * time.Minute
const projectRuntimeTaskActiveWindow = 3 * time.Minute

type projectRuntimeTaskStat struct {
	WorkflowRunID  int64 `orm:"workflow_run_id"`
	TotalTasks     int   `orm:"total_tasks"`
	CompletedTasks int   `orm:"completed_tasks"`
	FailedTasks    int   `orm:"failed_tasks"`
	RunningTasks   int   `orm:"running_tasks"`
}

type projectRuntimeTaskActivityRow struct {
	WorkflowRunID int64       `orm:"workflow_run_id"`
	Status        string      `orm:"status"`
	BatchNo       int         `orm:"batch_no"`
	HeartbeatAt   *gtime.Time `orm:"heartbeat_at"`
	StartedAt     *gtime.Time `orm:"started_at"`
	CompletedAt   *gtime.Time `orm:"completed_at"`
	UpdatedAt     *gtime.Time `orm:"updated_at"`
}

type projectRuntimeTaskActivityStat struct {
	LastActiveAt       *gtime.Time
	ActiveBatch        int
	ActiveRunningTasks int
	StalledTaskCount   int
}

type projectRuntimeSnapshot struct {
	CreatedAt *gtime.Time
	Situation autonomy.Situation
}

type projectRuntimeLatestID struct {
	ProjectID int64 `orm:"project_id"`
	ID        int64 `orm:"id"`
}

type workflowRuntimeSnapshotLatestID struct {
	WorkflowRunID int64 `orm:"workflow_run_id"`
	ID            int64 `orm:"id"`
}

func loadLatestWorkflowRuns(ctx context.Context, projectIDs []int64) (map[int64]gdb.Record, error) {
	result := make(map[int64]gdb.Record, len(projectIDs))
	if len(projectIDs) == 0 {
		return result, nil
	}

	records, err := repo.NewWorkflowRunRepo().ListLatestByProjects(ctx, projectIDs, "wr.id", "wr.project_id", "wr.current_stage", "wr.status")
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return result, nil
	}
	for _, run := range records {
		record := mapToDBRecord(run)
		result[record["project_id"].Int64()] = record
	}
	return result, nil
}

func loadLatestSituationSnapshots(ctx context.Context, workflowRunIDs []int64) (map[int64]*projectRuntimeSnapshot, error) {
	result := make(map[int64]*projectRuntimeSnapshot, len(workflowRunIDs))
	if len(workflowRunIDs) == 0 {
		return result, nil
	}

	snapshots, err := repo.NewSituationSnapshotRepo().ListLatestByWorkflowRunIDs(ctx, workflowRunIDs, "ss.id", "ss.workflow_run_id", "ss.snapshot_data", "ss.created_at")
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return result, nil
	}

	for _, item := range snapshots {
		record := mapToDBRecord(item)
		var sit autonomy.Situation
		if err := json.Unmarshal([]byte(record["snapshot_data"].String()), &sit); err != nil {
			continue
		}
		result[record["workflow_run_id"].Int64()] = &projectRuntimeSnapshot{
			CreatedAt: record["created_at"].GTime(),
			Situation: sit,
		}
	}
	return result, nil
}

func loadTaskStats(ctx context.Context, workflowRunIDs []int64) (map[int64]projectRuntimeTaskStat, error) {
	result := make(map[int64]projectRuntimeTaskStat, len(workflowRunIDs))
	if len(workflowRunIDs) == 0 {
		return result, nil
	}

	rows, err := repo.NewDomainTaskRepo().ListStatRowsByWorkflowRunIDs(ctx, workflowRunIDs)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		record := mapToDBRecord(row)
		workflowRunID := record["workflow_run_id"].Int64()
		result[workflowRunID] = projectRuntimeTaskStat{
			WorkflowRunID:  workflowRunID,
			TotalTasks:     record["total_tasks"].Int(),
			CompletedTasks: record["completed_tasks"].Int(),
			FailedTasks:    record["failed_tasks"].Int(),
			RunningTasks:   record["running_tasks"].Int(),
		}
	}
	return result, nil
}

func loadTaskActivityStats(ctx context.Context, workflowRunIDs []int64) (map[int64]projectRuntimeTaskActivityStat, error) {
	result := make(map[int64]projectRuntimeTaskActivityStat, len(workflowRunIDs))
	if len(workflowRunIDs) == 0 {
		return result, nil
	}

	rows, err := repo.NewDomainTaskRepo().ListByWorkflowRunIDs(ctx, workflowRunIDs, "workflow_run_id", "status", "batch_no", "heartbeat_at", "started_at", "completed_at", "updated_at")
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		record := mapToDBRecord(row)
		workflowRunID := record["workflow_run_id"].Int64()
		stat := result[workflowRunID]
		heartbeatAt := normalizeDBUTCGTime(record["heartbeat_at"].GTime())
		completedAt := normalizeDBUTCGTime(record["completed_at"].GTime())
		startedAt := normalizeDBUTCGTime(record["started_at"].GTime())
		updatedAt := normalizeDBUTCGTime(record["updated_at"].GTime())

		stat.LastActiveAt = latestNonNilTime(stat.LastActiveAt, heartbeatAt, completedAt, startedAt, updatedAt)
		if record["status"].String() != "running" {
			result[workflowRunID] = stat
			continue
		}

		lastRunningAt := latestNonNilTime(heartbeatAt, startedAt, updatedAt)
		if isRecentGTime(lastRunningAt, projectRuntimeTaskActiveWindow) {
			stat.ActiveRunningTasks++
			batchNo := record["batch_no"].Int()
			if stat.ActiveBatch == 0 || (batchNo > 0 && batchNo < stat.ActiveBatch) {
				stat.ActiveBatch = batchNo
			}
		} else {
			stat.StalledTaskCount++
		}
		result[workflowRunID] = stat
	}

	return result, nil
}

func shouldUseRuntimeSnapshot(snapshot *projectRuntimeSnapshot, workflowStatus string) bool {
	if snapshot == nil || snapshot.Situation.Progress == nil {
		return false
	}
	if snapshot.Situation.WorkflowStatus == workflowStatus {
		switch workflowStatus {
		case "completed", "failed", "canceled", "paused":
			return true
		}
	}
	if snapshot.CreatedAt == nil {
		return false
	}
	ageMillis := gtime.Now().TimestampMilli() - snapshot.CreatedAt.TimestampMilli()
	if ageMillis < 0 {
		ageMillis = 0
	}
	return time.Duration(ageMillis)*time.Millisecond <= projectRuntimeSnapshotFreshWindow
}

func taskStatFromProgress(progress *autonomy.ProgressMetrics) projectRuntimeTaskStat {
	if progress == nil {
		return projectRuntimeTaskStat{}
	}
	return projectRuntimeTaskStat{
		TotalTasks:     progress.TotalTasks,
		CompletedTasks: progress.CompletedTasks,
		FailedTasks:    progress.FailedTasks,
		RunningTasks:   progress.RunningTasks,
	}
}

func latestNonNilTime(items ...*gtime.Time) *gtime.Time {
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

func normalizeDBUTCGTime(value *gtime.Time) *gtime.Time {
	if value == nil || value.IsZero() {
		return nil
	}

	raw := strings.TrimSpace(value.String())
	switch raw {
	case "", "0000-00-00 00:00:00":
		return nil
	}

	for _, layout := range []string{
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05.999",
		time.DateTime,
		time.RFC3339Nano,
		time.RFC3339,
	} {
		if parsed, err := time.ParseInLocation(layout, raw, time.UTC); err == nil {
			return gtime.NewFromTime(parsed.In(time.Local))
		}
	}

	return value
}

func isRecentGTime(value *gtime.Time, window time.Duration) bool {
	if value == nil {
		return false
	}
	delta := gtime.Now().TimestampMilli() - value.TimestampMilli()
	if delta < 0 {
		delta = 0
	}
	return time.Duration(delta)*time.Millisecond <= window
}

// ProjectStatus 获取项目状态
func (c *cWorkflow) ProjectStatus(ctx context.Context, req *v1.WorkflowProjectStatusReq) (res *v1.WorkflowProjectStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	projectMap, err := repo.NewProjectRepo().GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(projectMap) == 0 {
		return nil, fmt.Errorf("项目不存在")
	}
	return projectStatusV2(ctx, mapToDBRecord(projectMap))
}

// projectStatusV2 V2 引擎的项目状态聚合。
func projectStatusV2(ctx context.Context, project gdb.Record) (*v1.WorkflowProjectStatusRes, error) {
	projectID := project["id"].Int64()

	var wfStatus, currentStage string
	var progressPercent int
	var totalTasks, completedTasks, failedTasks, runningTasks int
	var activity projectRuntimeTaskActivityStat

	wfRuns, err := loadLatestWorkflowRuns(ctx, []int64{projectID})
	if err != nil {
		return nil, err
	}
	wfRun := wfRuns[projectID]
	if !wfRun.IsEmpty() {
		wfRunID := wfRun["id"].Int64()
		wfStatus = wfRun["status"].String()
		currentStage = wfRun["current_stage"].String()

		stats := projectRuntimeTaskStat{}
		snapshots, snapshotErr := loadLatestSituationSnapshots(ctx, []int64{wfRunID})
		if snapshotErr != nil {
			g.Log().Warningf(ctx, "[ProjectStatus] 读取态势快照失败，回退实时聚合: workflowRunID=%d, err=%v", wfRunID, snapshotErr)
		}
		if snapshot := snapshots[wfRunID]; shouldUseRuntimeSnapshot(snapshot, wfStatus) {
			stats = taskStatFromProgress(snapshot.Situation.Progress)
			if currentStage == "" {
				currentStage = snapshot.Situation.ActiveStage
			}
		} else {
			taskStats, taskErr := loadTaskStats(ctx, []int64{wfRunID})
			if taskErr != nil {
				return nil, taskErr
			}
			stats = taskStats[wfRunID]
		}
		activityStats, activityErr := loadTaskActivityStats(ctx, []int64{wfRunID})
		if activityErr != nil {
			return nil, activityErr
		}
		activity = activityStats[wfRunID]

		totalTasks = stats.TotalTasks
		completedTasks = stats.CompletedTasks
		failedTasks = stats.FailedTasks
		runningTasks = stats.RunningTasks
		if totalTasks > 0 {
			progressPercent = completedTasks * 100 / totalTasks
		}
	}

	type statusCount struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	blueprintCounts, scanErr := repo.NewBlueprintRepo().CountStatusesByProjectDraftAndActive(ctx, projectID)
	if scanErr != nil {
		g.Log().Warningf(ctx, "[ProjectStatus] 蓝图统计查询失败: project=%d err=%v", projectID, scanErr)
	}

	statusCounts := make(map[string]int)
	bpTotal := 0
	for _, sc := range blueprintCounts {
		record := mapToDBRecord(sc)
		status := record["status"].String()
		count := record["count"].Int()
		statusCounts[status] = count
		bpTotal += count
	}

	if totalTasks > 0 {
		statusCounts["domain_total"] = totalTasks
		statusCounts["domain_completed"] = completedTasks
		statusCounts["domain_failed"] = failedTasks
		statusCounts["domain_running"] = runningTasks
	}

	displayTotal := bpTotal
	if totalTasks > 0 {
		displayTotal = totalTasks
	}

	res := &v1.WorkflowProjectStatusRes{
		Status:             project["status"].String(),
		PauseReason:        project["pause_reason"].String(),
		ActiveBatch:        activity.ActiveBatch,
		TotalTasks:         displayTotal,
		StatusCounts:       statusCounts,
		LastActiveAt:       activity.LastActiveAt,
		IsActuallyWorking:  activity.ActiveRunningTasks > 0,
		ActiveRunningTasks: activity.ActiveRunningTasks,
		StalledTaskCount:   activity.StalledTaskCount,
		WorkflowRunID:      snowflake.JsonInt64(wfRun["id"].Int64()),
		EngineVersion:      "workflow_v2",
		WorkflowStatus:     wfStatus,
		CurrentStage:       currentStage,
		ProgressPercent:    progressPercent,
	}

	if wfStatus != "" {
		res.Status = wfStatus
	}

	return res, nil
}

// BatchProjectStats 批量查询项目运行时统计（为项目列表页提供进度数据）
func (c *cWorkflow) BatchProjectStats(ctx context.Context, req *v1.WorkflowBatchProjectStatsReq) (res *v1.WorkflowBatchProjectStatsRes, err error) {
	if len(req.ProjectIDs) > 50 {
		return nil, fmt.Errorf("单次最多查询 50 个项目")
	}

	ids := make([]int64, 0, len(req.ProjectIDs))
	for _, id := range req.ProjectIDs {
		ids = append(ids, int64(id))
	}

	allowedRecords, err := repo.NewProjectRepo().ListByIDsWithScope(ctx, ids, func(model *gdb.Model) *gdb.Model {
		return middleware.ApplyDataScope(ctx, model, "created_by", "dept_id")
	}, "id")
	if err != nil {
		return nil, fmt.Errorf("权限过滤查询失败: %w", err)
	}
	allowedIDs := make(map[int64]bool, len(allowedRecords))
	for _, p := range allowedRecords {
		allowedIDs[mapToDBRecord(p)["id"].Int64()] = true
	}
	filtered := ids[:0]
	for _, id := range ids {
		if allowedIDs[id] {
			filtered = append(filtered, id)
		}
	}
	ids = filtered
	if len(ids) == 0 {
		return &v1.WorkflowBatchProjectStatsRes{Stats: []v1.ProjectRuntimeStat{}}, nil
	}

	wfMap, err := loadLatestWorkflowRuns(ctx, ids)
	if err != nil {
		return nil, err
	}

	wfRunIDs := make([]int64, 0, len(wfMap))
	wfStatusByRunID := make(map[int64]string, len(wfMap))
	for _, r := range wfMap {
		wfID := r["id"].Int64()
		wfRunIDs = append(wfRunIDs, wfID)
		wfStatusByRunID[wfID] = r["status"].String()
	}

	snapshotMap, snapshotErr := loadLatestSituationSnapshots(ctx, wfRunIDs)
	if snapshotErr != nil {
		g.Log().Warningf(ctx, "[BatchProjectStats] 读取态势快照失败，回退实时聚合: err=%v", snapshotErr)
	}

	fallbackRunIDs := make([]int64, 0, len(wfRunIDs))
	for _, wfID := range wfRunIDs {
		if !shouldUseRuntimeSnapshot(snapshotMap[wfID], wfStatusByRunID[wfID]) {
			fallbackRunIDs = append(fallbackRunIDs, wfID)
		}
	}

	taskStats, err := loadTaskStats(ctx, fallbackRunIDs)
	if err != nil {
		return nil, err
	}

	stats := make([]v1.ProjectRuntimeStat, 0, len(ids))
	for _, pid := range ids {
		stat := v1.ProjectRuntimeStat{
			ProjectID: snowflake.JsonInt64(pid),
		}
		if wf, ok := wfMap[pid]; ok {
			stat.CurrentStage = wf["current_stage"].String()
			wfID := wf["id"].Int64()
			if snapshot := snapshotMap[wfID]; shouldUseRuntimeSnapshot(snapshot, wf["status"].String()) {
				if stat.CurrentStage == "" {
					stat.CurrentStage = snapshot.Situation.ActiveStage
				}
				ts := taskStatFromProgress(snapshot.Situation.Progress)
				stat.TotalTasks = ts.TotalTasks
				stat.CompletedTasks = ts.CompletedTasks
				stat.FailedTasks = ts.FailedTasks
				stat.RunningTasks = ts.RunningTasks
			} else if ts, exists := taskStats[wfID]; exists {
				stat.TotalTasks = ts.TotalTasks
				stat.CompletedTasks = ts.CompletedTasks
				stat.FailedTasks = ts.FailedTasks
				stat.RunningTasks = ts.RunningTasks
			}
		}
		stats = append(stats, stat)
	}

	return &v1.WorkflowBatchProjectStatsRes{Stats: stats}, nil
}
