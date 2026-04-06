package chat

// workflow_telegram.go
// Telegram Bot 协作管理 API 控制器。
// 对标 workflow_feishu.go，提供配置、用户绑定、命令菜单管理。

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	collabRepo "easymvp/app/mvp/internal/collab/repo"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/utility/snowflake"
)

// defaultTelegramCommands 默认 Bot 命令菜单。
func defaultTelegramCommands() []v1.TelegramCommandItem {
	return []v1.TelegramCommandItem{
		{Command: "start", Description: "开始使用 / 帮助"},
		{Command: "help", Description: "查看所有功能"},
		{Command: "list", Description: "我的项目列表"},
		{Command: "quit", Description: "退出对话模式"},
	}
}

// tgPollingRunning 简单标记 Polling 是否已启动（进程级单例）。
var tgPollingRunning = false

// TelegramConfig 查询 Telegram Bot 配置。
func (c *cWorkflow) TelegramConfig(ctx context.Context, req *v1.WorkflowTelegramConfigReq) (res *v1.WorkflowTelegramConfigRes, err error) {
	token := engine.GetConfigString(ctx, "workflow.collab.telegram_bot_token", "", "")
	enabled := engine.GetConfigInt(ctx, "workflow.collab.telegram_enabled", "", 0)
	// Token 脱敏显示（只显示前8位）
	maskedToken := ""
	if len(token) > 8 {
		maskedToken = token[:8] + "****"
	} else if token != "" {
		maskedToken = "****"
	}
	return &v1.WorkflowTelegramConfigRes{
		Config: v1.TelegramConfigDTO{
			Enabled:    enabled,
			BotToken:   maskedToken,
			BotRunning: tgPollingRunning,
		},
	}, nil
}

// SaveTelegramConfig 保存 Telegram Bot 配置，并联动 Polling 启停。
func (c *cWorkflow) SaveTelegramConfig(ctx context.Context, req *v1.WorkflowSaveTelegramConfigReq) (res *v1.WorkflowSaveTelegramConfigRes, err error) {
	token := strings.TrimSpace(req.BotToken)
	// 若前端回传的是脱敏值（含****），则不更新 token
	if strings.Contains(token, "****") {
		token = ""
	}

	if err := saveMvpConfig(ctx, "workflow.collab.telegram_enabled", fmt.Sprintf("%d", req.Enabled), "int", "collab", "Telegram Bot 总开关(0关/1开)"); err != nil {
		return nil, err
	}
	if token != "" {
		if err := saveMvpConfig(ctx, "workflow.collab.telegram_bot_token", token, "string", "collab", "Telegram Bot Token"); err != nil {
			return nil, err
		}
	}

	// 联动 Polling
	realToken := engine.GetConfigString(ctx, "workflow.collab.telegram_bot_token", "", "")
	if req.Enabled == 1 && realToken != "" {
		if !tgPollingRunning {
			tgPollingRunning = true
			StartTelegramPolling(ctx)
		}
	}

	return &v1.WorkflowSaveTelegramConfigRes{}, nil
}

// TelegramBindings 查询 Telegram 绑定列表。
func (c *cWorkflow) TelegramBindings(ctx context.Context, req *v1.WorkflowTelegramBindingsReq) (res *v1.WorkflowTelegramBindingsRes, err error) {
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	bindings, err := bindingRepo.List(ctx, "telegram")
	if err != nil {
		return nil, err
	}
	items := make([]v1.FeishuBindingDTO, 0, len(bindings))
	for _, item := range bindings {
		items = append(items, mapToFeishuBindingDTO(item))
	}
	return &v1.WorkflowTelegramBindingsRes{Bindings: items}, nil
}

