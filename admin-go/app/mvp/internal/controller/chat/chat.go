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
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

var Chat = cChat{}

type cChat struct{}

// Send 发送消息
func (c *cChat) Send(ctx context.Context, req *v1.ChatSendReq) (res *v1.ChatSendRes, err error) {
	conversationRepo := repo.NewConversationRepo()

	userID := middleware.GetUserID(ctx)
	deptID := middleware.GetDeptID(ctx)

	conv, convErr := conversationRepo.GetByID(ctx, int64(req.ConversationID), "id", "created_by")
	if convErr != nil || len(conv) == 0 {
		return nil, fmt.Errorf("对话不存在")
	}
	if mapToDBRecord(conv)["created_by"].Int64() != userID && userID != 1 {
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
	var (
		conversationRepo = repo.NewConversationRepo()
		messageChunkRepo = repo.NewMessageChunkRepo()
		messageRepo      = repo.NewMessageRepo()
	)

	r := g.RequestFromCtx(ctx)
	messageID := int64(req.MessageID)

	// 设置 SSE 响应头
	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.Header().Set("Access-Control-Allow-Origin", "*")

	// 1. 先检查消息状态，如果已完成直接返回全部内容
	msg, msgErr := messageRepo.GetByID(ctx, messageID, "id", "conversation_id", "content", "status")
	if msgErr != nil || len(msg) == 0 {
		writeSSEEvent(r, "error", `{"error":"消息不存在"}`)
		writeSSEEvent(r, "done", `{"done":true}`)
		r.Response.Flush()
		return nil, nil
	}
	msgRecord := mapToDBRecord(msg)

	userID := middleware.GetUserID(ctx)
	if userID != 1 {
		convID := msgRecord["conversation_id"].Int64()
		conv, convErr := conversationRepo.GetByID(ctx, convID, "id", "created_by")
		if convErr != nil || len(conv) == 0 || mapToDBRecord(conv)["created_by"].Int64() != userID {
			writeSSEEvent(r, "error", `{"error":"无权访问"}`)
			writeSSEEvent(r, "done", `{"done":true}`)
			r.Response.Flush()
			return nil, nil
		}
	}

	status := msgRecord["status"].String()

	if status == "completed" || status == "failed" {
		// 消息已完成，发送全部内容
		writeSSEEvent(r, "full", fmt.Sprintf(`{"content":%q,"status":"%s"}`, msgRecord["content"].String(), status))
		writeSSEEvent(r, "done", `{"done":true}`)
		r.Response.Flush()
		return nil, nil
	}

	// 2. 消息正在 streaming，先发送已有的 chunks
	chunks, chunkErr := messageChunkRepo.ListByMessage(ctx, messageID)
	if chunkErr != nil {
		g.Log().Warningf(ctx, "[SSE] 查询 chunks 失败: msg=%d err=%v", messageID, chunkErr)
	}
	for _, chunk := range chunks {
		record := mapToDBRecord(chunk)
		writeSSEEvent(r, "chunk", fmt.Sprintf(`{"content":%q,"index":%d}`, record["content"].String(), record["chunk_index"].Int()))
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
	var (
		conversationRepo = repo.NewConversationRepo()
		messageRepo      = repo.NewMessageRepo()
	)

	userID := middleware.GetUserID(ctx)
	if userID != 1 {
		conv, convErr := conversationRepo.GetByID(ctx, int64(req.ConversationID), "id", "created_by")
		if convErr != nil || len(conv) == 0 || mapToDBRecord(conv)["created_by"].Int64() != userID {
			return nil, fmt.Errorf("无权访问该对话")
		}
	}

	records, err := messageRepo.ListHistoryByConversation(ctx, int64(req.ConversationID))
	if err != nil {
		return nil, err
	}

	list := make([]*v1.ChatMessageOutput, 0, len(records))
	for _, r := range records {
		record := mapToDBRecord(r)
		list = append(list, &v1.ChatMessageOutput{
			ID:          snowflake.JsonInt64(record["id"].Int64()),
			Role:        record["role"].String(),
			MessageType: record["message_type"].String(),
			Content:     record["content"].String(),
			Status:      record["status"].String(),
			ModelName:   record["model_name"].String(),
			CreatedAt:   gtime.New(record["created_at"]).String(),
		})
	}

	return &v1.ChatHistoryRes{List: list}, nil
}

// writeSSEEvent 写入 SSE 事件
func writeSSEEvent(r *ghttp.Request, event string, data string) {
	r.Response.Writef("event: %s\ndata: %s\n\n", event, data)
}
