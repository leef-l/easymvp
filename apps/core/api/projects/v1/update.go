package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type UpdateProjectReq struct {
	g.Meta          `path:"/api/v3/projects/{id}" tags:"Projects" method:"put" summary:"Update project"`
	Id              string `json:"id" in:"path" v:"required"`
	Name            string `json:"name"`
	GoalSummary     string `json:"goal_summary"`
	WorkspaceRoot   string `json:"workspace_root"`
	RepoRoot        string `json:"repo_root"`
}

type UpdateProjectRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}
