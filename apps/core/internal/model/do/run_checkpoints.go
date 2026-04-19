// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RunCheckpoints is the golang structure of table run_checkpoints for DAO operations like Where/Data.
type RunCheckpoints struct {
	g.Meta         `orm:"table:run_checkpoints, do:true"`
	Id             any //
	RunBindingId   any //
	CheckpointType any //
	PayloadJson    any //
	CreatedAt      any //
}
