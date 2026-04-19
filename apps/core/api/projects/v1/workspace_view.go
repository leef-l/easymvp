package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type ProjectWorkspaceViewReq struct {
	g.Meta    `path:"/api/v3/projects/{id}/workspace-view" tags:"Projects" method:"get" summary:"Project workspace view"`
	ProjectID string `json:"id" in:"path" v:"required"`
}

type ProjectWorkspaceViewRes struct {
	ProjectSnapshot      ProjectSnapshot      `json:"project_snapshot"`
	StageProgress        []StageProgressItem  `json:"stage_progress"`
	LiveActivity         []LiveActivityItem   `json:"live_activity"`
	ActionInbox          []ActionInboxItem    `json:"action_inbox"`
	AcceptanceCoverage   AcceptanceCoverage   `json:"acceptance_coverage"`
	WorkspaceExplanation WorkspaceExplanation `json:"workspace_explanation"`
}
