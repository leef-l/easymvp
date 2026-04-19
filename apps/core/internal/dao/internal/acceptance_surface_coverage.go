// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AcceptanceSurfaceCoverageDao is the data access object for the table acceptance_surface_coverage.
type AcceptanceSurfaceCoverageDao struct {
	table    string                           // table is the underlying table name of the DAO.
	group    string                           // group is the database configuration group name of the current DAO.
	columns  AcceptanceSurfaceCoverageColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler               // handlers for customized model modification.
}

// AcceptanceSurfaceCoverageColumns defines and stores column names for the table acceptance_surface_coverage.
type AcceptanceSurfaceCoverageColumns struct {
	Id              string //
	ProjectId       string //
	AcceptanceRunId string //
	Surface         string //
	CoverageStatus  string //
	EvidenceCount   string //
	CreatedAt       string //
	UpdatedAt       string //
}

// acceptanceSurfaceCoverageColumns holds the columns for the table acceptance_surface_coverage.
var acceptanceSurfaceCoverageColumns = AcceptanceSurfaceCoverageColumns{
	Id:              "id",
	ProjectId:       "project_id",
	AcceptanceRunId: "acceptance_run_id",
	Surface:         "surface",
	CoverageStatus:  "coverage_status",
	EvidenceCount:   "evidence_count",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
}

// NewAcceptanceSurfaceCoverageDao creates and returns a new DAO object for table data access.
func NewAcceptanceSurfaceCoverageDao(handlers ...gdb.ModelHandler) *AcceptanceSurfaceCoverageDao {
	return &AcceptanceSurfaceCoverageDao{
		group:    "default",
		table:    "acceptance_surface_coverage",
		columns:  acceptanceSurfaceCoverageColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AcceptanceSurfaceCoverageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AcceptanceSurfaceCoverageDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AcceptanceSurfaceCoverageDao) Columns() AcceptanceSurfaceCoverageColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AcceptanceSurfaceCoverageDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AcceptanceSurfaceCoverageDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AcceptanceSurfaceCoverageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
