package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"easymvp/app/mvp/internal/controller/config"
	"easymvp/app/mvp/internal/controller/conversation"
	"easymvp/app/mvp/internal/controller/message"
	"easymvp/app/mvp/internal/controller/project"
	"easymvp/app/mvp/internal/controller/project_category"
	"easymvp/app/mvp/internal/controller/project_role"
	"easymvp/app/mvp/internal/controller/role_preset"
	"easymvp/app/mvp/internal/controller/task"
	"easymvp/app/mvp/internal/controller/task_log"


	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/collab"
	"easymvp/app/mvp/internal/collab/notifier"
	"easymvp/app/mvp/internal/controller/chat"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/worker"
)

// autoStartFeishuWS 服务启动时自动从 DB 读取飞书配置，若启用了 websocket 模式则拉起长连接。
func autoStartFeishuWS(ctx context.Context) {
	enabled := engine.GetConfigInt(ctx, "workflow.collab.feishu_enabled", "workflow.collab.feishuEnabled", 0)
	if enabled != 1 {
		return
	}
	mode := engine.GetConfigString(ctx, "workflow.collab.feishu_connection_mode", "", "webhook")
	if mode != "websocket" {
		return
	}
	appID := engine.GetConfigString(ctx, "workflow.collab.feishu_app_id", "workflow.collab.feishuAppId", "")
	appSecret := engine.GetConfigString(ctx, "workflow.collab.feishu_app_secret", "workflow.collab.feishuAppSecret", "")
	encryptKey := engine.GetConfigString(ctx, "workflow.collab.feishu_encrypt_key", "workflow.collab.feishuEncryptKey", "")
	if appID == "" || appSecret == "" {
		return
	}
	g.Log().Infof(ctx, "[Startup] 飞书 WS 模式已启用，自动建立长连接 appID=%s", appID)
	collab.GetWSManager().StartWS(appID, appSecret, encryptKey, chat.FeishuWSEventHandlerExport)
}

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start mvp http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 注册飞书通知钩子
			n := notifier.GetNotifier()
			engine.RegisterFeishuNotifyAIReply(n.NotifyAIReply)
			engine.RegisterFeishuNotifyTaskFailed(n.NotifyTaskFailed)
			engine.RegisterFeishuNotifyProjectCompleted(n.NotifyProjectCompleted)

			// 启动异步删除 worker（Redis 队列消费）
			worker.StartDeleteWorker(ctx)

			// 自动恢复飞书 WS 长连接（读取 DB 配置）
			autoStartFeishuWS(ctx)

			// 自动启动 Telegram Bot Polling（读取 DB 配置）
			chat.StartTelegramPolling(ctx)

			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Group("/api/mvp", func(group *ghttp.RouterGroup) {
					group.Middleware(middleware.Auth)
					group.Bind(
						config.Config,
						conversation.Conversation,
						message.Message,
						project.Project,
						projectcategory.ProjectCategory,
						projectrole.ProjectRole,
						rolepreset.RolePreset,
						task.Task,
						tasklog.TaskLog,
					)
					// 注册手写的自定义路由（在 cmd_custom.go 中定义）
					registerCustomRoutes(group)
				})
			})
			s.Run()
			return nil
		},
	}
)
