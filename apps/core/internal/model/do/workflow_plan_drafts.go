// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowPlanDrafts is the golang structure of table workflow_plan_drafts for DAO operations like Where/Data.
type WorkflowPlanDrafts struct {
	g.Meta                `orm:"table:workflow_plan_drafts, do:true"`
	Id                    any //
	ProjectId             any //
	Version               any //
	SourceKind            any //
	SourceRunId           any //
	ProjectCategory       any //
	GoalSummary           any //
	InputRequirementsJson any //
	DraftTasksJson        any //
	Status                any //
	CreatedBy             any //
	CreatedAt             any //
	UpdatedAt             any //
}
