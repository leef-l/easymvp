// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BrainRunBindings is the golang structure of table brain_run_bindings for DAO operations like Where/Data.
type BrainRunBindings struct {
	g.Meta     `orm:"table:brain_run_bindings, do:true"`
	Id         any //
	ProjectId  any //
	TaskId     any //
	BrainKind  any //
	BrainRunId any //
	RunStatus  any //
	StartedAt  any //
	FinishedAt any //
	LastSyncAt any //
	CreatedAt  any //
	UpdatedAt  any //
}
