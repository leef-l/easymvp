// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowPlanReviewResults is the golang structure of table workflow_plan_review_results for DAO operations like Where/Data.
type WorkflowPlanReviewResults struct {
	g.Meta                  `orm:"table:workflow_plan_review_results, do:true"`
	Id                      any //
	ProjectId               any //
	PlanDraftId             any //
	ReviewVersion           any //
	ReviewRunId             any //
	Decision                any //
	BlockingIssueCount      any //
	AdvisoryIssueCount      any //
	IssuesJson              any //
	SplitSuggestionsJson    any //
	OverrideSuggestionsJson any //
	Status                  any //
	ReviewedAt              any //
}
