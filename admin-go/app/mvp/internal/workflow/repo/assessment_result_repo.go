package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// AssessmentResultRepo 元认知评估结果仓储。
type AssessmentResultRepo struct{}

func NewAssessmentResultRepo() *AssessmentResultRepo { return &AssessmentResultRepo{} }

func (r *AssessmentResultRepo) table() string { return "mvp_assessment_result" }

// Create 创建评估结果。
func (r *AssessmentResultRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetLatestByProject 查询项目最新评估结果。
func (r *AssessmentResultRepo) GetLatestByProject(ctx context.Context, projectID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("created_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByProject 查询项目评估历史。
func (r *AssessmentResultRepo) ListByProject(ctx context.Context, projectID int64, limit int, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
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
