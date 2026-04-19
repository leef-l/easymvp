package runtime

import (
	"context"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) GetRunBinding(ctx context.Context, req *runtimev1.GetRunBindingReq) (res *runtimev1.GetRunBindingRes, err error) {
	return service.Runtime().GetRunBindingView(ctx, req.BindingID)
}
