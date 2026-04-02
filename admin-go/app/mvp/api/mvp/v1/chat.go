package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// Chat 对话引擎 API（手写，独立于 codegen 生成的 CRUD）

// ChatSendReq 发送消息请求
type ChatSendReq struct {
	g.Meta         `path:"/chat/send" method:"post" tags:"对话引擎" summary:"发送消息"`
	ConversationID snowflake.JsonInt64 `json:"conversationID" v:"required" dc:"对话ID"`
	Content        string              `json:"content" v:"required" dc:"消息内容"`
}

// ChatSendRes 发送消息响应
type ChatSendRes struct {
	g.Meta    `mime:"application/json"`
	MessageID snowflake.JsonInt64 `json:"messageID"` // 用户消息 ID
	ReplyID   snowflake.JsonInt64 `json:"replyID"`   // AI 回复消息 ID（status=streaming）
}

// ChatSSEReq SSE 流式输出连接请求
type ChatSSEReq struct {
	g.Meta    `path:"/chat/sse" method:"get" tags:"对话引擎" summary:"SSE流式输出"`
	MessageID snowflake.JsonInt64 `json:"messageID" v:"required" dc:"AI回复消息ID"`
}

// ChatSSERes SSE 响应（实际通过 SSE 推流，此结构仅用于路由注册）
type ChatSSERes struct {
	g.Meta `mime:"text/event-stream"`
}

// ChatHistoryReq 获取对话历史请求
type ChatHistoryReq struct {
	g.Meta         `path:"/chat/history" method:"get" tags:"对话引擎" summary:"获取对话历史"`
	ConversationID snowflake.JsonInt64 `json:"conversationID" v:"required" dc:"对话ID"`
}

// ChatHistoryRes 获取对话历史响应
type ChatHistoryRes struct {
	g.Meta `mime:"application/json"`
	List   []*ChatMessageOutput `json:"list"`
}

// ChatMessageOutput 消息输出
type ChatMessageOutput struct {
	ID        snowflake.JsonInt64 `json:"id"`
	Role      string              `json:"role"`
	Content   string              `json:"content"`
	Status    string              `json:"status"`
	ModelName string              `json:"modelName,omitempty"`
	CreatedAt string              `json:"createdAt"`
}
