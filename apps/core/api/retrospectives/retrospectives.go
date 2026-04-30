package retrospectives

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/retrospectives/v1"
)

type IRetrospectivesV1 interface {
	Generate(ctx context.Context, req *v1.GenerateRetrospectiveReq) (res *v1.GenerateRetrospectiveRes, err error)
	Get(ctx context.Context, req *v1.GetRetrospectiveReq) (res *v1.GetRetrospectiveRes, err error)
	GetByProject(ctx context.Context, req *v1.GetProjectRetrospectiveReq) (res *v1.GetProjectRetrospectiveRes, err error)
}
