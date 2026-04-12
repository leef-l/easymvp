package hook

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// NotifyHook 发送完成通知（飞书等）（占位扩展点）。
type NotifyHook struct{}

// Name 返回 hook 名称。
func (h *NotifyHook) Name() string { return "notify" }

// Execute 发送完成通知，目前只记日志，留扩展点。
func (h *NotifyHook) Execute(ctx context.Context, workflowRunID int64) error {
	g.Log().Infof(ctx, "[Hook:notify] 工作流 %d 完成通知（占位）", workflowRunID)
	return nil
}