// BindTelegramUser 绑定 Telegram 用户（chat_id ↔ 系统用户）。
func (c *cWorkflow) BindTelegramUser(ctx context.Context, req *v1.WorkflowBindTelegramUserReq) (res *v1.WorkflowBindTelegramUserRes, err error) {
	currentUserID := middleware.GetUserID(ctx)
	if currentUserID != 1 && int64(req.UserID) != currentUserID {
		return nil, fmt.Errorf("普通用户只能绑定自己的 Telegram 账号")
	}
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	id, err := bindingRepo.Rebind(ctx, g.Map{
		"user_id":          int64(req.UserID),
		"platform":         "telegram",
		"platform_user_id": strings.TrimSpace(req.PlatformUserID),
		"platform_name":    strings.TrimSpace(req.PlatformName),
		"created_by":       currentUserID,
		"dept_id":          middleware.GetDeptID(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowBindTelegramUserRes{ID: snowflake.JsonInt64(id)}, nil
}

// UnbindTelegramUser 解绑 Telegram 用户。
func (c *cWorkflow) UnbindTelegramUser(ctx context.Context, req *v1.WorkflowUnbindTelegramUserReq) (res *v1.WorkflowUnbindTelegramUserRes, err error) {
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	binding, err := bindingRepo.GetByIDScoped(ctx, int64(req.BindingID))
	if err != nil || binding == nil {
		return nil, fmt.Errorf("绑定记录不存在或无权操作")
	}
	if err := bindingRepo.UnbindByID(ctx, int64(req.BindingID)); err != nil {
		return nil, err
	}
	return &v1.WorkflowUnbindTelegramUserRes{}, nil
}

// TestTelegramMessage 发送 Telegram 测试消息。
func (c *cWorkflow) TestTelegramMessage(ctx context.Context, req *v1.WorkflowTestTelegramMessageReq) (res *v1.WorkflowTestTelegramMessageRes, err error) {
	if engine.GetConfigInt(ctx, "workflow.collab.telegram_enabled", "", 0) != 1 {
		return nil, fmt.Errorf("Telegram Bot 未开启，请先保存并启用配置")
	}
	token := engine.GetConfigString(ctx, "workflow.collab.telegram_bot_token", "", "")
	if token == "" {
		return nil, fmt.Errorf("Telegram Bot Token 未配置")
	}
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	binding, err := bindingRepo.GetByIDScoped(ctx, int64(req.BindingID))
	if err != nil || binding == nil {
		return nil, fmt.Errorf("绑定记录不存在或无权操作")
	}

	chatIDStr := mapString(binding, "platform_user_id")
	var chatID int64
	if _, err := fmt.Sscanf(chatIDStr, "%d", &chatID); err != nil || chatID == 0 {
		return nil, fmt.Errorf("无效的 Telegram chat_id: %s", chatIDStr)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("初始化 Telegram Bot 失败: %v", err)
	}

	text := strings.TrimSpace(req.Content)
	if text == "" {
		text = "EasyMVP Telegram 联通测试成功 🎉\n后续项目通知会通过此账号发送。"
	}
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		return nil, fmt.Errorf("发送失败: %v", err)
	}
	return &v1.WorkflowTestTelegramMessageRes{}, nil
}

// SetTelegramCommands 设置 Telegram Bot 命令菜单。
func (c *cWorkflow) SetTelegramCommands(ctx context.Context, req *v1.WorkflowSetTelegramCommandsReq) (res *v1.WorkflowSetTelegramCommandsRes, err error) {
	token := engine.GetConfigString(ctx, "workflow.collab.telegram_bot_token", "", "")
	if token == "" {
		return nil, fmt.Errorf("Telegram Bot Token 未配置")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("初始化 Telegram Bot 失败: %v", err)
	}

	commands := req.Commands
	if req.UseDefault || len(commands) == 0 {
		commands = defaultTelegramCommands()
	}

	tgCmds := make([]tgbotapi.BotCommand, 0, len(commands))
	for _, cmd := range commands {
		tgCmds = append(tgCmds, tgbotapi.BotCommand{
			Command:     strings.TrimPrefix(cmd.Command, "/"),
			Description: cmd.Description,
		})
	}

	cfg := tgbotapi.NewSetMyCommands(tgCmds...)
	if _, err := bot.Request(cfg); err != nil {
		return nil, fmt.Errorf("设置命令菜单失败: %v", err)
	}

	return &v1.WorkflowSetTelegramCommandsRes{
		Message:  "命令菜单已更新",
		Commands: commands,
	}, nil
}
