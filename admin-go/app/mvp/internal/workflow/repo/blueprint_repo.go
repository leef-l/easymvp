package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/model/entity"
	"easymvp/utility/snowflake"
)

// BlueprintRepo 任务蓝图仓储。
type BlueprintRepo struct{}

func NewBlueprintRepo() *BlueprintRepo { return &BlueprintRepo{} }

func (r *BlueprintRepo) table() string { return "mvp_task_blueprint" }

// Create 创建蓝图。
func (r *BlueprintRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// ListByPlanVersion 查询版本下所有蓝图。
func (r *BlueprintRepo) ListByPlanVersion(ctx context.Context, planVersionID int64) ([]entity.MvpTaskBlueprint, error) {
	var list []entity.MvpTaskBlueprint
	err := g.DB().Model(r.table()).Ctx(ctx).
		Where("plan_version_id", planVersionID).
		WhereNull("deleted_at").
		OrderAsc("batch_no").
		OrderAsc("sort").
		Scan(&list)
	return list, err
}

// CountByPlanVersion 统计计划版本下的蓝图数量。
func (r *BlueprintRepo) CountByPlanVersion(ctx context.Context, planVersionID int64) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("plan_version_id", planVersionID).
		WhereNull("deleted_at").
		Count()
}

// CountByPlanVersionAndStatus 统计版本下指定状态蓝图数量。
func (r *BlueprintRepo) CountByPlanVersionAndStatus(ctx context.Context, planVersionID int64, blueprintStatus string) (int, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("plan_version_id", planVersionID).
		WhereNull("deleted_at")
	if blueprintStatus != "" {
		model = model.Where("blueprint_status", blueprintStatus)
	}
	return model.Count()
}

// UpdateDraftToConfirmedByPlanVersion 将 draft 蓝图标记为 confirmed。
func (r *BlueprintRepo) UpdateDraftToConfirmedByPlanVersion(ctx context.Context, planVersionID int64) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("plan_version_id", planVersionID).
		Where("blueprint_status", "draft").
		WhereNull("deleted_at").
		Data(g.Map{"blueprint_status": "confirmed"}).
		Update()
	return err
}

// UpdateByPlanVersionIDs 批量更新给定计划版本下的蓝图。
func (r *BlueprintRepo) UpdateByPlanVersionIDs(ctx context.Context, planVersionIDs []int64, data g.Map) error {
	if len(planVersionIDs) == 0 {
		return nil
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("plan_version_id", planVersionIDs).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}
