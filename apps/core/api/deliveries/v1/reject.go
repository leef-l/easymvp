package v1

import "github.com/gogf/gf/v2/frame/g"

type RejectDeliveryReq struct {
	g.Meta `path:"/api/v3/deliveries/{id}/reject" tags:"Deliveries" method:"post" summary:"Reject project delivery"`
	Id     string `json:"id" in:"path" v:"required"`
	Reason string `json:"reason"`
}

type RejectDeliveryRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}
