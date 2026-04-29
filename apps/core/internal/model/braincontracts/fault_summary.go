package braincontracts

// FaultSummary is the normalized runtime DTO for a failure event.
// It replaces raw logs with structured failure semantics so that the
// diagnostics page and repair_design contract can consume it directly.
type FaultSummary struct {
	FaultID                  string   `json:"fault_id"`
	SourceBrain              string   `json:"source_brain"`
	ReasonCode               string   `json:"reason_code"`
	ReasonSummary            string   `json:"reason_summary"`
	ReasonClass              string   `json:"reason_class"`      // execution_error | verification_failure | delivery_mismatch | policy_violation | environment_failure
	FailureClass             string   `json:"failure_class"`     // transient | permanent | cascading | unknown
	FailureStage             string   `json:"failure_stage"`     // design | review | execute | accept | complete
	ErrorSource              string   `json:"error_source"`      // brain | runtime | orchestration
	NonlinearityClass        string   `json:"nonlinearity_class,omitempty"` // oscillation | divergence | saturation | deadlock
	RepairStrategy           string   `json:"repair_strategy"`   // retry | redesign | replace | escalate | manual_checkpoint
	RecommendedAction        string   `json:"recommended_action"`
	RecoveryOptions          []string `json:"recovery_options"`
	AffectedTaskKeys         []string `json:"affected_task_keys"`
	DownstreamBlockedTasks   []string `json:"downstream_blocked_tasks,omitempty"`
	OccurredAt               string   `json:"occurred_at"`
}
