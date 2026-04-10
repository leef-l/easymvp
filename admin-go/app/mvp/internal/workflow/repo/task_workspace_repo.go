package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// TaskWorkspaceRepo 任务工作区仓储。
type TaskWorkspaceRepo struct{}

func NewTaskWorkspaceRepo() *TaskWorkspaceRepo { return &TaskWorkspaceRepo{} }

func (r *TaskWorkspaceRepo) table() string { return "mvp_task_workspace" }

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
