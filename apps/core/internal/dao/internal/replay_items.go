// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ReplayItemsDao is the data access object for the table replay_items.
type ReplayItemsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ReplayItemsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ReplayItemsColumns defines and stores column names for the table replay_items.
type ReplayItemsColumns struct {
	Id         string //
	ProjectId  string //
	RunId      string //
	ReplayType string //
	FilePath   string //
	Summary    string //
	CreatedAt  string //
}

// replayItemsColumns holds the columns for the table replay_items.
var replayItemsColumns = ReplayItemsColumns{
	Id:         "id",
	ProjectId:  "project_id",
	RunId:      "run_id",
	ReplayType: "replay_type",
	FilePath:   "file_path",
	Summary:    "summary",
	CreatedAt:  "created_at",
}

// NewReplayItemsDao creates and returns a new DAO object for table data access.
func NewReplayItemsDao(handlers ...gdb.ModelHandler) *ReplayItemsDao {
	return &ReplayItemsDao{
		group:    "default",
		table:    "replay_items",
		columns:  replayItemsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ReplayItemsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ReplayItemsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ReplayItemsDao) Columns() ReplayItemsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ReplayItemsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ReplayItemsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ReplayItemsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
