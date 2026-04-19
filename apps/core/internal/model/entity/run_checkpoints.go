// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// RunCheckpoints is the golang structure for table run_checkpoints.
type RunCheckpoints struct {
	Id             string `json:"id"             orm:"id"              ` //
	RunBindingId   string `json:"runBindingId"   orm:"run_binding_id"  ` //
	CheckpointType string `json:"checkpointType" orm:"checkpoint_type" ` //
	PayloadJson    string `json:"payloadJson"    orm:"payload_json"    ` //
	CreatedAt      string `json:"createdAt"      orm:"created_at"      ` //
}
