package braincontracts

// RuntimeEscalation is the normalized runtime DTO for a situation where
// control must be transferred from the current brain to another brain or
// to a human reviewer.
type RuntimeEscalation struct {
	EscalationID       string `json:"escalation_id"`
	EscalationType     string `json:"escalation_type"`     // unsupported_capability | policy_denied | verification_conflict | environment_unavailable | manual_review_required | fault_loop_detected
	SourceBrain        string `json:"source_brain"`
	TargetBrain        string `json:"target_brain,omitempty"`
	ReasonCode         string `json:"reason_code"`
	ReasonSummary      string `json:"reason_summary"`
	RecommendedAction  string `json:"recommended_action"`
	FailureClass       string `json:"failure_class"`
	FailureStage       string `json:"failure_stage"`
	RecoveryOptions    []string `json:"recovery_options"`
	RelatedFaultID     string `json:"related_fault_id,omitempty"`
	RelatedRunID       string `json:"related_run_id,omitempty"`
	CreatedAt          string `json:"created_at"`
}
