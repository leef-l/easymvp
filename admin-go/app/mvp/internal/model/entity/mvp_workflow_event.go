// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpWorkflowEvent is the golang structure for table mvp_workflow_event.
type MvpWorkflowEvent struct {
	Id            uint64      `orm:"id"              description:"雪花ID"`                                                               // 雪花ID
	WorkflowRunId uint64      `orm:"workflow_run_id" description:"所属工作流运行ID"`                                                          // 所属工作流运行ID
	StageRunId    uint64      `orm:"stage_run_id"    description:"关联阶段运行ID"`                                                           // 关联阶段运行ID
	EntityType    string      `orm:"entity_type"     description:"实体类型: workflow_run/stage_run/plan_version/domain_task/review_issue"` // 实体类型: workflow_run/stage_run/plan_version/domain_task/review_issue
	EntityId      uint64      `orm:"entity_id"       description:"实体ID"`                                                               // 实体ID
	EventType     string      `orm:"event_type"      description:"事件类型: workflow.created/stage.started/task.completed等"`               // 事件类型: workflow.created/stage.started/task.completed等
	Payload       string      `orm:"payload"         description:"事件载荷(JSON)"`                                                         // 事件载荷(JSON)
	CreatedAt     *gtime.Time `orm:"created_at"      description:"创建时间"`                                                               // 创建时间
}
