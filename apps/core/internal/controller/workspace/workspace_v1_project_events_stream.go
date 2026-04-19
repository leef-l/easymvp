package workspace

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/api/workspace/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ProjectEventsStream(ctx context.Context, req *v1.ProjectEventsStreamReq) (res *v1.ProjectEventsStreamRes, err error) {
	if err = service.Workspace().ProjectEventsStream(ctx, req); err != nil {
		return nil, err
	}
	g.RequestFromCtx(ctx).ExitAll()
	return nil, nil
}
