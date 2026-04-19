// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// EvidenceLinks is the golang structure for table evidence_links.
type EvidenceLinks struct {
	Id               string `json:"id"               orm:"id"                 ` //
	ProjectId        string `json:"projectId"        orm:"project_id"         ` //
	EvidenceItemId   string `json:"evidenceItemId"   orm:"evidence_item_id"   ` //
	LinkedObjectType string `json:"linkedObjectType" orm:"linked_object_type" ` //
	LinkedObjectId   string `json:"linkedObjectId"   orm:"linked_object_id"   ` //
	CreatedAt        string `json:"createdAt"        orm:"created_at"         ` //
}
