package chat

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

// ExecutionStatus 执行阶段实时状态
func (c *cWorkflow) ExecutionStatus(ctx context.Context, req *v1.WorkflowExecutionStatusReq) (res *v1.WorkflowExecutionStatusRes, err error) {
	var (
		domainTaskRepo = repo.NewDomainTaskRepo()
		stageRunRepo   = repo.NewStageRunRepo()
	)

	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	res = &v1.WorkflowExecutionStatusRes{
		Tasks:         []v1.DomainTaskItem{},
		ResourceLocks: []v1.ResourceLockItem{},
	}

	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return res, nil
	}
	wfRunID := wfRun["id"].Int64()
	res.WorkflowRunID = snowflake.JsonInt64(wfRunID)

	stageRun, stageErr := stageRunRepo.GetLatestByWorkflowAndType(ctx, wfRunID, "execute")
	if stageErr != nil {
		g.Log().Warningf(ctx, "[ExecutionStatus] 查询 stage_run 失败: wfRun=%d err=%v", wfRunID, stageErr)
	}
	if !stageRun.IsEmpty() {
		res.StageRunID = snowflake.JsonInt64(stageRun["id"].Int64())
		res.StageStatus = stageRun["status"].String()
	}

	taskMaps, taskErr := domainTaskRepo.ListByWorkflowOrdered(ctx, wfRunID)
	if taskErr != nil {
		g.Log().Warningf(ctx, "[ExecutionStatus] 查询领域任务失败: wfRun=%d err=%v", wfRunID, taskErr)
	}

	taskIDs := make([]int64, 0, len(taskMaps))
	tasks := make([]gdb.Record, 0, len(taskMaps))
	for _, t := range taskMaps {
		record := mapToDBRecord(t)
		tasks = append(tasks, record)
		taskIDs = append(taskIDs, record["id"].Int64())
	}
	workspaceMeta, wsErr := loadTaskWorkspaceMeta(ctx, taskIDs)
	if wsErr != nil {
		g.Log().Warningf(ctx, "[ExecutionStatus] 查询任务工作空间失败: wfRun=%d err=%v", wfRunID, wsErr)
	}

	for _, t := range tasks {
		res.Tasks = append(res.Tasks, buildDomainTaskItem(t, workspaceMeta[t["id"].Int64()]))
	}

	for _, t := range res.Tasks {
		res.TotalTasks++
		switch t.Status {
		case "completed":
			res.CompletedTasks++
		case "running":
			res.RunningTasks++
		case "failed":
			res.FailedTasks++
		case "pending":
			res.PendingTasks++
		case "escalated":
			res.EscalatedTasks++
		}
	}

	scheduler := orchestrator.GetTaskScheduler()
	if scheduler != nil {
		lockedRes := scheduler.GetLockedResources()
		for resource, taskID := range lockedRes {
			taskName := ""
			for _, t := range res.Tasks {
				if int64(t.ID) == taskID {
					taskName = t.Name
					break
				}
			}
			res.ResourceLocks = append(res.ResourceLocks, v1.ResourceLockItem{
				Resource: resource,
				TaskID:   snowflake.JsonInt64(taskID),
				TaskName: taskName,
			})
		}
	}

	for _, t := range res.Tasks {
		if t.Status == "running" || t.Status == "pending" {
			if t.BatchNo > 0 && (res.ActiveBatch == 0 || t.BatchNo < res.ActiveBatch) {
				res.ActiveBatch = t.BatchNo
			}
		}
	}

	return res, nil
}

// DomainTasks 领域任务列表
func (c *cWorkflow) DomainTasks(ctx context.Context, req *v1.WorkflowDomainTasksReq) (res *v1.WorkflowDomainTasksRes, err error) {
	domainTaskRepo := repo.NewDomainTaskRepo()

	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, wfErr := latestWorkflowRunForProject(ctx, projectID)
	if wfErr != nil {
		return &v1.WorkflowDomainTasksRes{Tasks: []v1.DomainTaskItem{}}, nil
	}

	taskMaps, err := domainTaskRepo.ListByWorkflowFiltered(ctx, wfRun["id"].Int64(), req.Status, req.BatchNo,
		"id", "name", "description", "status", "role_type", "role_level", "batch_no", "sort", "execution_mode", "affected_resources", "started_at", "completed_at", "result", "retry_count", "error_message")
	if err != nil {
		return nil, err
	}

	taskIDs := make([]int64, 0, len(taskMaps))
	tasks := make([]gdb.Record, 0, len(taskMaps))
	for _, t := range taskMaps {
		record := mapToDBRecord(t)
		tasks = append(tasks, record)
		taskIDs = append(taskIDs, record["id"].Int64())
	}
	workspaceMeta, wsErr := loadTaskWorkspaceMeta(ctx, taskIDs)
	if wsErr != nil {
		g.Log().Warningf(ctx, "[DomainTasks] 查询任务工作空间失败: project=%d err=%v", projectID, wsErr)
	}

	items := make([]v1.DomainTaskItem, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, buildDomainTaskItem(t, workspaceMeta[t["id"].Int64()]))
	}

	return &v1.WorkflowDomainTasksRes{Tasks: items, Total: len(items)}, nil
}

