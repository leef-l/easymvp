// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectProfilesDao is the data access object for the table project_profiles.
type ProjectProfilesDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  ProjectProfilesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// ProjectProfilesColumns defines and stores column names for the table project_profiles.
type ProjectProfilesColumns struct {
	Id                       string //
	ProjectId                string //
	CategoryProfileVersion   string //
	AcceptanceProfileVersion string //
	RoleProfileVersion       string //
	CreatedAt                string //
	UpdatedAt                string //
}

// projectProfilesColumns holds the columns for the table project_profiles.
var projectProfilesColumns = ProjectProfilesColumns{
	Id:                       "id",
	ProjectId:                "project_id",
	CategoryProfileVersion:   "category_profile_version",
	AcceptanceProfileVersion: "acceptance_profile_version",
	RoleProfileVersion:       "role_profile_version",
	CreatedAt:                "created_at",
	UpdatedAt:                "updated_at",
}

// NewProjectProfilesDao creates and returns a new DAO object for table data access.
func NewProjectProfilesDao(handlers ...gdb.ModelHandler) *ProjectProfilesDao {
	return &ProjectProfilesDao{
		group:    "default",
		table:    "project_profiles",
		columns:  projectProfilesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProjectProfilesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProjectProfilesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProjectProfilesDao) Columns() ProjectProfilesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProjectProfilesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProjectProfilesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProjectProfilesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
