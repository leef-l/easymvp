package hook

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// ReportHook 生成项目报告（占位扩展点）。
type ReportHook struct{}

// Name 返回 hook 名称。
func (h *ReportHook) Name() string { return "report" }

// Execute 生成项目报告，目前只记日志，留扩展点。
func (h *ReportHook) Execute(ctx context.Context, workflowRunID int64) error {
	g.Log().Infof(ctx, "[Hook:report] 工作流 %d 报告生成（占位）", workflowRunID)
	return nil
}
