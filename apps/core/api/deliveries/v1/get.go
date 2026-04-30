package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

type GetDeliveryReq struct {
	g.Meta `path:"/api/v3/deliveries/{id}" tags:"Deliveries" method:"get" summary:"Get project delivery by ID"`
	Id     string `json:"id" in:"path" v:"required"`
}

type GetDeliveryRes struct {
	Delivery *entity.ProjectDeliveries `json:"delivery"`
}
