// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpDomainTask is the golang structure of table mvp_domain_task for DAO operations like Where/Data.
type MvpDomainTask struct {
	g.Meta            `orm:"table:mvp_domain_task, do:true"`
	Id                any         // 雪花ID
	WorkflowRunId     any         // 所属工作流运行ID
	StageRunId        any         // 所属阶段运行ID
	PlanVersionId     any         // 来源计划版本ID
	BlueprintId       any         // 来源蓝图ID
	ParentTaskId      any         // 父任务ID
	SourceTaskId      any         // 来源任务ID(链路追踪)
	RootTaskId        any         // 根任务ID(链路追踪)
	TaskKind          any         // 任务种类: implement/audit/bug_analysis/failure_analysis
	Name              any         // 任务名称
	Description       any         // 任务描述
	RoleType          any         // 角色类型
	RoleLevel         any         // 角色等级
	ExecutionMode     any         // 执行方式: chat/aider/openhands
	Status            any         // 状态: pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched
	ConversationId    any         // 关联对话ID
	ModelId           any         // 使用的AI模型ID
	BatchNo           any         // 批次号
	Sort              any         // 排序
	RetryCount        any         // 重试次数
	AffectedResources any         // 影响资源列表(JSON)
	LockedResources   any         // 锁定资源列表(JSON)
	Result            any         // 执行结果
	ContextSummary    any         // 上下文摘要
	HeartbeatAt       *gtime.Time // 心跳时间
	StartedAt         *gtime.Time // 开始时间
	CompletedAt       *gtime.Time // 完成时间
	CreatedAt         *gtime.Time // 创建时间
	UpdatedAt         *gtime.Time // 更新时间
	DeletedAt         *gtime.Time // 软删除时间
}
