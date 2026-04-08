package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AnthropicProvider Anthropic Messages API Provider
// 覆盖：Claude Opus, Sonnet, Haiku 系列
type AnthropicProvider struct {
	config Config
	client *http.Client
}

// NewAnthropic 创建 Anthropic Provider
func NewAnthropic(cfg Config) *AnthropicProvider {
	return &AnthropicProvider{
		config: cfg,
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

// --- Anthropic Messages API 数据结构 ---

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content    []anthropicContent `json:"content"`
	StopReason string             `json:"stop_reason"`
	Usage      *anthropicUsage    `json:"usage,omitempty"`
	Type       string             `json:"type"`
	Error      *anthropicError    `json:"error,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Anthropic SSE 事件类型
type anthropicStreamEvent struct {
	Type  string          `json:"type"`
	Delta json.RawMessage `json:"delta,omitempty"`
	Usage json.RawMessage `json:"usage,omitempty"`
}

type anthropicContentDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicMessageDelta struct {
	StopReason string `json:"stop_reason"`
}

// Chat 非流式对话
func (p *AnthropicProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	body, err := p.buildRequest(req, false)
	if err != nil {
		return nil, err
	}

	httpReq, err := p.newHTTPRequest(ctx, body)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w, body: %s", err, string(respBody))
	}

	if result.Error != nil {
		return nil, fmt.Errorf("anthropic error: [%s] %s", result.Error.Type, result.Error.Message)
	}

	// 拼接所有 text 类型的 content
	var content strings.Builder
	for _, c := range result.Content {
		if c.Type == "text" {
			content.WriteString(c.Text)
		}
	}

	chatResp := &ChatResponse{
		Content:      content.String(),
		FinishReason: result.StopReason,
	}
	if result.Usage != nil {
		chatResp.Usage = &TokenUsage{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		}
	}
	return chatResp, nil
}

// ChatStream 流式对话
func (p *AnthropicProvider) ChatStream(ctx context.Context, req *ChatRequest, handler StreamHandler) error {
	body, err := p.buildRequest(req, true)
	if err != nil {
		return err
	}

	httpReq, err := p.newHTTPRequest(ctx, body)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("anthropic stream request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("anthropic stream error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var usage *TokenUsage
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var event anthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			// SSE 流中偶有非 JSON 事件（ping 等），跳过不阻断
			continue
		}

		switch event.Type {
		case "content_block_delta":
			var delta anthropicContentDelta
			if err := json.Unmarshal(event.Delta, &delta); err != nil {
				continue
			}
			if delta.Type == "text_delta" {
				if err := handler(&StreamChunk{Content: delta.Text}); err != nil {
					return err
				}
			}

		case "message_delta":
			var delta anthropicMessageDelta
			if err := json.Unmarshal(event.Delta, &delta); err != nil {
				continue
			}
			// 解析 usage（如果有）
			if event.Usage != nil {
				var u anthropicUsage
				if err := json.Unmarshal(event.Usage, &u); err == nil {
					usage = &TokenUsage{
						PromptTokens:     u.InputTokens,
						CompletionTokens: u.OutputTokens,
						TotalTokens:      u.InputTokens + u.OutputTokens,
					}
				}
			}
			// 发送结束 chunk
			if err := handler(&StreamChunk{
				FinishReason: delta.StopReason,
				Usage:        usage,
			}); err != nil {
				return err
			}

		case "message_start":
			// 解析初始 usage（prompt tokens）
			if event.Usage != nil {
				var u anthropicUsage
				if err := json.Unmarshal(event.Usage, &u); err == nil {
					usage = &TokenUsage{
						PromptTokens: u.InputTokens,
						TotalTokens:  u.InputTokens,
					}
				}
			}

		case "error":
			var errData anthropicError
			if err := json.Unmarshal(event.Delta, &errData); err == nil {
				return fmt.Errorf("anthropic stream error: [%s] %s", errData.Type, errData.Message)
			}
		}
	}

	return scanner.Err()
}

// buildRequest 构建 Anthropic 请求体
func (p *AnthropicProvider) buildRequest(req *ChatRequest, stream bool) ([]byte, error) {
	messages := make([]anthropicMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		// Anthropic Messages API 不支持 system 角色在 messages 中
		if m.Role == RoleSystem {
			continue
		}
		messages = append(messages, anthropicMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	aReq := anthropicRequest{
		Model:     req.Model,
		Messages:  messages,
		System:    req.SystemPrompt,
		MaxTokens: maxTokens,
		Stream:    stream,
	}

	return json.Marshal(aReq)
}

// newHTTPRequest 创建 HTTP 请求
func (p *AnthropicProvider) newHTTPRequest(ctx context.Context, body []byte) (*http.Request, error) {
	url := ResolveBaseURLForProtocol(p.config, TypeAnthropic) + "/v1/messages"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	return httpReq, nil
}
