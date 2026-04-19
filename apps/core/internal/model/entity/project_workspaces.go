// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProjectWorkspaces is the golang structure for table project_workspaces.
type ProjectWorkspaces struct {
	Id              string `json:"id"              orm:"id"               ` //
	ProjectId       string `json:"projectId"       orm:"project_id"       ` //
	WorkspaceRoot   string `json:"workspaceRoot"   orm:"workspace_root"   ` //
	EvidenceRoot    string `json:"evidenceRoot"    orm:"evidence_root"    ` //
	RunsRoot        string `json:"runsRoot"        orm:"runs_root"        ` //
	ReplayRoot      string `json:"replayRoot"      orm:"replay_root"      ` //
	DiagnosticsRoot string `json:"diagnosticsRoot" orm:"diagnostics_root" ` //
	CreatedAt       string `json:"createdAt"       orm:"created_at"       ` //
	UpdatedAt       string `json:"updatedAt"       orm:"updated_at"       ` //
}
