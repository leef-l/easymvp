// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AcceptanceJudgementsDao is the data access object for the table acceptance_judgements.
type AcceptanceJudgementsDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  AcceptanceJudgementsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// AcceptanceJudgementsColumns defines and stores column names for the table acceptance_judgements.
type AcceptanceJudgementsColumns struct {
	Id              string //
	ProjectId       string //
	AcceptanceRunId string //
	JudgementKind   string //
	JudgementResult string //
	Summary         string //
	DetailJson      string //
	CreatedAt       string //
}

// acceptanceJudgementsColumns holds the columns for the table acceptance_judgements.
var acceptanceJudgementsColumns = AcceptanceJudgementsColumns{
	Id:              "id",
	ProjectId:       "project_id",
	AcceptanceRunId: "acceptance_run_id",
	JudgementKind:   "judgement_kind",
	JudgementResult: "judgement_result",
	Summary:         "summary",
	DetailJson:      "detail_json",
	CreatedAt:       "created_at",
}

// NewAcceptanceJudgementsDao creates and returns a new DAO object for table data access.
func NewAcceptanceJudgementsDao(handlers ...gdb.ModelHandler) *AcceptanceJudgementsDao {
	return &AcceptanceJudgementsDao{
		group:    "default",
		table:    "acceptance_judgements",
		columns:  acceptanceJudgementsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AcceptanceJudgementsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AcceptanceJudgementsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AcceptanceJudgementsDao) Columns() AcceptanceJudgementsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AcceptanceJudgementsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AcceptanceJudgementsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AcceptanceJudgementsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
