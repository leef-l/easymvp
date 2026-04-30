package reviews

import (
	"context"

	reviewsv1 "github.com/leef-l/easymvp/apps/core/api/reviews/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) RunReviewLoop(ctx context.Context, req *reviewsv1.RunReviewLoopReq) (res *reviewsv1.RunReviewLoopRes, err error) {
	result, err := service.Review().RunReviewLoop(ctx, req.DesignID, req.ProjectID, req.MaxRounds)
	if err != nil {
		return nil, err
	}
	return &reviewsv1.RunReviewLoopRes{
		Passed:        result.Passed,
		Rounds:        result.Rounds,
		Reason:        result.Reason,
		FinalReviewID: result.FinalReviewID,
	}, nil
}
