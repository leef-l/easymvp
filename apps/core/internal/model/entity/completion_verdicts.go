// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// CompletionVerdicts is the golang structure for table completion_verdicts.
type CompletionVerdicts struct {
	Id                     string `json:"id"                       orm:"id"                       ` //
	ProjectId              string `json:"project_id"               orm:"project_id"               ` //
	AcceptanceRunId        string `json:"acceptance_run_id"        orm:"acceptance_run_id"        ` //
	Decision               string `json:"decision"                 orm:"decision"                 ` //
	FinalStatus            string `json:"final_status"             orm:"final_status"             ` //
	Reason                 string `json:"reason"                   orm:"reason"                   ` //
	Summary                string `json:"summary"                  orm:"summary"                  ` //
	NextAction             string `json:"next_action"              orm:"next_action"              ` //
	BlockerCount           int    `json:"blocker_count"            orm:"blocker_count"            ` //
	ReleaseReady           int    `json:"release_ready"            orm:"release_ready"            ` //
	Completed              int    `json:"completed"                orm:"completed"                ` //
	ManualReviewRequired   int    `json:"manual_review_required"   orm:"manual_review_required"   ` //
	ManualReleaseRequired  int    `json:"manual_release_required"  orm:"manual_release_required"  ` //
	ManualReleaseCompleted int    `json:"manual_release_completed" orm:"manual_release_completed" ` //
	SourceRunId            string `json:"source_run_id"            orm:"source_run_id"            ` //
	CreatedAt              string `json:"created_at"               orm:"created_at"               ` //
	UpdatedAt              string `json:"updated_at"               orm:"updated_at"               ` //
	ExecutorSucceeded      int    `json:"executor_succeeded"       orm:"executor_succeeded"       ` //
	DeliveryVerified       int    `json:"delivery_verified"        orm:"delivery_verified"        ` //
	AcceptancePassed       int    `json:"acceptance_passed"        orm:"acceptance_passed"        ` //
}
