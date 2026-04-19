package system

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/system/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ListProjectDiagnostics(ctx context.Context, req *v1.ListProjectDiagnosticsReq) (res *v1.ListProjectDiagnosticsRes, err error) {
	return service.System().ListProjectDiagnostics(ctx, req.ProjectID, req.Limit)
}
