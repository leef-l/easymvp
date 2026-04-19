package projects

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateProjectReq) (res *v1.CreateProjectRes, err error) {
	return service.Projects().CreateProject(ctx, service.CreateProjectCommand{
		Name:            req.Name,
		ProjectCategory: req.ProjectCategory,
		GoalSummary:     req.GoalSummary,
		WorkspaceRoot:   req.WorkspaceRoot,
		RepoRoot:        req.RepoRoot,
	})
}
