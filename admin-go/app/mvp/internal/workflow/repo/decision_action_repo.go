package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// DecisionActionRepo 决策动作仓储。
type DecisionActionRepo struct{}

// NewDecisionActionRepo 创建决策动作仓储。
func NewDecisionActionRepo() *DecisionActionRepo { return &DecisionActionRepo{} }

func (r *DecisionActionRepo) table() string { return "mvp_decision_action" }

// Create 创建决策动作记录。
func (r *DecisionActionRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	result, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

// GetByID 按 ID 查询。
func (r *DecisionActionRepo) GetByID(ctx context.Context, id int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).WhereNull("deleted_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// UpdateStatus 更新决策动作状态。
func (r *DecisionActionRepo) UpdateStatus(ctx context.Context, id int64, status string, extra g.Map) error {
	data := g.Map{"action_status": status, "updated_at": gtime.Now()}
	for k, v := range extra {
		data[k] = v
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Update(data)
	return err
}

// ListByWorkflow 按工作流查询决策记录。
func (r *DecisionActionRepo) ListByWorkflow(ctx context.Context, workflowRunID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		OrderDesc("created_at").
		All()
	if err != nil {
		return nil, err
	}
	result := make([]g.Map, 0, len(records))
	for _, rec := range records {
		result = append(result, rec.Map())
	}
	return result, nil
}

// ListByProject 按项目和动作类型查询。
func (r *DecisionActionRepo) ListByProject(ctx context.Context, projectID int64, actionType string) ([]g.Map, error) {
	q := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at")
	if actionType != "" {
		q = q.Where("decision_type", actionType)
	}
	records, err := q.OrderDesc("created_at").All()
	if err != nil {
		return nil, err
	}
	result := make([]g.Map, 0, len(records))
	for _, rec := range records {
		result = append(result, rec.Map())
	}
	return result, nil
}

// ListPending 查询待处理的决策。
func (r *DecisionActionRepo) ListPending(ctx context.Context, projectID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("action_status", g.Slice{"pending", "waiting_human"}).
		WhereNull("deleted_at").
		OrderDesc("created_at").
		All()
	if err != nil {
		return nil, err
	}
	result := make([]g.Map, 0, len(records))
	for _, rec := range records {
		result = append(result, rec.Map())
	}
	return result, nil
}

// CountByType 按类型统计。
func (r *DecisionActionRepo) CountByType(ctx context.Context, workflowRunID int64, actionType string) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("decision_type", actionType).
		WhereNull("deleted_at").
		Count()
}
