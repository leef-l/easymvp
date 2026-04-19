package runtime

import (
	"context"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) SyncRunBinding(ctx context.Context, req *runtimev1.SyncRunBindingReq) (res *runtimev1.SyncRunBindingRes, err error) {
	return service.Runtime().SyncRunBindingCommand(ctx, req.BindingID)
}
