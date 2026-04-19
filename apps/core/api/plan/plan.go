package plan

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/plan/v1"
)

type IPlanV1 interface {
	PlanView(ctx context.Context, req *v1.PlanViewReq) (res *v1.PlanViewRes, err error)
	RepairDraft(ctx context.Context, req *v1.RepairDraftReq) (res *v1.RepairDraftRes, err error)
	CreateRepairDraft(ctx context.Context, req *v1.CreateRepairDraftReq) (res *v1.CreateRepairDraftRes, err error)
	CompilePlan(ctx context.Context, req *v1.CompilePlanReq) (res *v1.CompilePlanRes, err error)
}
