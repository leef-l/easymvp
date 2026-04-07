package engine

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"easymvp/app/mvp/internal/workflow/repo"
)

// ResolveProjectRole 查找项目角色配置，找不到时自动从默认预设回退并写入 mvp_project_role。
// engine 包的入口，委托给 repo.GetProjectRole。
func ResolveProjectRole(ctx context.Context, projectID int64, roleType string) (gdb.Record, error) {
	return repo.GetProjectRole(ctx, projectID, roleType)
}
