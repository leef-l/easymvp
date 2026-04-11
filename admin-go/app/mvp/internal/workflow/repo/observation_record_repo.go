package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// ObservationRecordRepo 决策观测记录仓储。
type ObservationRecordRepo struct{}

func NewObservationRecordRepo() *ObservationRecordRepo { return &ObservationRecordRepo{} }

func (r *ObservationRecordRepo) table() string { return "mvp_observation_record" }

// Create 创建观测记录。
func (r *ObservationRecordRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// ListByProject 查询项目观测记录。
func (r *ObservationRecordRepo) ListByProject(ctx context.Context, projectID int64, limit int, fields ...string) ([]g.Map, error) {
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

// ListDistinctProjectIDs 查询存在观测记录的项目 ID 列表。
func (r *ObservationRecordRepo) ListDistinctProjectIDs(ctx context.Context) ([]int64, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		WhereNull("deleted_at").
		Fields("DISTINCT project_id").
		All()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(records))
	for _, record := range records {
		id := record["project_id"].Int64()
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// CountByProject 统计项目观测记录总数。
func (r *ObservationRecordRepo) CountByProject(ctx context.Context, projectID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Count()
}

// CountHumanOverrideByProject 统计项目人工干预记录数量。
func (r *ObservationRecordRepo) CountHumanOverrideByProject(ctx context.Context, projectID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		Where("human_override", 1).
		WhereNull("deleted_at").
		Count()
}

// CountGroupByField 按给定字段分组统计项目观测记录。
func (r *ObservationRecordRepo) CountGroupByField(ctx context.Context, projectID int64, field string) ([]g.Map, error) {
	return r.CountGroupByFieldInPeriod(ctx, projectID, field, nil, nil)
}

// CountGroupByFieldInPeriod 按给定字段分组统计项目观测记录，可选时间范围。
func (r *ObservationRecordRepo) CountGroupByFieldInPeriod(ctx context.Context, projectID int64, field string, periodStart, periodEnd *gtime.Time) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		WhereNull("deleted_at").
		Fields(field + ", COUNT(*) as cnt").
		Group(field)
	if projectID > 0 {
		model = model.Where("project_id", projectID)
	}
	if periodStart != nil {
		model = model.WhereGTE("created_at", periodStart)
	}
	if periodEnd != nil {
		model = model.WhereLTE("created_at", periodEnd)
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// CountByFiltersInPeriod 按条件统计观测记录，可选时间范围。
func (r *ObservationRecordRepo) CountByFiltersInPeriod(ctx context.Context, projectID int64, periodStart, periodEnd *gtime.Time, filters g.Map) (int, error) {
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
	for key, value := range filters {
		model = model.Where(key, value)
	}
	return model.Count()
}

// UpdateByDecisionAction 按决策动作 ID 更新观测记录。
func (r *ObservationRecordRepo) UpdateByDecisionAction(ctx context.Context, decisionActionID int64, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("decision_action_id", decisionActionID).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}
