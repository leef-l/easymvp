package projects

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateProjectReq) (res *v1.UpdateProjectRes, err error) {
	return service.Projects().UpdateProject(ctx, service.UpdateProjectCommand{
		ProjectID:     req.Id,
		Name:          req.Name,
		GoalSummary:   req.GoalSummary,
		WorkspaceRoot: req.WorkspaceRoot,
		RepoRoot:      req.RepoRoot,
	})
}
