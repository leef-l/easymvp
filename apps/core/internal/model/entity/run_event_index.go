// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// RunEventIndex is the golang structure for table run_event_index.
type RunEventIndex struct {
	Id           string `json:"id"           orm:"id"             ` //
	ProjectId    string `json:"projectId"    orm:"project_id"     ` //
	RunBindingId string `json:"runBindingId" orm:"run_binding_id" ` //
	SequenceNo   int    `json:"sequenceNo"   orm:"sequence_no"    ` //
	EventType    string `json:"eventType"    orm:"event_type"     ` //
	EventLevel   string `json:"eventLevel"   orm:"event_level"    ` //
	Summary      string `json:"summary"      orm:"summary"        ` //
	PayloadJson  string `json:"payloadJson"  orm:"payload_json"   ` //
	CreatedAt    string `json:"createdAt"    orm:"created_at"     ` //
}
