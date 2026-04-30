package projects

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/projects/v1"
)

type IProjectsV1 interface {
	Create(ctx context.Context, req *v1.CreateProjectReq) (res *v1.CreateProjectRes, err error)
	Update(ctx context.Context, req *v1.UpdateProjectReq) (res *v1.UpdateProjectRes, err error)
	Delete(ctx context.Context, req *v1.DeleteProjectReq) (res *v1.DeleteProjectRes, err error)
	WorkspaceView(ctx context.Context, req *v1.ProjectWorkspaceViewReq) (res *v1.ProjectWorkspaceViewRes, err error)
	ProjectProgressStream(ctx context.Context, req *v1.ProjectProgressStreamReq) (res *v1.ProjectProgressStreamRes, err error)
}
