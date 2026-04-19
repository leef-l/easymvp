// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkspaceSnapshotsDao is the data access object for the table workspace_snapshots.
type WorkspaceSnapshotsDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  WorkspaceSnapshotsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// WorkspaceSnapshotsColumns defines and stores column names for the table workspace_snapshots.
type WorkspaceSnapshotsColumns struct {
	Key          string //
	SnapshotJson string //
	GeneratedAt  string //
}

// workspaceSnapshotsColumns holds the columns for the table workspace_snapshots.
var workspaceSnapshotsColumns = WorkspaceSnapshotsColumns{
	Key:          "key",
	SnapshotJson: "snapshot_json",
	GeneratedAt:  "generated_at",
}

// NewWorkspaceSnapshotsDao creates and returns a new DAO object for table data access.
func NewWorkspaceSnapshotsDao(handlers ...gdb.ModelHandler) *WorkspaceSnapshotsDao {
	return &WorkspaceSnapshotsDao{
		group:    "default",
		table:    "workspace_snapshots",
		columns:  workspaceSnapshotsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WorkspaceSnapshotsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WorkspaceSnapshotsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WorkspaceSnapshotsDao) Columns() WorkspaceSnapshotsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WorkspaceSnapshotsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WorkspaceSnapshotsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WorkspaceSnapshotsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
