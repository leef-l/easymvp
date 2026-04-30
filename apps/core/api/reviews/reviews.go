package reviews

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/reviews/v1"
)

type IReviewsV1 interface {
	StartReview(ctx context.Context, req *v1.StartReviewReq) (res *v1.StartReviewRes, err error)
	RunReviewLoop(ctx context.Context, req *v1.RunReviewLoopReq) (res *v1.RunReviewLoopRes, err error)
	ListReviews(ctx context.Context, req *v1.ListReviewsReq) (res *v1.ListReviewsRes, err error)
	Intervene(ctx context.Context, req *v1.InterveneReq) (res *v1.InterveneRes, err error)
}
