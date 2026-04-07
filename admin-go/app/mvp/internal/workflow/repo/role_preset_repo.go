package repo

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// RolePresetQuery 统一的角色预设查询参数。
type RolePresetQuery struct {
	IDs              []int64
	CategoryCode     string
	ProjectCategory  string
	RoleType         string
	RoleLevel        string
	DefaultOnly      bool
	IncludeModelName bool
}

// ListRolePresets 统一查询角色预设列表。
func ListRolePresets(ctx context.Context, query RolePresetQuery) (gdb.Result, error) {
	fields := "p.id, p.role_type, p.role_level, p.model_id, p.system_prompt, p.execution_mode, p.is_default, p.project_category"
	m := buildPresetBaseQuery(ctx, query)
	if query.IncludeModelName {
		m = m.LeftJoin("ai_model AS m", "m.id = p.model_id")
		fields += ", m.name AS model_name"
	}
	return m.Fields(fields).OrderAsc("p.sort").OrderAsc("p.id").All()
}

// GetRolePreset 统一查询单条角色预设。
func GetRolePreset(ctx context.Context, query RolePresetQuery) (gdb.Record, error) {
	list, err := ListRolePresets(ctx, query)
	if err != nil || len(list) == 0 {
		return nil, err
	}
	return list[0], nil
}

// CountRolePresets 统一统计角色预设数量。
func CountRolePresets(ctx context.Context, query RolePresetQuery) (int, error) {
	return buildPresetBaseQuery(ctx, query).Count()
}

// buildPresetBaseQuery 构建角色预设通用查询条件（消除 List/Count 重复逻辑）。
func buildPresetBaseQuery(ctx context.Context, query RolePresetQuery) *gdb.Model {
	m := g.DB().Model("mvp_role_preset AS p").Ctx(ctx).
		Where("p.status", 1).
		Where("p.deleted_at IS NULL")

	if len(query.IDs) > 0 {
		m = m.WhereIn("p.id", query.IDs)
	}
	if query.CategoryCode != "" {
		m = m.Where("p.project_category", query.CategoryCode)
	} else if query.ProjectCategory != "" {
		m = m.Where("p.project_category", query.ProjectCategory)
	}
	if query.RoleType != "" {
		m = m.Where("p.role_type", query.RoleType)
	}
	if query.RoleLevel != "" {
		m = m.Where("p.role_level", query.RoleLevel)
	}
	if query.DefaultOnly {
		m = m.Where("p.is_default", 1)
	}

	return m
}
