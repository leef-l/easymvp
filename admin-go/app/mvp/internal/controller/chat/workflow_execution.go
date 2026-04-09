package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/utility/snowflake"
)

// ExecutionStatus 执行阶段实时状态
func (c *cWorkflow) ExecutionStatus(ctx context.Context, req *v1.WorkflowExecutionStatusReq) (res *v1.WorkflowExecutionStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	res = &v1.WorkflowExecutionStatusRes{
		Tasks:         []v1.DomainTaskItem{},
		ResourceLocks: []v1.ResourceLockItem{},
	}

	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	if err != nil || wfRun.IsEmpty() {
		return res, nil
	}
	wfRunID := wfRun["id"].Int64()
	res.WorkflowRunID = snowflake.JsonInt64(wfRunID)

	stageRun, stageErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", wfRunID).
		Where("stage_type", "execute").
		WhereNull("deleted_at").
		OrderDesc("stage_no").
		One()
	if stageErr != nil {
		g.Log().Warningf(ctx, "[ExecutionStatus] 查询 stage_run 失败: wfRun=%d err=%v", wfRunID, stageErr)
	}
	if !stageRun.IsEmpty() {
		res.StageRunID = snowflake.JsonInt64(stageRun["id"].Int64())
		res.StageStatus = stageRun["status"].String()
	}

	tasks, taskErr := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", wfRunID).
		WhereNull("deleted_at").
		OrderAsc("batch_no").
		OrderAsc("sort").
		All()
	if taskErr != nil {
		g.Log().Warningf(ctx, "[ExecutionStatus] 查询领域任务失败: wfRun=%d err=%v", wfRunID, taskErr)
	}

	taskIDs := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		taskIDs = append(taskIDs, t["id"].Int64())
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
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, wfErr := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	if wfErr != nil {
		return nil, fmt.Errorf("查询工作流运行失败: %w", wfErr)
	}
	if wfRun.IsEmpty() {
		return &v1.WorkflowDomainTasksRes{Tasks: []v1.DomainTaskItem{}}, nil
	}

	query := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", wfRun["id"].Int64()).
		WhereNull("deleted_at")

	if req.Status != "" {
		query = query.Where("status", req.Status)
	}
	if req.BatchNo > 0 {
		query = query.Where("batch_no", req.BatchNo)
	}

	tasks, err := query.Fields("id, name, description, status, role_type, role_level, batch_no, sort, execution_mode, affected_resources, started_at, completed_at, result, retry_count").OrderAsc("batch_no").OrderAsc("sort").All()
	if err != nil {
		return nil, err
	}

	taskIDs := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		taskIDs = append(taskIDs, t["id"].Int64())
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
	tasks, tErr := g.DB().Model("mvp_domain_task").Ctx(ctx).
		WhereIn("id", taskIDs).WhereNull("deleted_at").Fields("id, name").All()
	if tErr != nil {
		g.Log().Warningf(ctx, "[ResourceLocks] 查询任务名称失败: %v", tErr)
	}
	for _, t := range tasks {
		taskNames[t["id"].Int64()] = t["name"].String()
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

	records, err := queryTaskWorkspaceMetaRecords(ctx, taskIDs)
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

func queryTaskWorkspaceMetaRecords(ctx context.Context, taskIDs []int64) (gdb.Result, error) {
	records, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		WhereIn("task_id", taskIDs).
		WhereNull("deleted_at").
		Fields("id, task_id, status, cleanup_status, delivery_mode, delivery_status, sync_strategy, sync_status, risk_level, patch_ref, delivery_ref, delivery_title, diff_summary, updated_at").
		All()
	if err == nil || !isUnknownWorkspaceColumnErr(err) {
		return records, err
	}

	if isWorkspaceDeliveryRefColumnErr(err) {
		records, err = g.DB().Model("mvp_task_workspace").Ctx(ctx).
			WhereIn("task_id", taskIDs).
			WhereNull("deleted_at").
			Fields("id, task_id, status, cleanup_status, delivery_mode, delivery_status, sync_strategy, sync_status, risk_level, patch_ref, diff_summary, updated_at").
			All()
		if err == nil || !isUnknownWorkspaceColumnErr(err) {
			return records, err
		}
	}

	return g.DB().Model("mvp_task_workspace").Ctx(ctx).
		WhereIn("task_id", taskIDs).
		WhereNull("deleted_at").
		Fields("id, task_id, status, cleanup_status, diff_summary, updated_at").
		All()
}

func isUnknownWorkspaceColumnErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unknown column")
}

func isWorkspaceDeliveryRefColumnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "delivery_ref") || strings.Contains(msg, "delivery_title")
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
