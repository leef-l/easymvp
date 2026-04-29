package architect_chat

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/architect_chat/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ConfirmPlan(ctx context.Context, req *v1.ConfirmPlanReq) (res *v1.ConfirmPlanRes, err error) {
	result, err := service.ArchitectChat().ConfirmPlan(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}
	return &v1.ConfirmPlanRes{
		CommandID:      result.CommandID,
		Accepted:       result.Accepted,
		CompiledPlanID: result.CompiledPlanID,
		Reason:         result.Reason,
		ReviewID:       result.ReviewID,
	}, nil
}
