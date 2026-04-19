package runtime

import (
	"context"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) GetRunBindingDetail(ctx context.Context, req *runtimev1.GetRunBindingDetailReq) (res *runtimev1.GetRunBindingDetailRes, err error) {
	return service.Runtime().GetRunBindingDetail(ctx, req.BindingID, req.EventLimit)
}
