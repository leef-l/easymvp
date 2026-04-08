package chat

// bot_dispatcher.go
// 平台无关的 Bot 调度层。
// 各平台（飞书、钉钉、企业微信等）解析完消息后，统一调用 DispatchBotCommand，
// 通过 BotPlatform 接口回复消息，核心业务逻辑完全复用。

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/engine"
)

// BotPlatform 平台回调接口，屏蔽飞书/钉钉/企业微信的差异。
type BotPlatform interface {
	// Reply 回复消息（回复到原消息或发送到对话）
	Reply(ctx context.Context, text string)
	// PlatformName 平台名称（用于日志）
	PlatformName() string
}

// BotContext 一次 Bot 消息的上下文，平台无关。
type BotContext struct {
	OpenID   string // 发消息的用户在当前平台的唯一ID（飞书=open_id，钉钉=staffId）
	Content  string // 消息文本内容（已去除 @mention）
	Platform BotPlatform
}

// DispatchBotCommand 平台无关的统一 Bot 入口。
// 飞书/钉钉/企业微信 callback 各自解析消息后，构造 BotContext 调用此函数。
func DispatchBotCommand(ctx context.Context, bc *BotContext) {
	reply := func(text string) {
		bc.Platform.Reply(ctx, text)
	}

	startTime := time.Now()
	text := strings.TrimSpace(bc.Content)
	g.Log().Debugf(ctx, "[Bot/%s] 收到文本: openID=%s text_len=%d", bc.Platform.PlatformName(), bc.OpenID, len(text))
	defer func() {
		g.Log().Debugf(ctx, "[Bot/%s] 处理完成: openID=%s text_len=%d 耗时=%dms", bc.Platform.PlatformName(), bc.OpenID, len(text), time.Since(startTime).Milliseconds())
	}()
	if text == "" {
		reply(botHelpText())
		return
	}

	// 快速关键词匹配（无需调 AI，毫秒级响应）
	lower := strings.ToLower(strings.TrimSpace(text))
	switch lower {
	case "退出对话", "exit", "quit":
		clearBotSession(bc.Platform.PlatformName(), bc.OpenID)
		reply("✅ 已退出对话模式")
		return
	case "帮助", "help", "?", "？", "怎么用", "使用说明", "功能":
		reply(botHelpText())
		return
	case "我的项目", "项目列表", "所有项目", "list":
		systemUserID, deptID := lookupBotUser(ctx, bc.Platform.PlatformName(), bc.OpenID)
		handleBotListProjects(ctx, systemUserID, reply)
		_ = deptID
		return
	}

	// 反查绑定的系统用户（按平台区分）
	systemUserID, deptID := lookupBotUser(ctx, bc.Platform.PlatformName(), bc.OpenID)

	// 快速判断：不含项目操作关键词的消息直接走 fallback，不调 AI
	intent := fallbackParseIntent(text)
	if looksLikeCommand(lower) {
		var err error
		intent, err = parseIntentWithAI(ctx, text, systemUserID)
		if err != nil {
			intent = fallbackParseIntent(text)
		}
	}

	g.Log().Infof(ctx, "[Bot/%s] 意图: action=%s project=%s", bc.Platform.PlatformName(), intent.Action, intent.ProjectName)

	// 路由
	dispatchIntent(ctx, intent, text, bc.OpenID, systemUserID, deptID, reply, bc.Platform.PlatformName())
}

