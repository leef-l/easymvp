// executor_dispatch.go 任务执行与分发（Legacy 链路）。
//
// Deprecated: 本文件中的 Executor 及其方法将在后续 PR 中迁移到 workflow/executor/dispatcher.go。
// 当前仍是 legacy 任务调度的核心路径，请勿直接删除。
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/activity"
	"easymvp/app/mvp/internal/consts"
	mvpmodel "easymvp/app/mvp/internal/model"
	"easymvp/utility/provider"
	"easymvp/utility/snowflake"
)

// V2ExecutorFn V2 执行器回调：给定 executionMode，如果 V2 支持则执行并返回 true。
// 由外部注入（orchestrator 初始化时），避免循环依赖。
type V2ExecutorFn func(ctx context.Context, projectID, taskID int64, task gdb.Record, modelInfo *ModelInfo, executionMode string) (handled bool)

// Executor 任务执行器，为每个任务启动 goroutine 调用 AI
type Executor struct {
	scheduler  *Scheduler
	v2Executor V2ExecutorFn
}

// NewExecutor 创建执行器
func NewExecutor(scheduler *Scheduler) *Executor {
	return &Executor{scheduler: scheduler}
}

// SetV2Executor 注入 V2 执行器回调（由 orchestrator 在初始化时注入）。
func (e *Executor) SetV2Executor(fn V2ExecutorFn) { e.v2Executor = fn }

// Execute 执行单个任务
// 根据 execution_mode 分发执行方式：aider/chat/openhands
func (e *Executor) Execute(ctx context.Context, projectID int64, taskID int64) {
	// 防 panic 兜底：确保资源锁一定会被释放
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[Executor] panic recovered: task=%d, err=%v", taskID, r)
			e.scheduler.releaseTaskResources(taskID)
			e.scheduler.OnTaskFailed(projectID, taskID, fmt.Sprintf("内部错误(panic): %v", r))
		}
	}()

	// 1. 查询任务信息
	task, err := g.DB().Model("mvp_task").Ctx(ctx).Where("id", taskID).WhereNull("deleted_at").One()
	if err != nil || task.IsEmpty() {
		e.scheduler.OnTaskFailed(projectID, taskID, "任务不存在")
		return
	}

	roleType := task["role_type"].String()
	modelID := task["model_id"].Int64()

	// 2. 获取模型信息
	modelInfo, err := e.resolveTaskModel(ctx, projectID, taskID, roleType, modelID)
	if err != nil {
		e.handleTaskFailure(ctx, projectID, taskID, roleType, classifyTaskConfigError(err), err.Error())
		return
	}

	// 3. 根据 execution_mode 分发执行方式
	executionMode := e.getExecutionMode(ctx, projectID, roleType, task["role_level"].String())

	// 3a. 尝试 V2 执行器（如果已注入）；V2 支持 aider/openhands/claude_code/codex_cli/gemini_cli/chat 等
	if e.v2Executor != nil {
		if handled := e.v2Executor(ctx, projectID, taskID, task, modelInfo, executionMode); handled {
			return
		}
		g.Log().Infof(ctx, "[Executor] V2 执行器未处理 mode=%s，回退 legacy: task=%d", executionMode, taskID)
	}

	// 3b. Legacy 分发（V2 未注入或不支持时走旧链路）
	switch executionMode {
	case "aider":
		e.executeWithAider(ctx, projectID, taskID, task, modelInfo)
		return
	case "openhands":
		g.Log().Infof(ctx, "[Executor] openhands 模式（legacy）暂未实现，回退到 chat: task=%d", taskID)
	}

	// 4. 默认 chat 模式 → ChatStream 对话
	e.executeChatMode(ctx, projectID, taskID, task, roleType, modelInfo)
}

