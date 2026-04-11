package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// TuneRecommendationRepo 调参建议仓储。
type TuneRecommendationRepo struct{}

func NewTuneRecommendationRepo() *TuneRecommendationRepo { return &TuneRecommendationRepo{} }

func (r *TuneRecommendationRepo) table() string { return "mvp_tune_recommendation" }

// Create 创建调参建议。
func (r *TuneRecommendationRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询调参建议。
func (r *TuneRecommendationRepo) GetByID(ctx context.Context, recommendationID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", recommendationID).
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

// UpdateStatusIfCurrent 仅在当前状态匹配时更新状态。
func (r *TuneRecommendationRepo) UpdateStatusIfCurrent(ctx context.Context, recommendationID int64, from, to string, extra g.Map) (int64, error) {
	data := g.Map{"status": to}
	for k, v := range extra {
		data[k] = v
	}
	result, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", recommendationID).
		Where("status", from).
		WhereNull("deleted_at").
		Data(data).
		Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// ListByProjectStatus 查询项目调参建议，支持状态过滤。
func (r *TuneRecommendationRepo) ListByProjectStatus(ctx context.Context, projectID int64, status string, limit int, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		WhereNull("deleted_at").
		OrderDesc("created_at")
	if projectID > 0 {
		model = model.Where("project_id", projectID)
	}
	if status != "" {
		model = model.Where("status", status)
	}
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
