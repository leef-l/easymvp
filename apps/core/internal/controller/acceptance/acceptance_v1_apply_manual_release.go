package acceptance

import (
	"context"

	acceptancev1 "github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ApplyManualRelease(ctx context.Context, req *acceptancev1.ApplyManualReleaseReq) (res *acceptancev1.ApplyManualReleaseRes, err error) {
	return service.Acceptance().ApplyManualRelease(ctx, service.ApplyManualReleaseCommand{
		ProjectID: req.ProjectID,
		Comment:   req.Comment,
	})
}
