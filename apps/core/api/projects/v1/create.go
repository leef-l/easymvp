package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type CreateProjectReq struct {
	g.Meta          `path:"/api/v3/projects" tags:"Projects" method:"post" summary:"Create project"`
	Name            string `json:"name" v:"required"`
	ProjectCategory string `json:"project_category" v:"required"`
	GoalSummary     string `json:"goal_summary" v:"required"`
	WorkspaceRoot   string `json:"workspace_root" v:"required"`
	RepoRoot        string `json:"repo_root"`
}

type CreateProjectRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}
