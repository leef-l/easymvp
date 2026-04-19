// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AuditLogs is the golang structure for table audit_logs.
type AuditLogs struct {
	Id          string `json:"id"          orm:"id"           ` //
	ProjectId   string `json:"projectId"   orm:"project_id"   ` //
	EventType   string `json:"eventType"   orm:"event_type"   ` //
	ActorKind   string `json:"actorKind"   orm:"actor_kind"   ` //
	Summary     string `json:"summary"     orm:"summary"      ` //
	PayloadJson string `json:"payloadJson" orm:"payload_json" ` //
	CreatedAt   string `json:"createdAt"   orm:"created_at"   ` //
}
