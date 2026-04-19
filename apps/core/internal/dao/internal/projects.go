// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProjectsDao is the data access object for the table projects.
type ProjectsDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  ProjectsColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// ProjectsColumns defines and stores column names for the table projects.
type ProjectsColumns struct {
	Id                    string //
	Name                  string //
	ProjectCategory       string //
	GoalSummary           string //
	Status                string //
	ProductionStatus      string //
	WorkspaceRoot         string //
	RepoRoot              string //
	CurrentPlanDraftId    string //
	CurrentCompiledPlanId string //
	CreatedAt             string //
	UpdatedAt             string //
}

// projectsColumns holds the columns for the table projects.
var projectsColumns = ProjectsColumns{
	Id:                    "id",
	Name:                  "name",
	ProjectCategory:       "project_category",
	GoalSummary:           "goal_summary",
	Status:                "status",
	ProductionStatus:      "production_status",
	WorkspaceRoot:         "workspace_root",
	RepoRoot:              "repo_root",
	CurrentPlanDraftId:    "current_plan_draft_id",
	CurrentCompiledPlanId: "current_compiled_plan_id",
	CreatedAt:             "created_at",
	UpdatedAt:             "updated_at",
}

// NewProjectsDao creates and returns a new DAO object for table data access.
func NewProjectsDao(handlers ...gdb.ModelHandler) *ProjectsDao {
	return &ProjectsDao{
		group:    "default",
		table:    "projects",
		columns:  projectsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProjectsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProjectsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProjectsDao) Columns() ProjectsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProjectsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProjectsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProjectsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
