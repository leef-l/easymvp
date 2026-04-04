// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpTaskWorkspaceDao is the data access object for the table mvp_task_workspace.
type MvpTaskWorkspaceDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  MvpTaskWorkspaceColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// MvpTaskWorkspaceColumns defines and stores column names for the table mvp_task_workspace.
type MvpTaskWorkspaceColumns struct {
	Id            string // 雪花ID
	TaskId        string // 任务ID(domain_task或mvp_task)
	WorkflowRunId string // 所属工作流运行ID
	ProjectId     string // 项目ID
	WorkspaceType string // 工作空间类型: git_worktree
	WorkspacePath string // 工作空间绝对路径
	BaseRef       string // 基线引用(commit hash/branch)
	Status        string // 状态: creating/ready/running/completed/failed/canceled
	CleanupStatus string // 清理状态: pending/done/retained/failed
	DiffSummary   string // 变更摘要(diff统计)
	ErrorMessage  string // 错误信息
	CreatedAt     string // 创建时间
	UpdatedAt     string // 更新时间
	DeletedAt     string // 软删除时间
}

// mvpTaskWorkspaceColumns holds the columns for the table mvp_task_workspace.
var mvpTaskWorkspaceColumns = MvpTaskWorkspaceColumns{
	Id:            "id",
	TaskId:        "task_id",
	WorkflowRunId: "workflow_run_id",
	ProjectId:     "project_id",
	WorkspaceType: "workspace_type",
	WorkspacePath: "workspace_path",
	BaseRef:       "base_ref",
	Status:        "status",
	CleanupStatus: "cleanup_status",
	DiffSummary:   "diff_summary",
	ErrorMessage:  "error_message",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
	DeletedAt:     "deleted_at",
}

// NewMvpTaskWorkspaceDao creates and returns a new DAO object for table data access.
func NewMvpTaskWorkspaceDao(handlers ...gdb.ModelHandler) *MvpTaskWorkspaceDao {
	return &MvpTaskWorkspaceDao{
		group:    "default",
		table:    "mvp_task_workspace",
		columns:  mvpTaskWorkspaceColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpTaskWorkspaceDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpTaskWorkspaceDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpTaskWorkspaceDao) Columns() MvpTaskWorkspaceColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpTaskWorkspaceDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpTaskWorkspaceDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpTaskWorkspaceDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
