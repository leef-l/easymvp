package engine

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// defaultWorkDir 当项目未配置 work_dir 时的默认值。
// 从 chat_engine.go 和 SendFeishuMessage 中提取，消除硬编码。
var defaultWorkDir = "/www/wwwroot/project/easymvp"

// GetProjectWorkDir 获取项目工作目录，未配置时返回 defaultWorkDir。
func GetProjectWorkDir(ctx context.Context, projectID int64) string {
	project, err := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).
		WhereNull("deleted_at").
		Fields("work_dir").
		One()
	if err != nil {
		g.Log().Warningf(ctx, "[ProjectWorkDir] 查询项目 work_dir 失败: projectID=%d err=%v", projectID, err)
		return defaultWorkDir
	}
	workDir := project["work_dir"].String()
	if workDir == "" {
		return defaultWorkDir
	}
	return workDir
}
