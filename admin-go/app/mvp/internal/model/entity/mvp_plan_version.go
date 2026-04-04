// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpPlanVersion is the golang structure for table mvp_plan_version.
type MvpPlanVersion struct {
	Id                   uint64      `orm:"id"                     description:"雪花ID"`                            // 雪花ID
	ProjectId            uint64      `orm:"project_id"             description:"所属项目ID"`                          // 所属项目ID
	WorkflowRunId        uint64      `orm:"workflow_run_id"        description:"所属工作流运行ID"`                       // 所属工作流运行ID
	VersionNo            int         `orm:"version_no"             description:"版本号(项目内递增)"`                      // 版本号(项目内递增)
	SourceConversationId uint64      `orm:"source_conversation_id" description:"来源对话ID"`                          // 来源对话ID
	SourceMessageId      uint64      `orm:"source_message_id"      description:"来源消息ID"`                          // 来源消息ID
	Status               string      `orm:"status"                 description:"版本状态: draft/active/superseded"`   // 版本状态: draft/active/superseded
	ReviewStatus         string      `orm:"review_status"          description:"审核状态: pending/approved/rejected"` // 审核状态: pending/approved/rejected
	Summary              string      `orm:"summary"                description:"版本摘要"`                            // 版本摘要
	DiffSummary          string      `orm:"diff_summary"           description:"与上一版本的差异摘要"`                      // 与上一版本的差异摘要
	ApprovedAt           *gtime.Time `orm:"approved_at"            description:"审核通过时间"`                          // 审核通过时间
	RejectedAt           *gtime.Time `orm:"rejected_at"            description:"审核驳回时间"`                          // 审核驳回时间
	CreatedAt            *gtime.Time `orm:"created_at"             description:"创建时间"`                            // 创建时间
	UpdatedAt            *gtime.Time `orm:"updated_at"             description:"更新时间"`                            // 更新时间
	DeletedAt            *gtime.Time `orm:"deleted_at"             description:"软删除时间"`                           // 软删除时间
}
