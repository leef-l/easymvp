package workspace

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/workspace/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) HomeView(ctx context.Context, req *v1.HomeViewReq) (res *v1.HomeViewRes, err error) {
	_ = req

	return service.Workspace().GetHomeView(ctx)
}
