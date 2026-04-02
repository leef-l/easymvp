package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"easymvp/app/mvp/internal/controller/conversation"
	"easymvp/app/mvp/internal/controller/message"
	"easymvp/app/mvp/internal/controller/project"
	"easymvp/app/mvp/internal/controller/project_role"
	"easymvp/app/mvp/internal/controller/role_preset"
	"easymvp/app/mvp/internal/controller/task"
	"easymvp/app/mvp/internal/controller/task_log"


	"easymvp/app/mvp/internal/middleware"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start mvp http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Group("/api/mvp", func(group *ghttp.RouterGroup) {
					group.Middleware(middleware.Auth)
					group.Bind(
						conversation.Conversation,
						message.Message,
						project.Project,
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
