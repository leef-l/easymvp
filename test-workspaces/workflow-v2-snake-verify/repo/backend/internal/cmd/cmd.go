package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	_ "workflowv2snake/backend/internal/logic/game"

	"workflowv2snake/backend/internal/controller/game"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Use(func(r *ghttp.Request) {
				r.Response.CORSDefault()
				if r.Method == "OPTIONS" {
					r.ExitAll()
					return
				}
				r.Middleware.Next()
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					game.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)
