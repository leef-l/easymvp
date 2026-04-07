package repo

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/presetutil"
	"easymvp/utility/snowflake"
)

type projectPresetContext struct {
	Name            string
	Description     string
	ProjectCategory string
	CategoryCode    string
	CreatedBy       int64
	DeptID          int64
}

// GetProjectRole 查找项目角色配置；若项目未显式配置，则直接回退到分类默认预设。
func GetProjectRole(ctx context.Context, projectID int64, roleType string) (gdb.Record, error) {
	return GetProjectRoleByLevel(ctx, projectID, roleType, "")
}

// GetProjectRoleByLevel 统一解析项目角色配置，优先项目显式配置，再回退分类默认预设。
func GetProjectRoleByLevel(ctx context.Context, projectID int64, roleType string, roleLevel string) (gdb.Record, error) {
	projectCtx, err := loadProjectPresetContext(ctx, projectID)
	if err != nil {
		return nil, err
	}

	roles, err := loadProjectRoleRecords(ctx, projectID, roleType)
	if err != nil {
		return nil, fmt.Errorf("查询角色配置失败: %w", err)
	}
	if role := findRecordByRoleLevel(roles, roleLevel); role != nil {
		return role, nil
	}

	defaultPresets, err := loadDefaultPresets(ctx, projectCtx, roleType, "")
	if err != nil {
		return nil, fmt.Errorf("查询默认预设失败: %w", err)
	}
	if preset := findRecordByRoleLevel(defaultPresets, roleLevel); preset != nil {
		return buildPresetRoleRecord(ctx, projectID, projectCtx, preset), nil
	}
	if role := selectPreferredRoleRecord(roles, roleType); role != nil {
		return role, nil
	}
	if preset := selectPreferredRoleRecord(defaultPresets, roleType); preset != nil {
		return buildPresetRoleRecord(ctx, projectID, projectCtx, preset), nil
	}

	return nil, fmt.Errorf("分类 %s 没有 %s 的 V2 默认预设", projectCtx.CategoryCode, roleType)
}

// GetProjectRolesMap 批量获取项目角色配置（key: roleType/roleLevel）。
// 项目未显式配置的角色直接回退到默认预设，不写入 mvp_project_role。
func GetProjectRolesMap(ctx context.Context, projectID int64) (map[string]map[string]interface{}, error) {
	existing, err := loadProjectRoleRecords(ctx, projectID, "")
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]interface{})
	existingKeys := make(map[string]bool)
	for _, rc := range existing {
		key := rc["role_type"].String() + "/" + rc["role_level"].String()
		result[key] = map[string]interface{}{
			"execution_mode": rc["execution_mode"].String(),
			"model_id":       rc["model_id"].Int64(),
		}
		existingKeys[key] = true
	}

	projectCtx, err := loadProjectPresetContext(ctx, projectID)
	if err != nil {
		return result, nil
	}

	defaultPresets, err := loadDefaultPresets(ctx, projectCtx, "", "")
	if err != nil {
		return result, nil
	}

	for _, preset := range defaultPresets {
		key := preset["role_type"].String() + "/" + preset["role_level"].String()
		if existingKeys[key] {
			continue
		}
		role := buildPresetRoleRecord(ctx, projectID, projectCtx, preset)
		result[key] = map[string]interface{}{
			"execution_mode": role["execution_mode"].String(),
			"model_id":       role["model_id"].Int64(),
		}
		existingKeys[key] = true
	}

	return result, nil
}

func autoCreateFromPreset(ctx context.Context, projectID int64, roleType string) (gdb.Record, error) {
	return GetProjectRoleByLevel(ctx, projectID, roleType, "")
}

func loadProjectRoleRecords(ctx context.Context, projectID int64, roleType string) (gdb.Result, error) {
	m := g.DB().Model("mvp_project_role").Ctx(ctx).
		Where("project_id", projectID).
		Where("status", 1).
		WhereNull("deleted_at")
	if roleType != "" {
		m = m.Where("role_type", roleType)
	}
	return m.OrderAsc("id").All()
}

func loadProjectPresetContext(ctx context.Context, projectID int64) (*projectPresetContext, error) {
	project, err := g.DB().Model("mvp_project").Ctx(ctx).
		Fields("name, description, project_category, category_code, created_by, dept_id").
		Where("id", projectID).
		WhereNull("deleted_at").
		One()
	if err != nil {
		return nil, fmt.Errorf("查询项目 %d 失败: %w", projectID, err)
	}
	if project.IsEmpty() {
		return nil, fmt.Errorf("项目 %d 不存在", projectID)
	}

	categoryCode := project["category_code"].String()
	projectCategory := project["project_category"].String()
	if categoryCode == "" && projectCategory != "" {
		category, _ := g.DB().Model("mvp_project_category").Ctx(ctx).
			Fields("category_code").
			Where("display_name", projectCategory).
			Where("status", 1).
			WhereNull("deleted_at").
			One()
		if !category.IsEmpty() {
			categoryCode = category["category_code"].String()
		}
	}
	if categoryCode == "" {
		return nil, fmt.Errorf("项目 %d 缺少 category_code，无法走 V2 默认预设", projectID)
	}

	return &projectPresetContext{
		Name:            project["name"].String(),
		Description:     project["description"].String(),
		ProjectCategory: projectCategory,
		CategoryCode:    categoryCode,
		CreatedBy:       project["created_by"].Int64(),
		DeptID:          project["dept_id"].Int64(),
	}, nil
}

