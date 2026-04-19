// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SchemaMigrations is the golang structure for table schema_migrations.
type SchemaMigrations struct {
	Version      int    `json:"version"      orm:"version"       ` //
	Name         string `json:"name"         orm:"name"          ` //
	Checksum     string `json:"checksum"     orm:"checksum"      ` //
	AppliedAt    string `json:"appliedAt"    orm:"applied_at"    ` //
	DurationMs   int    `json:"durationMs"   orm:"duration_ms"   ` //
	Status       string `json:"status"       orm:"status"        ` //
	ErrorMessage string `json:"errorMessage" orm:"error_message" ` //
}
