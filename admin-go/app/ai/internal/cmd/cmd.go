package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"easymvp/app/ai/internal/controller/engine"
	"easymvp/app/ai/internal/controller/model"
	"easymvp/app/ai/internal/controller/plan"
	"easymvp/app/ai/internal/controller/provider"
	"easymvp/app/ai/internal/controller/task"

	"easymvp/app/ai/internal/middleware"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start ai http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Group("/api/ai", func(group *ghttp.RouterGroup) {
					group.Middleware(middleware.Auth)
					group.Bind(
						engine.Engine,
						model.Model,
						plan.Plan,
						provider.Provider,
						task.Task,
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
