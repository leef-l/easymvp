// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SolutionDesigns is the golang structure of table solution_designs for DAO operations like Where/Data.
type SolutionDesigns struct {
	g.Meta         `orm:"table:solution_designs, do:true"`
	Id             any //
	ProjectId      any //
	RequirementId  any //
	Version        any //
	Status         any //
	Architecture   any //
	ModulesJson    any //
	DataModelsJson any //
	PagesJson      any //
	TaskDraftsJson any //
	UserConfirmed  any //
	ConfirmedAt    any //
	CreatedAt      any //
	UpdatedAt      any //
}
