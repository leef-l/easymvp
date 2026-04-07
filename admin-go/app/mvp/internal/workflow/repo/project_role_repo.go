package repo

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// GetProjectRole 查找项目角色配置，找不到时自动从默认预设回退并写入 mvp_project_role。
func GetProjectRole(ctx context.Context, projectID int64, roleType string) (gdb.Record, error) {
	role, err := g.DB().Model("mvp_project_role").Ctx(ctx).
		Where("project_id", projectID).
		Where("role_type", roleType).
		WhereNull("deleted_at").
		Where("status", 1).
		One()
	if err != nil {
		return nil, fmt.Errorf("查询角色配置失败: %w", err)
	}
	if !role.IsEmpty() {
		return role, nil
	}
	return autoCreateFromPreset(ctx, projectID, roleType)
}

// GetProjectRolesMap 批量获取项目角色配置（key: roleType/roleLevel），缺失的自动补齐。
func GetProjectRolesMap(ctx context.Context, projectID int64) (map[string]map[string]interface{}, error) {
	existing, err := g.DB().Model("mvp_project_role").Ctx(ctx).
		Where("project_id", projectID).
		Where("status", 1).
		WhereNull("deleted_at").
		All()
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]interface{})
	existingTypes := make(map[string]bool)
	for _, rc := range existing {
		key := rc["role_type"].String() + "/" + rc["role_level"].String()
		result[key] = map[string]interface{}{
			"execution_mode": rc["execution_mode"].String(),
			"model_id":       rc["model_id"].Int64(),
		}
		existingTypes[rc["role_type"].String()] = true
	}

	// 补齐缺失角色
	for _, rt := range []string{"architect", "implementer", "auditor", "coordinator"} {
		if existingTypes[rt] {
			continue
		}
		role, err := autoCreateFromPreset(ctx, projectID, rt)
		if err != nil {
			g.Log().Warningf(ctx, "[ProjectRoleRepo] 自动创建 %s 角色失败: %v", rt, err)
			continue
		}
		key := role["role_type"].String() + "/" + role["role_level"].String()
		result[key] = map[string]interface{}{
			"execution_mode": role["execution_mode"].String(),
			"model_id":       role["model_id"].Int64(),
		}
	}

	return result, nil
}

func autoCreateFromPreset(ctx context.Context, projectID int64, roleType string) (gdb.Record, error) {
	project, err := g.DB().Model("mvp_project").Ctx(ctx).
		Fields("project_category, created_by, dept_id").
		Where("id", projectID).
		WhereNull("deleted_at").
		One()
	if err != nil || project.IsEmpty() {
		return nil, fmt.Errorf("项目 %d 不存在", projectID)
	}

	projectCategory := project["project_category"].String()
	userID := project["created_by"].Int64()
	deptID := project["dept_id"].Int64()

	preset, err := g.DB().Model("mvp_role_preset").Ctx(ctx).
		Where("project_category", projectCategory).
		Where("role_type", roleType).
		Where("is_default", 1).
		Where("status", 1).
		WhereNull("deleted_at").
		OrderAsc("sort").
		One()
	if err != nil {
		return nil, fmt.Errorf("查询默认预设失败: %w", err)
	}
	if preset.IsEmpty() {
		return nil, fmt.Errorf("分类 %s 没有 %s 的默认预设", projectCategory, roleType)
	}

	modelID := preset["model_id"].Int64()
	systemPrompt := preset["system_prompt"].String()
	if modelID > 0 {
		modelRec, _ := g.DB().Model("ai_model").Ctx(ctx).
			Fields("role_prompt").
			Where("id", modelID).
			WhereNull("deleted_at").
			One()
		if !modelRec.IsEmpty() && modelRec["role_prompt"].String() != "" {
			systemPrompt = modelRec["role_prompt"].String()
		}
	}

	executionMode := preset["execution_mode"].String()
	if executionMode == "" {
		executionMode = "chat"
	}

	roleID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_project_role").Ctx(ctx).Insert(g.Map{
		"id":               roleID,
		"project_id":       projectID,
		"project_category": projectCategory,
		"role_type":        roleType,
		"role_level":       preset["role_level"].String(),
		"model_id":         modelID,
		"system_prompt":    systemPrompt,
		"execution_mode":   executionMode,
		"status":           1,
		"created_by":       userID,
		"dept_id":          deptID,
		"created_at":       gtime.Now(),
		"updated_at":       gtime.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("自动创建 %s 角色失败: %w", roleType, err)
	}

	g.Log().Infof(ctx, "[ProjectRoleRepo] 项目 %d 自动从预设创建 %s 角色 (preset=%d, model=%d)",
		projectID, roleType, preset["id"].Int64(), modelID)

	created, _ := g.DB().Model("mvp_project_role").Ctx(ctx).Where("id", roleID).One()
	return created, nil
}
