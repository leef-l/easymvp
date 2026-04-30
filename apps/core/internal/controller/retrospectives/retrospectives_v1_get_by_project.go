package retrospectives

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/retrospectives/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) GetByProject(ctx context.Context, req *v1.GetProjectRetrospectiveReq) (res *v1.GetProjectRetrospectiveRes, err error) {
	retro, err := service.Retrospective().GetProjectRetrospective(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}
	return &v1.GetProjectRetrospectiveRes{Retrospective: retro}, nil
}
