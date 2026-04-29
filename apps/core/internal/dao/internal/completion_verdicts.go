// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CompletionVerdictsDao is the data access object for the table completion_verdicts.
type CompletionVerdictsDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  CompletionVerdictsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// CompletionVerdictsColumns defines and stores column names for the table completion_verdicts.
type CompletionVerdictsColumns struct {
	Id                     string //
	ProjectId              string //
	AcceptanceRunId        string //
	Decision               string //
	FinalStatus            string //
	Reason                 string //
	Summary                string //
	NextAction             string //
	BlockerCount           string //
	ReleaseReady           string //
	Completed              string //
	ManualReviewRequired   string //
	ManualReleaseRequired  string //
	ManualReleaseCompleted string //
	SourceRunId            string //
	CreatedAt              string //
	UpdatedAt              string //
	ExecutorSucceeded      string //
	DeliveryVerified       string //
	AcceptancePassed       string //
}

// completionVerdictsColumns holds the columns for the table completion_verdicts.
var completionVerdictsColumns = CompletionVerdictsColumns{
	Id:                     "id",
	ProjectId:              "project_id",
	AcceptanceRunId:        "acceptance_run_id",
	Decision:               "decision",
	FinalStatus:            "final_status",
	Reason:                 "reason",
	Summary:                "summary",
	NextAction:             "next_action",
	BlockerCount:           "blocker_count",
	ReleaseReady:           "release_ready",
	Completed:              "completed",
	ManualReviewRequired:   "manual_review_required",
	ManualReleaseRequired:  "manual_release_required",
	ManualReleaseCompleted: "manual_release_completed",
	SourceRunId:            "source_run_id",
	CreatedAt:              "created_at",
	UpdatedAt:              "updated_at",
	ExecutorSucceeded:      "executor_succeeded",
	DeliveryVerified:       "delivery_verified",
	AcceptancePassed:       "acceptance_passed",
}

// NewCompletionVerdictsDao creates and returns a new DAO object for table data access.
func NewCompletionVerdictsDao(handlers ...gdb.ModelHandler) *CompletionVerdictsDao {
	return &CompletionVerdictsDao{
		group:    "default",
		table:    "completion_verdicts",
		columns:  completionVerdictsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CompletionVerdictsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CompletionVerdictsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CompletionVerdictsDao) Columns() CompletionVerdictsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CompletionVerdictsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CompletionVerdictsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CompletionVerdictsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
