package chat

// tg_polling.go
// Telegram Bot Long Polling 循环。
// 服务启动时调用 StartTelegramPolling，在独立 goroutine 中阻塞轮询 Update。

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gogf/gf/v2/frame/g"
)

// StartTelegramPolling 启动 Telegram Bot Polling（非阻塞）。
// 若 token 未配置则静默跳过。
func StartTelegramPolling(ctx context.Context) {
	token := getTelegramBotToken(ctx)
	if token == "" {
		g.Log().Infof(ctx, "[TGBot] telegram_bot_token 未配置，跳过启动")
		return
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		g.Log().Errorf(ctx, "[TGBot] 初始化失败: %v", err)
		return
	}

	g.Log().Infof(ctx, "[TGBot] Polling 启动: bot=@%s", bot.Self.UserName)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(ctx, "[TGBot] Polling goroutine panic: %v", r)
			}
		}()
		for {
			if err := runPolling(ctx, bot); err != nil {
				g.Log().Warningf(ctx, "[TGBot] Polling 异常: %v，5s 后重试", err)
			}
			select {
			case <-ctx.Done():
				g.Log().Infof(ctx, "[TGBot] 收到退出信号，停止 Polling")
				return
			case <-time.After(5 * time.Second):
			}
		}
	}()
}

func runPolling(ctx context.Context, bot *tgbotapi.BotAPI) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			bot.StopReceivingUpdates()
			return nil
		case update, ok := <-updates:
			if !ok {
				return fmt.Errorf("update channel 关闭")
			}
			go DispatchTelegramUpdate(ctx, bot, update)
		}
	}
}
