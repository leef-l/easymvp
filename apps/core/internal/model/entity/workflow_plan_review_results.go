// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WorkflowPlanReviewResults is the golang structure for table workflow_plan_review_results.
type WorkflowPlanReviewResults struct {
	Id                      string `json:"id"                      orm:"id"                        ` //
	ProjectId               string `json:"projectId"               orm:"project_id"                ` //
	PlanDraftId             string `json:"planDraftId"             orm:"plan_draft_id"             ` //
	ReviewVersion           int    `json:"reviewVersion"           orm:"review_version"            ` //
	ReviewRunId             string `json:"reviewRunId"             orm:"review_run_id"             ` //
	Decision                string `json:"decision"                orm:"decision"                  ` //
	BlockingIssueCount      int    `json:"blockingIssueCount"      orm:"blocking_issue_count"      ` //
	AdvisoryIssueCount      int    `json:"advisoryIssueCount"      orm:"advisory_issue_count"      ` //
	IssuesJson              string `json:"issuesJson"              orm:"issues_json"               ` //
	SplitSuggestionsJson    string `json:"splitSuggestionsJson"    orm:"split_suggestions_json"    ` //
	OverrideSuggestionsJson string `json:"overrideSuggestionsJson" orm:"override_suggestions_json" ` //
	Status                  string `json:"status"                  orm:"status"                    ` //
	ReviewedAt              string `json:"reviewedAt"              orm:"reviewed_at"               ` //
}
