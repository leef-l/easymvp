package designs

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/designs/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Generate(ctx context.Context, req *v1.GenerateDesignReq) (res *v1.GenerateDesignRes, err error) {
	result, err := service.Design().GenerateDesign(ctx, req.ProjectID, req.RequirementID)
	if err != nil {
		return nil, err
	}
	return &v1.GenerateDesignRes{
		DesignID: result.DesignID,
		Status:   result.Status,
	}, nil
}
