// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// RunEventIndexDao is the data access object for the table run_event_index.
type RunEventIndexDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  RunEventIndexColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// RunEventIndexColumns defines and stores column names for the table run_event_index.
type RunEventIndexColumns struct {
	Id           string //
	ProjectId    string //
	RunBindingId string //
	SequenceNo   string //
	EventType    string //
	EventLevel   string //
	Summary      string //
	PayloadJson  string //
	CreatedAt    string //
}

// runEventIndexColumns holds the columns for the table run_event_index.
var runEventIndexColumns = RunEventIndexColumns{
	Id:           "id",
	ProjectId:    "project_id",
	RunBindingId: "run_binding_id",
	SequenceNo:   "sequence_no",
	EventType:    "event_type",
	EventLevel:   "event_level",
	Summary:      "summary",
	PayloadJson:  "payload_json",
	CreatedAt:    "created_at",
}

// NewRunEventIndexDao creates and returns a new DAO object for table data access.
func NewRunEventIndexDao(handlers ...gdb.ModelHandler) *RunEventIndexDao {
	return &RunEventIndexDao{
		group:    "default",
		table:    "run_event_index",
		columns:  runEventIndexColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *RunEventIndexDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *RunEventIndexDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *RunEventIndexDao) Columns() RunEventIndexColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *RunEventIndexDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *RunEventIndexDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *RunEventIndexDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
