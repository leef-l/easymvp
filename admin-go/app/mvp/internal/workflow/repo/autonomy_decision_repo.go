package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// AutonomyDecisionRepo 自治决策仓储。
type AutonomyDecisionRepo struct{}

func NewAutonomyDecisionRepo() *AutonomyDecisionRepo { return &AutonomyDecisionRepo{} }

func (r *AutonomyDecisionRepo) table() string { return "mvp_autonomy_decision" }

// Create 创建决策记录。
func (r *AutonomyDecisionRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	data["created_at"] = gtime.Now()
	data["updated_at"] = gtime.Now()
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *AutonomyDecisionRepo) GetByID(ctx context.Context, id int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).WhereNull("deleted_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByProject 按项目查询决策列表。
func (r *AutonomyDecisionRepo) ListByProject(ctx context.Context, projectID int64, decisionType string) ([]g.Map, error) {
	m := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("created_at")
	if decisionType != "" {
		m = m.Where("decision_type", decisionType)
	}
	records, err := m.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListPending 查询待审批决策。
func (r *AutonomyDecisionRepo) ListPending(ctx context.Context, projectID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		Where("decision_mode", "suggest").
		Where("human_action", "pending").
		WhereNull("deleted_at").
		OrderDesc("created_at").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// UpdateHumanAction 更新人工动作。
func (r *AutonomyDecisionRepo) UpdateHumanAction(ctx context.Context, id int64, action string) error {
	data := g.Map{"human_action": action, "updated_at": gtime.Now()}
	if action == "approved" {
		data["executed_at"] = gtime.Now()
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).WhereNull("deleted_at").
		Data(data).Update()
	return err
}

// UpdateResult 更新执行结果。
func (r *AutonomyDecisionRepo) UpdateResult(ctx context.Context, id int64, result string) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).WhereNull("deleted_at").
		Data(g.Map{"result": result, "updated_at": gtime.Now()}).Update()
	return err
}

// CountByType 统计某工作流的指定类型决策数。
func (r *AutonomyDecisionRepo) CountByType(ctx context.Context, workflowRunID int64, decisionType string) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("decision_type", decisionType).
		WhereNull("deleted_at").
		Count()
}
