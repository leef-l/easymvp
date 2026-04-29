package braincontracts

import "encoding/json"

// RuntimeTaskIntent is the normalized task intent sent from EasyMVP orchestrator
// to the runtime adapter layer. It maps Engineering Cybernetics ch.3 "input, output
// and transfer function" into a structured contract.
type RuntimeTaskIntent struct {
	TaskID              string          `json:"task_id"`
	ProjectID           string          `json:"project_id"`
	WorkflowRunID       string          `json:"workflow_run_id,omitempty"`
	Stage               string          `json:"stage"`
	BrainKind           string          `json:"brain_kind"`
	RoleType            string          `json:"role_type,omitempty"`
	Goal                string          `json:"goal"`
	InputSummary        string          `json:"input_summary,omitempty"`
	DeliveryContract    json.RawMessage `json:"delivery_contract,omitempty"`
	VerificationContract json.RawMessage `json:"verification_contract,omitempty"`
	ArtifactRefs        []ArtifactRef   `json:"artifact_refs,omitempty"`
	RiskLevel           string          `json:"risk_level,omitempty"`
	TimeoutPolicy       string          `json:"timeout_policy,omitempty"`
	ManualReviewRequired bool           `json:"manual_review_required,omitempty"`
}

// RunResult is the normalized runtime output consumed by the domain layer.
// It maps Engineering Cybernetics ch.4 "feedback servo system" into a
// structured feedback signal.
type RunResult struct {
	RunID         string          `json:"run_id"`
	TaskID        string          `json:"task_id,omitempty"`
	BrainKind     string          `json:"brain_kind,omitempty"`
	ExecutorBrain string          `json:"executor_brain,omitempty"`
	Status        string          `json:"status"` // completed | failed | unsupported | denied | cancelled | timeout
	StartedAt     string          `json:"started_at,omitempty"`
	EndedAt       string          `json:"ended_at,omitempty"`
	Summary       string          `json:"summary,omitempty"`
	ArtifactRefs  []ArtifactRef   `json:"artifact_refs,omitempty"`
	RuntimeFlags  json.RawMessage `json:"runtime_flags,omitempty"`
	RawEventRefs  []string        `json:"raw_event_refs,omitempty"`
}

// DeliveryResult is the normalized delivery output for tasks that generate artifacts.
// It maps Engineering Cybernetics ch.3 "system-level output" into a structured
// delivery report.
type DeliveryResult struct {
	TaskID            string          `json:"task_id"`
	DeliveryStatus    string          `json:"delivery_status"` // delivered | partially_delivered | not_delivered
	DeliveredArtifacts []string       `json:"delivered_artifacts,omitempty"`
	ChangedResources   []string       `json:"changed_resources,omitempty"`
	ContractSatisfied  bool           `json:"contract_satisfied"`
	DeliveryGaps      json.RawMessage `json:"delivery_gaps,omitempty"`
}
