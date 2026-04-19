package projects

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/projects/v1"
)

type IProjectsV1 interface {
	Create(ctx context.Context, req *v1.CreateProjectReq) (res *v1.CreateProjectRes, err error)
	WorkspaceView(ctx context.Context, req *v1.ProjectWorkspaceViewReq) (res *v1.ProjectWorkspaceViewRes, err error)
}
