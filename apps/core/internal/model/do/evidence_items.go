// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// EvidenceItems is the golang structure of table evidence_items for DAO operations like Where/Data.
type EvidenceItems struct {
	g.Meta       `orm:"table:evidence_items, do:true"`
	Id           any //
	ProjectId    any //
	RunId        any //
	Surface      any //
	Journey      any //
	EvidenceType any //
	FilePath     any //
	ContentHash  any //
	FileSize     any //
	CapturedAt   any //
	CreatedAt    any //
}
