// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SolutionDesigns is the golang structure for table solution_designs.
type SolutionDesigns struct {
	Id             string `json:"id"             orm:"id"               ` //
	ProjectId      string `json:"projectId"      orm:"project_id"       ` //
	RequirementId  string `json:"requirementId"  orm:"requirement_id"   ` //
	Version        int    `json:"version"        orm:"version"          ` //
	Status         string `json:"status"         orm:"status"           ` //
	Architecture   string `json:"architecture"   orm:"architecture"     ` //
	ModulesJson    string `json:"modulesJson"    orm:"modules_json"     ` //
	DataModelsJson string `json:"dataModelsJson" orm:"data_models_json" ` //
	PagesJson      string `json:"pagesJson"      orm:"pages_json"       ` //
	TaskDraftsJson string `json:"taskDraftsJson" orm:"task_drafts_json" ` //
	UserConfirmed  int    `json:"userConfirmed"  orm:"user_confirmed"   ` //
	ConfirmedAt    string `json:"confirmedAt"    orm:"confirmed_at"     ` //
	CreatedAt      string `json:"createdAt"      orm:"created_at"       ` //
	UpdatedAt      string `json:"updatedAt"      orm:"updated_at"       ` //
}
