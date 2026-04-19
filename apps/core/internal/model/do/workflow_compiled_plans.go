// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowCompiledPlans is the golang structure of table workflow_compiled_plans for DAO operations like Where/Data.
type WorkflowCompiledPlans struct {
	g.Meta             `orm:"table:workflow_compiled_plans, do:true"`
	Id                 any //
	ProjectId          any //
	PlanDraftId        any //
	PlanReviewResultId any //
	CompiledVersion    any //
	CompileRunId       any //
	ProjectCategory    any //
	Status             any //
	RiskSummaryJson    any //
	CompileDiffJson    any //
	GeneratedAt        any //
	ActivatedAt        any //
}
