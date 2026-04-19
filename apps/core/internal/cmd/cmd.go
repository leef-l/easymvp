package cmd

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"github.com/leef-l/easymvp/apps/core/internal/controller/acceptance"
	"github.com/leef-l/easymvp/apps/core/internal/controller/audit"
	"github.com/leef-l/easymvp/apps/core/internal/controller/plan"
	"github.com/leef-l/easymvp/apps/core/internal/controller/projects"
	"github.com/leef-l/easymvp/apps/core/internal/controller/replay"
	"github.com/leef-l/easymvp/apps/core/internal/controller/runtime"
	"github.com/leef-l/easymvp/apps/core/internal/controller/system"
	"github.com/leef-l/easymvp/apps/core/internal/controller/workspace"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			if err = service.Bootstrap(ctx); err != nil {
				return err
			}
			if err = service.Workers().Start(ctx); err != nil {
				return err
			}
			defer func() {
				stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = service.Workers().Stop(stopCtx)
			}()

			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					acceptance.NewV1(),
					audit.NewV1(),
					plan.NewV1(),
					projects.NewV1(),
					replay.NewV1(),
					runtime.NewV1(),
					system.NewV1(),
					workspace.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)
