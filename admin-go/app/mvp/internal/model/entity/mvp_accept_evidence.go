// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpAcceptEvidence is the golang structure for table mvp_accept_evidence.
type MvpAcceptEvidence struct {
	Id           int64       `orm:"id"            description:"主键ID"`                                                   // 主键ID
	AcceptRunId  int64       `orm:"accept_run_id" description:"验收运行ID"`                                                 // 验收运行ID
	EvidenceType string      `orm:"evidence_type" description:"task_output/file/log/diff/stage_output/handoff/summary"` // task_output/file/log/diff/stage_output/handoff/summary
	SourceType   string      `orm:"source_type"   description:"domain_task/stage_run/file/handoff_record/workflow_run"` // domain_task/stage_run/file/handoff_record/workflow_run
	SourceId     int64       `orm:"source_id"     description:"来源对象ID"`                                                 // 来源对象ID
	ContentRef   string      `orm:"content_ref"   description:"证据引用或JSON"`                                              // 证据引用或JSON
	Summary      string      `orm:"summary"       description:"证据摘要"`                                                   // 证据摘要
	CreatedAt    *gtime.Time `orm:"created_at"    description:"创建时间"`                                                   // 创建时间
	UpdatedAt    *gtime.Time `orm:"updated_at"    description:"更新时间"`                                                   // 更新时间
	DeletedAt    *gtime.Time `orm:"deleted_at"    description:"删除时间"`                                                   // 删除时间
}
