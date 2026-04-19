// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowPlanReviewResultsDao is the data access object for the table workflow_plan_review_results.
type WorkflowPlanReviewResultsDao struct {
	table    string                           // table is the underlying table name of the DAO.
	group    string                           // group is the database configuration group name of the current DAO.
	columns  WorkflowPlanReviewResultsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler               // handlers for customized model modification.
}

// WorkflowPlanReviewResultsColumns defines and stores column names for the table workflow_plan_review_results.
type WorkflowPlanReviewResultsColumns struct {
	Id                      string //
	ProjectId               string //
	PlanDraftId             string //
	ReviewVersion           string //
	ReviewRunId             string //
	Decision                string //
	BlockingIssueCount      string //
	AdvisoryIssueCount      string //
	IssuesJson              string //
	SplitSuggestionsJson    string //
	OverrideSuggestionsJson string //
	Status                  string //
	ReviewedAt              string //
}

// workflowPlanReviewResultsColumns holds the columns for the table workflow_plan_review_results.
var workflowPlanReviewResultsColumns = WorkflowPlanReviewResultsColumns{
	Id:                      "id",
	ProjectId:               "project_id",
	PlanDraftId:             "plan_draft_id",
	ReviewVersion:           "review_version",
	ReviewRunId:             "review_run_id",
	Decision:                "decision",
	BlockingIssueCount:      "blocking_issue_count",
	AdvisoryIssueCount:      "advisory_issue_count",
	IssuesJson:              "issues_json",
	SplitSuggestionsJson:    "split_suggestions_json",
	OverrideSuggestionsJson: "override_suggestions_json",
	Status:                  "status",
	ReviewedAt:              "reviewed_at",
}

// NewWorkflowPlanReviewResultsDao creates and returns a new DAO object for table data access.
func NewWorkflowPlanReviewResultsDao(handlers ...gdb.ModelHandler) *WorkflowPlanReviewResultsDao {
	return &WorkflowPlanReviewResultsDao{
		group:    "default",
		table:    "workflow_plan_review_results",
		columns:  workflowPlanReviewResultsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WorkflowPlanReviewResultsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WorkflowPlanReviewResultsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WorkflowPlanReviewResultsDao) Columns() WorkflowPlanReviewResultsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WorkflowPlanReviewResultsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WorkflowPlanReviewResultsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WorkflowPlanReviewResultsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
