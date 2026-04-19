// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// EvidenceItems is the golang structure for table evidence_items.
type EvidenceItems struct {
	Id           string `json:"id"           orm:"id"            ` //
	ProjectId    string `json:"projectId"    orm:"project_id"    ` //
	RunId        string `json:"runId"        orm:"run_id"        ` //
	Surface      string `json:"surface"      orm:"surface"       ` //
	Journey      string `json:"journey"      orm:"journey"       ` //
	EvidenceType string `json:"evidenceType" orm:"evidence_type" ` //
	FilePath     string `json:"filePath"     orm:"file_path"     ` //
	ContentHash  string `json:"contentHash"  orm:"content_hash"  ` //
	FileSize     int    `json:"fileSize"     orm:"file_size"     ` //
	CapturedAt   string `json:"capturedAt"   orm:"captured_at"   ` //
	CreatedAt    string `json:"createdAt"    orm:"created_at"    ` //
}
