// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpAcceptIssueDao is the data access object for the table mvp_accept_issue.
type MvpAcceptIssueDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  MvpAcceptIssueColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// MvpAcceptIssueColumns defines and stores column names for the table mvp_accept_issue.
type MvpAcceptIssueColumns struct {
	Id              string // 主键ID
	AcceptRunId     string // 验收运行ID
	WorkflowRunId   string // 工作流运行ID
	ProjectId       string // 项目ID
	DomainTaskId    string // 主关联任务ID
	IssueType       string // artifact/process/quality/risk
	RuleCode        string // 规则编码
	Severity        string // info/warn/error/blocker
	Title           string // 问题标题
	Detail          string // 问题详情
	ExpectedValue   string // 预期值
	ActualValue     string // 实际值
	SuggestedAction string // 建议动作
	ResourceRef     string // 关联资源引用(JSON)
	Status          string // open/resolved/ignored
	CreatedBy       string // 创建人
	DeptId          string // 部门ID
	CreatedAt       string // 创建时间
	UpdatedAt       string // 更新时间
	DeletedAt       string // 删除时间
}

// mvpAcceptIssueColumns holds the columns for the table mvp_accept_issue.
var mvpAcceptIssueColumns = MvpAcceptIssueColumns{
	Id:              "id",
	AcceptRunId:     "accept_run_id",
	WorkflowRunId:   "workflow_run_id",
	ProjectId:       "project_id",
	DomainTaskId:    "domain_task_id",
	IssueType:       "issue_type",
	RuleCode:        "rule_code",
	Severity:        "severity",
	Title:           "title",
	Detail:          "detail",
	ExpectedValue:   "expected_value",
	ActualValue:     "actual_value",
	SuggestedAction: "suggested_action",
	ResourceRef:     "resource_ref",
	Status:          "status",
	CreatedBy:       "created_by",
	DeptId:          "dept_id",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	DeletedAt:       "deleted_at",
}

// NewMvpAcceptIssueDao creates and returns a new DAO object for table data access.
func NewMvpAcceptIssueDao(handlers ...gdb.ModelHandler) *MvpAcceptIssueDao {
	return &MvpAcceptIssueDao{
		group:    "default",
		table:    "mvp_accept_issue",
		columns:  mvpAcceptIssueColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpAcceptIssueDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpAcceptIssueDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpAcceptIssueDao) Columns() MvpAcceptIssueColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpAcceptIssueDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpAcceptIssueDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpAcceptIssueDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
