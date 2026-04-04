// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpStageTask is the golang structure for table mvp_stage_task.
type MvpStageTask struct {
	Id            uint64      `orm:"id"             description:"雪花ID"`                                                              // 雪花ID
	StageRunId    uint64      `orm:"stage_run_id"   description:"所属阶段运行ID"`                                                          // 所属阶段运行ID
	TaskType      string      `orm:"task_type"      description:"任务类型: precheck/auditor_review/coordinator_optimize/review_summary"` // 任务类型: precheck/auditor_review/coordinator_optimize/review_summary
	RoleType      string      `orm:"role_type"      description:"执行角色"`                                                              // 执行角色
	Status        string      `orm:"status"         description:"状态: pending/running/completed/failed/skipped"`                      // 状态: pending/running/completed/failed/skipped
	InputPayload  string      `orm:"input_payload"  description:"输入载荷(JSON)"`                                                        // 输入载荷(JSON)
	OutputPayload string      `orm:"output_payload" description:"输出载荷(JSON)"`                                                        // 输出载荷(JSON)
	ErrorMessage  string      `orm:"error_message"  description:"错误信息"`                                                              // 错误信息
	StartedAt     *gtime.Time `orm:"started_at"     description:"开始时间"`                                                              // 开始时间
	CompletedAt   *gtime.Time `orm:"completed_at"   description:"完成时间"`                                                              // 完成时间
	CreatedAt     *gtime.Time `orm:"created_at"     description:"创建时间"`                                                              // 创建时间
	UpdatedAt     *gtime.Time `orm:"updated_at"     description:"更新时间"`                                                              // 更新时间
	DeletedAt     *gtime.Time `orm:"deleted_at"     description:"软删除时间"`                                                             // 软删除时间
}
