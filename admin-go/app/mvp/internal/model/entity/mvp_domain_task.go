// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpDomainTask is the golang structure for table mvp_domain_task.
type MvpDomainTask struct {
	Id                uint64      `orm:"id"                  description:"雪花ID"`                                                                             // 雪花ID
	WorkflowRunId     uint64      `orm:"workflow_run_id"     description:"所属工作流运行ID"`                                                                        // 所属工作流运行ID
	StageRunId        uint64      `orm:"stage_run_id"        description:"所属阶段运行ID"`                                                                         // 所属阶段运行ID
	PlanVersionId     uint64      `orm:"plan_version_id"     description:"来源计划版本ID"`                                                                         // 来源计划版本ID
	BlueprintId       uint64      `orm:"blueprint_id"        description:"来源蓝图ID"`                                                                           // 来源蓝图ID
	ParentTaskId      uint64      `orm:"parent_task_id"      description:"父任务ID"`                                                                            // 父任务ID
	DependsOnTaskIds  string      `orm:"depends_on_task_ids" description:"依赖任务ID列表(JSON数组)"`                                                                 // 依赖任务ID列表(JSON数组)
	SourceTaskId      uint64      `orm:"source_task_id"      description:"来源任务ID(链路追踪)"`                                                                     // 来源任务ID(链路追踪)
	RootTaskId        uint64      `orm:"root_task_id"        description:"根任务ID(链路追踪)"`                                                                      // 根任务ID(链路追踪)
	TaskKind          string      `orm:"task_kind"           description:"任务种类: implement/audit/bug_analysis/failure_analysis"`                              // 任务种类: implement/audit/bug_analysis/failure_analysis
	Name              string      `orm:"name"                description:"任务名称"`                                                                             // 任务名称
	Description       string      `orm:"description"         description:"任务描述"`                                                                             // 任务描述
	RoleType          string      `orm:"role_type"           description:"角色类型"`                                                                             // 角色类型
	RoleLevel         string      `orm:"role_level"          description:"角色等级"`                                                                             // 角色等级
	ExecutionMode     string      `orm:"execution_mode"      description:"执行方式: chat/aider/openhands"`                                                       // 执行方式: chat/aider/openhands
	Status            string      `orm:"status"              description:"状态: pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched"` // 状态: pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched
	ConversationId    uint64      `orm:"conversation_id"     description:"关联对话ID"`                                                                           // 关联对话ID
	ModelId           uint64      `orm:"model_id"            description:"使用的AI模型ID"`                                                                        // 使用的AI模型ID
	BatchNo           int         `orm:"batch_no"            description:"批次号"`                                                                              // 批次号
	Sort              int         `orm:"sort"                description:"排序"`                                                                               // 排序
	RetryCount        int         `orm:"retry_count"         description:"重试次数"`                                                                             // 重试次数
	AffectedResources string      `orm:"affected_resources"  description:"影响资源列表(JSON)"`                                                                     // 影响资源列表(JSON)
	LockedResources   string      `orm:"locked_resources"    description:"锁定资源列表(JSON)"`                                                                     // 锁定资源列表(JSON)
	Result            string      `orm:"result"              description:"执行结果"`                                                                             // 执行结果
	ContextSummary    string      `orm:"context_summary"     description:"上下文摘要"`                                                                            // 上下文摘要
	HeartbeatAt       *gtime.Time `orm:"heartbeat_at"        description:"心跳时间"`                                                                             // 心跳时间
	StartedAt         *gtime.Time `orm:"started_at"          description:"开始时间"`                                                                             // 开始时间
	CompletedAt       *gtime.Time `orm:"completed_at"        description:"完成时间"`                                                                             // 完成时间
	CreatedAt         *gtime.Time `orm:"created_at"          description:"创建时间"`                                                                             // 创建时间
	UpdatedAt         *gtime.Time `orm:"updated_at"          description:"更新时间"`                                                                             // 更新时间
	DeletedAt         *gtime.Time `orm:"deleted_at"          description:"软删除时间"`                                                                            // 软删除时间
	CreatedBy         int64       `orm:"created_by"          description:"创建人ID"`                                                                            // 创建人ID
	DeptId            int64       `orm:"dept_id"             description:"部门ID"`                                                                             // 部门ID
}
