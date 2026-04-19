// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectWorkspaces is the golang structure of table project_workspaces for DAO operations like Where/Data.
type ProjectWorkspaces struct {
	g.Meta          `orm:"table:project_workspaces, do:true"`
	Id              any //
	ProjectId       any //
	WorkspaceRoot   any //
	EvidenceRoot    any //
	RunsRoot        any //
	ReplayRoot      any //
	DiagnosticsRoot any //
	CreatedAt       any //
	UpdatedAt       any //
}
