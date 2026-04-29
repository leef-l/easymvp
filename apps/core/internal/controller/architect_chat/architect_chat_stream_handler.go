package architect_chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/leef-l/easymvp/apps/core/internal/service"
)

// SendMessageStreamHandler handles SSE streaming for architect chat.
// If Brain SSE stream breaks before execution.done, it auto-fallback to sync call.
// Bound manually in cmd.go at /api/v3/projects/{id}/architect-chat/messages/stream
func SendMessageStreamHandler(r *ghttp.Request) {
	ctx := r.GetCtx()
	projectID := r.Get("id").String()
	content := r.Get("content").String()
	if content == "" {
		// Try JSON body parsing (GoFrame r.Parse does not parse JSON body)
		if jsonBody, err := r.GetJson(); err == nil && jsonBody != nil {
			content = jsonBody.Get("content").String()
		}
	}
	if projectID == "" || content == "" {
		r.Response.WriteStatus(http.StatusBadRequest, `{"code":400,"message":"project id and content are required"}`)
		return
	}

	if r.Response.Writer == nil {
		r.Response.WriteStatus(http.StatusInternalServerError, `{"code":500,"message":"streaming not supported"}`)
		return
	}

	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.Header().Set("X-Accel-Buffering", "no")

	eventChan := make(chan service.ArchitectChatStreamEvent, 32)
	go func() {
		defer close(eventChan)
		if err := service.ArchitectChat().SendMessageStream(ctx, service.SendMessageCommand{
			ProjectID: projectID,
			Content:   content,
		}, eventChan); err != nil {
			g.Log().Warningf(ctx, "architect chat stream failed: %v", err)
			eventChan <- service.ArchitectChatStreamEvent{
				Type:    "error",
				Content: err.Error(),
				Done:    true,
			}
		}
	}()

	gotDone := false
	for ev := range eventChan {
		data, _ := json.Marshal(ev)
		fmt.Fprint(r.Response.Writer, "data: ", string(data), "\n\n")
		r.Response.Flush()
		if ev.Done {
			gotDone = true
			break
		}
	}

	// Fallback: if stream broke before execution.done, call sync API to complete.
	if !gotDone {
		g.Log().Infof(ctx, "SSE stream broke before done, falling back to sync call for project=%s", projectID)
		// Give Brain a brief moment to recover before sync call
		time.Sleep(500 * time.Millisecond)
		result, err := service.ArchitectChat().SendMessage(ctx, service.SendMessageCommand{
			ProjectID: projectID,
			Content:   content,
		})
		if err != nil {
			data, _ := json.Marshal(service.ArchitectChatStreamEvent{
				Type:    "error",
				Content: "Stream interrupted and sync fallback also failed: " + err.Error(),
				Done:    true,
			})
			fmt.Fprint(r.Response.Writer, "data: ", string(data), "\n\n")
			r.Response.Flush()
			return
		}
		data, _ := json.Marshal(service.ArchitectChatStreamEvent{
			Type:      "execution.done",
			Content:   result.ArchitectReply,
			Done:      true,
			CommandID: result.CommandID,
			MessageID: result.MessageID,
		})
		fmt.Fprint(r.Response.Writer, "data: ", string(data), "\n\n")
		r.Response.Flush()
	}
}
