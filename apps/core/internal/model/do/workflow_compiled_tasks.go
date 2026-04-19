// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowCompiledTasks is the golang structure of table workflow_compiled_tasks for DAO operations like Where/Data.
type WorkflowCompiledTasks struct {
	g.Meta                   `orm:"table:workflow_compiled_tasks, do:true"`
	Id                       any //
	CompiledPlanId           any //
	TaskKey                  any //
	Name                     any //
	Description              any //
	Phase                    any //
	TaskKind                 any //
	RoleType                 any //
	BrainKind                any //
	RiskLevel                any //
	AffectedResourcesJson    any //
	DeliveryContractJson     any //
	VerificationContractJson any //
	ManualReviewRequired     any //
	DependsOnTaskKeysJson    any //
	Status                   any //
	CreatedAt                any //
	UpdatedAt                any //
}
