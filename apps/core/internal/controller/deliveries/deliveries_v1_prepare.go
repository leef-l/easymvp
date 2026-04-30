package deliveries

import (
	"context"

	v1 "github.com/leef-l/easymvp/apps/core/api/deliveries/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

func (c *ControllerV1) Prepare(ctx context.Context, req *v1.PrepareDeliveryReq) (res *v1.PrepareDeliveryRes, err error) {
	result, err := service.Delivery().PrepareDelivery(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}
	return &v1.PrepareDeliveryRes{
		DeliveryID: result.DeliveryID,
		Status:     result.Status,
	}, nil
}
