package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
{{range .Modules}}
	"easymvp/app/{{$.AppName}}/internal/controller/{{.}}"{{end}}


	"easymvp/app/{{.AppName}}/internal/middleware"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start {{.AppName}} http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Group("/api/{{.AppName}}", func(group *ghttp.RouterGroup) {
					group.Middleware(middleware.Auth)
					group.Bind({{range .Modules}}
						{{PackageName .}}.{{ModuleCamel .}},{{end}}
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