// executeChatMode 执行 chat 模式任务（ChatStream 对话）
func (e *Executor) executeChatMode(ctx context.Context, projectID int64, taskID int64, task gdb.Record, roleType string, modelInfo *ModelInfo) {
	// 查找或创建任务对话
	conversationID, err := e.ensureConversation(ctx, projectID, taskID, roleType)
	if err != nil {
		e.handleTaskFailure(ctx, projectID, taskID, roleType, taskFailureExecution, err.Error())
		return
	}

	// 回写 conversation_id 到 task，方便前端和 watchdog 检测
	if _, err := g.DB().Model("mvp_task").Ctx(ctx).Where("id", taskID).Update(g.Map{
		"conversation_id": conversationID,
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 回写 conversation_id 失败: task=%d, err=%v", taskID, err)
	}

	// 构建任务指令消息
	taskPrompt := e.buildTaskPrompt(task)

	// 保存指令消息
	userMsgID := int64(snowflake.Generate())
	if _, err := g.DB().Model("mvp_message").Insert(g.Map{
		"id":              userMsgID,
		"conversation_id": conversationID,
		"role":            "user",
		"message_type":    mvpmodel.MessageTypeTaskPrompt,
		"content":         taskPrompt,
		"status":          "completed",
		"created_by":      0,
		"dept_id":         0,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 保存用户消息失败: task=%d, err=%v", taskID, err)
	}

	// 创建 AI 回复消息
	replyID := int64(snowflake.Generate())
	if _, err := g.DB().Model("mvp_message").Insert(g.Map{
		"id":              replyID,
		"conversation_id": conversationID,
		"role":            "assistant",
		"message_type":    mvpmodel.MessageTypeTaskReply,
		"content":         "",
		"model_id":        modelInfo.ModelID,
		"status":          "streaming",
		"created_by":      0,
		"dept_id":         0,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 创建AI回复消息失败: task=%d, err=%v", taskID, err)
	}

	// 调用 AI
	p, err := provider.GetProvider(provider.Config{
		ProviderType:       modelInfo.ProviderType,
		SupportedProtocols: modelInfo.SupportedProtocols,
		BaseURL:            modelInfo.BaseURL,
		APIKey:             modelInfo.APIKey,
		APISecret:          modelInfo.APISecret,
	})
	if err != nil {
		e.handleTaskFailure(ctx, projectID, taskID, roleType, taskFailureExecution, err.Error())
		return
	}

	// 加载对话历史
	history, histErr := e.loadConversationHistory(ctx, conversationID, replyID)
	if histErr != nil {
		g.Log().Warningf(ctx, "[Executor] 加载对话历史失败: conv=%d err=%v", conversationID, histErr)
	}

	// 构建包含上下文摘要的 system prompt
	enrichedPrompt := BuildTaskSystemPrompt(ctx, projectID, taskID, roleType, task["role_level"].String(), modelInfo.SystemPrompt)

	req := &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     history,
		MaxTokens:    modelInfo.MaxTokens,
		Temperature:  0.7,
		Stream:       true,
		SystemPrompt: enrichedPrompt,
	}

	var fullContent strings.Builder
	chunkIndex := 0
	hub := GetHub()
	var lastFinishReason string

	streamHandler := func(chunk *provider.StreamChunk) error {
		if chunk.Content != "" {
			fullContent.WriteString(chunk.Content)
			chunkIndex++

			// 每 10 个 chunk 更新一次心跳，供看门狗检测
			if chunkIndex%10 == 0 {
				TouchHeartbeat(ctx, taskID)
			}

			if _, insertErr := g.DB().Model("mvp_message_chunk").Insert(g.Map{
				"message_id":  replyID,
				"chunk_index": chunkIndex,
				"content":     chunk.Content,
				"created_at":  gtime.Now(),
			}); insertErr != nil {
				g.Log().Errorf(ctx, "[Executor] 写入 chunk 失败: msg=%d, err=%v", replyID, insertErr)
			}
			activity.TouchMessageActivity(ctx, replyID)
			activity.TouchConversationActivity(ctx, conversationID)
			activity.TouchTaskActivity(ctx, taskID)

			chunkJSON, cErr := json.Marshal(map[string]interface{}{
				"content": chunk.Content,
				"index":   chunkIndex,
			})
			if cErr != nil {
				chunkJSON = []byte(`{"content":"","index":0}`)
			}
			hub.Publish(replyID, string(chunkJSON))
		}

		if chunk.FinishReason != "" {
			lastFinishReason = chunk.FinishReason
			if chunk.Usage != nil {
				usageJSON, uErr := json.Marshal(chunk.Usage)
				if uErr != nil {
					usageJSON = []byte("{}")
				}
				if _, err := g.DB().Model("mvp_message").Ctx(ctx).Where("id", replyID).Update(g.Map{
					"token_usage": string(usageJSON),
				}); err != nil {
					g.Log().Errorf(ctx, "[Executor] 更新 token_usage 失败: msg=%d, err=%v", replyID, err)
				}
			}
		}

		return nil
	}

	// 带重试 + 自动续写的 AI 调用
	const maxRetries = 2
	const maxContinueRounds = 5
	var callErr error

	for round := 0; round <= maxContinueRounds; round++ {
		lastFinishReason = ""

		// 瞬时错误重试
		callErr = nil
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if attempt > 0 {
				g.Log().Warningf(ctx, "[Executor] AI 调用第 %d 次重试 (task=%d): %v", attempt, taskID, callErr)
				time.Sleep(time.Duration(attempt*2) * time.Second)
			}
			callErr = p.ChatStream(ctx, req, streamHandler)
			if callErr == nil {
				break
			}
			// context canceled/deadline exceeded 不重试，直接退出
			if ctx.Err() != nil {
				g.Log().Debugf(ctx, "[Executor] AI 调用因 context 取消而中断 (task=%d): %v", taskID, callErr)
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
			break
		}

		// 检查是否被截断
		isTruncated := lastFinishReason == "length" || lastFinishReason == "max_tokens"
		if !isTruncated || round == maxContinueRounds {
			break
		}

		// 被截断：自动续写
		g.Log().Infof(ctx, "[Executor] 回复被截断(reason=%s)，自动续写第 %d 轮 (task=%d)",
			lastFinishReason, round+1, taskID)
		TouchHeartbeat(ctx, taskID)

		req.Messages = append(req.Messages,
			provider.Message{Role: provider.RoleAssistant, Content: fullContent.String()},
			provider.Message{Role: provider.RoleUser, Content: "继续，从上次中断的地方接着输出，不要重复已输出的内容。"},
		)
	}

	if callErr != nil {
		// AI 调用失败
		if _, dbErr := g.DB().Model("mvp_message").Ctx(ctx).Where("id", replyID).Update(g.Map{
			"content":      "AI调用失败: " + callErr.Error(),
			"message_type": mvpmodel.MessageTypePoison,
			"status":       "failed",
			"updated_at":   gtime.Now(),
		}); dbErr != nil {
			g.Log().Errorf(ctx, "[Executor] 更新失败消息状态失败: msg=%d, err=%v", replyID, dbErr)
		}
		e.handleTaskFailure(ctx, projectID, taskID, roleType, taskFailureExecution, callErr.Error())
		return
	}

	// 空内容检测：AI 返回空内容视为失败
	result := fullContent.String()
	if strings.TrimSpace(result) == "" {
		if _, dbErr := g.DB().Model("mvp_message").Ctx(ctx).Where("id", replyID).Update(g.Map{
			"content":      "AI返回空内容",
			"message_type": mvpmodel.MessageTypePoison,
			"status":       "failed",
			"updated_at":   gtime.Now(),
		}); dbErr != nil {
			g.Log().Errorf(ctx, "[Executor] 更新空内容消息失败: msg=%d, err=%v", replyID, dbErr)
		}
		hub.Publish(replyID, `{"done":true}`)
		hub.Done(replyID)
		e.handleTaskFailure(ctx, projectID, taskID, roleType, taskFailureExecution, "AI返回空内容，可能被模型内容过滤或请求异常")
		return
	}

	// 完成消息
	if _, err := g.DB().Model("mvp_message").Ctx(ctx).Where("id", replyID).Update(g.Map{
		"content":    result,
		"status":     "completed",
		"updated_at": gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 更新完成消息失败: msg=%d, err=%v", replyID, err)
	}

	hub.Publish(replyID, `{"done":true}`)
	hub.Done(replyID)
	if _, err := updateTaskStatus(ctx, taskID, "running", "completed", g.Map{
		"result":       result,
		"completed_at": gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 保存任务结果失败: task=%d, err=%v", taskID, err)
	}

	// 压缩任务上下文为摘要（同步执行，确保不丢失）
	compressCtx, compressCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer compressCancel()
	if err := GetCompressor().CompressTaskContext(compressCtx, projectID, taskID); err != nil {
		g.Log().Errorf(ctx, "[Executor] 压缩任务上下文失败（非致命）: task=%d, err=%v", taskID, err)
	}

	// 如果是实施员任务，完成后需要创建对应的审计任务（如果还没有）
	// 同步执行，确保审计任务一定被创建
	if roleType == "implementer" {
		e.createAuditTask(ctx, projectID, taskID, task)
	} else if roleType == "architect" {
		// 双读：优先 task_kind，兼容旧数据名称前缀
		// 使用项目级 ctx 传播取消信号
		projCtx := e.scheduler.getProjectContext(projectID)
		switch task["task_kind"].String() {
		case consts.TaskKindBugAnalysis:
			e.scheduler.AutoDispatchBugFix(projCtx, projectID, taskID)
		case consts.TaskKindFailureAnalysis:
			e.scheduler.AutoDispatchFailureFix(projCtx, projectID, taskID)
		default:
			name := task["name"].String()
			switch {
			case strings.HasPrefix(name, "Bug分析:"):
				e.scheduler.AutoDispatchBugFix(projCtx, projectID, taskID)
			case strings.HasPrefix(name, "失败分析:"):
				e.scheduler.AutoDispatchFailureFix(projCtx, projectID, taskID)
			}
		}
	}

	// 通知调度器
	e.scheduler.OnTaskCompleted(projectID, taskID)
}

// resolveTaskModel 解析任务使用的 AI 模型
func (e *Executor) resolveTaskModel(ctx context.Context, projectID int64, taskID int64, roleType string, modelID int64) (*ModelInfo, error) {
	task, err := g.DB().Model("mvp_task").Ctx(ctx).Where("id", taskID).WhereNull("deleted_at").Fields("role_level").One()
	if err != nil {
		g.Log().Warningf(ctx, "[Executor] 查询 task role_level 失败: task=%d, err=%v", taskID, err)
	}
	roleLevel := ""
	if !task.IsEmpty() {
		roleLevel = task["role_level"].String()
	}

	return ResolveProjectModelInfo(ctx, projectID, roleType, roleLevel, modelID)
}

// getModelInfo 查询模型详情
func (e *Executor) getModelInfo(ctx context.Context, modelID int64, systemPrompt string) (*ModelInfo, error) {
	model, err := g.DB().Model("ai_model m").
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.model_code, m.max_tokens, pv.provider_type, pv.supported_protocols, pv.base_url, p.api_key, p.api_secret, m.role_prompt").
		Where("m.id", modelID).
		Where("m.deleted_at IS NULL").
		One()
	if err != nil || model.IsEmpty() {
		return nil, fmt.Errorf("AI模型 %d 不存在", modelID)
	}

	prompt := systemPrompt
	if prompt == "" {
		prompt = model["role_prompt"].String()
	}

	return &ModelInfo{
		ModelID:            modelID,
		ModelCode:          model["model_code"].String(),
		ProviderType:       model["provider_type"].String(),
		SupportedProtocols: decodeProviderProtocols(model["supported_protocols"].String(), model["provider_type"].String()),
		BaseURL:            model["base_url"].String(),
		APIKey:             model["api_key"].String(),
		APISecret:          model["api_secret"].String(),
		SystemPrompt:       prompt,
		MaxTokens:          model["max_tokens"].Int(),
	}, nil
}

// getExecutionMode 从项目角色配置中获取执行方式
func (e *Executor) getExecutionMode(ctx context.Context, projectID int64, roleType string, roleLevel string) string {
	return ResolveProjectExecutionMode(ctx, projectID, roleType, roleLevel)
}

// handleTaskFailure 处理任务失败，根据错误分类决定升级还是直接失败
func (e *Executor) handleTaskFailure(ctx context.Context, projectID int64, taskID int64, roleType string, category taskFailureCategory, errMsg string) {
	switch category {
	case taskFailurePlanning, taskFailurePolicyGuard:
		updateTaskStatus(ctx, taskID, "running", "escalated", g.Map{
			"error_message": errMsg,
		})
		e.scheduler.OnTaskEscalated(projectID, taskID, errMsg)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					g.Log().Errorf(context.Background(), "[Executor] EscalateFailedTask panic: project=%d task=%d err=%v", projectID, taskID, r)
				}
			}()
			e.scheduler.EscalateFailedTask(e.scheduler.getProjectContext(projectID), projectID, taskID, roleType, errMsg)
		}()
	default:
		e.failTask(ctx, projectID, taskID, errMsg)
	}
}

// failTask 标记任务失败
func (e *Executor) failTask(ctx context.Context, projectID int64, taskID int64, errMsg string) {
	updateTaskStatus(ctx, taskID, "running", "failed", g.Map{
		"error_message": errMsg,
	})
	e.scheduler.OnTaskFailed(projectID, taskID, errMsg)
}
