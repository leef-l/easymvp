package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/model/entity"
	"easymvp/utility/snowflake"
)

// PlanVersionRepo 计划版本仓储。
type PlanVersionRepo struct{}

func NewPlanVersionRepo() *PlanVersionRepo { return &PlanVersionRepo{} }

func (r *PlanVersionRepo) table() string { return "mvp_plan_version" }

// Create 创建计划版本。
func (r *PlanVersionRepo) Create(ctx context.Context, data g.Map) (int64, error) {
	id := snowflake.Generate()
	data["id"] = id
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return int64(id), err
}

// GetByID 按 ID 查询。
func (r *PlanVersionRepo) GetByID(ctx context.Context, id int64) (*entity.MvpPlanVersion, error) {
	var ent entity.MvpPlanVersion
	err := g.DB().Model(r.table()).Ctx(ctx).Where("id", id).Scan(&ent)
	return &ent, err
}

// NextVersionNo 获取项目下一个版本号。
func (r *PlanVersionRepo) NextVersionNo(ctx context.Context, projectID int64) (int, error) {
	val, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Max("version_no")
	if err != nil {
		return 0, err
	}
	return int(val) + 1, nil
}

// ListByProject 查询项目所有版本。
func (r *PlanVersionRepo) ListByProject(ctx context.Context, projectID int64) ([]entity.MvpPlanVersion, error) {
	var list []entity.MvpPlanVersion
	err := g.DB().Model(r.table()).Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("version_no").
		Scan(&list)
	return list, err
}
