// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Projects is the golang structure of table projects for DAO operations like Where/Data.
type Projects struct {
	g.Meta                `orm:"table:projects, do:true"`
	Id                    any //
	Name                  any //
	ProjectCategory       any //
	GoalSummary           any //
	Status                any //
	ProductionStatus      any //
	WorkspaceRoot         any //
	RepoRoot              any //
	CurrentPlanDraftId    any //
	CurrentCompiledPlanId any //
	CreatedAt             any //
	UpdatedAt             any //
}
