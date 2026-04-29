package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type DeleteProjectReq struct {
	g.Meta `path:"/api/v3/projects/{id}" tags:"Projects" method:"delete" summary:"Delete project"`
	Id     string `json:"id" in:"path" v:"required"`
}

type DeleteProjectRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}
