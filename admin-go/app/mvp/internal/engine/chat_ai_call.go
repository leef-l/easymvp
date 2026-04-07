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

// runAICall 异步调用 AI（goroutine 中运行，不依赖前端连接）
func (e *ChatEngine) runAICall(conversationID int64, replyID int64, modelInfo *ModelInfo) {
	ctx := context.Background()

	// 1. 获取对话历史
	messages, err := e.loadHistory(ctx, conversationID, replyID)
	if err != nil {
		e.failMessage(ctx, replyID, err.Error())
		return
	}

	// 2. 创建 Provider
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
	})
	if err != nil {
		e.failMessage(ctx, replyID, err.Error())
		return
	}

	// 3. 构建请求
	req := &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     messages,
		MaxTokens:    modelInfo.MaxTokens,
		Temperature:  0.7,
		Stream:       true,
		SystemPrompt: modelInfo.SystemPrompt,
	}

	// 4. 流式调用 AI（带重试 + 自动续写：截断时自动发"继续"续写，最多 5 轮）
	const maxRetries = 2
	const maxContinueRounds = 5 // 自动续写上限
	var fullContent strings.Builder
	chunkIndex := 0
	var lastFinishReason string

	streamHandler := func(chunk *provider.StreamChunk) error {
		if chunk.Content != "" {
			fullContent.WriteString(chunk.Content)
			chunkIndex++

			// 写入 message_chunk 表
			_, insertErr := g.DB().Model("mvp_message_chunk").Insert(g.Map{
				"message_id":  replyID,
				"chunk_index": chunkIndex,
				"content":     chunk.Content,
				"created_at":  gtime.Now(),
			})
			if insertErr != nil {
				g.Log().Errorf(ctx, "写入 chunk 失败: %v", insertErr)
			}
			activity.TouchMessageActivity(ctx, replyID)
			activity.TouchConversationActivity(ctx, conversationID)

			// 推送到 SSE Hub
			chunkJSON, _ := json.Marshal(map[string]interface{}{
				"content": chunk.Content,
				"index":   chunkIndex,
			})
			e.hub.Publish(replyID, string(chunkJSON))
		}

		// 最后一个 chunk（有 finish_reason）
		if chunk.FinishReason != "" {
			lastFinishReason = chunk.FinishReason
			// 更新 token 用量
			if chunk.Usage != nil {
				usageJSON, _ := json.Marshal(chunk.Usage)
				if _, err := g.DB().Model("mvp_message").Where("id", replyID).Update(g.Map{
					"token_usage": string(usageJSON),
				}); err != nil {
					g.Log().Errorf(ctx, "[ChatEngine] 更新 token_usage 失败: msg=%d, err=%v", replyID, err)
				}
			}
		}

		return nil
	}

	// 外层循环：自动续写（被截断时追加"继续"重新调用）
	for round := 0; round <= maxContinueRounds; round++ {
		lastFinishReason = ""

		// 内层循环：瞬时错误重试
		var lastErr error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if attempt > 0 {
				g.Log().Warningf(ctx, "[ChatEngine] AI 调用第 %d 次重试 (messageID=%d): %v", attempt, replyID, lastErr)
				time.Sleep(time.Duration(attempt*2) * time.Second)
			}
			lastErr = p.ChatStream(ctx, req, streamHandler)
			if lastErr == nil {
				break
			}
			errMsg := lastErr.Error()
			isRetryable := strings.Contains(errMsg, "status 500") ||
				strings.Contains(errMsg, "EOF") ||
				strings.Contains(errMsg, "deadline exceeded") ||
				strings.Contains(errMsg, "connection reset")
			if !isRetryable {
				break
			}
		}

		if lastErr != nil {
			// 续写失败：如果已有部分内容，保留为 completed 而不是 failed
			if fullContent.Len() > 0 {
				g.Log().Warningf(ctx, "[ChatEngine] 续写第 %d 轮失败但已有内容，保留为 completed: messageID=%d err=%v", round, replyID, lastErr)
				break // 跳出续写循环，走正常完成流程
			}
			e.failMessage(ctx, replyID, lastErr.Error())
			return
		}

		// 检查是否被截断（finish_reason=length 或 max_tokens）
		isTruncated := lastFinishReason == "length" || lastFinishReason == "max_tokens"
		if !isTruncated || round == maxContinueRounds {
			break
		}

		// 被截断：追加当前已有内容为 assistant 消息，再加"继续"指令，重新调用
		g.Log().Infof(ctx, "[ChatEngine] 回复被截断(reason=%s)，自动续写第 %d 轮 (messageID=%d)",
			lastFinishReason, round+1, replyID)

		// 通知前端正在续写
		hub.Publish(replyID, fmt.Sprintf(`{"event":"continue","round":%d}`, round+1))

		// 更新请求的消息列表：追加已有回复 + "继续"指令
		req.Messages = append(req.Messages,
			provider.Message{Role: provider.RoleAssistant, Content: fullContent.String()},
			provider.Message{Role: provider.RoleUser, Content: "继续，从上次中断的地方接着输出，不要重复已输出的内容。"},
		)
	}

	// 5. 更新消息为完成状态
	_, updateErr := g.DB().Model("mvp_message").Where("id", replyID).Update(g.Map{
		"content":    fullContent.String(),
		"status":     "completed",
		"updated_at": gtime.Now(),
	})
	if updateErr != nil {
		g.Log().Errorf(ctx, "更新消息状态失败: %v", updateErr)
	}

	// 6. 通知 SSE Hub 流式输出完成
	doneJSON, _ := json.Marshal(map[string]interface{}{
		"done": true,
	})
	e.hub.Publish(replyID, string(doneJSON))

	// 7. 飞书主动推送：将 AI 回复发给对话绑定用户（异步，不阻塞）
	go func() {
		feishuNotifyAIReply(ctx, conversationID, fullContent.String())
	}()

	// 短暂延迟后关闭 channel，让前端有时间接收最后的消息
	time.Sleep(100 * time.Millisecond)
	e.hub.Done(replyID)
}

