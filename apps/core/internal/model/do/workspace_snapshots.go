// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WorkspaceSnapshots is the golang structure of table workspace_snapshots for DAO operations like Where/Data.
type WorkspaceSnapshots struct {
	g.Meta       `orm:"table:workspace_snapshots, do:true"`
	Key          any //
	SnapshotJson any //
	GeneratedAt  any //
}
