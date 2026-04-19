package braincontracts

import "encoding/json"

type PlanCompileInput struct {
	PlanDraftJSON        json.RawMessage `json:"plan_draft_json"`
	PlanReviewResultJSON json.RawMessage `json:"plan_review_result_json"`
	CategoryProfileJSON  json.RawMessage `json:"category_profile_json"`
	RoleContextJSON      json.RawMessage `json:"role_context_json"`
	ArtifactRefs         []ArtifactRef   `json:"artifact_refs,omitempty"`
}

type PlanCompileResult struct {
	CompiledPlanID  string             `json:"compiled_plan_id"`
	CompiledVersion int                `json:"compiled_version"`
	CompiledTasks   []CompiledTaskItem `json:"compiled_tasks"`
	RiskSummary     json.RawMessage    `json:"risk_summary"`
}

type CompiledTaskItem struct {
	CompiledTaskID       string          `json:"compiled_task_id"`
	Name                 string          `json:"name"`
	RoleType             string          `json:"role_type"`
	BrainKind            string          `json:"brain_kind"`
	DeliveryContract     json.RawMessage `json:"delivery_contract"`
	VerificationContract json.RawMessage `json:"verification_contract"`
	RiskLevel            string          `json:"risk_level"`
}
