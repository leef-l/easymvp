package v1

import "github.com/gogf/gf/v2/frame/g"

type ApplyManualDecisionReq struct {
	g.Meta     `path:"/api/v3/manual-decisions" tags:"Decisions" method:"post" summary:"Apply manual decision"`
	ProjectID  string `json:"project_id" v:"required"`
	TargetKind string `json:"target_kind" v:"required"` // e.g. "plan", "task", "acceptance_run"
	TargetID   string `json:"target_id" v:"required"`
	Decision   string `json:"decision" v:"required"`    // "approved", "rejected", "needs_review"
	Reason     string `json:"reason"`
	Comment    string `json:"comment"`
}

type ApplyManualDecisionRes struct {
	CommandID  string `json:"command_id"`
	Accepted   bool   `json:"accepted"`
	ResourceID string `json:"resource_id"`
	NextAction string `json:"next_action"`
}
