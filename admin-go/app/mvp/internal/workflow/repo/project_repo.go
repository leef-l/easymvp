package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// ProjectRepo 项目仓储。
type ProjectRepo struct{}

func NewProjectRepo() *ProjectRepo { return &ProjectRepo{} }

func (r *ProjectRepo) table() string { return "mvp_project" }

// GetByID 按 ID 查询项目；可选传入字段列表减少读取范围。
func (r *ProjectRepo) GetByID(ctx context.Context, projectID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", projectID).
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

// BackfillCategoryCodeIfEmpty 在 category_code 为空时回填稳定分类编码。
func (r *ProjectRepo) BackfillCategoryCodeIfEmpty(ctx context.Context, projectID int64, categoryCode string) error {
	categoryCode = strings.TrimSpace(categoryCode)
	if projectID == 0 || categoryCode == "" {
		return nil
	}
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", projectID).
		Where("category_code IS NULL OR category_code = ''").
		Data(g.Map{"category_code": categoryCode}).
		Update()
	return err
}

// UpdateStatus 更新项目状态。
func (r *ProjectRepo) UpdateStatus(ctx context.Context, projectID int64, status string) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", projectID).
		WhereNull("deleted_at").
		Data(g.Map{"status": status}).
		Update()
	return err
}

// UpdateStatusIfCurrent 仅在当前状态匹配时更新项目状态。
func (r *ProjectRepo) UpdateStatusIfCurrent(ctx context.Context, projectID int64, from, to string) (int64, error) {
	result, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", projectID).
		Where("status", from).
		WhereNull("deleted_at").
		Data(g.Map{"status": to}).
		Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}
