// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpStageRun is the golang structure of table mvp_stage_run for DAO operations like Where/Data.
type MvpStageRun struct {
	g.Meta        `orm:"table:mvp_stage_run, do:true"`
	Id            any         // 雪花ID
	WorkflowRunId any         // 所属工作流运行ID
	StageType     any         // 阶段类型: design/review/execute/rework/complete
	StageNo       any         // 同类型阶段序号(支持多轮)
	Status        any         // 状态: pending/running/completed/failed/skipped
	InputRef      any         // 阶段输入引用(JSON)
	OutputRef     any         // 阶段输出引用(JSON)
	Decision      any         // 阶段决策结果(JSON)
	ErrorMessage  any         // 错误信息
	StartedAt     *gtime.Time // 开始时间
	FinishedAt    *gtime.Time // 结束时间
	CreatedAt     *gtime.Time // 创建时间
	UpdatedAt     *gtime.Time // 更新时间
	DeletedAt     *gtime.Time // 软删除时间
}
