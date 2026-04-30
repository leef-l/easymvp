package designs

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/designs/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Reject(ctx context.Context, req *v1.RejectDesignReq) (res *v1.RejectDesignRes, err error) {
	if err = service.Design().RejectDesign(ctx, req.Id, req.Reason); err != nil {
		return nil, err
	}
	return &v1.RejectDesignRes{
		Accepted:   true,
		ResourceID: req.Id,
		NextAction: "regenerate_design",
	}, nil
}
