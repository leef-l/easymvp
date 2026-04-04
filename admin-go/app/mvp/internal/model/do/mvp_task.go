// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpTask is the golang structure of table mvp_task for DAO operations like Where/Data.
type MvpTask struct {
	g.Meta            `orm:"table:mvp_task, do:true"`
	Id                any         // 雪花ID
	ProjectId         any         // 项目ID
	ParentId          any         // 父任务ID，0=顶级
	Name              any         // 任务名称
	Description       any         // 任务描述
	RoleType          any         // 角色类型：architect/implementer/auditor/coordinator
	RoleLevel         any         // 角色等级：lite/pro/max
	TaskKind          any         // 任务记录类型：implement/audit/bug_analysis/failure_analysis
	SourceTaskId      any         // 直接来源任务ID，原始任务为NULL
	RootTaskId        any         // 所属主链根任务ID
	ModelId           any         // 使用的AI模型ID
	ConversationId    any         // 任务对话ID，用于检测任务状态
	Status            any         // 任务状态: draft/pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched/submit_error
	Sort              any         // 排序
	BatchNo           any         // 执行批次号，同批次内可并行，批次间串行
	AffectedResources any         // 涉及的资源范围（文件/模块），用于并发冲突检测
	LockedResources   any         // 任务持有的资源锁（JSON数组，持久化防泄露）
	HeartbeatAt       *gtime.Time // 最近心跳时间（执行器定期更新，看门狗检测超时）
	DependsOn         any         // 依赖的任务ID列表
	Result            any         // 任务执行结果
	ContextSummary    any         // 任务完成后的上下文压缩摘要，供后续AI读取
	ErrorMessage      any         // 错误信息
	StartedAt         *gtime.Time // 开始时间
	CompletedAt       *gtime.Time // 完成时间
	CreatedBy         any         // 创建人ID
	DeptId            any         // 所属部门ID
	CreatedAt         *gtime.Time // 创建时间
	UpdatedAt         *gtime.Time // 更新时间
	DeletedAt         *gtime.Time // 软删除时间
}
