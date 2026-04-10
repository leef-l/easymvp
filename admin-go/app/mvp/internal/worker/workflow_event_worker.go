package worker

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/eventstream"
)

// StartWorkflowEventWorker 启动工作流事件流消费 worker。
// 具体 consumer 构建与接线由上层负责（如 registry 初始化阶段）。
func StartWorkflowEventWorker(ctx context.Context, consumer *eventstream.Consumer) {
	if consumer == nil {
		g.Log().Warningf(ctx, "[WorkflowEventWorker] consumer is nil, skip start")
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		defer func() {
			if rec := recover(); rec != nil {
				g.Log().Errorf(runCtx, "[WorkflowEventWorker] panic: %v", rec)
			}
		}()
		consumer.Start(runCtx)
	}()
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				if consumer.IsStarted() {
					consumer.PulseHeartbeat()
				}
			}
		}
	}()
}