// dispatchIntent 意图路由（被各平台共同调用）。
func dispatchIntent(
	ctx context.Context,
	intent *botIntent,
	rawText string,
	openID string,
	systemUserID, deptID int64,
	reply func(string),
	platform ...string,
) {
	platformName := "feishu"
	if len(platform) > 0 && platform[0] != "" {
		platformName = platform[0]
	}
	switch intent.Action {
	// ── 项目管理 ──
	case "create_project":
		handleBotCreateProject(ctx, intent.ProjectName, intent.Category, systemUserID, deptID, openID, platformName, reply)
	case "list_projects":
		handleBotListProjects(ctx, systemUserID, reply)
	case "project_status":
		handleBotProjectStatus(ctx, intent.ProjectName, systemUserID, reply)
	case "pause_project":
		handleBotPauseProject(ctx, intent.ProjectName, systemUserID, reply)
	case "resume_project":
		handleBotResumeProject(ctx, intent.ProjectName, systemUserID, reply)
	case "cancel_project":
		handleBotCancelProject(ctx, intent.ProjectName, systemUserID, reply)
	case "force_stage":
		handleBotForceStage(ctx, intent.ProjectName, intent.TargetStage, intent.TaskID, systemUserID, reply)
	case "confirm_plan":
		handleBotConfirmPlan(ctx, openID, platformName, systemUserID, reply)

	// ── 任务管理 ──
	case "list_tasks":
		handleBotListTasks(ctx, intent.ProjectName, systemUserID, reply)
	case "retry_task":
		handleBotRetryTask(ctx, intent.ProjectName, intent.TaskID, systemUserID, reply)
	case "skip_task":
		handleBotSkipTask(ctx, intent.ProjectName, intent.TaskID, systemUserID, reply)
	case "update_task":
		handleBotUpdateTask(ctx, intent, systemUserID, reply)

	// ── 审核管理 ──
	case "review_status":
		handleBotReviewStatus(ctx, intent.ProjectName, systemUserID, reply)
	case "approve_review":
		handleBotApproveReview(ctx, intent.ProjectName, systemUserID, reply)
	case "reject_review":
		handleBotRejectReview(ctx, intent.ProjectName, systemUserID, reply)

	// ── 验收管理 ──
	case "accept_status":
		handleBotAcceptStatus(ctx, intent.ProjectName, systemUserID, reply)
	case "approve_accept":
		handleBotApproveAccept(ctx, intent.ProjectName, systemUserID, reply)
	case "reject_accept":
		handleBotRejectAccept(ctx, intent.ProjectName, systemUserID, reply)

	// ── 自治管理 ──
	case "autonomy_status":
		handleBotAutonomyStatus(ctx, intent.ProjectName, systemUserID, reply)
	case "approve_checkpoint":
		handleBotApproveCheckpoint(ctx, intent.ProjectName, systemUserID, reply)
	case "reject_checkpoint":
		handleBotRejectCheckpoint(ctx, intent.ProjectName, systemUserID, reply)

	// ── 通用 ──
	case "help":
		reply(botHelpText())
	case "chat":
		// 先检查是否有活跃对话会话
		if convID, ok := getBotSession(platformName, openID); ok {
			conv, _ := g.DB().Ctx(ctx).Model("mvp_conversation").
				Where("id", convID).WhereNull("deleted_at").One()
			if !conv.IsEmpty() {
				projectID := conv["project_id"].Int64()
				chatEng := engine.NewChatEngine()
				replyMsgID, err := chatEng.SendFeishuMessage(ctx, convID, projectID, rawText, systemUserID, deptID)
				if err == nil {
					aiReply := waitForAIReply(ctx, replyMsgID, 30*time.Second)
					if aiReply != "" {
						reply(aiReply)
						return
					}
				}
				reply("⏳ AI 正在思考中，请稍后发送「项目状态」查看进展")
				return
			}
		}
		if intent.Reply != "" {
			reply(intent.Reply)
		} else {
			reply(botHelpText())
		}
	default:
		reply(botHelpText())
	}
}

// looksLikeCommand 判断消息是否包含项目操作关键词，是则值得调 AI 解析意图。
// 不含关键词的闲聊/问候直接走 fallback，避免不必要的 AI 调用。
func looksLikeCommand(lower string) bool {
	keywords := []string{
		"项目", "任务", "创建", "新建", "暂停", "恢复", "继续", "取消", "终止", "重试", "跳过", "修改", "调整", "更新",
		"强制", "回到设计", "重开审核", "重开执行", "验收", "返工",
		"状态", "进度", "完成", "失败", "审核", "验收", "通过", "驳回", "拒绝",
		"批准", "确认", "方案", "自治", "检查点",
		"create", "project", "task", "pause", "resume", "cancel", "retry", "skip", "force", "reset",
		"status", "review", "accept", "approve", "reject", "confirm",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// botHelpText 平台无关的帮助文本。
func botHelpText() string {
	return `🤖 EasyMVP 机器人
───────────────
直接用自然语言和我说就行，不需要记命令。

我能帮你做：
📁 创建项目 / 查看进度 / 暂停继续 / 直接取消
🧭 强制回设计 / 强制重审 / 强制重跑执行
📋 查看任务 / 重试失败 / 跳过阻塞 / 直接改任务
🔍 查看审核结果 / 人工审核通过或驳回
🎯 查看验收状态 / 验收通过或打回
🤖 查看自治检查点 / 批准或拒绝

示例：
  "帮我建个电商后台项目"
  "电商后台跑到哪里了"
  "把电商后台回到设计阶段"
  "强制把电商后台重开到执行阶段"
  "把电商后台的任务123改成 codex_cli 并重新开始"
  "有失败任务帮我重试一下"
  "审核通过了吗，没问题就批了"
  "自治那边有啥需要我确认的吗"
───────────────
群聊中需要 @EasyMVP`
}
