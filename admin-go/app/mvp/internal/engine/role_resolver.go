package engine

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
)

// ResolveProjectRole 统一查找项目角色配置。
func ResolveProjectRole(ctx context.Context, projectID int64, roleType string) (gdb.Record, error) {
	return repo.GetProjectRole(ctx, projectID, roleType)
}

// ResolveProjectRoleByLevel 统一按 role_type + role_level 查找项目角色配置。
func ResolveProjectRoleByLevel(ctx context.Context, projectID int64, roleType string, roleLevel string) (gdb.Record, error) {
	return repo.GetProjectRoleByLevel(ctx, projectID, roleType, roleLevel)
}

// ResolveProjectRolesMap 统一获取项目可用角色映射。
func ResolveProjectRolesMap(ctx context.Context, projectID int64) (map[string]map[string]interface{}, error) {
	return repo.GetProjectRolesMap(ctx, projectID)
}

// ResolveProjectModelInfo 统一解析项目角色对应的模型信息。
func ResolveProjectModelInfo(ctx context.Context, projectID int64, roleType string, roleLevel string, modelID int64) (*ModelInfo, error) {
	systemPrompt := ""
	if projectID > 0 && roleType != "" {
		role, err := ResolveProjectRoleByLevel(ctx, projectID, roleType, roleLevel)
		if err != nil {
			if modelID == 0 {
				return nil, fmt.Errorf("解析项目角色模型失败: %w", err)
			}
		} else if role != nil {
			systemPrompt = role["system_prompt"].String()
			if modelID == 0 {
				modelID = role["model_id"].Int64()
			}
		}
	}
	if modelID == 0 {
		return nil, fmt.Errorf("未解析到可用模型: project=%d role=%s/%s", projectID, roleType, roleLevel)
	}

	return getModelInfoStatic(ctx, modelID, systemPrompt)
}

// ResolveProjectExecutionMode 统一解析项目角色执行方式。
func ResolveProjectExecutionMode(ctx context.Context, projectID int64, roleType string, roleLevel string) string {
	role, err := ResolveProjectRoleByLevel(ctx, projectID, roleType, roleLevel)
	if err != nil {
		g.Log().Warningf(ctx, "[RoleResolver] 解析执行模式失败，回退 chat: project=%d role=%s/%s err=%v",
			projectID, roleType, roleLevel, err)
		return "chat"
	}
	if role == nil {
		g.Log().Debugf(ctx, "[RoleResolver] 未找到角色配置，回退 chat: project=%d role=%s/%s",
			projectID, roleType, roleLevel)
		return "chat"
	}

	mode := role["execution_mode"].String()
	if mode == "" {
		return "chat"
	}
	return mode
}
