// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectSnapshotsDao is the data access object for the table project_snapshots.
type ProjectSnapshotsDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  ProjectSnapshotsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// ProjectSnapshotsColumns defines and stores column names for the table project_snapshots.
type ProjectSnapshotsColumns struct {
	ProjectId    string //
	SnapshotJson string //
	GeneratedAt  string //
}

// projectSnapshotsColumns holds the columns for the table project_snapshots.
var projectSnapshotsColumns = ProjectSnapshotsColumns{
	ProjectId:    "project_id",
	SnapshotJson: "snapshot_json",
	GeneratedAt:  "generated_at",
}

// NewProjectSnapshotsDao creates and returns a new DAO object for table data access.
func NewProjectSnapshotsDao(handlers ...gdb.ModelHandler) *ProjectSnapshotsDao {
	return &ProjectSnapshotsDao{
		group:    "default",
		table:    "project_snapshots",
		columns:  projectSnapshotsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProjectSnapshotsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProjectSnapshotsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProjectSnapshotsDao) Columns() ProjectSnapshotsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProjectSnapshotsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProjectSnapshotsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProjectSnapshotsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
