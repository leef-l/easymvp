// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RepairPlanDrafts is the golang structure of table repair_plan_drafts for DAO operations like Where/Data.
type RepairPlanDrafts struct {
	g.Meta                  `orm:"table:repair_plan_drafts, do:true"`
	Id                      any //
	ProjectId               any //
	FailedTaskContextJson   any //
	FailureReasonJson       any //
	OriginalContractsJson   any //
	RuntimeSummaryJson      any //
	ArtifactRefsJson        any //
	RepairPlanJson          any //
	RepairReasoningSummary  any //
	ReplacedConstraintsJson any //
	Status                  any //
	CreatedBy               any //
	CreatedAt               any //
	UpdatedAt               any //
}
