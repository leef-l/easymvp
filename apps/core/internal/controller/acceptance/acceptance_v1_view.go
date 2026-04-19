package acceptance

import (
	"context"

	acceptancev1 "github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) AcceptanceView(ctx context.Context, req *acceptancev1.AcceptanceViewReq) (res *acceptancev1.AcceptanceViewRes, err error) {
	return service.Acceptance().GetAcceptanceView(ctx, req.ProjectID)
}
