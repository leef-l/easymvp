package reviews

import (
	"context"

	reviewsv1 "github.com/leef-l/easymvp/apps/core/api/reviews/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Intervene(ctx context.Context, req *reviewsv1.InterveneReq) (res *reviewsv1.InterveneRes, err error) {
	if err = service.Review().Intervene(ctx, req.DesignID, req.Action, req.Reason); err != nil {
		return nil, err
	}
	return &reviewsv1.InterveneRes{
		DesignID: req.DesignID,
		Action:   req.Action,
		Applied:  true,
		Message:  "intervention applied successfully",
	}, nil
}
