package acceptance

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
)

type IAcceptanceV1 interface {
	AcceptanceView(ctx context.Context, req *v1.AcceptanceViewReq) (res *v1.AcceptanceViewRes, err error)
	StartAcceptance(ctx context.Context, req *v1.StartAcceptanceReq) (res *v1.StartAcceptanceRes, err error)
	AdjudicateAcceptance(ctx context.Context, req *v1.AdjudicateAcceptanceReq) (res *v1.AdjudicateAcceptanceRes, err error)
	ApplyManualRelease(ctx context.Context, req *v1.ApplyManualReleaseReq) (res *v1.ApplyManualReleaseRes, err error)
	RefreshAcceptanceProfiles(ctx context.Context, req *v1.RefreshAcceptanceProfilesReq) (res *v1.RefreshAcceptanceProfilesRes, err error)
}
