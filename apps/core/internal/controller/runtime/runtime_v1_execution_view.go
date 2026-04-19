package runtime

import (
	"context"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ExecutionView(ctx context.Context, req *runtimev1.ExecutionViewReq) (res *runtimev1.ExecutionViewRes, err error) {
	return service.Runtime().GetExecutionView(ctx, req.ProjectID, req.BindingID, req.EventLimit, req.ReplayLimit, req.LogLimit)
}
