package requirements

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/requirements/v1"
)

type IRequirementsV1 interface {
	Analyze(ctx context.Context, req *v1.AnalyzeRequirementReq) (res *v1.AnalyzeRequirementRes, err error)
	Confirm(ctx context.Context, req *v1.ConfirmRequirementReq) (res *v1.ConfirmRequirementRes, err error)
	Get(ctx context.Context, req *v1.GetRequirementReq) (res *v1.GetRequirementRes, err error)
	GetByProject(ctx context.Context, req *v1.GetProjectRequirementReq) (res *v1.GetProjectRequirementRes, err error)
}
