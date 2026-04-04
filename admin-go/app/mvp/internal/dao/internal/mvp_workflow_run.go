// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpWorkflowRunDao is the data access object for the table mvp_workflow_run.
type MvpWorkflowRunDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  MvpWorkflowRunColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// MvpWorkflowRunColumns defines and stores column names for the table mvp_workflow_run.
type MvpWorkflowRunColumns struct {
	Id                  string // 雪花ID
	ProjectId           string // 所属项目ID
	RunNo               string // 项目内运行序号(从1递增)
	Status              string // 状态: designing/reviewing/executing/reworking/paused/completed/failed/canceled
	CurrentStage        string // 当前阶段: design/review/execute/rework/complete
	CurrentStageRunId   string // 当前阶段运行ID
	ActivePlanVersionId string // 当前活跃计划版本ID
	PauseReason         string // 暂停原因
	StatusBeforePause   string // 暂停前的阶段状态（恢复时回退）
	CancelReason        string // 取消原因
	RuntimeToken        string // 运行时令牌(防重入)
	StartedAt           string // 开始时间
	FinishedAt          string // 结束时间
	CreatedBy           string // 创建人ID
	DeptId              string // 所属部门ID
	CreatedAt           string // 创建时间
	UpdatedAt           string // 更新时间
	DeletedAt           string // 软删除时间
}

// mvpWorkflowRunColumns holds the columns for the table mvp_workflow_run.
var mvpWorkflowRunColumns = MvpWorkflowRunColumns{
	Id:                  "id",
	ProjectId:           "project_id",
	RunNo:               "run_no",
	Status:              "status",
	CurrentStage:        "current_stage",
	CurrentStageRunId:   "current_stage_run_id",
	ActivePlanVersionId: "active_plan_version_id",
	PauseReason:         "pause_reason",
	StatusBeforePause:   "status_before_pause",
	CancelReason:        "cancel_reason",
	RuntimeToken:        "runtime_token",
	StartedAt:           "started_at",
	FinishedAt:          "finished_at",
	CreatedBy:           "created_by",
	DeptId:              "dept_id",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
	DeletedAt:           "deleted_at",
}

// NewMvpWorkflowRunDao creates and returns a new DAO object for table data access.
func NewMvpWorkflowRunDao(handlers ...gdb.ModelHandler) *MvpWorkflowRunDao {
	return &MvpWorkflowRunDao{
		group:    "default",
		table:    "mvp_workflow_run",
		columns:  mvpWorkflowRunColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpWorkflowRunDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpWorkflowRunDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpWorkflowRunDao) Columns() MvpWorkflowRunColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpWorkflowRunDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpWorkflowRunDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpWorkflowRunDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
