package requirements

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/requirements/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Analyze(ctx context.Context, req *v1.AnalyzeRequirementReq) (res *v1.AnalyzeRequirementRes, err error) {
	result, err := service.Requirement().AnalyzeRequirement(ctx, req.ProjectID, req.RawInput)
	if err != nil {
		return nil, err
	}
	return &v1.AnalyzeRequirementRes{
		RequirementID:  result.RequirementID,
		Status:         result.Status,
		Summary:        result.Summary,
		RequirementDoc: result.RequirementDoc,
		NextAction:     "confirm_requirement",
	}, nil
}
