// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpWorkflowRun is the golang structure for table mvp_workflow_run.
type MvpWorkflowRun struct {
	Id                  uint64      `orm:"id"                     description:"雪花ID"`                                                                         // 雪花ID
	ProjectId           uint64      `orm:"project_id"             description:"所属项目ID"`                                                                       // 所属项目ID
	RunNo               int         `orm:"run_no"                 description:"项目内运行序号(从1递增)"`                                                                // 项目内运行序号(从1递增)
	Status              string      `orm:"status"                 description:"状态: designing/reviewing/executing/reworking/paused/completed/failed/canceled"` // 状态: designing/reviewing/executing/reworking/paused/completed/failed/canceled
	TokensConsumed      int64       `orm:"tokens_consumed"        description:"已消耗Token总量"`                                                                   // 已消耗Token总量
	ReplanCount         int         `orm:"replan_count"           description:"重规划次数"`                                                                        // 重规划次数
	CurrentStage        string      `orm:"current_stage"          description:"当前阶段: design/review/execute/rework/complete"`                                  // 当前阶段: design/review/execute/rework/complete
	CurrentStageRunId   uint64      `orm:"current_stage_run_id"   description:"当前阶段运行ID"`                                                                     // 当前阶段运行ID
	ActivePlanVersionId uint64      `orm:"active_plan_version_id" description:"当前活跃计划版本ID"`                                                                   // 当前活跃计划版本ID
	PauseReason         string      `orm:"pause_reason"           description:"暂停原因"`                                                                         // 暂停原因
	StatusBeforePause   string      `orm:"status_before_pause"    description:"暂停前的阶段状态（恢复时回退）"`                                                              // 暂停前的阶段状态（恢复时回退）
	CancelReason        string      `orm:"cancel_reason"          description:"取消原因"`                                                                         // 取消原因
	RuntimeToken        string      `orm:"runtime_token"          description:"运行时令牌(防重入)"`                                                                   // 运行时令牌(防重入)
	StartedAt           *gtime.Time `orm:"started_at"             description:"开始时间"`                                                                         // 开始时间
	FinishedAt          *gtime.Time `orm:"finished_at"            description:"结束时间"`                                                                         // 结束时间
	CreatedBy           uint64      `orm:"created_by"             description:"创建人ID"`                                                                        // 创建人ID
	DeptId              uint64      `orm:"dept_id"                description:"所属部门ID"`                                                                       // 所属部门ID
	CreatedAt           *gtime.Time `orm:"created_at"             description:"创建时间"`                                                                         // 创建时间
	UpdatedAt           *gtime.Time `orm:"updated_at"             description:"更新时间"`                                                                         // 更新时间
	DeletedAt           *gtime.Time `orm:"deleted_at"             description:"软删除时间"`                                                                        // 软删除时间
}
