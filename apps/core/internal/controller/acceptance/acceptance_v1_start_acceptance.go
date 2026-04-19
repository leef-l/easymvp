package acceptance

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) StartAcceptance(ctx context.Context, req *v1.StartAcceptanceReq) (res *v1.StartAcceptanceRes, err error) {
	return service.Acceptance().StartAcceptanceRun(ctx, service.StartAcceptanceCommand{
		ProjectID:      req.ProjectID,
		TaskID:         req.TaskID,
		ProfileVersion: req.ProfileVersion,
		Mode:           req.Mode,
	})
}

func (c *ControllerV1) AdjudicateAcceptance(ctx context.Context, req *v1.AdjudicateAcceptanceReq) (res *v1.AdjudicateAcceptanceRes, err error) {
	return service.Acceptance().AdjudicateAcceptanceRun(ctx, req.ProjectID)
}

func (c *ControllerV1) RefreshAcceptanceProfiles(ctx context.Context, req *v1.RefreshAcceptanceProfilesReq) (res *v1.RefreshAcceptanceProfilesRes, err error) {
	return service.Acceptance().RefreshAcceptanceProfiles(ctx, req.ProjectID)
}
