package projects

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteProjectReq) (res *v1.DeleteProjectRes, err error) {
	return service.Projects().DeleteProject(ctx, req.Id)
}
