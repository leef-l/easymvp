// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// EvidenceLinks is the golang structure of table evidence_links for DAO operations like Where/Data.
type EvidenceLinks struct {
	g.Meta           `orm:"table:evidence_links, do:true"`
	Id               any //
	ProjectId        any //
	EvidenceItemId   any //
	LinkedObjectType any //
	LinkedObjectId   any //
	CreatedAt        any //
}
