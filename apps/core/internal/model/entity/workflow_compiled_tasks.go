// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WorkflowCompiledTasks is the golang structure for table workflow_compiled_tasks.
type WorkflowCompiledTasks struct {
	Id                       string `json:"id"                       orm:"id"                         ` //
	CompiledPlanId           string `json:"compiledPlanId"           orm:"compiled_plan_id"           ` //
	TaskKey                  string `json:"taskKey"                  orm:"task_key"                   ` //
	Name                     string `json:"name"                     orm:"name"                       ` //
	Description              string `json:"description"              orm:"description"                ` //
	Phase                    string `json:"phase"                    orm:"phase"                      ` //
	TaskKind                 string `json:"taskKind"                 orm:"task_kind"                  ` //
	RoleType                 string `json:"roleType"                 orm:"role_type"                  ` //
	BrainKind                string `json:"brainKind"                orm:"brain_kind"                 ` //
	RiskLevel                string `json:"riskLevel"                orm:"risk_level"                 ` //
	AffectedResourcesJson    string `json:"affectedResourcesJson"    orm:"affected_resources_json"    ` //
	DeliveryContractJson     string `json:"deliveryContractJson"     orm:"delivery_contract_json"     ` //
	VerificationContractJson string `json:"verificationContractJson" orm:"verification_contract_json" ` //
	ManualReviewRequired     int    `json:"manualReviewRequired"     orm:"manual_review_required"     ` //
	DependsOnTaskKeysJson    string `json:"dependsOnTaskKeysJson"    orm:"depends_on_task_keys_json"  ` //
	Status                   string `json:"status"                   orm:"status"                     ` //
	CreatedAt                string `json:"createdAt"                orm:"created_at"                 ` //
	UpdatedAt                string `json:"updatedAt"                orm:"updated_at"                 ` //
}
