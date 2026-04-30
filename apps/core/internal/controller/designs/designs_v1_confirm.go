package designs

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/designs/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Confirm(ctx context.Context, req *v1.ConfirmDesignReq) (res *v1.ConfirmDesignRes, err error) {
	if err = service.Design().ConfirmDesign(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.ConfirmDesignRes{
		Accepted:   true,
		ResourceID: req.Id,
		NextAction: "generate_plan",
	}, nil
}
