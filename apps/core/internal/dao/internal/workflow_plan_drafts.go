// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowPlanDraftsDao is the data access object for the table workflow_plan_drafts.
type WorkflowPlanDraftsDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  WorkflowPlanDraftsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// WorkflowPlanDraftsColumns defines and stores column names for the table workflow_plan_drafts.
type WorkflowPlanDraftsColumns struct {
	Id                    string //
	ProjectId             string //
	Version               string //
	SourceKind            string //
	SourceRunId           string //
	ProjectCategory       string //
	GoalSummary           string //
	InputRequirementsJson string //
	DraftTasksJson        string //
	Status                string //
	CreatedBy             string //
	CreatedAt             string //
	UpdatedAt             string //
}

// workflowPlanDraftsColumns holds the columns for the table workflow_plan_drafts.
var workflowPlanDraftsColumns = WorkflowPlanDraftsColumns{
	Id:                    "id",
	ProjectId:             "project_id",
	Version:               "version",
	SourceKind:            "source_kind",
	SourceRunId:           "source_run_id",
	ProjectCategory:       "project_category",
	GoalSummary:           "goal_summary",
	InputRequirementsJson: "input_requirements_json",
	DraftTasksJson:        "draft_tasks_json",
	Status:                "status",
	CreatedBy:             "created_by",
	CreatedAt:             "created_at",
	UpdatedAt:             "updated_at",
}

// NewWorkflowPlanDraftsDao creates and returns a new DAO object for table data access.
func NewWorkflowPlanDraftsDao(handlers ...gdb.ModelHandler) *WorkflowPlanDraftsDao {
	return &WorkflowPlanDraftsDao{
		group:    "default",
		table:    "workflow_plan_drafts",
		columns:  workflowPlanDraftsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WorkflowPlanDraftsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WorkflowPlanDraftsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WorkflowPlanDraftsDao) Columns() WorkflowPlanDraftsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WorkflowPlanDraftsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WorkflowPlanDraftsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WorkflowPlanDraftsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
