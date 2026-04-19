package workspace

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/workspace/v1"
)

type IWorkspaceV1 interface {
	HomeView(ctx context.Context, req *v1.HomeViewReq) (res *v1.HomeViewRes, err error)
	ProjectEventsStream(ctx context.Context, req *v1.ProjectEventsStreamReq) (res *v1.ProjectEventsStreamRes, err error)
}
