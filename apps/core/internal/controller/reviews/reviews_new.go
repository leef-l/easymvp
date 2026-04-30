package reviews

import api "github.com/leef-l/easymvp/apps/core/api/reviews"

func NewV1() api.IReviewsV1 {
	return &ControllerV1{}
}
