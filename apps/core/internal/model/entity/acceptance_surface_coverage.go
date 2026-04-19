// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AcceptanceSurfaceCoverage is the golang structure for table acceptance_surface_coverage.
type AcceptanceSurfaceCoverage struct {
	Id              string `json:"id"              orm:"id"                ` //
	ProjectId       string `json:"projectId"       orm:"project_id"        ` //
	AcceptanceRunId string `json:"acceptanceRunId" orm:"acceptance_run_id" ` //
	Surface         string `json:"surface"         orm:"surface"           ` //
	CoverageStatus  string `json:"coverageStatus"  orm:"coverage_status"   ` //
	EvidenceCount   int    `json:"evidenceCount"   orm:"evidence_count"    ` //
	CreatedAt       string `json:"createdAt"       orm:"created_at"        ` //
	UpdatedAt       string `json:"updatedAt"       orm:"updated_at"        ` //
}
