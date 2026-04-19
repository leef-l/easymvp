package braincontracts

import "encoding/json"

type WorkspaceExplanationInput struct {
	WorkspaceContextJSON      json.RawMessage `json:"workspace_context_json"`
	RiskSummaryJSON           json.RawMessage `json:"risk_summary_json"`
	LatestDecisionSummaryJSON json.RawMessage `json:"latest_decision_summary_json"`
}

type WorkspaceExplanationResult struct {
	Headline           string                  `json:"headline"`
	Summary            string                  `json:"summary"`
	TopBlockers        []string                `json:"top_blockers"`
	RecommendedActions []RecommendedActionItem `json:"recommended_actions"`
	ExplainLinks       []string                `json:"explain_links,omitempty"`
}

type RecommendedActionItem struct {
	ActionKey string `json:"action_key"`
	Label     string `json:"label"`
	Reason    string `json:"reason"`
	DeepLink  string `json:"deep_link,omitempty"`
}
