// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AcceptanceJourneyCoverage is the golang structure of table acceptance_journey_coverage for DAO operations like Where/Data.
type AcceptanceJourneyCoverage struct {
	g.Meta          `orm:"table:acceptance_journey_coverage, do:true"`
	Id              any //
	ProjectId       any //
	AcceptanceRunId any //
	Journey         any //
	CoverageStatus  any //
	EvidenceCount   any //
	CreatedAt       any //
	UpdatedAt       any //
}
