package designs

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/designs/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) GetProjectLatest(ctx context.Context, req *v1.GetProjectLatestDesignReq) (res *v1.GetProjectLatestDesignRes, err error) {
	design, err := service.Design().GetProjectLatestDesign(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}
	return &v1.GetProjectLatestDesignRes{Design: design}, nil
}
