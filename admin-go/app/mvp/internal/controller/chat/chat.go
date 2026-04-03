package chat

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/utility/snowflake"
)

var Chat = cChat{}

type cChat struct{}

// Send 发送消息
func (c *cChat) Send(ctx context.Context, req *v1.ChatSendReq) (res *v1.ChatSendRes, err error) {
	userID := middleware.GetUserID(ctx)
	deptID := middleware.GetDeptID(ctx)

	conv, convErr := g.DB().Model("mvp_conversation").Where("id", req.ConversationID).Where("deleted_at IS NULL").One()
	if convErr != nil || conv.IsEmpty() {
		return nil, fmt.Errorf("对话不存在")
	}
	if conv["created_by"].Int64() != userID && userID != 1 {
		return nil, fmt.Errorf("无权操作该对话")
	}

	msgID, replyID, err := engine.GetEngine().SendMessage(ctx, int64(req.ConversationID), req.Content, userID, deptID)
	if err != nil {
		return nil, err
	}

	return &v1.ChatSendRes{
		MessageID: snowflake.JsonInt64(msgID),
		ReplyID:   snowflake.JsonInt64(replyID),
	}, nil
}

// SSE 流式输出端点
func (c *cChat) SSE(ctx context.Context, req *v1.ChatSSEReq) (res *v1.ChatSSERes, err error) {
	r := g.RequestFromCtx(ctx)
	messageID := int64(req.MessageID)

	// 设置 SSE 响应头
	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.Header().Set("Access-Control-Allow-Origin", "*")

	// 1. 先检查消息状态，如果已完成直接返回全部内容
	msg, msgErr := g.DB().Model("mvp_message").Where("id", messageID).Where("deleted_at IS NULL").One()
	if msgErr != nil || msg.IsEmpty() {
		writeSSEEvent(r, "error", `{"error":"消息不存在"}`)
		writeSSEEvent(r, "done", `{"done":true}`)
		r.Response.Flush()
		return nil, nil
	}

	userID := middleware.GetUserID(ctx)
	if userID != 1 {
		convID := msg["conversation_id"].Int64()
		conv, _ := g.DB().Model("mvp_conversation").Where("id", convID).Where("deleted_at IS NULL").One()
		if conv.IsEmpty() || conv["created_by"].Int64() != userID {
			writeSSEEvent(r, "error", `{"error":"无权访问"}`)
			writeSSEEvent(r, "done", `{"done":true}`)
			r.Response.Flush()
			return nil, nil
		}
	}

	status := msg["status"].String()

	if status == "completed" || status == "failed" {
		// 消息已完成，发送全部内容
		writeSSEEvent(r, "full", fmt.Sprintf(`{"content":%q,"status":"%s"}`, msg["content"].String(), status))
		writeSSEEvent(r, "done", `{"done":true}`)
		r.Response.Flush()
		return nil, nil
	}

	// 2. 消息正在 streaming，先发送已有的 chunks
	chunks, _ := g.DB().Model("mvp_message_chunk").
		Where("message_id", messageID).
		Order("chunk_index ASC").
		All()
	for _, chunk := range chunks {
		writeSSEEvent(r, "chunk", fmt.Sprintf(`{"content":%q,"index":%d}`, chunk["content"].String(), chunk["chunk_index"].Int()))
	}
	r.Response.Flush()

	// 3. 订阅后续的实时 chunks
	hub := engine.GetHub()
	ch, unsubscribe := hub.Subscribe(messageID)
	defer unsubscribe()

	// 使用 HTTP 请求的 context 来检测客户端断开
	httpCtx := r.Context()

	for {
		select {
		case data, ok := <-ch:
			if !ok {
				// channel 关闭，流式输出完成
				writeSSEEvent(r, "done", `{"done":true}`)
				r.Response.Flush()
				return nil, nil
			}
			writeSSEEvent(r, "chunk", data)
			r.Response.Flush()
		case <-httpCtx.Done():
			// 客户端断开连接（不影响后端 goroutine 继续运行）
			return nil, nil
		}
	}
}

// History 获取对话历史
func (c *cChat) History(ctx context.Context, req *v1.ChatHistoryReq) (res *v1.ChatHistoryRes, err error) {
	userID := middleware.GetUserID(ctx)
	if userID != 1 {
		conv, convErr := g.DB().Model("mvp_conversation").Where("id", req.ConversationID).Where("deleted_at IS NULL").One()
		if convErr != nil || conv.IsEmpty() || conv["created_by"].Int64() != userID {
			return nil, fmt.Errorf("无权访问该对话")
		}
	}

	records, err := g.DB().Model("mvp_message m").
		LeftJoin("ai_model am", "am.id = m.model_id").
		Fields("m.id, m.role, m.message_type, m.content, m.status, m.created_at, am.name as model_name").
		Where("m.conversation_id", req.ConversationID).
		Where("m.deleted_at IS NULL").
		Order("m.created_at ASC").
		All()
	if err != nil {
		return nil, err
	}

	list := make([]*v1.ChatMessageOutput, 0, len(records))
	for _, r := range records {
		list = append(list, &v1.ChatMessageOutput{
			ID:          snowflake.JsonInt64(r["id"].Int64()),
			Role:        r["role"].String(),
			MessageType: r["message_type"].String(),
			Content:     r["content"].String(),
			Status:      r["status"].String(),
			ModelName:   r["model_name"].String(),
			CreatedAt:   gtime.New(r["created_at"]).String(),
		})
	}

	return &v1.ChatHistoryRes{List: list}, nil
}

// writeSSEEvent 写入 SSE 事件
func writeSSEEvent(r *ghttp.Request, event string, data string) {
	r.Response.Writef("event: %s\ndata: %s\n\n", event, data)
}
