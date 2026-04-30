package cmd

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"github.com/leef-l/easymvp/apps/core/internal/controller/acceptance"
	"github.com/leef-l/easymvp/apps/core/internal/controller/architect_chat"
	"github.com/leef-l/easymvp/apps/core/internal/controller/audit"
	"github.com/leef-l/easymvp/apps/core/internal/controller/decisions"
	"github.com/leef-l/easymvp/apps/core/internal/controller/deliveries"
	"github.com/leef-l/easymvp/apps/core/internal/controller/designs"
	"github.com/leef-l/easymvp/apps/core/internal/controller/evidence"
	"github.com/leef-l/easymvp/apps/core/internal/controller/plan"
	"github.com/leef-l/easymvp/apps/core/internal/controller/projects"
	"github.com/leef-l/easymvp/apps/core/internal/controller/replay"
	"github.com/leef-l/easymvp/apps/core/internal/controller/requirements"
	"github.com/leef-l/easymvp/apps/core/internal/controller/retrospectives"
	"github.com/leef-l/easymvp/apps/core/internal/controller/reviews"
	"github.com/leef-l/easymvp/apps/core/internal/controller/runtime"
	"github.com/leef-l/easymvp/apps/core/internal/controller/system"
	"github.com/leef-l/easymvp/apps/core/internal/controller/workspace"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func middlewareCORS(r *ghttp.Request) {
	// Use permissive CORS for local dev; all API clients are first-party.
	corsOptions := r.Response.DefaultCORSOptions()
	corsOptions.AllowOrigin = "*"
	corsOptions.AllowMethods = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
	corsOptions.AllowHeaders = "Origin,Content-Type,Accept,Authorization,X-Request-Id"
	corsOptions.AllowCredentials = "true"
	r.Response.CORS(corsOptions)
	r.Middleware.Next()
}

var (
	Main = gcmd.Command{
		Name:        "main",
		Usage:       "main [--data-root=PATH] [--db-path=PATH] [--migration-path=PATH] [--brain-serve-base-url=URL] [--port=8000] [--safe-mode]",
		Brief:       "start http server",
		Description: "Start EasyMVP core service. Use --safe-mode to skip background workers and only expose recovery/diagnostic-safe APIs.",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			startup := service.ResolveStartupConfig(ctx, parser)
			service.SetStartupConfig(startup)

			if err = service.Bootstrap(ctx); err != nil {
				return err
			}
			if !startup.SafeMode {
				if err = service.Workers().Start(ctx); err != nil {
					return err
				}
			} else {
				g.Log().Warning(ctx, "EasyMVP core started in safe-mode; background workers are disabled")
			}
			defer func() {
				stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = service.Workers().Stop(stopCtx)
			}()

			s := g.Server()
			s.SetAddr(startup.ServerAddress)
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(middlewareCORS)
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					acceptance.NewV1(),
					architect_chat.NewV1(),
					audit.NewV1(),
					decisions.NewV1(),
					deliveries.NewV1(),
					designs.NewV1(),
					evidence.NewV1(),
					plan.NewV1(),
					projects.NewV1(),
					replay.NewV1(),
					requirements.NewV1(),
					retrospectives.NewV1(),
					reviews.NewV1(),
					runtime.NewV1(),
					system.NewV1(),
					workspace.NewV1(),
				)
				// SSE streaming endpoint for architect chat (manually bound)
				group.POST("/api/v3/projects/{id}/architect-chat/messages/stream", architect_chat.SendMessageStreamHandler)
			})
			// 启用 OpenAPI/Swagger 文档
			s.SetOpenApiPath("/api.json")
			s.SetSwaggerPath("/swagger")
			s.Run()
			return nil
		},
	}
)
