package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetRequirementReq — GET /api/v3/requirements/{id}
type GetRequirementReq struct {
	g.Meta `path:"/api/v3/requirements/{id}" tags:"Requirements" method:"get" summary:"Get requirement by ID"`
	ID     string `json:"id" in:"path" v:"required"`
}

type GetRequirementRes struct {
	Requirement *RequirementDetail `json:"requirement"`
}

// GetProjectRequirementReq — GET /api/v3/requirements/project/{projectId}
type GetProjectRequirementReq struct {
	g.Meta    `path:"/api/v3/requirements/project/{projectId}" tags:"Requirements" method:"get" summary:"Get latest requirement for a project"`
	ProjectID string `json:"projectId" in:"path" v:"required"`
}

type GetProjectRequirementRes struct {
	Requirement *RequirementDetail `json:"requirement"`
}
