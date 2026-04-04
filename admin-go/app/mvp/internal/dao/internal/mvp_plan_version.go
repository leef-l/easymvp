// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpPlanVersionDao is the data access object for the table mvp_plan_version.
type MvpPlanVersionDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  MvpPlanVersionColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// MvpPlanVersionColumns defines and stores column names for the table mvp_plan_version.
type MvpPlanVersionColumns struct {
	Id                   string // 雪花ID
	ProjectId            string // 所属项目ID
	WorkflowRunId        string // 所属工作流运行ID
	VersionNo            string // 版本号(项目内递增)
	SourceConversationId string // 来源对话ID
	SourceMessageId      string // 来源消息ID
	Status               string // 版本状态: draft/active/superseded
	ReviewStatus         string // 审核状态: pending/approved/rejected
	Summary              string // 版本摘要
	DiffSummary          string // 与上一版本的差异摘要
	ApprovedAt           string // 审核通过时间
	RejectedAt           string // 审核驳回时间
	CreatedAt            string // 创建时间
	UpdatedAt            string // 更新时间
	DeletedAt            string // 软删除时间
}

// mvpPlanVersionColumns holds the columns for the table mvp_plan_version.
var mvpPlanVersionColumns = MvpPlanVersionColumns{
	Id:                   "id",
	ProjectId:            "project_id",
	WorkflowRunId:        "workflow_run_id",
	VersionNo:            "version_no",
	SourceConversationId: "source_conversation_id",
	SourceMessageId:      "source_message_id",
	Status:               "status",
	ReviewStatus:         "review_status",
	Summary:              "summary",
	DiffSummary:          "diff_summary",
	ApprovedAt:           "approved_at",
	RejectedAt:           "rejected_at",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
	DeletedAt:            "deleted_at",
}

// NewMvpPlanVersionDao creates and returns a new DAO object for table data access.
func NewMvpPlanVersionDao(handlers ...gdb.ModelHandler) *MvpPlanVersionDao {
	return &MvpPlanVersionDao{
		group:    "default",
		table:    "mvp_plan_version",
		columns:  mvpPlanVersionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpPlanVersionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpPlanVersionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpPlanVersionDao) Columns() MvpPlanVersionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpPlanVersionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpPlanVersionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpPlanVersionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
