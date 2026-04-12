package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/utility/provider"
)

// runAICall 异步调用 AI（goroutine 中运行，不依赖前端连接）
func (e *ChatEngine) runAICall(conversationID int64, replyID int64, modelInfo *ModelInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[ChatEngine] runAICall panic: conversationID=%d replyID=%d err=%v", conversationID, replyID, r)
			e.failMessage(ctx, replyID, fmt.Sprintf("AI 调用内部错误: %v", r))
		}
	}()

	// 1. 获取对话历史
	messages, err := e.loadHistory(ctx, conversationID, replyID)
	if err != nil {
		e.failMessage(ctx, replyID, err.Error())
		return
	}

	// 2. 创建 Provider
	p, err := provider.GetProvider(provider.Config{
		ProviderType:       modelInfo.ProviderType,
		SupportedProtocols: modelInfo.SupportedProtocols,
		BaseURL:            modelInfo.BaseURL,
		APIKey:             modelInfo.APIKey,
		APISecret:          modelInfo.APISecret,
	})
	if err != nil {
		e.failMessage(ctx, replyID, err.Error())
		return
	}

	// 3. 统一流式调用（重试 + 续写 + chunk 落盘由 StreamCall 处理）
	result, callErr := StreamCall(ctx, &StreamCallConfig{
		Provider: p,
		Request: &provider.ChatRequest{
			Model:        modelInfo.ModelCode,
			Messages:     messages,
			MaxTokens:    modelInfo.MaxTokens,
			Temperature:  0.7,
			Stream:       true,
			SystemPrompt: modelInfo.SystemPrompt,
		},
		ReplyID:        replyID,
		Hub:            e.hub,
		ConversationID: conversationID,
	})

	if callErr != nil {
		e.failMessage(ctx, replyID, callErr.Error())
		return
	}

	// 4. 更新消息为完成状态
	if _, updateErr := g.DB().Model("mvp_message").Ctx(ctx).Where("id", replyID).Update(g.Map{
		"content":    result.Content,
		"status":     "completed",
		"updated_at": gtime.Now(),
	}); updateErr != nil {
		g.Log().Errorf(ctx, "更新消息状态失败: %v", updateErr)
	}

	if architectReplyRequestsContinuation(result.Content) {
		if e.shouldAutoContinueArchitectReply(ctx, conversationID) {
			e.autoContinueArchitectReply(ctx, conversationID)
		}
	} else {
		e.tryParseArchitectTasks(conversationID, replyID, result.Content)
	}

	// 5. 通知 SSE Hub 流式输出完成
	e.hub.Publish(replyID, `{"done":true}`)

	// 6. 飞书主动推送
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[ChatEngine] feishuNotifyAIReply panic: %v", r)
			}
		}()
		feishuNotifyAIReply(ctx, conversationID, result.Content)
	}()

	// 短暂延迟后关闭 channel，让前端有时间接收最后的消息
	time.Sleep(100 * time.Millisecond)
	e.hub.Done(replyID)
}

// feishuNotifyAIReply 推送 AI 回复到飞书（避免循环引用，用函数变量注入）。
var feishuNotifyAIReply = func(ctx context.Context, conversationID int64, content string) {}

