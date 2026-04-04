// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpWorkflowRun is the golang structure of table mvp_workflow_run for DAO operations like Where/Data.
type MvpWorkflowRun struct {
	g.Meta              `orm:"table:mvp_workflow_run, do:true"`
	Id                  any         // 雪花ID
	ProjectId           any         // 所属项目ID
	RunNo               any         // 项目内运行序号(从1递增)
	Status              any         // 状态: pending/running/paused/completed/canceled
	CurrentStage        any         // 当前阶段: design/review/execute/rework/complete
	CurrentStageRunId   any         // 当前阶段运行ID
	ActivePlanVersionId any         // 当前活跃计划版本ID
	PauseReason         any         // 暂停原因
	CancelReason        any         // 取消原因
	RuntimeToken        any         // 运行时令牌(防重入)
	StartedAt           *gtime.Time // 开始时间
	FinishedAt          *gtime.Time // 结束时间
	CreatedBy           any         // 创建人ID
	DeptId              any         // 所属部门ID
	CreatedAt           *gtime.Time // 创建时间
	UpdatedAt           *gtime.Time // 更新时间
	DeletedAt           *gtime.Time // 软删除时间
}
