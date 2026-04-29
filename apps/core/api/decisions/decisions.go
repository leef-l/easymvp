package decisions

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/decisions/v1"
)

type IDecisionsV1 interface {
	Apply(ctx context.Context, req *v1.ApplyManualDecisionReq) (res *v1.ApplyManualDecisionRes, err error)
}
