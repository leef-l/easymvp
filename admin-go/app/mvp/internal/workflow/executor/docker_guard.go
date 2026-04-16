package executor

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
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
	cfg, err := repo.NewAIEngineConfigRepo().GetEnabledByCode(ctx, engineCode, "command_template")
	if err != nil {
		g.Log().Warningf(ctx, "[DockerGuard] 查询引擎配置失败 engine=%s err=%v，按无 Docker 处理", engineCode, err)
		return false
	}
	if len(cfg) == 0 {
		return false
	}
	return commandTemplateUsesDocker(g.NewVar(cfg["command_template"]).String())
}
