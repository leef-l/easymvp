// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpAcceptRun is the golang structure of table mvp_accept_run for DAO operations like Where/Data.
type MvpAcceptRun struct {
	g.Meta           `orm:"table:mvp_accept_run, do:true"`
	Id               any         // 主键ID
	WorkflowRunId    any         // 工作流运行ID
	StageRunId       any         // accept阶段stage_run_id
	ProjectId        any         // 项目ID
	PlanVersionId    any         // 关联方案版本ID
	AcceptRound      any         // 第几轮验收
	Status           any         // pending/running/completed/failed/canceled
	Decision         any         // passed/failed/manual_review
	Score            any         // 验收评分
	Summary          any         // 验收摘要
	RulesVersion     any         // 规则版本号
	RulesSnapshotRef any         // 规则快照引用或JSON
	CreatedBy        any         // 创建人
	DeptId           any         // 部门ID
	StartedAt        *gtime.Time // 开始时间
	FinishedAt       *gtime.Time // 结束时间
	CreatedAt        *gtime.Time // 创建时间
	UpdatedAt        *gtime.Time // 更新时间
	DeletedAt        *gtime.Time // 删除时间
}
