package braincontracts

import "encoding/json"

type CompletionAdjudicationInput struct {
	ExecutionSummaryJSON    json.RawMessage `json:"execution_summary_json"`
	DeliverySummaryJSON     json.RawMessage `json:"delivery_summary_json"`
	VerificationSummaryJSON json.RawMessage `json:"verification_summary_json"`
	AcceptanceSummaryJSON   json.RawMessage `json:"acceptance_summary_json"`
	ManualReleaseStateJSON  json.RawMessage `json:"manual_release_state_json"`
	// BrowserValidationResultJSON carries the structured result of the browser
	// brain's automated UI verification (DOM snapshot, screenshot, anomalies).
	BrowserValidationResultJSON json.RawMessage `json:"browser_validation_result_json,omitempty"`
	// VerifierCheckResultJSON carries the structured result of the verifier
	// brain's automated test execution (unit tests, assertions, static checks).
	VerifierCheckResultJSON json.RawMessage `json:"verifier_check_result_json,omitempty"`
}

type CompletionAdjudicationResult struct {
	FunctionalPassed       bool   `json:"functional_passed"`
	ProductionPassed       bool   `json:"production_passed"`
	ManualReleaseRequired  bool   `json:"manual_release_required"`
	ManualReleaseCompleted bool   `json:"manual_release_completed"`
	FinalStatus            string `json:"final_status"`
	DecisionReason         string `json:"decision_reason"`
	// Four-layer completion state flags (mapped from Engineering Cybernetics ch.4 feedback loop)
	ExecutorSucceeded bool   `json:"executor_succeeded"`
	DeliveryVerified  bool   `json:"delivery_verified"`
	AcceptancePassed  bool   `json:"acceptance_passed"`
	Completed         bool   `json:"completed"`
	Decision          string `json:"decision"` // complete | rework | blocked | manual_checkpoint
}
