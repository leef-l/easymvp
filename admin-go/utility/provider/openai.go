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

// OpenAIProvider OpenAI 兼容的 Provider
// 覆盖：OpenAI, DeepSeek, Qwen(阿里), Doubao(字节), GLM(智谱), Moonshot(Kimi), Yi(零一万物), Ollama, Google(Gemini代理)
type OpenAIProvider struct {
	config Config
	client *http.Client
}

// NewOpenAI 创建 OpenAI 兼容 Provider
func NewOpenAI(cfg Config) *OpenAIProvider {
	return &OpenAIProvider{
		config: cfg,
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

// --- OpenAI API 数据结构 ---

type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []openaiChoice `json:"choices"`
	Usage   *openaiUsage   `json:"usage,omitempty"`
	Error   *openaiError   `json:"error,omitempty"`
}

type openaiChoice struct {
	Message      *openaiMessage     `json:"message,omitempty"`
	Delta        *openaiMessage     `json:"delta,omitempty"`
	FinishReason string             `json:"finish_reason,omitempty"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openaiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Chat 非流式对话
func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
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
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result openaiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w, body: %s", err, string(respBody))
	}

	if result.Error != nil {
		return nil, fmt.Errorf("openai error: [%s] %s", result.Error.Type, result.Error.Message)
	}

	if len(result.Choices) == 0 || result.Choices[0].Message == nil {
		return nil, fmt.Errorf("AI 返回空响应")
	}

	chatResp := &ChatResponse{
		Content:      result.Choices[0].Message.Content,
		FinishReason: result.Choices[0].FinishReason,
	}
	if result.Usage != nil {
		chatResp.Usage = &TokenUsage{
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		}
	}
	return chatResp, nil
}

// ChatStream 流式对话
func (p *OpenAIProvider) ChatStream(ctx context.Context, req *ChatRequest, handler StreamHandler) error {
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
		return fmt.Errorf("openai stream request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openai stream error (status %d): %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var streamResp openaiResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) == 0 {
			continue
		}

		choice := streamResp.Choices[0]
		chunk := &StreamChunk{
			FinishReason: choice.FinishReason,
		}
		if choice.Delta != nil {
			chunk.Content = choice.Delta.Content
		}
		if streamResp.Usage != nil {
			chunk.Usage = &TokenUsage{
				PromptTokens:     streamResp.Usage.PromptTokens,
				CompletionTokens: streamResp.Usage.CompletionTokens,
				TotalTokens:      streamResp.Usage.TotalTokens,
			}
		}

		if err := handler(chunk); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// buildRequest 构建 OpenAI 请求体
func (p *OpenAIProvider) buildRequest(req *ChatRequest, stream bool) ([]byte, error) {
	messages := make([]openaiMessage, 0, len(req.Messages)+1)

	// system prompt 作为第一条消息
	if req.SystemPrompt != "" {
		messages = append(messages, openaiMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	for _, m := range req.Messages {
		messages = append(messages, openaiMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}

	oaiReq := openaiRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   stream,
	}
	if req.MaxTokens > 0 {
		oaiReq.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		oaiReq.Temperature = &req.Temperature
	}
	if req.TopP > 0 {
		oaiReq.TopP = &req.TopP
	}

	return json.Marshal(oaiReq)
}

// newHTTPRequest 创建 HTTP 请求
func (p *OpenAIProvider) newHTTPRequest(ctx context.Context, body []byte) (*http.Request, error) {
	url := strings.TrimRight(p.config.BaseURL, "/") + "/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	return httpReq, nil
}
