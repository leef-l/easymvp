// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpHandoffRecordDao is the data access object for the table mvp_handoff_record.
type MvpHandoffRecordDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  MvpHandoffRecordColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// MvpHandoffRecordColumns defines and stores column names for the table mvp_handoff_record.
type MvpHandoffRecordColumns struct {
	Id            string // 雪花ID
	WorkflowRunId string // 所属工作流运行ID
	FromTaskId    string // 来源任务ID
	ToTaskId      string // 目标任务ID
	HandoffType   string // 交接类型: bug_fix/failure_escalation/rework/audit
	Reason        string // 交接原因
	Payload       string // 交接载荷(JSON)
	CreatedAt     string // 创建时间
}

// mvpHandoffRecordColumns holds the columns for the table mvp_handoff_record.
var mvpHandoffRecordColumns = MvpHandoffRecordColumns{
	Id:            "id",
	WorkflowRunId: "workflow_run_id",
	FromTaskId:    "from_task_id",
	ToTaskId:      "to_task_id",
	HandoffType:   "handoff_type",
	Reason:        "reason",
	Payload:       "payload",
	CreatedAt:     "created_at",
}

// NewMvpHandoffRecordDao creates and returns a new DAO object for table data access.
func NewMvpHandoffRecordDao(handlers ...gdb.ModelHandler) *MvpHandoffRecordDao {
	return &MvpHandoffRecordDao{
		group:    "default",
		table:    "mvp_handoff_record",
		columns:  mvpHandoffRecordColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpHandoffRecordDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpHandoffRecordDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpHandoffRecordDao) Columns() MvpHandoffRecordColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpHandoffRecordDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpHandoffRecordDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpHandoffRecordDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
