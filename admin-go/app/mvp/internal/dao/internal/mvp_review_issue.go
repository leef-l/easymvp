// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpReviewIssueDao is the data access object for the table mvp_review_issue.
type MvpReviewIssueDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  MvpReviewIssueColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// MvpReviewIssueColumns defines and stores column names for the table mvp_review_issue.
type MvpReviewIssueColumns struct {
	Id            string // 雪花ID
	WorkflowRunId string // 所属工作流运行ID
	StageRunId    string // 所属阶段运行ID
	PlanVersionId string // 所属计划版本ID
	BlueprintId   string // 关联蓝图ID
	Severity      string // 严重级别: error/warning/info
	IssueCode     string // 问题代码
	IssueType     string // 问题类型
	SourceRole    string // 发现角色
	TaskName      string // 关联任务名
	Message       string // 问题描述
	Suggestion    string // 修复建议
	Status        string // 状态: open/resolved/ignored
	ResolvedAt    string // 解决时间
	CreatedAt     string // 创建时间
	UpdatedAt     string // 更新时间
	DeletedAt     string // 软删除时间
}

// mvpReviewIssueColumns holds the columns for the table mvp_review_issue.
var mvpReviewIssueColumns = MvpReviewIssueColumns{
	Id:            "id",
	WorkflowRunId: "workflow_run_id",
	StageRunId:    "stage_run_id",
	PlanVersionId: "plan_version_id",
	BlueprintId:   "blueprint_id",
	Severity:      "severity",
	IssueCode:     "issue_code",
	IssueType:     "issue_type",
	SourceRole:    "source_role",
	TaskName:      "task_name",
	Message:       "message",
	Suggestion:    "suggestion",
	Status:        "status",
	ResolvedAt:    "resolved_at",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
	DeletedAt:     "deleted_at",
}

// NewMvpReviewIssueDao creates and returns a new DAO object for table data access.
func NewMvpReviewIssueDao(handlers ...gdb.ModelHandler) *MvpReviewIssueDao {
	return &MvpReviewIssueDao{
		group:    "default",
		table:    "mvp_review_issue",
		columns:  mvpReviewIssueColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpReviewIssueDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpReviewIssueDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpReviewIssueDao) Columns() MvpReviewIssueColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpReviewIssueDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpReviewIssueDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpReviewIssueDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
