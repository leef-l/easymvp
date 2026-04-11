package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// LearningRecordRepo EMA 学习记录仓储。
type LearningRecordRepo struct{}

func NewLearningRecordRepo() *LearningRecordRepo { return &LearningRecordRepo{} }

func (r *LearningRecordRepo) table() string { return "mvp_learning_record" }

// Create 创建学习记录。
func (r *LearningRecordRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByMetric 查询指标学习记录。
func (r *LearningRecordRepo) GetByMetric(ctx context.Context, metricKey string, projectID int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("metric_key", metricKey).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByProject 查询项目学习记录。
func (r *LearningRecordRepo) ListByProject(ctx context.Context, projectID int64, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderAsc("metric_key")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// UpdateByMetric 按指标键更新学习记录。
func (r *LearningRecordRepo) UpdateByMetric(ctx context.Context, metricKey string, projectID int64, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("metric_key", metricKey).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}