// feishuNotifyAIReply 推送 AI 回复到飞书（避免循环引用，用函数变量注入）。
var feishuNotifyAIReply = func(ctx context.Context, conversationID int64, content string) {}

// tryParseArchitectTasks 尝试从架构师回复中解析任务清单
func (e *ChatEngine) tryParseArchitectTasks(conversationID int64, aiReply string) {
	ctx := context.Background()

	// 查对话的角色类型和项目ID
	conv, err := g.DB().Model("mvp_conversation").Where("id", conversationID).One()
	if err != nil || conv.IsEmpty() {
		return
	}

	// 只有架构师对话且是项目级对话（task_id 为空）才解析
	if conv["role_type"].String() != "architect" || conv["task_id"].Int64() != 0 {
		return
	}

	projectID := conv["project_id"].Int64()

	// 判断引擎版本
	ev, _ := g.DB().Model("mvp_project").Where("id", projectID).Value("engine_version")
	if ev.String() == "workflow_v2" {
		e.tryParseArchitectBlueprints(ctx, projectID, conv["id"].Int64(), conversationID, aiReply)
		return
	}

	// Legacy：写入 mvp_task
	count, err := GetParser().ParseAndCreateTasks(ctx, projectID, aiReply)
	if err != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 解析任务失败: %v", err)
		return
	}
	if count > 0 {
		g.Log().Infof(ctx, "[ChatEngine] 架构师回复中解析出 %d 个任务（draft），项目 %d", count, projectID)
	}
}

// BlueprintCreator V2 蓝图创建回调，由 orchestrator 包注册。
// 避免 engine→orchestrator 循环依赖。
type BlueprintCreator func(ctx context.Context, projectID, workflowRunID, conversationID, messageID int64, tasks []ArchitectTask) (planVersionID int64, count int, err error)

var blueprintCreatorFn BlueprintCreator

// RegisterBlueprintCreator 注册 V2 蓝图创建回调（应用启动时由 orchestrator.Init 调用）。
func RegisterBlueprintCreator(fn BlueprintCreator) {
	blueprintCreatorFn = fn
}

// tryParseArchitectBlueprints V2 专用：解析 AI 回复并创建蓝图。
func (e *ChatEngine) tryParseArchitectBlueprints(ctx context.Context, projectID, conversationID, messageID int64, aiReply string) {
	if blueprintCreatorFn == nil {
		g.Log().Warningf(ctx, "[ChatEngine] V2 蓝图创建回调未注册，跳过")
		return
	}

	// 获取项目分类
	projectCategory, _ := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).Value("project_category")

	tasks, err := GetParser().ExtractAndNormalize(ctx, aiReply, projectCategory.String())
	if err != nil || len(tasks) == 0 {
		return
	}

	// 查活跃的 workflow_run
	wfRun, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"pending", "running", "paused"}).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	var wfRunID int64
	if !wfRun.IsEmpty() {
		wfRunID = wfRun["id"].Int64()
	}

	pvID, bpCount, err := blueprintCreatorFn(ctx, projectID, wfRunID, conversationID, messageID, tasks)
	if err != nil {
		g.Log().Warningf(ctx, "[ChatEngine] V2 创建蓝图失败: %v", err)
		return
	}
	g.Log().Infof(ctx, "[ChatEngine] V2 架构师回复解析出 %d 个蓝图, planVersion=%d, 项目 %d", bpCount, pvID, projectID)
}