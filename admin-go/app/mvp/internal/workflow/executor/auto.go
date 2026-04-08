package executor

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
)

// autoPreference auto 模式的执行器优先级（按推荐度从高到低）。
// 实际选择时检查哪个已在 ai_engine_config 中启用且 status=1。
var autoPreference = []string{
	"claude_code",
	"aider",
	"codex_cli",
	"gemini_cli",
	"openhands",
	"chat", // 兜底
}

// AutoExecutor 自动选择执行器。
// 根据 ai_engine_config 中已启用的执行器，按优先级自动分派到最合适的执行器。
type AutoExecutor struct {
	registry *Registry
}

// NewAutoExecutor 创建自动选择执行器。
func NewAutoExecutor(registry *Registry) *AutoExecutor {
	return &AutoExecutor{registry: registry}
}

func (e *AutoExecutor) Name() string { return "auto" }

func (e *AutoExecutor) NeedsWorkspace() bool { return true }

func (e *AutoExecutor) Execute(ctx context.Context, req *Request) *Result {
	// 查询已启用的执行器配置
	enabledEngines := make(map[string]bool)
	rows, err := g.DB().Model("ai_engine_config").Ctx(ctx).
		Fields("engine_code").
		Where("status", 1).
		WhereNull("deleted_at").
		All()
	if err == nil {
		for _, row := range rows {
			enabledEngines[row["engine_code"].String()] = true
		}
	}

	// 按优先级选择第一个已启���且已注册的执行器
	for _, mode := range autoPreference {
		if mode == "chat" {
			// chat 模式不需要 ai_engine_config 启用
			if exec := e.registry.Get(mode); exec != nil {
				g.Log().Infof(ctx, "[AutoExecutor] 选择执行器: %s（兜底）", mode)
				return exec.Execute(ctx, req)
			}
			continue
		}
		if !enabledEngines[mode] {
			continue
		}
		if engineConfiguredWithDocker(ctx, mode) {
			g.Log().Warningf(ctx, "[AutoExecutor] 跳过执行器: %s（命令模板依赖 Docker，当前已禁用）", mode)
			continue
		}
		exec := e.registry.Get(mode)
		if exec == nil {
			continue
		}
		g.Log().Infof(ctx, "[AutoExecutor] 自动选择执行器: %s", mode)
		return exec.Execute(ctx, req)
	}

	// 兜底：直接用 chat
	if exec := e.registry.Get("chat"); exec != nil {
		g.Log().Warningf(ctx, "[AutoExecutor] 无可用执行器，退化为 chat 模式")
		return exec.Execute(ctx, req)
	}

	return &Result{
		Success: false,
		Output:  "",
		Error:   fmt.Errorf("auto 模式找不到任何可用的执行器"),
	}
}
