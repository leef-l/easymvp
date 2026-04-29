package braincontracts

import "encoding/json"

type RepairDesignInput struct {
	FailedTaskContextJSON json.RawMessage `json:"failed_task_context_json"`
	FailureReasonJSON     json.RawMessage `json:"failure_reason_json"`
	OriginalContractsJSON json.RawMessage `json:"original_contracts_json"`
	RuntimeSummaryJSON    json.RawMessage `json:"runtime_summary_json"`
	ArtifactRefs          []ArtifactRef   `json:"artifact_refs,omitempty"`
}

type RepairDesignResult struct {
	RepairPlanDraftID         string          `json:"repair_plan_draft_id"`
	RepairPlanJSON            json.RawMessage `json:"repair_plan_json"`
	RepairReasoningSummary    string          `json:"repair_reasoning_summary"`
	ReplacedConstraints       []string        `json:"replaced_constraints,omitempty"`
	ReasonClass               string          `json:"reason_class"`               // execution_error | verification_failure | delivery_mismatch | policy_violation | environment_failure
	RepairStrategy            string          `json:"repair_strategy"`            // retry | redesign | replace | escalate | manual_checkpoint
	UpdatedTasks              []RepairTaskItem `json:"updated_tasks"`
	VerificationAdjustments   []string        `json:"verification_adjustments"`
	DeliveryAdjustments       []string        `json:"delivery_adjustments"`
	HumanCheckpointRequired   bool            `json:"human_checkpoint_required"`
}

type RepairTaskItem struct {
	TaskKey   string `json:"task_key"`
	Name      string `json:"name"`
	Summary   string `json:"summary"`
	BrainKind string `json:"brain_kind"`
}
