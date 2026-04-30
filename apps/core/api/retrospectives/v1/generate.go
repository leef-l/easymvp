package v1

import "github.com/gogf/gf/v2/frame/g"

type GenerateRetrospectiveReq struct {
	g.Meta    `path:"/api/v3/retrospectives/generate" tags:"Retrospectives" method:"post" summary:"Generate project retrospective"`
	ProjectID string `json:"project_id" v:"required"`
}

type GenerateRetrospectiveRes struct {
	RetrospectiveID string `json:"retrospective_id"`
	TotalTasks      int    `json:"total_tasks"`
	CompletedTasks  int    `json:"completed_tasks"`
	FailedTasks     int    `json:"failed_tasks"`
}
