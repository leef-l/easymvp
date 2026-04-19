// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReplayItems is the golang structure of table replay_items for DAO operations like Where/Data.
type ReplayItems struct {
	g.Meta     `orm:"table:replay_items, do:true"`
	Id         any //
	ProjectId  any //
	RunId      any //
	ReplayType any //
	FilePath   any //
	Summary    any //
	CreatedAt  any //
}
