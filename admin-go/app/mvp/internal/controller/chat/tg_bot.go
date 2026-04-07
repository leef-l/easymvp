package chat

// tg_bot.go
// Telegram Bot 平台适配层。
// 实现 BotPlatform 接口，解析 Telegram Update 后转发到统一 Bot 调度器。

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gogf/gf/v2/frame/g"

	collabRepo "easymvp/app/mvp/internal/collab/repo"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/orchestrator"
)

// tgBotPlatform 实现 BotPlatform 接口，封装 Telegram 消息回复。
type tgBotPlatform struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func (t *tgBotPlatform) Reply(ctx context.Context, text string) {
	g.Log().Debugf(ctx, "[TGBot] 发送消息: chatID=%d text_len=%d", t.chatID, len(text))
	msg := tgbotapi.NewMessage(t.chatID, text)
	if _, err := t.bot.Send(msg); err != nil {
		g.Log().Warningf(ctx, "[TGBot] 发送消息失败: chatID=%d err=%v", t.chatID, err)
	}
}

func (t *tgBotPlatform) PlatformName() string { return "telegram" }

// DispatchTelegramUpdate Telegram Update 入口，解析消息后转发到统一 Bot 调度器。
func DispatchTelegramUpdate(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	msg := update.Message
	chatID := msg.Chat.ID
	chatIDStr := fmt.Sprintf("%d", chatID)

	// 群聊只响应 @bot 的消息
	if !msg.Chat.IsPrivate() {
		botUsername := "@" + bot.Self.UserName
		if !strings.Contains(msg.Text, botUsername) {
			return
		}
	}

	text := normalizeTGCommand(msg.Text, bot.Self.UserName)
	g.Log().Debugf(ctx, "[TGBot] 收到消息: chatID=%d text_len=%d", chatID, len(text))

	DispatchBotCommand(ctx, &BotContext{
		OpenID:  chatIDStr,
		Content: text,
		Platform: &tgBotPlatform{
			bot:    bot,
			chatID: chatID,
		},
	})
}

// normalizeTGCommand 将 /command 风格转为自然语言，并去掉 @mention。
func normalizeTGCommand(text, botUsername string) string {
	// 去掉 @mention
	text = strings.ReplaceAll(text, "@"+botUsername, "")
	text = strings.TrimSpace(text)

	lower := strings.ToLower(text)
	switch {
	case lower == "/start" || lower == "/help":
		return "帮助"
	case lower == "/list" || lower == "/projects":
		return "我的项目"
	case lower == "/quit" || lower == "/exit":
		return "退出对话"
	}
	// 去掉普通 /command 前缀，保留后面的内容
	if strings.HasPrefix(text, "/") {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}
	return text
}

// lookupSystemUserByTG 根据 TG chatID 查找绑定的系统用户。
func lookupSystemUserByTG(ctx context.Context, chatIDStr string) (userID int64, deptID int64) {
	if chatIDStr == "" {
		return 0, 0
	}
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	binding, err := bindingRepo.GetByPlatformUserID(ctx, "telegram", chatIDStr)
	if err != nil || binding == nil {
		return 0, 0
	}
	userID = toCallbackInt64(binding["user_id"])
	deptID = toCallbackInt64(binding["dept_id"])
	return
}

// getTelegramBotToken 从引擎配置读取 Telegram Bot Token。
func getTelegramBotToken(ctx context.Context) string {
	return engine.GetConfigString(ctx, "workflow.collab.telegram_bot_token", "", "")
}
