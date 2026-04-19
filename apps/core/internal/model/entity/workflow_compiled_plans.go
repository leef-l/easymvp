// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WorkflowCompiledPlans is the golang structure for table workflow_compiled_plans.
type WorkflowCompiledPlans struct {
	Id                 string `json:"id"                 orm:"id"                    ` //
	ProjectId          string `json:"projectId"          orm:"project_id"            ` //
	PlanDraftId        string `json:"planDraftId"        orm:"plan_draft_id"         ` //
	PlanReviewResultId string `json:"planReviewResultId" orm:"plan_review_result_id" ` //
	CompiledVersion    int    `json:"compiledVersion"    orm:"compiled_version"      ` //
	CompileRunId       string `json:"compileRunId"       orm:"compile_run_id"        ` //
	ProjectCategory    string `json:"projectCategory"    orm:"project_category"      ` //
	Status             string `json:"status"             orm:"status"                ` //
	RiskSummaryJson    string `json:"riskSummaryJson"    orm:"risk_summary_json"     ` //
	CompileDiffJson    string `json:"compileDiffJson"    orm:"compile_diff_json"     ` //
	GeneratedAt        string `json:"generatedAt"        orm:"generated_at"          ` //
	ActivatedAt        string `json:"activatedAt"        orm:"activated_at"          ` //
}
