// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProjectDeliveries is the golang structure for table project_deliveries.
type ProjectDeliveries struct {
	Id              string `json:"id"              orm:"id"                ` //
	ProjectId       string `json:"projectId"       orm:"project_id"        ` //
	Status          string `json:"status"           orm:"status"            ` //
	WorkspacePath   string `json:"workspacePath"   orm:"workspace_path"    ` //
	Readme          string `json:"readme"           orm:"readme"            ` //
	ArchitectureDoc string `json:"architectureDoc" orm:"architecture_doc"  ` //
	ApiDocs         string `json:"apiDocs"          orm:"api_docs"          ` //
	DeploymentDoc   string `json:"deploymentDoc"   orm:"deployment_doc"    ` //
	TestReportJson  string `json:"testReportJson"  orm:"test_report_json"  ` //
	StatisticsJson  string `json:"statisticsJson"  orm:"statistics_json"   ` //
	UserAccepted    int    `json:"userAccepted"    orm:"user_accepted"     ` //
	AcceptedAt      string `json:"acceptedAt"      orm:"accepted_at"       ` //
	DeliveredAt     string `json:"deliveredAt"     orm:"delivered_at"      ` //
	CreatedAt       string `json:"createdAt"       orm:"created_at"        ` //
	UpdatedAt       string `json:"updatedAt"       orm:"updated_at"        ` //
}
