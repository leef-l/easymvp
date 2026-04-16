package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowEventRepo 工作流事件仓储。
type WorkflowEventRepo struct{}

func NewWorkflowEventRepo() *WorkflowEventRepo { return &WorkflowEventRepo{} }

func (r *WorkflowEventRepo) table() string { return "mvp_workflow_event" }

// Insert 创建工作流事件记录。
func (r *WorkflowEventRepo) Insert(ctx context.Context, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return err
}

// CountByWorkflow 统计工作流事件数。
func (r *WorkflowEventRepo) CountByWorkflow(ctx context.Context, workflowRunID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Count()
}

// ListRecentByWorkflow 查询工作流最近事件。
func (r *WorkflowEventRepo) ListRecentByWorkflow(ctx context.Context, workflowRunID int64, limit int) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		OrderDesc("created_at")
	if limit > 0 {
		model = model.Limit(limit)
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByWorkflowAndEntity 查询工作流下指定实体的事件。
func (r *WorkflowEventRepo) ListByWorkflowAndEntity(ctx context.Context, workflowRunID int64, entityType string, entityID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("entity_type", entityType).
		Where("entity_id", entityID).
		OrderAsc("created_at").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByWorkflowIDs 查询多个工作流下的事件列表。
func (r *WorkflowEventRepo) ListByWorkflowIDs(ctx context.Context, workflowRunIDs []int64, limit int, fields ...string) ([]g.Map, error) {
	if len(workflowRunIDs) == 0 {
		return nil, nil
	}
	model := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("workflow_run_id", workflowRunIDs).
		OrderDesc("created_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	if limit > 0 {
		model = model.Limit(limit)
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// InspectMetadataColumns 检查 workflow_event 幂等元数据列是否存在。
func (r *WorkflowEventRepo) InspectMetadataColumns(ctx context.Context) error {
	_, err := g.DB().Ctx(ctx).Model(r.table()).
		Fields("event_id,idempotency_key,attempt").
		Limit(1).
		All()
	return err
}
