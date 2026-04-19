// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// DiagnosticRecords is the golang structure for table diagnostic_records.
type DiagnosticRecords struct {
	Id         string `json:"id"         orm:"id"          ` //
	Scope      string `json:"scope"      orm:"scope"       ` //
	Severity   string `json:"severity"   orm:"severity"    ` //
	ErrorCode  string `json:"errorCode"  orm:"error_code"  ` //
	Summary    string `json:"summary"    orm:"summary"     ` //
	DetailJson string `json:"detailJson" orm:"detail_json" ` //
	CreatedAt  string `json:"createdAt"  orm:"created_at"  ` //
}
