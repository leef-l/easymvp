// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// RequirementsDao is the data access object for the table requirements.
type RequirementsDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  RequirementsColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// RequirementsColumns defines and stores column names for the table requirements.
type RequirementsColumns struct {
	Id                 string //
	ProjectId          string //
	RawInput           string //
	Status             string //
	RequirementDocJson string //
	UserConfirmed      string //
	ConfirmedAt        string //
	CreatedAt          string //
	UpdatedAt          string //
}

// requirementsColumns holds the columns for the table requirements.
var requirementsColumns = RequirementsColumns{
	Id:                 "id",
	ProjectId:          "project_id",
	RawInput:           "raw_input",
	Status:             "status",
	RequirementDocJson: "requirement_doc_json",
	UserConfirmed:      "user_confirmed",
	ConfirmedAt:        "confirmed_at",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

// NewRequirementsDao creates and returns a new DAO object for table data access.
func NewRequirementsDao(handlers ...gdb.ModelHandler) *RequirementsDao {
	return &RequirementsDao{
		group:    "default",
		table:    "requirements",
		columns:  requirementsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *RequirementsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *RequirementsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *RequirementsDao) Columns() RequirementsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *RequirementsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *RequirementsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *RequirementsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
