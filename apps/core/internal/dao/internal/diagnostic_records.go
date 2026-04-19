// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// DiagnosticRecordsDao is the data access object for the table diagnostic_records.
type DiagnosticRecordsDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  DiagnosticRecordsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// DiagnosticRecordsColumns defines and stores column names for the table diagnostic_records.
type DiagnosticRecordsColumns struct {
	Id         string //
	Scope      string //
	Severity   string //
	ErrorCode  string //
	Summary    string //
	DetailJson string //
	CreatedAt  string //
}

// diagnosticRecordsColumns holds the columns for the table diagnostic_records.
var diagnosticRecordsColumns = DiagnosticRecordsColumns{
	Id:         "id",
	Scope:      "scope",
	Severity:   "severity",
	ErrorCode:  "error_code",
	Summary:    "summary",
	DetailJson: "detail_json",
	CreatedAt:  "created_at",
}

// NewDiagnosticRecordsDao creates and returns a new DAO object for table data access.
func NewDiagnosticRecordsDao(handlers ...gdb.ModelHandler) *DiagnosticRecordsDao {
	return &DiagnosticRecordsDao{
		group:    "default",
		table:    "diagnostic_records",
		columns:  diagnosticRecordsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *DiagnosticRecordsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *DiagnosticRecordsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *DiagnosticRecordsDao) Columns() DiagnosticRecordsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *DiagnosticRecordsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *DiagnosticRecordsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *DiagnosticRecordsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
