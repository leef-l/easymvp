// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// DomainTasks is the golang structure for table domain_tasks.
type DomainTasks struct {
	Id                   string `json:"id"                   orm:"id"                      ` //
	ProjectId            string `json:"projectId"            orm:"project_id"              ` //
	SourceCompiledPlanId string `json:"sourceCompiledPlanId" orm:"source_compiled_plan_id" ` //
	SourceCompiledTaskId string `json:"sourceCompiledTaskId" orm:"source_compiled_task_id" ` //
	SourceTaskKey        string `json:"sourceTaskKey"        orm:"source_task_key"         ` //
	CompiledVersion      int    `json:"compiledVersion"      orm:"compiled_version"        ` //
	Name                 string `json:"name"                 orm:"name"                    ` //
	Phase                string `json:"phase"                orm:"phase"                   ` //
	TaskKind             string `json:"taskKind"             orm:"task_kind"               ` //
	RoleType             string `json:"roleType"             orm:"role_type"               ` //
	BrainKind            string `json:"brainKind"            orm:"brain_kind"              ` //
	RiskLevel            string `json:"riskLevel"            orm:"risk_level"              ` //
	Status               string `json:"status"               orm:"status"                  ` //
	ManualReviewRequired int    `json:"manualReviewRequired" orm:"manual_review_required"  ` //
	CreatedAt            string `json:"createdAt"            orm:"created_at"              ` //
	UpdatedAt            string `json:"updatedAt"            orm:"updated_at"              ` //
}
