// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Projects is the golang structure for table projects.
type Projects struct {
	Id                    string `json:"id"                    orm:"id"                       ` //
	Name                  string `json:"name"                  orm:"name"                     ` //
	ProjectCategory       string `json:"projectCategory"       orm:"project_category"         ` //
	GoalSummary           string `json:"goalSummary"           orm:"goal_summary"             ` //
	Status                string `json:"status"                orm:"status"                   ` //
	ProductionStatus      string `json:"productionStatus"      orm:"production_status"        ` //
	WorkspaceRoot         string `json:"workspaceRoot"         orm:"workspace_root"           ` //
	RepoRoot              string `json:"repoRoot"              orm:"repo_root"                ` //
	CurrentPlanDraftId    string `json:"currentPlanDraftId"    orm:"current_plan_draft_id"    ` //
	CurrentCompiledPlanId string `json:"currentCompiledPlanId" orm:"current_compiled_plan_id" ` //
	CreatedAt             string `json:"createdAt"             orm:"created_at"               ` //
	UpdatedAt             string `json:"updatedAt"             orm:"updated_at"               ` //
}
