// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpDomainTaskDao is the data access object for the table mvp_domain_task.
type MvpDomainTaskDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  MvpDomainTaskColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// MvpDomainTaskColumns defines and stores column names for the table mvp_domain_task.
type MvpDomainTaskColumns struct {
	Id                string // 雪花ID
	WorkflowRunId     string // 所属工作流运行ID
	StageRunId        string // 所属阶段运行ID
	PlanVersionId     string // 来源计划版本ID
	BlueprintId       string // 来源蓝图ID
	ParentTaskId      string // 父任务ID
	DependsOnTaskIds  string // 依赖任务ID列表(JSON数组)
	SourceTaskId      string // 来源任务ID(链路追踪)
	RootTaskId        string // 根任务ID(链路追踪)
	TaskKind          string // 任务种类: implement/audit/bug_analysis/failure_analysis
	Name              string // 任务名称
	Description       string // 任务描述
	RoleType          string // 角色类型
	RoleLevel         string // 角色等级
	ExecutionMode     string // 执行方式: chat/aider/openhands
	Status            string // 状态: pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched
	ConversationId    string // 关联对话ID
	ModelId           string // 使用的AI模型ID
	BatchNo           string // 批次号
	Sort              string // 排序
	RetryCount        string // 重试次数
	AffectedResources string // 影响资源列表(JSON)
	LockedResources   string // 锁定资源列表(JSON)
	Result            string // 执行结果
	ContextSummary    string // 上下文摘要
	HeartbeatAt       string // 心跳时间
	StartedAt         string // 开始时间
	CompletedAt       string // 完成时间
	CreatedAt         string // 创建时间
	UpdatedAt         string // 更新时间
	DeletedAt         string // 软删除时间
}

// mvpDomainTaskColumns holds the columns for the table mvp_domain_task.
var mvpDomainTaskColumns = MvpDomainTaskColumns{
	Id:                "id",
	WorkflowRunId:     "workflow_run_id",
	StageRunId:        "stage_run_id",
	PlanVersionId:     "plan_version_id",
	BlueprintId:       "blueprint_id",
	ParentTaskId:      "parent_task_id",
	DependsOnTaskIds:  "depends_on_task_ids",
	SourceTaskId:      "source_task_id",
	RootTaskId:        "root_task_id",
	TaskKind:          "task_kind",
	Name:              "name",
	Description:       "description",
	RoleType:          "role_type",
	RoleLevel:         "role_level",
	ExecutionMode:     "execution_mode",
	Status:            "status",
	ConversationId:    "conversation_id",
	ModelId:           "model_id",
	BatchNo:           "batch_no",
	Sort:              "sort",
	RetryCount:        "retry_count",
	AffectedResources: "affected_resources",
	LockedResources:   "locked_resources",
	Result:            "result",
	ContextSummary:    "context_summary",
	HeartbeatAt:       "heartbeat_at",
	StartedAt:         "started_at",
	CompletedAt:       "completed_at",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
	DeletedAt:         "deleted_at",
}

// NewMvpDomainTaskDao creates and returns a new DAO object for table data access.
func NewMvpDomainTaskDao(handlers ...gdb.ModelHandler) *MvpDomainTaskDao {
	return &MvpDomainTaskDao{
		group:    "default",
		table:    "mvp_domain_task",
		columns:  mvpDomainTaskColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpDomainTaskDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpDomainTaskDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpDomainTaskDao) Columns() MvpDomainTaskColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpDomainTaskDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpDomainTaskDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *MvpDomainTaskDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
