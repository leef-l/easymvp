// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectWorkspacesDao is the data access object for the table project_workspaces.
type ProjectWorkspacesDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  ProjectWorkspacesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// ProjectWorkspacesColumns defines and stores column names for the table project_workspaces.
type ProjectWorkspacesColumns struct {
	Id              string //
	ProjectId       string //
	WorkspaceRoot   string //
	EvidenceRoot    string //
	RunsRoot        string //
	ReplayRoot      string //
	DiagnosticsRoot string //
	CreatedAt       string //
	UpdatedAt       string //
}

// projectWorkspacesColumns holds the columns for the table project_workspaces.
var projectWorkspacesColumns = ProjectWorkspacesColumns{
	Id:              "id",
	ProjectId:       "project_id",
	WorkspaceRoot:   "workspace_root",
	EvidenceRoot:    "evidence_root",
	RunsRoot:        "runs_root",
	ReplayRoot:      "replay_root",
	DiagnosticsRoot: "diagnostics_root",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
}

// NewProjectWorkspacesDao creates and returns a new DAO object for table data access.
func NewProjectWorkspacesDao(handlers ...gdb.ModelHandler) *ProjectWorkspacesDao {
	return &ProjectWorkspacesDao{
		group:    "default",
		table:    "project_workspaces",
		columns:  projectWorkspacesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProjectWorkspacesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProjectWorkspacesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProjectWorkspacesDao) Columns() ProjectWorkspacesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProjectWorkspacesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProjectWorkspacesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProjectWorkspacesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
