package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type AnalyzeRequirementReq struct {
	g.Meta    `path:"/api/v3/requirements/analyze" tags:"Requirements" method:"post" summary:"Analyze requirement from raw natural-language input"`
	ProjectID string `json:"project_id" v:"required"`
	RawInput  string `json:"raw_input"  v:"required"`
}

type AnalyzeRequirementRes struct {
	RequirementID  string `json:"requirement_id"`
	Status         string `json:"status"`
	Summary        string `json:"summary"`
	RequirementDoc string `json:"requirement_doc_json"`
	NextAction     string `json:"next_action"`
}
