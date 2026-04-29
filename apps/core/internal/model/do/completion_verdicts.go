// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CompletionVerdicts is the golang structure of table completion_verdicts for DAO operations like Where/Data.
type CompletionVerdicts struct {
	g.Meta                 `orm:"table:completion_verdicts, do:true"`
	Id                     any //
	ProjectId              any //
	AcceptanceRunId        any //
	Decision               any //
	FinalStatus            any //
	Reason                 any //
	Summary                any //
	NextAction             any //
	BlockerCount           any //
	ReleaseReady           any //
	Completed              any //
	ManualReviewRequired   any //
	ManualReleaseRequired  any //
	ManualReleaseCompleted any //
	SourceRunId            any //
	CreatedAt              any //
	UpdatedAt              any //
	ExecutorSucceeded      any //
	DeliveryVerified       any //
	AcceptancePassed       any //
}
