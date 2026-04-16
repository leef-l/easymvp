package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskWorkspaceRepo 任务工作区仓储。
type TaskWorkspaceRepo struct{}

func NewTaskWorkspaceRepo() *TaskWorkspaceRepo { return &TaskWorkspaceRepo{} }

func (r *TaskWorkspaceRepo) table() string { return "mvp_task_workspace" }

// ListByWorkflow 查询工作流下的工作空间记录。
func (r *TaskWorkspaceRepo) ListByWorkflow(ctx context.Context, workflowRunID int64, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListDeliveriesByWorkflow 查询工作流下的交付结果。
func (r *TaskWorkspaceRepo) ListDeliveriesByWorkflow(ctx context.Context, workflowRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("task_id, delivery_mode, delivery_status, sync_status, risk_level, patch_ref").
		OrderAsc("task_id").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListWorkspaceSummaryByWorkflow 查询工作流下的工作空间交付摘要字段。
func (r *TaskWorkspaceRepo) ListWorkspaceSummaryByWorkflow(ctx context.Context, workflowRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("delivery_mode, delivery_status, sync_status, risk_level").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListMetaByTasks 查询任务对应的工作空间元信息。
func (r *TaskWorkspaceRepo) ListMetaByTasks(ctx context.Context, taskIDs []int64) (gdb.Result, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}

	records, err := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("task_id", taskIDs).
		WhereNull("deleted_at").
		Fields("id, task_id, status, cleanup_status, delivery_mode, delivery_status, sync_strategy, sync_status, risk_level, patch_ref, delivery_ref, delivery_title, diff_summary, updated_at").
		All()
	if err == nil || !isUnknownTaskWorkspaceColumnErr(err) {
		return records, err
	}

	if isTaskWorkspaceDeliveryRefColumnErr(err) {
		records, err = g.DB().Model(r.table()).Ctx(ctx).
			WhereIn("task_id", taskIDs).
			WhereNull("deleted_at").
			Fields("id, task_id, status, cleanup_status, delivery_mode, delivery_status, sync_strategy, sync_status, risk_level, patch_ref, diff_summary, updated_at").
			All()
		if err == nil || !isUnknownTaskWorkspaceColumnErr(err) {
			return records, err
		}
	}

	return g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("task_id", taskIDs).
		WhereNull("deleted_at").
		Fields("id, task_id, status, cleanup_status, diff_summary, updated_at").
		All()
}

// ListArtifactRecordsByWorkflow 查询工作流下 workspace 交付记录，兼容旧表结构。
func (r *TaskWorkspaceRepo) ListArtifactRecordsByWorkflow(ctx context.Context, workflowRunID int64) (gdb.Result, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("id, task_id, delivery_mode, delivery_status, sync_status, patch_ref, delivery_ref, delivery_title, diff_summary").
		OrderAsc("task_id").
		All()
	if err == nil || !isUnknownTaskWorkspaceColumnErr(err) {
		return records, err
	}

	if isTaskWorkspaceDeliveryRefColumnErr(err) {
		records, err = g.DB().Model(r.table()).Ctx(ctx).
			Where("workflow_run_id", workflowRunID).
			WhereNull("deleted_at").
			Fields("id, task_id, delivery_mode, delivery_status, sync_status, patch_ref, diff_summary").
			OrderAsc("task_id").
			All()
		if err == nil || !isUnknownTaskWorkspaceColumnErr(err) {
			return records, err
		}
	}

	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Fields("id, task_id, diff_summary").
		OrderAsc("task_id").
		All()
}

// GetLatestIDByTask 查询任务最近一条工作空间记录 ID。
func (r *TaskWorkspaceRepo) GetLatestIDByTask(ctx context.Context, taskID int64) (int64, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("task_id", taskID).
		WhereNull("deleted_at").
		Fields("id").
		OrderDesc("created_at").
		One()
	if err != nil || record.IsEmpty() {
		return 0, err
	}
	return record["id"].Int64(), nil
}

// ListReadyPendingSyncByWorkflow 查询待人工回写的交付记录。
func (r *TaskWorkspaceRepo) ListReadyPendingSyncByWorkflow(ctx context.Context, workflowRunID int64, taskIDs []int64) ([]g.Map, error) {
	model := g.DB().Model(r.table()+" w").Ctx(ctx).
		InnerJoin("mvp_domain_task t", "t.id = w.task_id").
		Where("w.workflow_run_id", workflowRunID).
		WhereNull("w.deleted_at").
		WhereNull("t.deleted_at").
		Where("t.status", "completed").
		Where("w.delivery_status", "ready").
		WhereIn("w.sync_status", g.Slice{"pending", "failed"}).
		Fields("w.id, w.task_id, w.sync_status, w.delivery_status, t.name, t.batch_no, t.sort")
	if len(taskIDs) > 0 {
		model = model.WhereIn("w.task_id", taskIDs)
	}
	records, err := model.OrderAsc("t.batch_no").OrderAsc("t.sort").All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

func isUnknownTaskWorkspaceColumnErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unknown column")
}

func isTaskWorkspaceDeliveryRefColumnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "delivery_ref") || strings.Contains(msg, "delivery_title")
}

// InspectColumns 检查任务工作区指定列是否可用。
func (r *TaskWorkspaceRepo) InspectColumns(ctx context.Context, columns []string) error {
	if len(columns) == 0 {
		return nil
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Fields(strings.Join(columns, ",")).
		Limit(1).
		All()
	return err
}

// SoftDeleteByWorkflow 软删除工作流下的工作空间记录。
func (r *TaskWorkspaceRepo) SoftDeleteByWorkflow(ctx context.Context, workflowRunID int64, deletedAt *gtime.Time) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Data(g.Map{
			"deleted_at": deletedAt,
			"updated_at": deletedAt,
		}).
		Update()
	return err
}

// SoftDeleteByTask 软删除指定任务的工作空间记录。
func (r *TaskWorkspaceRepo) SoftDeleteByTask(ctx context.Context, taskID int64, deletedAt *gtime.Time) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("task_id", taskID).
		WhereNull("deleted_at").
		Data(g.Map{
			"deleted_at": deletedAt,
			"updated_at": deletedAt,
		}).
		Update()
	return err
}
