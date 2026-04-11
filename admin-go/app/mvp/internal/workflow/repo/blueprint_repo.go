package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
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

// Insert 使用给定数据插入蓝图记录，不自动生成 ID。
func (r *BlueprintRepo) Insert(ctx context.Context, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return err
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

// ListByPlanVersionStatus 查询版本下指定状态蓝图。
func (r *BlueprintRepo) ListByPlanVersionStatus(ctx context.Context, planVersionID int64, blueprintStatus string) (gdb.Result, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("plan_version_id", planVersionID).
		WhereNull("deleted_at").
		OrderAsc("sort")
	if blueprintStatus != "" {
		model = model.Where("blueprint_status", blueprintStatus)
	}
	return model.All()
}

// ListByPlanVersionStatuses 查询版本下指定状态集合的蓝图。
func (r *BlueprintRepo) ListByPlanVersionStatuses(ctx context.Context, planVersionID int64, statuses []string, fields ...string) ([]g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("plan_version_id", planVersionID).
		WhereNull("deleted_at").
		OrderAsc("sort")
	if len(statuses) > 0 {
		model = model.WhereIn("blueprint_status", statuses)
	}
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
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

// UpdateFields 按 ID 更新蓝图字段。
func (r *BlueprintRepo) UpdateFields(ctx context.Context, blueprintID int64, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", blueprintID).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
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

// UpdateStatusByPlanVersion 按计划版本批量更新指定状态的蓝图。
func (r *BlueprintRepo) UpdateStatusByPlanVersion(ctx context.Context, planVersionID int64, fromStatus string, data g.Map) error {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("plan_version_id", planVersionID).
		WhereNull("deleted_at")
	if fromStatus != "" {
		model = model.Where("blueprint_status", fromStatus)
	}
	_, err := model.Data(data).Update()
	return err
}

// UpdateByPlanVersionStatuses 按计划版本和蓝图状态集合批量更新蓝图。
func (r *BlueprintRepo) UpdateByPlanVersionStatuses(ctx context.Context, planVersionID int64, statuses []string, data g.Map) (int64, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("plan_version_id", planVersionID).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("blueprint_status", statuses)
	}
	result, err := model.Data(data).Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
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

// CountStatusesByProjectDraftAndActive 查询项目 draft/active 方案下的蓝图状态统计。
func (r *BlueprintRepo) CountStatusesByProjectDraftAndActive(ctx context.Context, projectID int64) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()+" AS bp").Ctx(ctx).
		InnerJoin("mvp_plan_version AS pv", "pv.id = bp.plan_version_id").
		Where("pv.project_id", projectID).
		WhereIn("pv.status", g.Slice{"draft", "active"}).
		WhereNull("bp.deleted_at").
		Fields("bp.blueprint_status AS status, COUNT(*) AS count").
		Group("bp.blueprint_status").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}
