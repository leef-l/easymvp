package v1

import "github.com/gogf/gf/v2/frame/g"

type PrepareDeliveryReq struct {
	g.Meta    `path:"/api/v3/deliveries/prepare" tags:"Deliveries" method:"post" summary:"Prepare project delivery"`
	ProjectID string `json:"project_id" v:"required"`
}

type PrepareDeliveryRes struct {
	DeliveryID string `json:"delivery_id"`
	Status     string `json:"status"`
}
