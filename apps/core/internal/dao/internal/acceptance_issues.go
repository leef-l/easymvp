// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AcceptanceIssuesDao is the data access object for the table acceptance_issues.
type AcceptanceIssuesDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  AcceptanceIssuesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// AcceptanceIssuesColumns defines and stores column names for the table acceptance_issues.
type AcceptanceIssuesColumns struct {
	Id              string //
	ProjectId       string //
	AcceptanceRunId string //
	Severity        string //
	IssueKind       string //
	Blocking        string //
	Summary         string //
	DetailJson      string //
	CreatedAt       string //
}

// acceptanceIssuesColumns holds the columns for the table acceptance_issues.
var acceptanceIssuesColumns = AcceptanceIssuesColumns{
	Id:              "id",
	ProjectId:       "project_id",
	AcceptanceRunId: "acceptance_run_id",
	Severity:        "severity",
	IssueKind:       "issue_kind",
	Blocking:        "blocking",
	Summary:         "summary",
	DetailJson:      "detail_json",
	CreatedAt:       "created_at",
}

// NewAcceptanceIssuesDao creates and returns a new DAO object for table data access.
func NewAcceptanceIssuesDao(handlers ...gdb.ModelHandler) *AcceptanceIssuesDao {
	return &AcceptanceIssuesDao{
		group:    "default",
		table:    "acceptance_issues",
		columns:  acceptanceIssuesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AcceptanceIssuesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AcceptanceIssuesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AcceptanceIssuesDao) Columns() AcceptanceIssuesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AcceptanceIssuesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AcceptanceIssuesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AcceptanceIssuesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
