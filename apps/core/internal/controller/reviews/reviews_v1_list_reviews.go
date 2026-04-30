package reviews

import (
	"context"

	reviewsv1 "github.com/leef-l/easymvp/apps/core/api/reviews/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) ListReviews(ctx context.Context, req *reviewsv1.ListReviewsReq) (res *reviewsv1.ListReviewsRes, err error) {
	records, err := service.Review().GetDesignReviews(ctx, req.DesignID)
	if err != nil {
		return nil, err
	}

	items := make([]reviewsv1.ReviewItem, 0, len(records))
	for _, r := range records {
		items = append(items, reviewsv1.ReviewItem{
			ID:              r.Id,
			DesignID:        r.DesignId,
			ProjectID:       r.ProjectId,
			Round:           r.Round,
			Passed:          r.Passed == 1,
			Score:           r.Score,
			DimensionsJSON:  r.DimensionsJson,
			IssuesJSON:      r.IssuesJson,
			SuggestionsJSON: r.SuggestionsJson,
			FixTasksJSON:    r.FixTasksJson,
			BrainRunID:      r.BrainRunId,
			CreatedAt:       r.CreatedAt,
		})
	}

	return &reviewsv1.ListReviewsRes{
		Items: items,
		Total: len(items),
	}, nil
}
