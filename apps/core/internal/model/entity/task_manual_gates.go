// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TaskManualGates is the golang structure for table task_manual_gates.
type TaskManualGates struct {
	Id         string `json:"id"         orm:"id"          ` //
	ProjectId  string `json:"projectId"  orm:"project_id"  ` //
	TaskId     string `json:"taskId"     orm:"task_id"     ` //
	GateKind   string `json:"gateKind"   orm:"gate_kind"   ` //
	GateStatus string `json:"gateStatus" orm:"gate_status" ` //
	Comment    string `json:"comment"    orm:"comment"     ` //
	CreatedAt  string `json:"createdAt"  orm:"created_at"  ` //
	UpdatedAt  string `json:"updatedAt"  orm:"updated_at"  ` //
}
