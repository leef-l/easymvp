package retrospectives

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/retrospectives/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Get(ctx context.Context, req *v1.GetRetrospectiveReq) (res *v1.GetRetrospectiveRes, err error) {
	retro, err := service.Retrospective().GetRetrospective(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRetrospectiveRes{Retrospective: retro}, nil
}
