package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// PolicyRuleRepo 策略规则仓储。
type PolicyRuleRepo struct{}

// NewPolicyRuleRepo 创建策略规则仓储。
func NewPolicyRuleRepo() *PolicyRuleRepo { return &PolicyRuleRepo{} }

func (r *PolicyRuleRepo) table() string { return "mvp_policy_rule" }

// Create 创建策略规则。
func (r *PolicyRuleRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	result, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

// GetByID 按 ID 查询。
func (r *PolicyRuleRepo) GetByID(ctx context.Context, id int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).WhereNull("deleted_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetByCode 按规则编码查询。
func (r *PolicyRuleRepo) GetByCode(ctx context.Context, code string) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("rule_code", code).WhereNull("deleted_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByTriggerAndScope 按触发源和项目作用域查询策略规则（按优先级排序）。
// family 和 categoryCode 为空时匹配全局规则。
func (r *PolicyRuleRepo) ListByTriggerAndScope(ctx context.Context, triggerSource, family, categoryCode string) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("trigger_source", triggerSource).
		Where("enabled", 1).
		WhereNull("deleted_at").
		Where(g.DB().Raw("(project_family IS NULL OR project_family = ? OR project_family = '')", family)).
		Where(g.DB().Raw("(project_category_code IS NULL OR project_category_code = ? OR project_category_code = '')", categoryCode)).
		OrderAsc("priority").
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

// ListByScope 按项目作用域查询全部启用的策略规则（不限触发源）。
func (r *PolicyRuleRepo) ListByScope(ctx context.Context, family, categoryCode string) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("enabled", 1).
		WhereNull("deleted_at").
		Where(g.DB().Raw("(project_family IS NULL OR project_family = ? OR project_family = '')", family)).
		Where(g.DB().Raw("(project_category_code IS NULL OR project_category_code = ? OR project_category_code = '')", categoryCode)).
		OrderAsc("priority").
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

// Update 更新策略规则。
func (r *PolicyRuleRepo) Update(ctx context.Context, id int64, data g.Map) error {
	data["updated_at"] = gtime.Now()
	_, err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Update(data)
	return err
}

// SoftDelete 软删除。
func (r *PolicyRuleRepo) SoftDelete(ctx context.Context, id int64) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).
		Update(g.Map{"deleted_at": gtime.Now()})
	return err
}
