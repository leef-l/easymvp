// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpTaskDao is the data access object for the table mvp_task.
type MvpTaskDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MvpTaskColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MvpTaskColumns defines and stores column names for the table mvp_task.
type MvpTaskColumns struct {
	Id                string // 雪花ID
	ProjectId         string // 项目ID
	ParentId          string // 父任务ID，0=顶级
	Name              string // 任务名称
	Description       string // 任务描述
	RoleType          string // 角色类型：architect/implementer/auditor/coordinator
	RoleLevel         string // 角色等级：lite/pro/max
	TaskKind          string // 任务记录类型：implement/audit/bug_analysis/failure_analysis
	SourceTaskId      string // 直接来源任务ID，原始任务为NULL
	RootTaskId        string // 所属主链根任务ID
	ModelId           string // 使用的AI模型ID
	ConversationId    string // 任务对话ID，用于检测任务状态
	Status            string // 任务状态: draft/pending/running/completed/failed/escalated/auditing/bug_found/bug_dispatched/submit_error
	Sort              string // 排序
	BatchNo           string // 执行批次号，同批次内可并行，批次间串行
	AffectedResources string // 涉及的资源范围（文件/模块），用于并发冲突检测
	LockedResources   string // 任务持有的资源锁（JSON数组，持久化防泄露）
	HeartbeatAt       string // 最近心跳时间（执行器定期更新，看门狗检测超时）
	DependsOn         string // 依赖的任务ID列表
	Result            string // 任务执行结果
	ContextSummary    string // 任务完成后的上下文压缩摘要，供后续AI读取
	ErrorMessage      string // 错误信息
	StartedAt         string // 开始时间
	CompletedAt       string // 完成时间
	CreatedBy         string // 创建人ID
	DeptId            string // 所属部门ID
	CreatedAt         string // 创建时间
	UpdatedAt         string // 更新时间
	DeletedAt         string // 软删除时间
}

// mvpTaskColumns holds the columns for the table mvp_task.
var mvpTaskColumns = MvpTaskColumns{
	Id:                "id",
	ProjectId:         "project_id",
	ParentId:          "parent_id",
	Name:              "name",
	Description:       "description",
	RoleType:          "role_type",
	RoleLevel:         "role_level",
	TaskKind:          "task_kind",
	SourceTaskId:      "source_task_id",
	RootTaskId:        "root_task_id",
	ModelId:           "model_id",
	ConversationId:    "conversation_id",
	Status:            "status",
	Sort:              "sort",
	BatchNo:           "batch_no",
	AffectedResources: "affected_resources",
	LockedResources:   "locked_resources",
	HeartbeatAt:       "heartbeat_at",
	DependsOn:         "depends_on",
	Result:            "result",
	ContextSummary:    "context_summary",
	ErrorMessage:      "error_message",
	StartedAt:         "started_at",
	CompletedAt:       "completed_at",
	CreatedBy:         "created_by",
	DeptId:            "dept_id",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
	DeletedAt:         "deleted_at",
}

// NewMvpTaskDao creates and returns a new DAO object for table data access.
func NewMvpTaskDao(handlers ...gdb.ModelHandler) *MvpTaskDao {
	return &MvpTaskDao{
		group:    "default",
		table:    "mvp_task",
		columns:  mvpTaskColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpTaskDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpTaskDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpTaskDao) Columns() MvpTaskColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpTaskDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpTaskDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpTaskDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
