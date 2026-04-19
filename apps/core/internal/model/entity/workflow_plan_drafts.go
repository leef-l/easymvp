// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WorkflowPlanDrafts is the golang structure for table workflow_plan_drafts.
type WorkflowPlanDrafts struct {
	Id                    string `json:"id"                    orm:"id"                      ` //
	ProjectId             string `json:"projectId"             orm:"project_id"              ` //
	Version               int    `json:"version"               orm:"version"                 ` //
	SourceKind            string `json:"sourceKind"            orm:"source_kind"             ` //
	SourceRunId           string `json:"sourceRunId"           orm:"source_run_id"           ` //
	ProjectCategory       string `json:"projectCategory"       orm:"project_category"        ` //
	GoalSummary           string `json:"goalSummary"           orm:"goal_summary"            ` //
	InputRequirementsJson string `json:"inputRequirementsJson" orm:"input_requirements_json" ` //
	DraftTasksJson        string `json:"draftTasksJson"        orm:"draft_tasks_json"        ` //
	Status                string `json:"status"                orm:"status"                  ` //
	CreatedBy             string `json:"createdBy"             orm:"created_by"              ` //
	CreatedAt             string `json:"createdAt"             orm:"created_at"              ` //
	UpdatedAt             string `json:"updatedAt"             orm:"updated_at"              ` //
}
