package executor

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

func commandTemplateUsesDocker(template string) bool {
	lower := strings.ToLower(strings.TrimSpace(template))
	if lower == "" {
		return false
	}
	return strings.HasPrefix(lower, "docker ") ||
		strings.Contains(lower, "docker run") ||
		strings.Contains(lower, " docker ")
}

func engineConfiguredWithDocker(ctx context.Context, engineCode string) bool {
	if engineCode == "" {
		return false
	}
	cfg, err := g.DB().Model("ai_engine_config").Ctx(ctx).
		Fields("command_template").
		Where("engine_code", engineCode).
		Where("status", 1).
		WhereNull("deleted_at").
		One()
	if err != nil {
		g.Log().Warningf(ctx, "[DockerGuard] 查询引擎配置失败 engine=%s err=%v，按无 Docker 处理", engineCode, err)
		return false
	}
	if cfg.IsEmpty() {
		return false
	}
	return commandTemplateUsesDocker(cfg["command_template"].String())
}
