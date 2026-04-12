// Package hook 提供工作流完成时的 hook 链机制。
// 每个 hook 失败只记日志不中断后续 hook 执行。
package hook

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// CompletionHook 工作流完成时执行的 hook。
type CompletionHook interface {
	Name() string
	Execute(ctx context.Context, workflowRunID int64) error
}

// Chain 按顺序执行 hook 链，每个 hook 失败只记日志不中断。
type Chain struct {
	hooks []CompletionHook
}

// NewChain 创建 hook 链。
func NewChain(hooks ...CompletionHook) *Chain {
	return &Chain{hooks: hooks}
}

// Run 顺序执行所有 hook，单个 hook 失败记日志后继续执行下一个。
func (c *Chain) Run(ctx context.Context, workflowRunID int64) {
	for _, h := range c.hooks {
		if err := h.Execute(ctx, workflowRunID); err != nil {
			g.Log().Warningf(ctx, "[Hook] %s 执行失败（不中断）: wfRun=%d err=%v", h.Name(), workflowRunID, err)
		}
	}
}
