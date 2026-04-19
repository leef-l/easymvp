// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AcceptanceJourneyCoverageDao is the data access object for the table acceptance_journey_coverage.
type AcceptanceJourneyCoverageDao struct {
	table    string                           // table is the underlying table name of the DAO.
	group    string                           // group is the database configuration group name of the current DAO.
	columns  AcceptanceJourneyCoverageColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler               // handlers for customized model modification.
}

// AcceptanceJourneyCoverageColumns defines and stores column names for the table acceptance_journey_coverage.
type AcceptanceJourneyCoverageColumns struct {
	Id              string //
	ProjectId       string //
	AcceptanceRunId string //
	Journey         string //
	CoverageStatus  string //
	EvidenceCount   string //
	CreatedAt       string //
	UpdatedAt       string //
}

// acceptanceJourneyCoverageColumns holds the columns for the table acceptance_journey_coverage.
var acceptanceJourneyCoverageColumns = AcceptanceJourneyCoverageColumns{
	Id:              "id",
	ProjectId:       "project_id",
	AcceptanceRunId: "acceptance_run_id",
	Journey:         "journey",
	CoverageStatus:  "coverage_status",
	EvidenceCount:   "evidence_count",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
}

// NewAcceptanceJourneyCoverageDao creates and returns a new DAO object for table data access.
func NewAcceptanceJourneyCoverageDao(handlers ...gdb.ModelHandler) *AcceptanceJourneyCoverageDao {
	return &AcceptanceJourneyCoverageDao{
		group:    "default",
		table:    "acceptance_journey_coverage",
		columns:  acceptanceJourneyCoverageColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AcceptanceJourneyCoverageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AcceptanceJourneyCoverageDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AcceptanceJourneyCoverageDao) Columns() AcceptanceJourneyCoverageColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AcceptanceJourneyCoverageDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AcceptanceJourneyCoverageDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AcceptanceJourneyCoverageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
