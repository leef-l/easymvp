package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// AcceptRuleRepo 验收规则仓储。
type AcceptRuleRepo struct{}

func NewAcceptRuleRepo() *AcceptRuleRepo { return &AcceptRuleRepo{} }

func (r *AcceptRuleRepo) table() string { return "mvp_accept_rule" }

// ListByProjectType 按项目类型加载已启用的验收规则，按优先级排序。
func (r *AcceptRuleRepo) ListByProjectType(ctx context.Context, projectType string) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_type", projectType).
		Where("enabled", 1).
		WhereNull("deleted_at").
		OrderAsc("priority").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// GetByCode 按项目类型和规则编码查询单条规则。
func (r *AcceptRuleRepo) GetByCode(ctx context.Context, projectType, ruleCode string) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_type", projectType).
		Where("rule_code", ruleCode).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}
