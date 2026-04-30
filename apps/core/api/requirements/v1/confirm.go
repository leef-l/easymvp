package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type ConfirmRequirementReq struct {
	g.Meta `path:"/api/v3/requirements/{id}/confirm" tags:"Requirements" method:"post" summary:"Confirm a requirement"`
	ID     string `json:"id" in:"path" v:"required"`
}

type ConfirmRequirementRes struct {
	RequirementID string `json:"requirement_id"`
	Status        string `json:"status"`
	NextAction    string `json:"next_action"`
}