func loadDefaultPresets(ctx context.Context, projectCtx *projectPresetContext, roleType string, roleLevel string) (gdb.Result, error) {
	return ListRolePresets(ctx, RolePresetQuery{
		CategoryCode:    projectCtx.CategoryCode,
		ProjectCategory: projectCtx.ProjectCategory,
		RoleType:        roleType,
		RoleLevel:       roleLevel,
		DefaultOnly:     true,
	})
}

func selectPreferredRoleRecord(records gdb.Result, roleType string) gdb.Record {
	if len(records) == 0 {
		return nil
	}
	for _, level := range presetutil.PreferredRoleLevels(roleType) {
		for _, record := range records {
			if record["role_level"].String() == level {
				return record
			}
		}
	}
	return records[0]
}

func findRecordByRoleLevel(records gdb.Result, roleLevel string) gdb.Record {
	if roleLevel == "" {
		return nil
	}
	for _, record := range records {
		if record["role_level"].String() == roleLevel {
			return record
		}
	}
	return nil
}

func createProjectRoleFromPreset(ctx context.Context, projectID int64, projectCtx *projectPresetContext, preset gdb.Record) (gdb.Record, error) {
	modelID := preset["model_id"].Int64()
	modelPrompt := loadModelRolePrompt(ctx, modelID)
	roleType := preset["role_type"].String()
	roleLevel := preset["role_level"].String()

	systemPrompt := presetutil.BuildRoleSystemPrompt(projectCtx.CategoryCode, roleType, roleLevel, preset["system_prompt"].String(), modelPrompt)
	if roleType == "architect" {
		systemPrompt = presetutil.BuildArchitectSystemPrompt(projectCtx.Name, projectCtx.Description, projectCtx.CategoryCode, systemPrompt)
	}

	executionMode := preset["execution_mode"].String()
	if executionMode == "" {
		executionMode = "chat"
	}

	roleID := int64(snowflake.Generate())
	_, err := g.DB().Model("mvp_project_role").Ctx(ctx).Insert(g.Map{
		"id":               roleID,
		"project_id":       projectID,
		"project_category": projectCtx.ProjectCategory,
		"role_type":        roleType,
		"role_level":       roleLevel,
		"model_id":         modelID,
		"system_prompt":    systemPrompt,
		"execution_mode":   executionMode,
		"status":           1,
		"created_by":       projectCtx.CreatedBy,
		"dept_id":          projectCtx.DeptID,
		"created_at":       gtime.Now(),
		"updated_at":       gtime.Now(),
	})
	if err != nil {
		return nil, err
	}

	return g.DB().Model("mvp_project_role").Ctx(ctx).Where("id", roleID).One()
}

func buildPresetRoleRecord(ctx context.Context, projectID int64, projectCtx *projectPresetContext, preset gdb.Record) gdb.Record {
	modelID := preset["model_id"].Int64()
	modelPrompt := loadModelRolePrompt(ctx, modelID)
	roleType := preset["role_type"].String()
	roleLevel := preset["role_level"].String()

	systemPrompt := presetutil.BuildRoleSystemPrompt(projectCtx.CategoryCode, roleType, roleLevel, preset["system_prompt"].String(), modelPrompt)
	if roleType == "architect" {
		systemPrompt = presetutil.BuildArchitectSystemPrompt(projectCtx.Name, projectCtx.Description, projectCtx.CategoryCode, systemPrompt)
	}

	executionMode := preset["execution_mode"].String()
	if executionMode == "" {
		executionMode = "chat"
	}

	return gdb.Record{
		"id":               gvar.New(0),
		"project_id":       gvar.New(projectID),
		"project_category": gvar.New(projectCtx.ProjectCategory),
		"role_type":        gvar.New(roleType),
		"role_level":       gvar.New(roleLevel),
		"model_id":         gvar.New(modelID),
		"system_prompt":    gvar.New(systemPrompt),
		"execution_mode":   gvar.New(executionMode),
		"status":           gvar.New(1),
	}
}

func loadModelRolePrompt(ctx context.Context, modelID int64) string {
	if modelID == 0 {
		return ""
	}
	modelRec, _ := g.DB().Model("ai_model").Ctx(ctx).
		Fields("role_prompt").
		Where("id", modelID).
		WhereNull("deleted_at").
		One()
	if modelRec.IsEmpty() {
		return ""
	}
	return modelRec["role_prompt"].String()
}
