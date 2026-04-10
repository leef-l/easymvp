// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpWorkflowEvent is the golang structure of table mvp_workflow_event for DAO operations like Where/Data.
type MvpWorkflowEvent struct {
	g.Meta         `orm:"table:mvp_workflow_event, do:true"`
	Id             any         // 雪花ID
	EventId        any         // 事件元ID
	WorkflowRunId  any         // 所属工作流运行ID
	StageRunId     any         // 关联阶段运行ID
	EntityType     any         // 实体类型: workflow_run/stage_run/plan_version/domain_task/review_issue
	EntityId       any         // 实体ID
	EventType      any         // 事件类型: workflow.created/stage.started/task.completed等
	Attempt        any         // 事件尝试次数
	IdempotencyKey any         // 幂等键
	Payload        any         // 事件载荷(JSON)
	CreatedAt      *gtime.Time // 创建时间
}
