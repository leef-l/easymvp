package deliveries

import (
	"context"

	"github.com/leef-l/easymvp/apps/core/api/deliveries/v1"
)

type IDeliveriesV1 interface {
	Prepare(ctx context.Context, req *v1.PrepareDeliveryReq) (res *v1.PrepareDeliveryRes, err error)
	Accept(ctx context.Context, req *v1.AcceptDeliveryReq) (res *v1.AcceptDeliveryRes, err error)
	Reject(ctx context.Context, req *v1.RejectDeliveryReq) (res *v1.RejectDeliveryRes, err error)
	Get(ctx context.Context, req *v1.GetDeliveryReq) (res *v1.GetDeliveryRes, err error)
}
