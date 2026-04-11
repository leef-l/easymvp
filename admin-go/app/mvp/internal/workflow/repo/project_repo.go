package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectRepo 项目仓储。
type ProjectRepo struct{}

func NewProjectRepo() *ProjectRepo { return &ProjectRepo{} }

func (r *ProjectRepo) table() string { return "mvp_project" }

// ProjectScopeFilter 项目查询权限范围。
type ProjectScopeFilter struct {
	All         bool
	IncludeSelf bool
	UserID      int64
	DeptIDs     []int64
}

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

// GetLatestByCreator 查询用户最近创建的项目。
func (r *ProjectRepo) GetLatestByCreator(ctx context.Context, createdBy int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("created_by", createdBy).
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

// ListRecentByScope 查询权限范围内最近的项目列表。
func (r *ProjectRepo) ListRecentByScope(ctx context.Context, scope ProjectScopeFilter, limit int, fields ...string) ([]g.Map, error) {
	model := applyProjectScopeFilter(g.DB().Model(r.table()).Ctx(ctx).WhereNull("deleted_at"), scope).
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

// FindByKeywordWithScope 按 ID 或名称在权限范围内查找项目。
func (r *ProjectRepo) FindByKeywordWithScope(ctx context.Context, keyword string, scope ProjectScopeFilter, fields ...string) (g.Map, error) {
	model := applyProjectScopeFilter(g.DB().Model(r.table()).Ctx(ctx).WhereNull("deleted_at"), scope)
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}

	var numID int64
	if _, err := fmt.Sscanf(keyword, "%d", &numID); err == nil && numID > 0 {
		record, err := model.Clone().Where("id", numID).One()
		if err != nil {
			return nil, err
		}
		if !record.IsEmpty() {
			return record.Map(), nil
		}
	}

	record, err := model.WhereLike("name", "%"+keyword+"%").OrderDesc("created_at").One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// ListByIDs 查询指定项目集合；可选传入字段列表减少读取范围。
func (r *ProjectRepo) ListByIDs(ctx context.Context, projectIDs []int64, fields ...string) ([]g.Map, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	model := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("id", projectIDs).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	records, err := model.All()
	if err != nil {
		return nil, err
	}
	return records.List(), nil
}

// ListByIDsWithScope 查询指定项目集合，并允许调用方追加数据权限条件。
func (r *ProjectRepo) ListByIDsWithScope(ctx context.Context, projectIDs []int64, applyScope func(model *gdb.Model) *gdb.Model, fields ...string) ([]g.Map, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	model := g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("id", projectIDs).
		WhereNull("deleted_at")
	if applyScope != nil {
		model = applyScope(model)
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

// UpdateFields 按 ID 更新项目字段。
func (r *ProjectRepo) UpdateFields(ctx context.Context, projectID int64, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", projectID).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}

// UpdateFieldsInTx 在事务中按 ID 更新项目字段。
func (r *ProjectRepo) UpdateFieldsInTx(ctx context.Context, tx gdb.TX, projectID int64, data g.Map) error {
	_, err := tx.Model(r.table()).Ctx(ctx).
		Where("id", projectID).
		WhereNull("deleted_at").
		Data(data).
		Update()
	return err
}

// UpdateFieldsIfStatuses 在当前状态命中集合时更新项目字段。
func (r *ProjectRepo) UpdateFieldsIfStatuses(ctx context.Context, projectID int64, statuses []string, data g.Map) (int64, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", projectID).
		WhereNull("deleted_at")
	if len(statuses) > 0 {
		model = model.WhereIn("status", statuses)
	}
	result, err := model.Data(data).Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
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

func applyProjectScopeFilter(model *gdb.Model, scope ProjectScopeFilter) *gdb.Model {
	switch {
	case scope.All:
		return model
	case len(scope.DeptIDs) == 0 && scope.IncludeSelf:
		return model.Where("created_by", scope.UserID)
	case len(scope.DeptIDs) == 0:
		return model.Where("id", -1)
	case scope.IncludeSelf:
		return model.Where("(created_by = ? OR dept_id IN (?))", scope.UserID, scope.DeptIDs)
	default:
		return model.WhereIn("dept_id", scope.DeptIDs)
	}
}
