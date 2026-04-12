// stream_caller.go 流式 AI 调用统一封装：重试 + 自动续写 + chunk 落盘 + SSE 推送。
//
// 从 chat_ai_call.go 和 executor_dispatch.go 中提取的公共逻辑。
// 两处原始代码有 ~150 行几乎相同的重试+续写+chunk写入，现在统一为 StreamCall。
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/activity"
	"easymvp/utility/provider"
)

// StreamCallConfig 流式调用配置
type StreamCallConfig struct {
	Provider      provider.Provider
	Request       *provider.ChatRequest
	ReplyID       int64 // 消息 ID（chunk 关联 + SSE 推送 key）
	Hub           *SSEHub
	MaxRetries    int // 瞬时错误重试次数（默认 2）
	MaxContinue   int // 自动续写轮次上限（默认 5）
	HeartbeatFn   func() // 每 10 个 chunk 调用一次（可选，executor 用来喂 watchdog）
	ConversationID int64 // 用于 activity 追踪（可选）
	TaskID         int64 // 用于 activity 追踪（可选）
}

// StreamCallResult 流式调用结果
type StreamCallResult struct {
	Content      string
	TokenUsage   json.RawMessage
	Truncated    bool // 最终是否仍被截断（达到续写上限）
}

// StreamCall 执行流式 AI 调用，带重试和自动续写。
// 返回完整内容或错误。部分内容 + 错误时，Content 非空且 err 非 nil。
func StreamCall(ctx context.Context, cfg *StreamCallConfig) (*StreamCallResult, error) {
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 2
	}
	if cfg.MaxContinue <= 0 {
		cfg.MaxContinue = 5
	}

	var fullContent strings.Builder
	chunkIndex := 0
	var lastFinishReason string
	var lastUsage json.RawMessage

	streamHandler := func(chunk *provider.StreamChunk) error {
		if chunk.Content != "" {
			fullContent.WriteString(chunk.Content)
			chunkIndex++

			// 心跳
			if cfg.HeartbeatFn != nil && chunkIndex%10 == 0 {
				cfg.HeartbeatFn()
			}

			// chunk 落盘
			if _, insertErr := g.DB().Model("mvp_message_chunk").Insert(g.Map{
				"message_id":  cfg.ReplyID,
				"chunk_index": chunkIndex,
				"content":     chunk.Content,
				"created_at":  gtime.Now(),
			}); insertErr != nil {
				g.Log().Errorf(ctx, "[StreamCall] 写入 chunk 失败: msg=%d, err=%v", cfg.ReplyID, insertErr)
			}

			// activity 追踪
			activity.TouchMessageActivity(ctx, cfg.ReplyID)
			if cfg.ConversationID > 0 {
				activity.TouchConversationActivity(ctx, cfg.ConversationID)
			}
			if cfg.TaskID > 0 {
				activity.TouchTaskActivity(ctx, cfg.TaskID)
			}

			// SSE 推送
			chunkJSON, cErr := json.Marshal(map[string]interface{}{
				"content": chunk.Content,
				"index":   chunkIndex,
			})
			if cErr != nil {
				chunkJSON = []byte(`{"content":"","index":0}`)
			}
			cfg.Hub.Publish(cfg.ReplyID, string(chunkJSON))
		}

		if chunk.FinishReason != "" {
			lastFinishReason = chunk.FinishReason
			if chunk.Usage != nil {
				usageJSON, uErr := json.Marshal(chunk.Usage)
				if uErr != nil {
					usageJSON = []byte("{}")
				}
				lastUsage = usageJSON
				if _, err := g.DB().Model("mvp_message").Ctx(ctx).Where("id", cfg.ReplyID).Update(g.Map{
					"token_usage": string(usageJSON),
				}); err != nil {
					g.Log().Errorf(ctx, "[StreamCall] 更新 token_usage 失败: msg=%d, err=%v", cfg.ReplyID, err)
				}
			}
		}

		return nil
	}

	req := cfg.Request

	for round := 0; round <= cfg.MaxContinue; round++ {
		lastFinishReason = ""

		// 瞬时错误重试
		var callErr error
		for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
			if attempt > 0 {
				g.Log().Warningf(ctx, "[StreamCall] AI 调用第 %d 次重试 (msg=%d): %v", attempt, cfg.ReplyID, callErr)
				time.Sleep(time.Duration(attempt*2) * time.Second)
			}
			callErr = cfg.Provider.ChatStream(ctx, req, streamHandler)
			if callErr == nil {
				break
			}
			if ctx.Err() != nil {
				break
			}
			errMsg := callErr.Error()
			isRetryable := strings.Contains(errMsg, "status 500") ||
				strings.Contains(errMsg, "EOF") ||
				strings.Contains(errMsg, "connection reset")
			if !isRetryable {
				break
			}
		}

		if callErr != nil {
			// 续写失败但已有部分内容 → 保留内容，返回部分结果
			if fullContent.Len() > 0 {
				g.Log().Warningf(ctx, "[StreamCall] 续写第 %d 轮失败但已有内容，保留: msg=%d err=%v", round, cfg.ReplyID, callErr)
				return &StreamCallResult{
					Content:    fullContent.String(),
					TokenUsage: lastUsage,
					Truncated:  true,
				}, nil
			}
			return nil, callErr
		}

		// 检查是否被截断
		isTruncated := lastFinishReason == "length" || lastFinishReason == "max_tokens"
		if !isTruncated || round == cfg.MaxContinue {
			return &StreamCallResult{
				Content:    fullContent.String(),
				TokenUsage: lastUsage,
				Truncated:  isTruncated,
			}, nil
		}

		// 被截断：自动续写
		g.Log().Infof(ctx, "[StreamCall] 回复被截断(reason=%s)，自动续写第 %d 轮 (msg=%d)",
			lastFinishReason, round+1, cfg.ReplyID)

		if cfg.HeartbeatFn != nil {
			cfg.HeartbeatFn()
		}

		// 通知前端正在续写
		cfg.Hub.Publish(cfg.ReplyID, fmt.Sprintf(`{"event":"continue","round":%d}`, round+1))

		req.Messages = append(req.Messages,
			provider.Message{Role: provider.RoleAssistant, Content: fullContent.String()},
			provider.Message{Role: provider.RoleUser, Content: "继续，从上次中断的地方接着输出，不要重复已输出的内容。"},
		)
	}

	return &StreamCallResult{
		Content:    fullContent.String(),
		TokenUsage: lastUsage,
		Truncated:  false,
	}, nil
}
