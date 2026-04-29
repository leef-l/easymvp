package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type CollectEvidenceReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/evidence/collect" method:"post" tags:"Evidence" summary:"Collect project evidence"`
	ProjectID string `json:"id" in:"path" v:"required"`
}

type CollectEvidenceRes struct {
	CommandID     string `json:"command_id"`
	Accepted      bool   `json:"accepted"`
	ResourceID    string `json:"resource_id"`
	NextAction    string `json:"next_action"`
	InsertedCount int    `json:"inserted_count"`
}