// tryParseArchitectTasks 尝试从架构师回复中解析任务清单
func (e *ChatEngine) tryParseArchitectTasks(conversationID, messageID int64, aiReply string) {
	ctx := context.Background()

	// 查对话的角色类型和项目ID
	conv, err := g.DB().Model("mvp_conversation").Ctx(ctx).Where("id", conversationID).WhereNull("deleted_at").One()
	if err != nil || conv.IsEmpty() {
		return
	}

	// 只有架构师对话且是项目级对话（task_id 为空）才解析
	if conv["role_type"].String() != "architect" || conv["task_id"].Int64() != 0 {
		return
	}
	userContents, userErr := loadRecentArchitectUserMessages(ctx, conversationID, 12)
	replyPolicy := architectReplyPolicy{}
	if userErr != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 查询架构师最近用户消息失败: conversationID=%d err=%v", conversationID, userErr)
	} else if !shouldParseArchitectReply(ctx, userContents) {
		g.Log().Infof(ctx, "[ChatEngine] 跳过解析架构师回复: conversationID=%d messageID=%d 原因=最近触发消息属于系统状态通知", conversationID, messageID)
		return
	} else {
		replyPolicy = resolveArchitectReplyPolicy(ctx, userContents)
	}

	projectID := conv["project_id"].Int64()
	replyForParse := aiReply
	if combinedReply, windowErr := collectArchitectReplyWindow(ctx, conversationID); windowErr != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 汇总架构师回复窗口失败: conversationID=%d err=%v", conversationID, windowErr)
	} else if strings.TrimSpace(combinedReply) != "" {
		replyForParse = combinedReply
	}

	// 判断引擎版本
	ev, evErr := g.DB().Model("mvp_project").Ctx(ctx).Where("id", projectID).WhereNull("deleted_at").Value("engine_version")
	if evErr != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 查询 engine_version 失败: projectID=%d err=%v", projectID, evErr)
	}
	if ev.String() == "workflow_v2" {
		e.tryParseArchitectBlueprints(ctx, projectID, conv["id"].Int64(), messageID, replyForParse, replyPolicy)
		return
	}

	// Legacy：写入 mvp_task
	count, err := GetParser().ParseAndCreateTasks(ctx, projectID, replyForParse)
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
func (e *ChatEngine) tryParseArchitectBlueprints(ctx context.Context, projectID, conversationID, messageID int64, aiReply string, replyPolicy architectReplyPolicy) {
	if blueprintCreatorFn == nil {
		g.Log().Warningf(ctx, "[ChatEngine] V2 蓝图创建回调未注册，跳过")
		return
	}

	// 获取项目分类
	projectCategory, pcErr := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).Value("project_category")
	if pcErr != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 查询项目分类失败: projectID=%d err=%v", projectID, pcErr)
	}

	// 查活跃的 workflow_run
	wfRun, wfErr := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNotIn("status", g.Slice{consts.WorkflowRunStatusCompleted, consts.WorkflowRunStatusCanceled}).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	if wfErr != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 查询活跃 workflow_run 失败: projectID=%d err=%v", projectID, wfErr)
	}
	var wfRunID int64
	var currentStage string
	if !wfRun.IsEmpty() {
		wfRunID = wfRun["id"].Int64()
		currentStage = wfRun["current_stage"].String()
	}
	if !shouldApplyArchitectBlueprintMutation(currentStage, replyPolicy) {
		g.Log().Infof(ctx, "[ChatEngine] 跳过晚到的架构师蓝图回写: project=%d workflowRun=%d currentStage=%s messageID=%d",
			projectID, wfRunID, currentStage, messageID)
		return
	}

	fastTasks, fastReport, fastErr := GetParser().FastExtractWithReport(ctx, aiReply, projectCategory.String())
	if fastErr != nil {
		g.Log().Warningf(ctx, "[ChatEngine] V2 FastExtractWithReport 失败: %v", fastErr)
	}
	if fastReport != nil && fastReport.HasBlockingIssue() {
		g.Log().Warningf(ctx, "[ChatEngine] V2 架构师任务清单存在阻断问题: project=%d summary=%s", projectID, fastReport.Summary())
		NotifyProjectArchitectConversation(ctx, projectID, fastReport.BuildContinuationPrompt())
		return
	}
	if len(fastTasks) > 0 {
		pvID, bpCount, createErr := blueprintCreatorFn(ctx, projectID, wfRunID, conversationID, messageID, fastTasks)
		if createErr != nil {
			g.Log().Warningf(ctx, "[ChatEngine] V2 创建蓝图失败: %v", createErr)
			return
		}
		g.Log().Infof(ctx, "[ChatEngine] V2 架构师回复解析出 %d 个蓝图, planVersion=%d, 项目 %d, report=%s",
			bpCount, pvID, projectID, fastReport.Summary())
		e.autoResubmitArchitectReviewIfNeeded(ctx, projectID, conversationID, pvID)
		return
	}

	patches, patchErr := extractArchitectTaskPatches(aiReply)
	if patchErr == nil && len(patches) > 0 {
		if blueprintPatchApplierFn == nil {
			g.Log().Warningf(ctx, "[ChatEngine] V2 蓝图 patch 回调未注册，跳过")
			return
		}
		pvID, patchedCount, applyErr := blueprintPatchApplierFn(ctx, projectID, wfRunID, conversationID, messageID, patches)
		if applyErr != nil {
			g.Log().Warningf(ctx, "[ChatEngine] V2 应用蓝图修订失败: %v", applyErr)
			NotifyProjectArchitectConversation(ctx, projectID,
				fmt.Sprintf("## 方案修订未生效\n\n系统已识别到局部修订 JSON，但未能回写到当前方案：%v\n\n请确认 `task_name` 与现有蓝图名称一致，或直接输出完整 `{\"tasks\": [...]}` 新方案。", applyErr))
			return
		}
		g.Log().Infof(ctx, "[ChatEngine] V2 架构师局部修订已应用 %d 个蓝图, planVersion=%d, 项目 %d", patchedCount, pvID, projectID)
		e.autoResubmitArchitectReviewIfNeeded(ctx, projectID, conversationID, pvID)
		return
	}

	tasks, report, err := GetParser().ExtractAndNormalizeWithReport(ctx, aiReply, projectCategory.String())
	if report != nil && report.HasBlockingIssue() {
		g.Log().Warningf(ctx, "[ChatEngine] V2 架构师任务清单存在阻断问题: project=%d summary=%s", projectID, report.Summary())
		NotifyProjectArchitectConversation(ctx, projectID, report.BuildContinuationPrompt())
		return
	}
	if err != nil || len(tasks) == 0 {
		return
	}

	pvID, bpCount, err := blueprintCreatorFn(ctx, projectID, wfRunID, conversationID, messageID, tasks)
	if err != nil {
		g.Log().Warningf(ctx, "[ChatEngine] V2 创建蓝图失败: %v", err)
		return
	}
	g.Log().Infof(ctx, "[ChatEngine] V2 架构师回复解析出 %d 个蓝图, planVersion=%d, 项目 %d, report=%s",
		bpCount, pvID, projectID, report.Summary())
	e.autoResubmitArchitectReviewIfNeeded(ctx, projectID, conversationID, pvID)
}

func (e *ChatEngine) autoResubmitArchitectReviewIfNeeded(ctx context.Context, projectID, conversationID, planVersionID int64) {
	if architectReviewResubmitterFn == nil {
		return
	}
	if !e.shouldAutoResubmitArchitectReview(ctx, conversationID) {
		return
	}
	if err := architectReviewResubmitterFn(ctx, projectID); err != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 架构师修订后自动重提审核失败: project=%d planVersion=%d err=%v", projectID, planVersionID, err)
		return
	}
	g.Log().Infof(ctx, "[ChatEngine] 架构师修订后已自动重提审核: project=%d planVersion=%d", projectID, planVersionID)
}
