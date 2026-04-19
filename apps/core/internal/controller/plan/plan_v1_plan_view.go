package plan

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/plan/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) PlanView(ctx context.Context, req *v1.PlanViewReq) (res *v1.PlanViewRes, err error) {
	return service.Plan().GetPlanView(ctx, req.ProjectID)
}
