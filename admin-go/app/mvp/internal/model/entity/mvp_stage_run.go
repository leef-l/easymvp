// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpStageRun is the golang structure for table mvp_stage_run.
type MvpStageRun struct {
	Id            uint64      `orm:"id"              description:"雪花ID"`                                         // 雪花ID
	WorkflowRunId uint64      `orm:"workflow_run_id" description:"所属工作流运行ID"`                                    // 所属工作流运行ID
	StageType     string      `orm:"stage_type"      description:"阶段类型: design/review/execute/rework/complete"`  // 阶段类型: design/review/execute/rework/complete
	StageNo       int         `orm:"stage_no"        description:"同类型阶段序号(支持多轮)"`                                // 同类型阶段序号(支持多轮)
	Status        string      `orm:"status"          description:"状态: pending/running/completed/failed/skipped"` // 状态: pending/running/completed/failed/skipped
	CreatedBy     int64       `orm:"created_by"      description:"创建人ID"`                                        // 创建人ID
	DeptId        int64       `orm:"dept_id"         description:"部门ID"`                                         // 部门ID
	InputRef      string      `orm:"input_ref"       description:"阶段输入引用(JSON)"`                                 // 阶段输入引用(JSON)
	OutputRef     string      `orm:"output_ref"      description:"阶段输出引用(JSON)"`                                 // 阶段输出引用(JSON)
	Decision      string      `orm:"decision"        description:"阶段决策结果(JSON)"`                                 // 阶段决策结果(JSON)
	ErrorMessage  string      `orm:"error_message"   description:"错误信息"`                                         // 错误信息
	StartedAt     *gtime.Time `orm:"started_at"      description:"开始时间"`                                         // 开始时间
	FinishedAt    *gtime.Time `orm:"finished_at"     description:"结束时间"`                                         // 结束时间
	CreatedAt     *gtime.Time `orm:"created_at"      description:"创建时间"`                                         // 创建时间
	UpdatedAt     *gtime.Time `orm:"updated_at"      description:"更新时间"`                                         // 更新时间
	DeletedAt     *gtime.Time `orm:"deleted_at"      description:"软删除时间"`                                        // 软删除时间
}
