// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Requirements is the golang structure of table requirements for DAO operations like Where/Data.
type Requirements struct {
	g.Meta             `orm:"table:requirements, do:true"`
	Id                 any //
	ProjectId          any //
	RawInput           any //
	Status             any //
	RequirementDocJson any //
	UserConfirmed      any //
	ConfirmedAt        any //
	CreatedAt          any //
	UpdatedAt          any //
}
