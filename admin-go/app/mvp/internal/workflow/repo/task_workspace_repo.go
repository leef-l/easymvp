package repo

import (
	"context"
	"strings"

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
