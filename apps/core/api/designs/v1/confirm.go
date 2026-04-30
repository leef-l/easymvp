package v1

import "github.com/gogf/gf/v2/frame/g"

type ConfirmDesignReq struct {
	g.Meta `path:"/api/v3/designs/{id}/confirm" tags:"Designs" method:"post" summary:"Confirm solution design"`
	Id     string `json:"id" in:"path" v:"required"`
}

type ConfirmDesignRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}
