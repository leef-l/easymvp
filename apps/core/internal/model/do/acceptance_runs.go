// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AcceptanceRuns is the golang structure of table acceptance_runs for DAO operations like Where/Data.
type AcceptanceRuns struct {
	g.Meta                `orm:"table:acceptance_runs, do:true"`
	Id                    any //
	ProjectId             any //
	TaskId                any //
	ProfileVersion        any //
	Status                any //
	FunctionalStatus      any //
	ProductionStatus      any //
	ManualReleaseRequired any //
	BrowserRunID          any //
	VerifierRunID         any //
	ValidationResultsJSON any //
	CreatedAt             any //
	FinishedAt            any //
}
