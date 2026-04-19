// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DiagnosticRecords is the golang structure of table diagnostic_records for DAO operations like Where/Data.
type DiagnosticRecords struct {
	g.Meta     `orm:"table:diagnostic_records, do:true"`
	Id         any //
	Scope      any //
	Severity   any //
	ErrorCode  any //
	Summary    any //
	DetailJson any //
	CreatedAt  any //
}
