package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// ActionOutcomeRepo 动作效果记录仓储。
type ActionOutcomeRepo struct{}

func NewActionOutcomeRepo() *ActionOutcomeRepo { return &ActionOutcomeRepo{} }

func (r *ActionOutcomeRepo) table() string { return "mvp_action_outcome" }

// Create 创建动作效果记录。
func (r *ActionOutcomeRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询动作效果记录。
func (r *ActionOutcomeRepo) GetByID(ctx context.Context, outcomeID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", outcomeID).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// UpdateFields 按 ID 更新动作效果记录字段。
func (r *ActionOutcomeRepo) UpdateFields(ctx context.Context, outcomeID int64, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", outcomeID).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}

// GetAverageEffectScore 查询指定范围内的平均效果得分与样本数。
func (r *ActionOutcomeRepo) GetAverageEffectScore(ctx context.Context, projectID int64, periodStart, periodEnd *gtime.Time) (float64, int, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		WhereNull("deleted_at")
	if projectID > 0 {
		model = model.Where("project_id", projectID)
	}
	if periodStart != nil {
		model = model.WhereGTE("created_at", periodStart)
	}
	if periodEnd != nil {
		model = model.WhereLTE("created_at", periodEnd)
	}

	record, err := model.Fields("AVG(effect_score) AS avg_score, COUNT(*) AS cnt").One()
	if err != nil {
		return 0, 0, err
	}
	if record.IsEmpty() {
		return 0, 0, nil
	}
	return record["avg_score"].Float64(), record["cnt"].Int(), nil
}
