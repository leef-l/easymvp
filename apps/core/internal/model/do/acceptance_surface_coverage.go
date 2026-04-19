// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AcceptanceSurfaceCoverage is the golang structure of table acceptance_surface_coverage for DAO operations like Where/Data.
type AcceptanceSurfaceCoverage struct {
	g.Meta          `orm:"table:acceptance_surface_coverage, do:true"`
	Id              any //
	ProjectId       any //
	AcceptanceRunId any //
	Surface         any //
	CoverageStatus  any //
	EvidenceCount   any //
	CreatedAt       any //
	UpdatedAt       any //
}
