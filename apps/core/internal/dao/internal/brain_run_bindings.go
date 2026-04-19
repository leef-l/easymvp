// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BrainRunBindingsDao is the data access object for the table brain_run_bindings.
type BrainRunBindingsDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  BrainRunBindingsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// BrainRunBindingsColumns defines and stores column names for the table brain_run_bindings.
type BrainRunBindingsColumns struct {
	Id         string //
	ProjectId  string //
	TaskId     string //
	BrainKind  string //
	BrainRunId string //
	RunStatus  string //
	StartedAt  string //
	FinishedAt string //
	LastSyncAt string //
	CreatedAt  string //
	UpdatedAt  string //
}

// brainRunBindingsColumns holds the columns for the table brain_run_bindings.
var brainRunBindingsColumns = BrainRunBindingsColumns{
	Id:         "id",
	ProjectId:  "project_id",
	TaskId:     "task_id",
	BrainKind:  "brain_kind",
	BrainRunId: "brain_run_id",
	RunStatus:  "run_status",
	StartedAt:  "started_at",
	FinishedAt: "finished_at",
	LastSyncAt: "last_sync_at",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
}

// NewBrainRunBindingsDao creates and returns a new DAO object for table data access.
func NewBrainRunBindingsDao(handlers ...gdb.ModelHandler) *BrainRunBindingsDao {
	return &BrainRunBindingsDao{
		group:    "default",
		table:    "brain_run_bindings",
		columns:  brainRunBindingsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BrainRunBindingsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BrainRunBindingsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BrainRunBindingsDao) Columns() BrainRunBindingsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BrainRunBindingsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BrainRunBindingsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BrainRunBindingsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
