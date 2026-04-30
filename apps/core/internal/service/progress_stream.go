package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	projectsv1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
)

const (
	progressStreamKeepaliveInterval = 15 * time.Second
)

// StreamProjectProgress 处理 MACCS 工作流实时进度 SSE 流。
// 它订阅 ProgressBus 中指定项目的事件，将其以 SSE 格式推送给客户端，
// 直到客户端断开连接或 context 取消。
func StreamProjectProgress(ctx context.Context, req *projectsv1.ProjectProgressStreamReq) error {
	httpReq := g.RequestFromCtx(ctx)
	if httpReq == nil {
		return gerror.New("progress stream request context is unavailable")
	}

	projectID := strings.TrimSpace(req.Id)
	if projectID == "" {
		return gerror.New("project id is required for progress stream")
	}

	// 设置 SSE 响应头。
	w := httpReq.Response.RawWriter()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	flusher, hasFlusher := w.(http.Flusher)

	// 写入重连指令。
	if err := writeProgressSSEControlFrame(httpReq, "retry: 3000\n\n"); err != nil {
		return gerror.Wrap(err, "write progress retry frame failed")
	}
	if hasFlusher {
		flusher.Flush()
	}

	// 订阅进度事件。
	bus := ProgressBusInstance()
	ch := bus.Subscribe(projectID)
	defer bus.Unsubscribe(projectID, ch)

	keepaliveTicker := time.NewTicker(progressStreamKeepaliveInterval)
	defer keepaliveTicker.Stop()

	reqCtx := httpReq.Request.Context()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-reqCtx.Done():
			return nil
		case evt, ok := <-ch:
			if !ok {
				return nil
			}
			if err := writeProgressSSEEvent(httpReq, evt); err != nil {
				g.Log().Warningf(ctx, "[progress-stream] write event failed for project %s: %v", projectID, err)
				return nil
			}
			if hasFlusher {
				flusher.Flush()
			}
		case <-keepaliveTicker.C:
			if err := writeProgressSSEControlFrame(httpReq, ": keepalive\n\n"); err != nil {
				return gerror.Wrap(err, "write progress keepalive frame failed")
			}
			if hasFlusher {
				flusher.Flush()
			}
		}
	}
}

// writeProgressSSEEvent 将 ProgressEvent 序列化为 SSE data 帧写入响应。
func writeProgressSSEEvent(httpReq *ghttp.Request, evt ProgressEvent) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return gerror.Wrap(err, "marshal progress sse payload failed")
	}

	var frame strings.Builder
	frame.WriteString("event: ")
	frame.WriteString(cleanProgressSSELine(evt.Type))
	frame.WriteString("\n")
	for _, line := range strings.Split(string(payload), "\n") {
		frame.WriteString("data: ")
		frame.WriteString(line)
		frame.WriteString("\n")
	}
	frame.WriteString("\n")
	_, err = httpReq.Response.RawWriter().Write([]byte(frame.String()))
	return err
}

// writeProgressSSEControlFrame 写入原始 SSE 控制帧（如 retry:、注释行）。
func writeProgressSSEControlFrame(httpReq *ghttp.Request, frame string) error {
	_, err := httpReq.Response.RawWriter().Write([]byte(frame))
	return err
}

func cleanProgressSSELine(value string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(strings.TrimSpace(value))
}
