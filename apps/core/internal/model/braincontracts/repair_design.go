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
	RepairPlanDraftID      string          `json:"repair_plan_draft_id"`
	RepairPlanJSON         json.RawMessage `json:"repair_plan_json"`
	RepairReasoningSummary string          `json:"repair_reasoning_summary"`
	ReplacedConstraints    []string        `json:"replaced_constraints,omitempty"`
}
