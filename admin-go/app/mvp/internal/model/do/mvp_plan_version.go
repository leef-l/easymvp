// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpPlanVersion is the golang structure of table mvp_plan_version for DAO operations like Where/Data.
type MvpPlanVersion struct {
	g.Meta               `orm:"table:mvp_plan_version, do:true"`
	Id                   any         // 雪花ID
	ProjectId            any         // 所属项目ID
	WorkflowRunId        any         // 所属工作流运行ID
	VersionNo            any         // 版本号(项目内递增)
	SourceConversationId any         // 来源对话ID
	SourceMessageId      any         // 来源消息ID
	Status               any         // 版本状态: draft/active/superseded
	ReviewStatus         any         // 审核状态: pending/approved/rejected
	Summary              any         // 版本摘要
	DiffSummary          any         // 与上一版本的差异摘要
	ApprovedAt           *gtime.Time // 审核通过时间
	RejectedAt           *gtime.Time // 审核驳回时间
	CreatedAt            *gtime.Time // 创建时间
	UpdatedAt            *gtime.Time // 更新时间
	DeletedAt            *gtime.Time // 软删除时间
	CreatedBy            any         // 创建人ID
	DeptId               any         // 部门ID
}
