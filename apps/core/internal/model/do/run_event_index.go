// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RunEventIndex is the golang structure of table run_event_index for DAO operations like Where/Data.
type RunEventIndex struct {
	g.Meta       `orm:"table:run_event_index, do:true"`
	Id           any //
	ProjectId    any //
	RunBindingId any //
	SequenceNo   any //
	EventType    any //
	EventLevel   any //
	Summary      any //
	PayloadJson  any //
	CreatedAt    any //
}
