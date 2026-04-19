// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AcceptanceIssues is the golang structure of table acceptance_issues for DAO operations like Where/Data.
type AcceptanceIssues struct {
	g.Meta          `orm:"table:acceptance_issues, do:true"`
	Id              any //
	ProjectId       any //
	AcceptanceRunId any //
	Severity        any //
	IssueKind       any //
	Blocking        any //
	Summary         any //
	DetailJson      any //
	CreatedAt       any //
}
