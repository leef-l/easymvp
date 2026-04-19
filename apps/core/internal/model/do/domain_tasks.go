// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DomainTasks is the golang structure of table domain_tasks for DAO operations like Where/Data.
type DomainTasks struct {
	g.Meta               `orm:"table:domain_tasks, do:true"`
	Id                   any //
	ProjectId            any //
	SourceCompiledPlanId any //
	SourceCompiledTaskId any //
	SourceTaskKey        any //
	CompiledVersion      any //
	Name                 any //
	Phase                any //
	TaskKind             any //
	RoleType             any //
	BrainKind            any //
	RiskLevel            any //
	Status               any //
	ManualReviewRequired any //
	CreatedAt            any //
	UpdatedAt            any //
}
