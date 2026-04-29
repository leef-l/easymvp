package braincontracts

import "encoding/json"

// CentralCoordinationIntent is the structured input object sent to the central
// brain when multiple specialist brains need to collaborate. It replaces ad-hoc
// string prompts with explicit routing instructions.
type CentralCoordinationIntent struct {
	IntentID          string            `json:"intent_id"`
	IntentKind        string            `json:"intent_kind"`        // "plan_compilation" | "execution_delegation" | "verification_aggregation" | "fault_analysis" | "completion_arbitration"
	SourceStage       string            `json:"source_stage"`       // designing | reviewing | executing | accepting | reworking | completed
	TargetBrains      []string          `json:"target_brains"`      // e.g. ["code", "browser", "verifier"]
	PrimaryBrain      string            `json:"primary_brain"`
	TaskContextJSON   json.RawMessage   `json:"task_context_json"`
	RequiredOutputs   []string          `json:"required_outputs"`   // e.g. ["VerificationResult", "EvidenceSummary"]
	Budget            CoordinationBudget `json:"budget"`
	EscalationPolicy  EscalationPolicy  `json:"escalation_policy"`
	TimeoutSeconds    int               `json:"timeout_seconds"`
}

// CoordinationBudget defines resource limits for a multi-brain coordination.
type CoordinationBudget struct {
	MaxTotalTurns   int `json:"max_total_turns"`
	MaxTurnsPerBrain int `json:"max_turns_per_brain"`
	MaxParallelBrains int `json:"max_parallel_brains"`
}

// EscalationPolicy defines when and how to escalate during coordination.
type EscalationPolicy struct {
	OnUnsupportedCapability string `json:"on_unsupported_capability"` // "fallback" | "escalate" | "abort"
	OnPolicyDenied          string `json:"on_policy_denied"`
	OnVerificationConflict  string `json:"on_verification_conflict"`
	OnEnvironmentUnavailable string `json:"on_environment_unavailable"`
	MaxRetryPerBrain        int    `json:"max_retry_per_brain"`
}
