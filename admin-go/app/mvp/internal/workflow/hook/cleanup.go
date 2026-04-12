package hook

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// CleanupHook 清理运行时资源（worktree 等）（占位扩展点）。
type CleanupHook struct{}

// Name 返回 hook 名称。
func (h *CleanupHook) Name() string { return "cleanup" }

// Execute 清理运行时资源，目前只记日志，留扩展点。
func (h *CleanupHook) Execute(ctx context.Context, workflowRunID int64) error {
	g.Log().Infof(ctx, "[Hook:cleanup] 工作流 %d 资源清理（占位）", workflowRunID)
	return nil
}
