package decisions

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/decisions/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Apply(ctx context.Context, req *v1.ApplyManualDecisionReq) (res *v1.ApplyManualDecisionRes, err error) {
	return service.Decisions().ApplyManualDecision(ctx, service.ApplyManualDecisionCommand{
		ProjectID:  req.ProjectID,
		TargetKind: req.TargetKind,
		TargetID:   req.TargetID,
		Decision:   req.Decision,
		Reason:     req.Reason,
		Comment:    req.Comment,
	})
}
