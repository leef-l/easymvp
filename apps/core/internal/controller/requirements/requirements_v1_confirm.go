package requirements

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/requirements/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Confirm(ctx context.Context, req *v1.ConfirmRequirementReq) (res *v1.ConfirmRequirementRes, err error) {
	if err = service.Requirement().ConfirmRequirement(ctx, req.ID); err != nil {
		return nil, err
	}
	return &v1.ConfirmRequirementRes{
		RequirementID: req.ID,
		Status:        "confirmed",
		NextAction:    "create_plan",
	}, nil
}
