// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpStageRunDao is the data access object for the table mvp_stage_run.
type MvpStageRunDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MvpStageRunColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MvpStageRunColumns defines and stores column names for the table mvp_stage_run.
type MvpStageRunColumns struct {
	Id            string // 雪花ID
	WorkflowRunId string // 所属工作流运行ID
	StageType     string // 阶段类型: design/review/execute/rework/complete
	StageNo       string // 同类型阶段序号(支持多轮)
	Status        string // 状态: pending/running/completed/failed/skipped
	InputRef      string // 阶段输入引用(JSON)
	OutputRef     string // 阶段输出引用(JSON)
	Decision      string // 阶段决策结果(JSON)
	ErrorMessage  string // 错误信息
	StartedAt     string // 开始时间
	FinishedAt    string // 结束时间
	CreatedAt     string // 创建时间
	UpdatedAt     string // 更新时间
	DeletedAt     string // 软删除时间
}

// mvpStageRunColumns holds the columns for the table mvp_stage_run.
var mvpStageRunColumns = MvpStageRunColumns{
	Id:            "id",
	WorkflowRunId: "workflow_run_id",
	StageType:     "stage_type",
	StageNo:       "stage_no",
	Status:        "status",
	InputRef:      "input_ref",
	OutputRef:     "output_ref",
	Decision:      "decision",
	ErrorMessage:  "error_message",
	StartedAt:     "started_at",
	FinishedAt:    "finished_at",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
	DeletedAt:     "deleted_at",
}

// NewMvpStageRunDao creates and returns a new DAO object for table data access.
func NewMvpStageRunDao(handlers ...gdb.ModelHandler) *MvpStageRunDao {
	return &MvpStageRunDao{
		group:    "default",
		table:    "mvp_stage_run",
		columns:  mvpStageRunColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpStageRunDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpStageRunDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpStageRunDao) Columns() MvpStageRunColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpStageRunDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpStageRunDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpStageRunDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
