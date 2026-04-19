// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowCompiledPlansDao is the data access object for the table workflow_compiled_plans.
type WorkflowCompiledPlansDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  WorkflowCompiledPlansColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// WorkflowCompiledPlansColumns defines and stores column names for the table workflow_compiled_plans.
type WorkflowCompiledPlansColumns struct {
	Id                 string //
	ProjectId          string //
	PlanDraftId        string //
	PlanReviewResultId string //
	CompiledVersion    string //
	CompileRunId       string //
	ProjectCategory    string //
	Status             string //
	RiskSummaryJson    string //
	CompileDiffJson    string //
	GeneratedAt        string //
	ActivatedAt        string //
}

// workflowCompiledPlansColumns holds the columns for the table workflow_compiled_plans.
var workflowCompiledPlansColumns = WorkflowCompiledPlansColumns{
	Id:                 "id",
	ProjectId:          "project_id",
	PlanDraftId:        "plan_draft_id",
	PlanReviewResultId: "plan_review_result_id",
	CompiledVersion:    "compiled_version",
	CompileRunId:       "compile_run_id",
	ProjectCategory:    "project_category",
	Status:             "status",
	RiskSummaryJson:    "risk_summary_json",
	CompileDiffJson:    "compile_diff_json",
	GeneratedAt:        "generated_at",
	ActivatedAt:        "activated_at",
}

// NewWorkflowCompiledPlansDao creates and returns a new DAO object for table data access.
func NewWorkflowCompiledPlansDao(handlers ...gdb.ModelHandler) *WorkflowCompiledPlansDao {
	return &WorkflowCompiledPlansDao{
		group:    "default",
		table:    "workflow_compiled_plans",
		columns:  workflowCompiledPlansColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WorkflowCompiledPlansDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WorkflowCompiledPlansDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WorkflowCompiledPlansDao) Columns() WorkflowCompiledPlansColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WorkflowCompiledPlansDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WorkflowCompiledPlansDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WorkflowCompiledPlansDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
