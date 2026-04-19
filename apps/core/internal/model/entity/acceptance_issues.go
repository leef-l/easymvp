// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AcceptanceIssues is the golang structure for table acceptance_issues.
type AcceptanceIssues struct {
	Id              string `json:"id"              orm:"id"                ` //
	ProjectId       string `json:"projectId"       orm:"project_id"        ` //
	AcceptanceRunId string `json:"acceptanceRunId" orm:"acceptance_run_id" ` //
	Severity        string `json:"severity"        orm:"severity"          ` //
	IssueKind       string `json:"issueKind"       orm:"issue_kind"        ` //
	Blocking        int    `json:"blocking"        orm:"blocking"          ` //
	Summary         string `json:"summary"         orm:"summary"           ` //
	DetailJson      string `json:"detailJson"      orm:"detail_json"       ` //
	CreatedAt       string `json:"createdAt"       orm:"created_at"        ` //
}
