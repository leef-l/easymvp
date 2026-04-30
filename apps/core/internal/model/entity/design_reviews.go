// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// DesignReviews is the golang structure for table design_reviews.
type DesignReviews struct {
	Id              string `json:"id"              orm:"id"               ` //
	DesignId        string `json:"designId"        orm:"design_id"        ` //
	ProjectId       string `json:"projectId"       orm:"project_id"       ` //
	Round           int    `json:"round"           orm:"round"            ` //
	Passed          int    `json:"passed"          orm:"passed"           ` //
	Score           int    `json:"score"           orm:"score"            ` //
	DimensionsJson  string `json:"dimensionsJson"  orm:"dimensions_json"  ` //
	IssuesJson      string `json:"issuesJson"      orm:"issues_json"      ` //
	SuggestionsJson string `json:"suggestionsJson" orm:"suggestions_json" ` //
	FixTasksJson    string `json:"fixTasksJson"    orm:"fix_tasks_json"   ` //
	BrainRunId      string `json:"brainRunId"      orm:"brain_run_id"     ` //
	CreatedAt       string `json:"createdAt"       orm:"created_at"       ` //
}
