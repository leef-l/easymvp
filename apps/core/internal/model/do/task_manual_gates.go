// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TaskManualGates is the golang structure of table task_manual_gates for DAO operations like Where/Data.
type TaskManualGates struct {
	g.Meta     `orm:"table:task_manual_gates, do:true"`
	Id         any //
	ProjectId  any //
	TaskId     any //
	GateKind   any //
	GateStatus any //
	Comment    any //
	CreatedAt  any //
	UpdatedAt  any //
}
