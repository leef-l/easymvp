// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpHandoffRecord is the golang structure of table mvp_handoff_record for DAO operations like Where/Data.
type MvpHandoffRecord struct {
	g.Meta        `orm:"table:mvp_handoff_record, do:true"`
	Id            any         // 雪花ID
	WorkflowRunId any         // 所属工作流运行ID
	FromTaskId    any         // 来源任务ID
	ToTaskId      any         // 目标任务ID
	HandoffType   any         // 交接类型: bug_fix/failure_escalation/rework/audit
	Reason        any         // 交接原因
	Payload       any         // 交接载荷(JSON)
	CreatedAt     *gtime.Time // 创建时间
}
