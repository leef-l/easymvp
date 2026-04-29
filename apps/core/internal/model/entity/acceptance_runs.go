// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AcceptanceRuns is the golang structure for table acceptance_runs.
type AcceptanceRuns struct {
	Id                    string `json:"id"                    orm:"id"                      ` //
	ProjectId             string `json:"projectId"             orm:"project_id"              ` //
	TaskId                string `json:"taskId"                orm:"task_id"                 ` //
	ProfileVersion        string `json:"profileVersion"        orm:"profile_version"         ` //
	Status                string `json:"status"                orm:"status"                  ` //
	FunctionalStatus      string `json:"functionalStatus"      orm:"functional_status"       ` //
	ProductionStatus      string `json:"productionStatus"      orm:"production_status"       ` //
	ManualReleaseRequired int    `json:"manualReleaseRequired" orm:"manual_release_required" ` //
	BrowserRunID          string `json:"browserRunId"          orm:"browser_run_id"          ` //
	VerifierRunID         string `json:"verifierRunId"         orm:"verifier_run_id"         ` //
	ValidationResultsJSON string `json:"validationResultsJson" orm:"validation_results_json" ` //
	CreatedAt             string `json:"createdAt"             orm:"created_at"              ` //
	FinishedAt            string `json:"finishedAt"            orm:"finished_at"             ` //
}
