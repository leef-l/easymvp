package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// RiskGateRuleRepo 风险闸门规则仓储。
type RiskGateRuleRepo struct{}

// NewRiskGateRuleRepo 创建闸门规则仓储。
func NewRiskGateRuleRepo() *RiskGateRuleRepo { return &RiskGateRuleRepo{} }

func (r *RiskGateRuleRepo) table() string { return "mvp_risk_gate_rule" }

// Create 创建闸门规则。
func (r *RiskGateRuleRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	result, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

// GetByID 按 ID 查询。
func (r *RiskGateRuleRepo) GetByID(ctx context.Context, id int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", id).WhereNull("deleted_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListEnabled 查询适用的已启用闸门规则（按优先级排序）。
func (r *RiskGateRuleRepo) ListEnabled(ctx context.Context, family, categoryCode string) ([]g.Map, error) {
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

// ListByType 按闸门类型查询。
func (r *RiskGateRuleRepo) ListByType(ctx context.Context, gateType string, enabled bool) ([]g.Map, error) {
	q := g.DB().Model(r.table()).Ctx(ctx).
		Where("gate_type", gateType).
		WhereNull("deleted_at")
	if enabled {
		q = q.Where("enabled", 1)
	}
	records, err := q.OrderAsc("priority").All()
	if err != nil {
		return nil, err
	}
	result := make([]g.Map, 0, len(records))
	for _, rec := range records {
		result = append(result, rec.Map())
	}
	return result, nil
}

// Update 更新闸门规则。
func (r *RiskGateRuleRepo) Update(ctx context.Context, id int64, data g.Map) error {
	data["updated_at"] = gtime.Now()
	_, err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Update(data)
	return err
}
