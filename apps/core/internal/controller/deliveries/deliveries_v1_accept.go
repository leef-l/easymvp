package deliveries

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/deliveries/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Accept(ctx context.Context, req *v1.AcceptDeliveryReq) (res *v1.AcceptDeliveryRes, err error) {
	if err = service.Delivery().AcceptDelivery(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.AcceptDeliveryRes{
		Accepted:   true,
		ResourceID: req.Id,
		NextAction: "generate_retrospective",
	}, nil
}
