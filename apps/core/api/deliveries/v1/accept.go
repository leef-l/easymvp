package v1

import "github.com/gogf/gf/v2/frame/g"

type AcceptDeliveryReq struct {
	g.Meta `path:"/api/v3/deliveries/{id}/accept" tags:"Deliveries" method:"post" summary:"Accept project delivery"`
	Id     string `json:"id" in:"path" v:"required"`
}

type AcceptDeliveryRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}
