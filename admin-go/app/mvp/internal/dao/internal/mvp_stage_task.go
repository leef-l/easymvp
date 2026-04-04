// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpStageTaskDao is the data access object for the table mvp_stage_task.
type MvpStageTaskDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  MvpStageTaskColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// MvpStageTaskColumns defines and stores column names for the table mvp_stage_task.
type MvpStageTaskColumns struct {
	Id            string // 雪花ID
	StageRunId    string // 所属阶段运行ID
	TaskType      string // 任务类型: precheck/auditor_review/coordinator_optimize/review_summary
	RoleType      string // 执行角色
	Status        string // 状态: pending/running/completed/failed/skipped
	InputPayload  string // 输入载荷(JSON)
	OutputPayload string // 输出载荷(JSON)
	ErrorMessage  string // 错误信息
	StartedAt     string // 开始时间
	CompletedAt   string // 完成时间
	CreatedAt     string // 创建时间
	UpdatedAt     string // 更新时间
	DeletedAt     string // 软删除时间
}

// mvpStageTaskColumns holds the columns for the table mvp_stage_task.
var mvpStageTaskColumns = MvpStageTaskColumns{
	Id:            "id",
	StageRunId:    "stage_run_id",
	TaskType:      "task_type",
	RoleType:      "role_type",
	Status:        "status",
	InputPayload:  "input_payload",
	OutputPayload: "output_payload",
	ErrorMessage:  "error_message",
	StartedAt:     "started_at",
	CompletedAt:   "completed_at",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
	DeletedAt:     "deleted_at",
}

// NewMvpStageTaskDao creates and returns a new DAO object for table data access.
func NewMvpStageTaskDao(handlers ...gdb.ModelHandler) *MvpStageTaskDao {
	return &MvpStageTaskDao{
		group:    "default",
		table:    "mvp_stage_task",
		columns:  mvpStageTaskColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpStageTaskDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpStageTaskDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpStageTaskDao) Columns() MvpStageTaskColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpStageTaskDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpStageTaskDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpStageTaskDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
