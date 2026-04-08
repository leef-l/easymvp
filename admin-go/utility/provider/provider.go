package provider

import (
	"context"
)

// Role 消息角色
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message 对话消息
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// ChatRequest AI 对话请求
type ChatRequest struct {
	Model        string    `json:"model"`         // 模型代码
	Messages     []Message `json:"messages"`      // 对话历史
	MaxTokens    int       `json:"max_tokens"`    // 最大输出 token
	Temperature  float64   `json:"temperature"`   // 温度（0-2）
	TopP         float64   `json:"top_p"`         // Top-P 采样
	Stream       bool      `json:"stream"`        // 是否流式
	SystemPrompt string    `json:"system_prompt"` // 系统提示词（部分 API 单独传递）
}

// ChatResponse AI 对话响应
type ChatResponse struct {
	Content      string      `json:"content"`       // 回复内容
	FinishReason string      `json:"finish_reason"` // 结束原因：stop/length/error
	Usage        *TokenUsage `json:"usage"`         // token 用量
}

// TokenUsage token 消耗统计
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk 流式输出的单个分片
type StreamChunk struct {
	Content      string      `json:"content"`                 // 增量内容
	FinishReason string      `json:"finish_reason,omitempty"` // 结束原因（最后一个 chunk 才有）
	Usage        *TokenUsage `json:"usage,omitempty"`         // token 用量（最后一个 chunk 才有）
}

// StreamHandler 流式输出回调函数
// 每收到一个 chunk 调用一次，返回 error 可中止流式传输
type StreamHandler func(chunk *StreamChunk) error

// Provider AI 模型统一调用接口
type Provider interface {
	// Chat 非流式对话，返回完整响应
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream 流式对话，通过 handler 回调逐步返回
	ChatStream(ctx context.Context, req *ChatRequest, handler StreamHandler) error
}

// Config Provider 配置
type Config struct {
	ProviderType       string   // 类型：openai_compatible/anthropic/google
	SupportedProtocols []string // 供应商支持的协议类型列表
	BaseURL            string   // API 基础地址
	APIKey             string   // API Key
	APISecret          string   // API Secret（部分供应商需要）
}
