// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AcceptanceJudgements is the golang structure of table acceptance_judgements for DAO operations like Where/Data.
type AcceptanceJudgements struct {
	g.Meta          `orm:"table:acceptance_judgements, do:true"`
	Id              any //
	ProjectId       any //
	AcceptanceRunId any //
	JudgementKind   any //
	JudgementResult any //
	Summary         any //
	DetailJson      any //
	CreatedAt       any //
}
