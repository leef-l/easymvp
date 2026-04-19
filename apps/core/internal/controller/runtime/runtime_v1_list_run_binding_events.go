package runtime

import (
	"context"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ListRunBindingEvents(ctx context.Context, req *runtimev1.ListRunBindingEventsReq) (res *runtimev1.ListRunBindingEventsRes, err error) {
	return service.Runtime().ListRunBindingEvents(ctx, req.BindingID, req.Limit)
}
