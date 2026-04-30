// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DesignReviews is the golang structure of table design_reviews for DAO operations like Where/Data.
type DesignReviews struct {
	g.Meta          `orm:"table:design_reviews, do:true"`
	Id              any //
	DesignId        any //
	ProjectId       any //
	Round           any //
	Passed          any //
	Score           any //
	DimensionsJson  any //
	IssuesJson      any //
	SuggestionsJson any //
	FixTasksJson    any //
	BrainRunId      any //
	CreatedAt       any //
}
