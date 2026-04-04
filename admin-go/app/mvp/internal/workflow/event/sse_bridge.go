package event

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
)

// SSEBridge 将内存 Bus 事件桥接到 SSE HTTP 流。
type SSEBridge struct {
	bus *Bus
}

// NewSSEBridge 创建 SSE 桥接器。
func NewSSEBridge(bus *Bus) *SSEBridge {
	return &SSEBridge{bus: bus}
}

// ServeWorkflowEvents 为指定工作流输出 SSE 事件流。
// 调用方需保证 w 支持 Flush（通常由 http.ResponseWriter 实现）。
func (b *SSEBridge) ServeWorkflowEvents(ctx context.Context, w http.ResponseWriter, workflowRunID int64) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		g.Log().Error(ctx, "[SSEBridge] ResponseWriter does not support Flush")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan Event, 64)
	b.bus.Subscribe("*", func(evt Event) {
		if evt.WorkflowRunID == workflowRunID {
			select {
			case ch <- evt:
			default:
				// 缓冲满则丢弃，避免阻塞
			}
		}
	})

	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-ch:
			data, _ := json.Marshal(evt)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.EventType, data)
			flusher.Flush()
		}
	}
}
