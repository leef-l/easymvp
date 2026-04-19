package runtime

import (
	"context"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ResumeRunBinding(ctx context.Context, req *runtimev1.ResumeRunBindingReq) (res *runtimev1.ResumeRunBindingRes, err error) {
	return service.Runtime().ResumeRunBindingCommand(ctx, req.BindingID)
}
