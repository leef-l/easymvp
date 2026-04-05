package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// ProjectCategoryRepo 项目分类配置仓储。
type ProjectCategoryRepo struct{}

func NewProjectCategoryRepo() *ProjectCategoryRepo { return &ProjectCategoryRepo{} }

func (r *ProjectCategoryRepo) table() string { return "mvp_project_category" }

// GetByCode 按 category_code 查询分类配置。
func (r *ProjectCategoryRepo) GetByCode(ctx context.Context, categoryCode string) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("category_code", categoryCode).
		Where("status", 1).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetByDisplayName 按展示名称查询分类配置（兼容映射用）。
func (r *ProjectCategoryRepo) GetByDisplayName(ctx context.Context, displayName string) (g.Map, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("display_name", displayName).
		Where("status", 1).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListAll 查询所有启用的分类，按 sort 排序。
func (r *ProjectCategoryRepo) ListAll(ctx context.Context) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("status", 1).
		WhereNull("deleted_at").
		OrderAsc("sort").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByFamily 按 family_code 查询分类列表。
func (r *ProjectCategoryRepo) ListByFamily(ctx context.Context, familyCode string) ([]g.Map, error) {
	records, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("family_code", familyCode).
		Where("status", 1).
		WhereNull("deleted_at").
		OrderAsc("sort").
		All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}
