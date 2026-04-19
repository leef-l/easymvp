// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ReplayItems is the golang structure for table replay_items.
type ReplayItems struct {
	Id         string `json:"id"         orm:"id"          ` //
	ProjectId  string `json:"projectId"  orm:"project_id"  ` //
	RunId      string `json:"runId"      orm:"run_id"      ` //
	ReplayType string `json:"replayType" orm:"replay_type" ` //
	FilePath   string `json:"filePath"   orm:"file_path"   ` //
	Summary    string `json:"summary"    orm:"summary"     ` //
	CreatedAt  string `json:"createdAt"  orm:"created_at"  ` //
}
