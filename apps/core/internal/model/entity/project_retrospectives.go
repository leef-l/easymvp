// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProjectRetrospectives is the golang structure for table project_retrospectives.
type ProjectRetrospectives struct {
	Id                 string  `json:"id"                 orm:"id"                   ` //
	ProjectId          string  `json:"projectId"          orm:"project_id"            ` //
	PlanVsActualJson   string  `json:"planVsActualJson"   orm:"plan_vs_actual_json"   ` //
	SuccessFactorsJson string  `json:"successFactorsJson" orm:"success_factors_json"  ` //
	FailureLessonsJson string  `json:"failureLessonsJson" orm:"failure_lessons_json"  ` //
	PatternsJson       string  `json:"patternsJson"       orm:"patterns_json"         ` //
	TotalTasks         int     `json:"totalTasks"         orm:"total_tasks"           ` //
	CompletedTasks     int     `json:"completedTasks"     orm:"completed_tasks"       ` //
	FailedTasks        int     `json:"failedTasks"        orm:"failed_tasks"          ` //
	RetriedTasks       int     `json:"retriedTasks"       orm:"retried_tasks"         ` //
	TotalTurns         int     `json:"totalTurns"         orm:"total_turns"           ` //
	TotalTokens        int     `json:"totalTokens"        orm:"total_tokens"          ` //
	TotalCostUsd       float64 `json:"totalCostUsd"       orm:"total_cost_usd"        ` //
	DurationSeconds    int     `json:"durationSeconds"    orm:"duration_seconds"      ` //
	ReviewRounds       int     `json:"reviewRounds"       orm:"review_rounds"         ` //
	BrainsUsedJson     string  `json:"brainsUsedJson"     orm:"brains_used_json"      ` //
	CreatedAt          string  `json:"createdAt"          orm:"created_at"            ` //
}
