// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BrainRunBindings is the golang structure for table brain_run_bindings.
type BrainRunBindings struct {
	Id         string `json:"id"         orm:"id"           ` //
	ProjectId  string `json:"projectId"  orm:"project_id"   ` //
	TaskId     string `json:"taskId"     orm:"task_id"      ` //
	BrainKind  string `json:"brainKind"  orm:"brain_kind"   ` //
	BrainRunId string `json:"brainRunId" orm:"brain_run_id" ` //
	RunStatus  string `json:"runStatus"  orm:"run_status"   ` //
	StartedAt  string `json:"startedAt"  orm:"started_at"   ` //
	FinishedAt string `json:"finishedAt" orm:"finished_at"  ` //
	LastSyncAt string `json:"lastSyncAt" orm:"last_sync_at" ` //
	CreatedAt  string `json:"createdAt"  orm:"created_at"   ` //
	UpdatedAt  string `json:"updatedAt"  orm:"updated_at"   ` //
}
