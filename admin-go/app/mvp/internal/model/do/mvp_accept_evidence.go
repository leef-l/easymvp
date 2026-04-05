// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpAcceptEvidence is the golang structure of table mvp_accept_evidence for DAO operations like Where/Data.
type MvpAcceptEvidence struct {
	g.Meta       `orm:"table:mvp_accept_evidence, do:true"`
	Id           any         // 主键ID
	AcceptRunId  any         // 验收运行ID
	EvidenceType any         // task_output/file/log/diff/stage_output/handoff/summary
	SourceType   any         // domain_task/stage_run/file/handoff_record/workflow_run
	SourceId     any         // 来源对象ID
	ContentRef   any         // 证据引用或JSON
	Summary      any         // 证据摘要
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
	DeletedAt    *gtime.Time // 删除时间
}
