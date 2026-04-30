// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectRetrospectives is the golang structure of table project_retrospectives for DAO operations like Where/Data.
type ProjectRetrospectives struct {
	g.Meta             `orm:"table:project_retrospectives, do:true"`
	Id                 any //
	ProjectId          any //
	PlanVsActualJson   any //
	SuccessFactorsJson any //
	FailureLessonsJson any //
	PatternsJson       any //
	TotalTasks         any //
	CompletedTasks     any //
	FailedTasks        any //
	RetriedTasks       any //
	TotalTurns         any //
	TotalTokens        any //
	TotalCostUsd       any //
	DurationSeconds    any //
	ReviewRounds       any //
	BrainsUsedJson     any //
	CreatedAt          any //
}
