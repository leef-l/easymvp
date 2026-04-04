// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpStageTask is the golang structure of table mvp_stage_task for DAO operations like Where/Data.
type MvpStageTask struct {
	g.Meta        `orm:"table:mvp_stage_task, do:true"`
	Id            any         // 雪花ID
	StageRunId    any         // 所属阶段运行ID
	TaskType      any         // 任务类型: precheck/auditor_review/coordinator_optimize/review_summary
	RoleType      any         // 执行角色
	Status        any         // 状态: pending/running/completed/failed/skipped
	InputPayload  any         // 输入载荷(JSON)
	OutputPayload any         // 输出载荷(JSON)
	ErrorMessage  any         // 错误信息
	StartedAt     *gtime.Time // 开始时间
	CompletedAt   *gtime.Time // 完成时间
	CreatedAt     *gtime.Time // 创建时间
	UpdatedAt     *gtime.Time // 更新时间
	DeletedAt     *gtime.Time // 软删除时间
}
