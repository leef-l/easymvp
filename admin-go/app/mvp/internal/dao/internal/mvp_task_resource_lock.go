// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpTaskResourceLockDao is the data access object for the table mvp_task_resource_lock.
type MvpTaskResourceLockDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  MvpTaskResourceLockColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// MvpTaskResourceLockColumns defines and stores column names for the table mvp_task_resource_lock.
type MvpTaskResourceLockColumns struct {
	Id            string // 雪花ID
	WorkflowRunId string // 所属工作流运行ID
	TaskId        string // 持锁任务ID
	ResourcePath  string // 资源路径
	LockStatus    string // 锁状态: held/released/leaked
	LockedAt      string // 加锁时间
	ReleasedAt    string // 释放时间
}

// mvpTaskResourceLockColumns holds the columns for the table mvp_task_resource_lock.
var mvpTaskResourceLockColumns = MvpTaskResourceLockColumns{
	Id:            "id",
	WorkflowRunId: "workflow_run_id",
	TaskId:        "task_id",
	ResourcePath:  "resource_path",
	LockStatus:    "lock_status",
	LockedAt:      "locked_at",
	ReleasedAt:    "released_at",
}

// NewMvpTaskResourceLockDao creates and returns a new DAO object for table data access.
func NewMvpTaskResourceLockDao(handlers ...gdb.ModelHandler) *MvpTaskResourceLockDao {
	return &MvpTaskResourceLockDao{
		group:    "default",
		table:    "mvp_task_resource_lock",
		columns:  mvpTaskResourceLockColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpTaskResourceLockDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpTaskResourceLockDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpTaskResourceLockDao) Columns() MvpTaskResourceLockColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpTaskResourceLockDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpTaskResourceLockDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpTaskResourceLockDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
