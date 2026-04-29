package braincontracts

import "encoding/json"

type PlanRedesignInput struct {
	PlanDraftID      string          `json:"plan_draft_id"`
	PlanDraftJSON    json.RawMessage `json:"plan_draft_json"`
	ReviewResultID   string          `json:"review_result_id"`
	ReviewResultJSON json.RawMessage `json:"review_result_json"`
	RewriteHints     []string        `json:"rewrite_hints"`
	Feedback         string          `json:"feedback,omitempty"`
}

type PlanRedesignResult struct {
	RedesignedPlanDraftID string          `json:"redesigned_plan_draft_id"`
	RedesignedPlanJSON    json.RawMessage `json:"redesigned_plan_json"`
	ChangesSummary        string          `json:"changes_summary"`
}
