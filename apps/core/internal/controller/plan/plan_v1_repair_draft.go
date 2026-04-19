package plan

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/plan/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) RepairDraft(ctx context.Context, req *v1.RepairDraftReq) (res *v1.RepairDraftRes, err error) {
	return service.Plan().GetRepairDraftView(ctx, req.ProjectID)
}

func (c *ControllerV1) CreateRepairDraft(ctx context.Context, req *v1.CreateRepairDraftReq) (res *v1.CreateRepairDraftRes, err error) {
	result, err := service.Plan().CreateRepairDraft(ctx, service.CreateRepairDraftCommand{
		ProjectID:             req.ProjectID,
		FailedTaskContextJSON: req.FailedTaskContextJSON,
		FailureReasonJSON:     req.FailureReasonJSON,
		OriginalContractsJSON: req.OriginalContractsJSON,
		RuntimeSummaryJSON:    req.RuntimeSummaryJSON,
		ArtifactRefs:          req.ArtifactRefs,
		CreatedBy:             req.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRepairDraftRes{
		CommandID:  result.CommandID,
		Accepted:   result.Accepted,
		ResourceID: result.ResourceID,
		NextAction: result.NextAction,
	}, nil
}
