// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpHandoffRecord is the golang structure for table mvp_handoff_record.
type MvpHandoffRecord struct {
	Id            uint64      `orm:"id"              description:"雪花ID"`                                          // 雪花ID
	WorkflowRunId uint64      `orm:"workflow_run_id" description:"所属工作流运行ID"`                                     // 所属工作流运行ID
	FromTaskId    uint64      `orm:"from_task_id"    description:"来源任务ID"`                                        // 来源任务ID
	ToTaskId      uint64      `orm:"to_task_id"      description:"目标任务ID"`                                        // 目标任务ID
	HandoffType   string      `orm:"handoff_type"    description:"交接类型: bug_fix/failure_escalation/rework/audit"` // 交接类型: bug_fix/failure_escalation/rework/audit
	Reason        string      `orm:"reason"          description:"交接原因"`                                          // 交接原因
	Payload       string      `orm:"payload"         description:"交接载荷(JSON)"`                                    // 交接载荷(JSON)
	CreatedAt     *gtime.Time `orm:"created_at"      description:"创建时间"`                                          // 创建时间
}
