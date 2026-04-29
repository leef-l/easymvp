package projects

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) WorkspaceView(ctx context.Context, req *v1.ProjectWorkspaceViewReq) (res *v1.ProjectWorkspaceViewRes, err error) {
	return service.Projects().GetProjectWorkspaceView(ctx, req.Id)
}
