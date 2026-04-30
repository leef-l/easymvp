package designs

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/designs/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Get(ctx context.Context, req *v1.GetDesignReq) (res *v1.GetDesignRes, err error) {
	design, err := service.Design().GetDesign(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetDesignRes{Design: design}, nil
}
