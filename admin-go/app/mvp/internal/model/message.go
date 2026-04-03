package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

const (
	MessageTypeGeneral      = "general"
	MessageTypeChatUser     = "chat_user"
	MessageTypeChatReply    = "chat_reply"
	MessageTypeTaskPrompt   = "task_prompt"
	MessageTypeTaskReply    = "task_reply"
	MessageTypeSystemNotice = "system_notice"
	MessageTypePoison       = "poison"
)

// Message DTO 模型

// MessageCreateInput 创建MVP消息表输入
type MessageCreateInput struct {
	ConversationID snowflake.JsonInt64 `json:"conversationID"`
	Role string `json:"role"`
	MessageType string `json:"messageType"`
	Content string `json:"content"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	TokenUsage string `json:"tokenUsage"`
	Status string `json:"status"`
}

// MessageUpdateInput 更新MVP消息表输入
type MessageUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ConversationID snowflake.JsonInt64 `json:"conversationID"`
	Role string `json:"role"`
	MessageType string `json:"messageType"`
	Content string `json:"content"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	TokenUsage string `json:"tokenUsage"`
	Status string `json:"status"`
}

// MessageDetailOutput MVP消息表详情输出
type MessageDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ConversationID snowflake.JsonInt64 `json:"conversationID"`
	ConversationTitle string `json:"conversationTitle"`
	Role string `json:"role"`
	MessageType string `json:"messageType"`
	Content string `json:"content"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	TokenUsage string `json:"tokenUsage"`
	Status string `json:"status"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// MessageListOutput MVP消息表列表输出
type MessageListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ConversationID snowflake.JsonInt64 `json:"conversationID"`
	ConversationTitle string `json:"conversationTitle"`
	Role string `json:"role"`
	MessageType string `json:"messageType"`
	Content string `json:"content"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	TokenUsage string `json:"tokenUsage"`
	Status string `json:"status"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// MessageListInput MVP消息表列表查询输入
type MessageListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	MessageType string `json:"messageType"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}


// MessageBatchUpdateInput 批量编辑MVP消息表输入
type MessageBatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *string               `json:"status"`
}

