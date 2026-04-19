package plan

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/plan/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) CompilePlan(ctx context.Context, req *v1.CompilePlanReq) (res *v1.CompilePlanRes, err error) {
	return service.Plan().CompilePlan(ctx, service.CompilePlanCommand{
		ProjectID:      req.ProjectID,
		PlanDraftID:    req.PlanDraftID,
		ForceRecompile: req.ForceRecompile,
	})
}
