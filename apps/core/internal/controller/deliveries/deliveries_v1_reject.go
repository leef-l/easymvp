package deliveries

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/deliveries/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Reject(ctx context.Context, req *v1.RejectDeliveryReq) (res *v1.RejectDeliveryRes, err error) {
	if err = service.Delivery().RejectDelivery(ctx, req.Id, req.Reason); err != nil {
		return nil, err
	}
	return &v1.RejectDeliveryRes{
		Accepted:   true,
		ResourceID: req.Id,
		NextAction: "revise_delivery",
	}, nil
}
