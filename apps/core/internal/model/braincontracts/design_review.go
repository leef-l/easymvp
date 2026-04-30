package braincontracts

import "encoding/json"

// DesignReviewInput is the structured input for the design_review brain contract.
// The Verifier Brain reviews a solution design and returns structured feedback.
type DesignReviewInput struct {
	DesignID       string          `json:"design_id"`
	ProjectID      string          `json:"project_id"`
	DesignVersion  int             `json:"design_version"`
	Architecture   string          `json:"architecture"`
	ModulesJSON    json.RawMessage `json:"modules_json,omitempty"`
	DataModelsJSON json.RawMessage `json:"data_models_json,omitempty"`
	PagesJSON      json.RawMessage `json:"pages_json,omitempty"`
	TaskDraftsJSON json.RawMessage `json:"task_drafts_json,omitempty"`
	GoalSummary    string          `json:"goal_summary"`
	Round          int             `json:"round"`
	PreviousIssues []string        `json:"previous_issues,omitempty"`
}

// DesignReviewResult is the structured output from the design_review brain contract.
type DesignReviewResult struct {
	ReviewResultID string                  `json:"review_result_id"`
	Passed         bool                    `json:"passed"`
	Score          int                     `json:"score"`
	Dimensions     []DesignReviewDimension `json:"dimensions"`
	Issues         []DesignReviewIssue     `json:"issues"`
	Suggestions    []string                `json:"suggestions"`
}

// DesignReviewDimension represents a scored dimension in the design review.
type DesignReviewDimension struct {
	Name    string `json:"name"`
	Score   int    `json:"score"`
	Weight  int    `json:"weight"`
	Comment string `json:"comment,omitempty"`
}

// DesignReviewIssue represents a single issue found during design review.
type DesignReviewIssue struct {
	Code     string `json:"code,omitempty"`
	Severity string `json:"severity"` // critical | major | minor
	Summary  string `json:"summary"`
	Location string `json:"location,omitempty"`
	Fix      string `json:"fix,omitempty"`
}

// DesignFixInput is the structured input for the design_fix brain contract.
// The Central Brain fixes a solution design based on review issues.
type DesignFixInput struct {
	DesignID       string              `json:"design_id"`
	ProjectID      string              `json:"project_id"`
	DesignVersion  int                 `json:"design_version"`
	Architecture   string              `json:"architecture"`
	ModulesJSON    json.RawMessage     `json:"modules_json,omitempty"`
	DataModelsJSON json.RawMessage     `json:"data_models_json,omitempty"`
	PagesJSON      json.RawMessage     `json:"pages_json,omitempty"`
	TaskDraftsJSON json.RawMessage     `json:"task_drafts_json,omitempty"`
	GoalSummary    string              `json:"goal_summary"`
	Issues         []DesignReviewIssue `json:"issues"`
	Suggestions    []string            `json:"suggestions"`
}

// DesignFixResult is the structured output from the design_fix brain contract.
type DesignFixResult struct {
	FixResultID    string          `json:"fix_result_id"`
	Architecture   string          `json:"architecture"`
	ModulesJSON    json.RawMessage `json:"modules_json,omitempty"`
	DataModelsJSON json.RawMessage `json:"data_models_json,omitempty"`
	PagesJSON      json.RawMessage `json:"pages_json,omitempty"`
	TaskDraftsJSON json.RawMessage `json:"task_drafts_json,omitempty"`
	ChangesSummary string          `json:"changes_summary"`
	FixedIssues    []string        `json:"fixed_issues"`
}
