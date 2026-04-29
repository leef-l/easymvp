// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// RepairPlanDrafts is the golang structure for table repair_plan_drafts.
type RepairPlanDrafts struct {
	Id                      string `json:"id"                        orm:"id"                        ` //
	ProjectId               string `json:"project_id"                orm:"project_id"                ` //
	FailedTaskContextJson   string `json:"failed_task_context_json"  orm:"failed_task_context_json"  ` //
	FailureReasonJson       string `json:"failure_reason_json"       orm:"failure_reason_json"       ` //
	OriginalContractsJson   string `json:"original_contracts_json"   orm:"original_contracts_json"   ` //
	RuntimeSummaryJson      string `json:"runtime_summary_json"      orm:"runtime_summary_json"      ` //
	ArtifactRefsJson        string `json:"artifact_refs_json"        orm:"artifact_refs_json"        ` //
	RepairPlanJson          string `json:"repair_plan_json"          orm:"repair_plan_json"          ` //
	RepairReasoningSummary  string `json:"repair_reasoning_summary"  orm:"repair_reasoning_summary"  ` //
	ReplacedConstraintsJson string `json:"replaced_constraints_json" orm:"replaced_constraints_json" ` //
	Status                  string `json:"status"                    orm:"status"                    ` //
	CreatedBy               string `json:"created_by"                orm:"created_by"                ` //
	CreatedAt               string `json:"created_at"                orm:"created_at"                ` //
	UpdatedAt               string `json:"updated_at"                orm:"updated_at"                ` //
}
