package braincontracts

import "encoding/json"

type CompletionAdjudicationInput struct {
	ExecutionSummaryJSON    json.RawMessage `json:"execution_summary_json"`
	DeliverySummaryJSON     json.RawMessage `json:"delivery_summary_json"`
	VerificationSummaryJSON json.RawMessage `json:"verification_summary_json"`
	AcceptanceSummaryJSON   json.RawMessage `json:"acceptance_summary_json"`
	ManualReleaseStateJSON  json.RawMessage `json:"manual_release_state_json"`
}

type CompletionAdjudicationResult struct {
	FunctionalPassed       bool   `json:"functional_passed"`
	ProductionPassed       bool   `json:"production_passed"`
	ManualReleaseRequired  bool   `json:"manual_release_required"`
	ManualReleaseCompleted bool   `json:"manual_release_completed"`
	FinalStatus            string `json:"final_status"`
	DecisionReason         string `json:"decision_reason"`
}
