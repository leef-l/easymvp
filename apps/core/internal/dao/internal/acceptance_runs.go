// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AcceptanceRunsDao is the data access object for the table acceptance_runs.
type AcceptanceRunsDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  AcceptanceRunsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// AcceptanceRunsColumns defines and stores column names for the table acceptance_runs.
type AcceptanceRunsColumns struct {
	Id                    string //
	ProjectId             string //
	TaskId                string //
	ProfileVersion        string //
	Status                string //
	FunctionalStatus      string //
	ProductionStatus      string //
	ManualReleaseRequired string //
	CreatedAt             string //
	FinishedAt            string //
}

// acceptanceRunsColumns holds the columns for the table acceptance_runs.
var acceptanceRunsColumns = AcceptanceRunsColumns{
	Id:                    "id",
	ProjectId:             "project_id",
	TaskId:                "task_id",
	ProfileVersion:        "profile_version",
	Status:                "status",
	FunctionalStatus:      "functional_status",
	ProductionStatus:      "production_status",
	ManualReleaseRequired: "manual_release_required",
	CreatedAt:             "created_at",
	FinishedAt:            "finished_at",
}

// NewAcceptanceRunsDao creates and returns a new DAO object for table data access.
func NewAcceptanceRunsDao(handlers ...gdb.ModelHandler) *AcceptanceRunsDao {
	return &AcceptanceRunsDao{
		group:    "default",
		table:    "acceptance_runs",
		columns:  acceptanceRunsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AcceptanceRunsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AcceptanceRunsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AcceptanceRunsDao) Columns() AcceptanceRunsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AcceptanceRunsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AcceptanceRunsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AcceptanceRunsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
