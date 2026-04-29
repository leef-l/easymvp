package entity

// FaultSummaries is the golang structure for table fault_summaries.
type FaultSummaries struct {
	Id                 string `json:"id" orm:"id"`
	ProjectId          string `json:"project_id" orm:"project_id"`
	AcceptanceRunId    string `json:"acceptance_run_id" orm:"acceptance_run_id"`
	Status             string `json:"status" orm:"status"`
	BlockingIssueCount int    `json:"blocking_issue_count" orm:"blocking_issue_count"`
	AdvisoryIssueCount int    `json:"advisory_issue_count" orm:"advisory_issue_count"`
	TopIssue           string `json:"top_issue" orm:"top_issue"`
	FaultLoopDetected  int    `json:"fault_loop_detected" orm:"fault_loop_detected"`
	FaultKind          string `json:"fault_kind" orm:"fault_kind"`
	Severity           string `json:"severity" orm:"severity"`
	Summary            string `json:"summary" orm:"summary"`
	FailedChecksJson   string `json:"failed_checks_json" orm:"failed_checks_json"`
	AffectedTasksJson  string `json:"affected_tasks_json" orm:"affected_tasks_json"`
	CreatedAt          string `json:"created_at" orm:"created_at"`
	UpdatedAt          string `json:"updated_at" orm:"updated_at"`
}
