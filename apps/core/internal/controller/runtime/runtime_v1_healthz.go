package runtime

import (
	"context"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Healthz(ctx context.Context, req *runtimev1.HealthzReq) (res *runtimev1.HealthzRes, err error) {
	return service.Runtime().Healthz(ctx)
}