// ResourceLocks 资源锁列表
func (c *cWorkflow) ResourceLocks(ctx context.Context, req *v1.WorkflowResourceLocksReq) (res *v1.WorkflowResourceLocksRes, err error) {
	domainTaskRepo := repo.NewDomainTaskRepo()

	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	res = &v1.WorkflowResourceLocksRes{Locks: []v1.ResourceLockItem{}}

	scheduler := orchestrator.GetTaskScheduler()
	if scheduler == nil {
		return res, nil
	}

	lockedRes := scheduler.GetLockedResources()
	if len(lockedRes) == 0 {
		return res, nil
	}

	taskIDs := make([]int64, 0, len(lockedRes))
	for _, tid := range lockedRes {
		taskIDs = append(taskIDs, tid)
	}
	taskNames := make(map[int64]string)
	taskMaps, tErr := domainTaskRepo.ListByIDs(ctx, taskIDs, "id", "name")
	if tErr != nil {
		g.Log().Warningf(ctx, "[ResourceLocks] 查询任务名称失败: %v", tErr)
	}
	for _, t := range taskMaps {
		taskNames[g.NewVar(t["id"]).Int64()] = g.NewVar(t["name"]).String()
	}

	for resource, taskID := range lockedRes {
		res.Locks = append(res.Locks, v1.ResourceLockItem{
			Resource: resource,
			TaskID:   snowflake.JsonInt64(taskID),
			TaskName: taskNames[taskID],
		})
	}

	return res, nil
}

// buildDomainTaskItem 构建领域任务响应项。
type taskWorkspaceMeta struct {
	WorkspaceID    int64
	Status         string
	CleanupStatus  string
	DeliveryMode   string
	DeliveryStatus string
	SyncStrategy   string
	SyncStatus     string
	RiskLevel      string
	PatchRef       string
	DeliveryRef    string
	DeliveryTitle  string
	DiffSummary    string
	UpdatedAt      *gtime.Time
}

func loadTaskWorkspaceMeta(ctx context.Context, taskIDs []int64) (map[int64]taskWorkspaceMeta, error) {
	result := make(map[int64]taskWorkspaceMeta, len(taskIDs))
	if len(taskIDs) == 0 {
		return result, nil
	}

	records, err := repo.NewTaskWorkspaceRepo().ListMetaByTasks(ctx, taskIDs)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		result[record["task_id"].Int64()] = taskWorkspaceMeta{
			WorkspaceID:    record["id"].Int64(),
			Status:         record["status"].String(),
			CleanupStatus:  record["cleanup_status"].String(),
			DeliveryMode:   record["delivery_mode"].String(),
			DeliveryStatus: record["delivery_status"].String(),
			SyncStrategy:   record["sync_strategy"].String(),
			SyncStatus:     record["sync_status"].String(),
			RiskLevel:      record["risk_level"].String(),
			PatchRef:       record["patch_ref"].String(),
			DeliveryRef:    record["delivery_ref"].String(),
			DeliveryTitle:  record["delivery_title"].String(),
			DiffSummary:    record["diff_summary"].String(),
			UpdatedAt:      normalizeDBUTCGTime(record["updated_at"].GTime()),
		}
	}

	return result, nil
}

func buildDomainTaskItem(t gdb.Record, ws taskWorkspaceMeta) v1.DomainTaskItem {
	var resources []string
	resJSON := t["affected_resources"].String()
	if resJSON != "" && resJSON != "[]" && resJSON != "null" {
		if umErr := json.Unmarshal([]byte(resJSON), &resources); umErr != nil {
			g.Log().Warningf(context.Background(), "[buildDomainTaskItem] 解析 affected_resources 失败: task=%d err=%v", t["id"].Int64(), umErr)
		}
	}
	startedAt := normalizeDBUTCGTime(t["started_at"].GTime())
	completedAt := normalizeDBUTCGTime(t["completed_at"].GTime())
	return v1.DomainTaskItem{
		ID:                snowflake.JsonInt64(t["id"].Int64()),
		Name:              t["name"].String(),
		Description:       t["description"].String(),
		Status:            t["status"].String(),
		RoleType:          t["role_type"].String(),
		RoleLevel:         t["role_level"].String(),
		BatchNo:           t["batch_no"].Int(),
		Sort:              t["sort"].Int(),
		ExecutionMode:     t["execution_mode"].String(),
		AffectedResources: resources,
		StartedAt:         startedAt,
		CompletedAt:       completedAt,
		ErrorMessage:      domainTaskErrorMessage(t),
		Result:            t["result"].String(),
		RetryCount:        t["retry_count"].Int(),
		WorkspaceStatus:   ws.Status,
		CleanupStatus:     ws.CleanupStatus,
		DeliveryMode:      ws.DeliveryMode,
		DeliveryStatus:    ws.DeliveryStatus,
		SyncStrategy:      ws.SyncStrategy,
		SyncStatus:        ws.SyncStatus,
		RiskLevel:         ws.RiskLevel,
		PatchRef:          ws.PatchRef,
		DeliveryRef:       ws.DeliveryRef,
		DeliveryTitle:     ws.DeliveryTitle,
		DiffSummary:       ws.DiffSummary,
	}
}

func domainTaskErrorMessage(t gdb.Record) string {
	if msg := t["error_message"].String(); msg != "" {
		return msg
	}
	status := t["status"].String()
	if status == "failed" || status == "escalated" {
		return t["result"].String()
	}
	return ""
}
