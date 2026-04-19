// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectSnapshots is the golang structure of table project_snapshots for DAO operations like Where/Data.
type ProjectSnapshots struct {
	g.Meta       `orm:"table:project_snapshots, do:true"`
	ProjectId    any //
	SnapshotJson any //
	GeneratedAt  any //
}
