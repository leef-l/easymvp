package reviews

import (
	"context"

	reviewsv1 "github.com/leef-l/easymvp/apps/core/api/reviews/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) StartReview(ctx context.Context, req *reviewsv1.StartReviewReq) (res *reviewsv1.StartReviewRes, err error) {
	result, err := service.Review().StartDesignReview(ctx, req.DesignID, req.ProjectID)
	if err != nil {
		return nil, err
	}
	return &reviewsv1.StartReviewRes{
		ReviewID: result.ReviewID,
		Passed:   result.Passed,
		Score:    result.Score,
		Issues:   result.Issues,
		Round:    result.Round,
	}, nil
}
