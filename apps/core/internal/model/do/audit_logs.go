// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AuditLogs is the golang structure of table audit_logs for DAO operations like Where/Data.
type AuditLogs struct {
	g.Meta      `orm:"table:audit_logs, do:true"`
	Id          any //
	ProjectId   any //
	EventType   any //
	ActorKind   any //
	Summary     any //
	PayloadJson any //
	CreatedAt   any //
}
