package plan

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/plan/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) RedesignPlan(ctx context.Context, req *v1.RedesignPlanReq) (res *v1.RedesignPlanRes, err error) {
	return service.Plan().RedesignPlan(ctx, req.Id, req.Feedback)
}
