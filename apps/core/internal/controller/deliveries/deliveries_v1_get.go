package deliveries

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/deliveries/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Get(ctx context.Context, req *v1.GetDeliveryReq) (res *v1.GetDeliveryRes, err error) {
	delivery, err := service.Delivery().GetDelivery(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetDeliveryRes{Delivery: delivery}, nil
}
