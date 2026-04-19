package braincontracts

import "encoding/json"

type PlanReviewInput struct {
	PlanDraftID            string          `json:"plan_draft_id"`
	PlanDraftVersion       int             `json:"plan_draft_version"`
	PlanDraftJSON          json.RawMessage `json:"plan_draft_json"`
	ProjectCategory        string          `json:"project_category"`
	CategoryProfileVersion int             `json:"category_profile_version"`
	CategoryProfileJSON    json.RawMessage `json:"category_profile_json"`
	ProjectContextJSON     json.RawMessage `json:"project_context_json,omitempty"`
	ArtifactRefs           []ArtifactRef   `json:"artifact_refs,omitempty"`
}

type PlanReviewResult struct {
	ReviewResultID string      `json:"review_result_id"`
	ReviewVersion  int         `json:"review_version"`
	Decision       string      `json:"decision"`
	CompileAllowed bool        `json:"compile_allowed"`
	BlockingIssues []IssueItem `json:"blocking_issues"`
	AdvisoryIssues []IssueItem `json:"advisory_issues"`
	RewriteHints   []string    `json:"rewrite_hints,omitempty"`
}

type IssueItem struct {
	Code     string `json:"code,omitempty"`
	Severity string `json:"severity,omitempty"`
	Summary  string `json:"summary,omitempty"`
}
