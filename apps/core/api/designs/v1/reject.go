package v1

import "github.com/gogf/gf/v2/frame/g"

type RejectDesignReq struct {
	g.Meta `path:"/api/v3/designs/{id}/reject" tags:"Designs" method:"post" summary:"Reject solution design"`
	Id     string `json:"id" in:"path" v:"required"`
	Reason string `json:"reason"`
}

type RejectDesignRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}
