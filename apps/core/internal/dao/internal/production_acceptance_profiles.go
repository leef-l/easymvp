// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductionAcceptanceProfilesDao is the data access object for the table production_acceptance_profiles.
type ProductionAcceptanceProfilesDao struct {
	table    string                              // table is the underlying table name of the DAO.
	group    string                              // group is the database configuration group name of the current DAO.
	columns  ProductionAcceptanceProfilesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                  // handlers for customized model modification.
}

// ProductionAcceptanceProfilesColumns defines and stores column names for the table production_acceptance_profiles.
type ProductionAcceptanceProfilesColumns struct {
}

// productionAcceptanceProfilesColumns holds the columns for the table production_acceptance_profiles.
var productionAcceptanceProfilesColumns = ProductionAcceptanceProfilesColumns{}

// NewProductionAcceptanceProfilesDao creates and returns a new DAO object for table data access.
func NewProductionAcceptanceProfilesDao(handlers ...gdb.ModelHandler) *ProductionAcceptanceProfilesDao {
	return &ProductionAcceptanceProfilesDao{
		group:    "default",
		table:    "production_acceptance_profiles",
		columns:  productionAcceptanceProfilesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductionAcceptanceProfilesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductionAcceptanceProfilesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductionAcceptanceProfilesDao) Columns() ProductionAcceptanceProfilesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductionAcceptanceProfilesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductionAcceptanceProfilesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductionAcceptanceProfilesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
