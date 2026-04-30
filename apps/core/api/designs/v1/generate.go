package v1

import "github.com/gogf/gf/v2/frame/g"

type GenerateDesignReq struct {
	g.Meta        `path:"/api/v3/designs/generate" tags:"Designs" method:"post" summary:"Generate solution design"`
	ProjectID     string `json:"project_id" v:"required"`
	RequirementID string `json:"requirement_id" v:"required"`
}

type GenerateDesignRes struct {
	DesignID string `json:"design_id"`
	Status   string `json:"status"`
}
