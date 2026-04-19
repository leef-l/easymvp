package runtime

import (
	"context"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) CancelRunBinding(ctx context.Context, req *runtimev1.CancelRunBindingReq) (res *runtimev1.CancelRunBindingRes, err error) {
	return service.Runtime().CancelRunBindingCommand(ctx, req.BindingID)
}
