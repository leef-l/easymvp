// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpAcceptRuleDao is the data access object for the table mvp_accept_rule.
type MvpAcceptRuleDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  MvpAcceptRuleColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// MvpAcceptRuleColumns defines and stores column names for the table mvp_accept_rule.
type MvpAcceptRuleColumns struct {
	Id          string // 主键ID
	ProjectType string // 项目类型模板
	RuleCode    string // 规则编码
	RuleName    string // 规则名称
	RuleType    string // artifact/process/quality
	ScopeType   string // project/task/file/stage
	ConfigJson  string // 规则配置
	Enabled     string // 是否启用
	Priority    string // 优先级(越小越先执行)
	CreatedAt   string // 创建时间
	UpdatedAt   string // 更新时间
	DeletedAt   string // 删除时间
}

// mvpAcceptRuleColumns holds the columns for the table mvp_accept_rule.
var mvpAcceptRuleColumns = MvpAcceptRuleColumns{
	Id:          "id",
	ProjectType: "project_type",
	RuleCode:    "rule_code",
	RuleName:    "rule_name",
	RuleType:    "rule_type",
	ScopeType:   "scope_type",
	ConfigJson:  "config_json",
	Enabled:     "enabled",
	Priority:    "priority",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewMvpAcceptRuleDao creates and returns a new DAO object for table data access.
func NewMvpAcceptRuleDao(handlers ...gdb.ModelHandler) *MvpAcceptRuleDao {
	return &MvpAcceptRuleDao{
		group:    "default",
		table:    "mvp_accept_rule",
		columns:  mvpAcceptRuleColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpAcceptRuleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpAcceptRuleDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpAcceptRuleDao) Columns() MvpAcceptRuleColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpAcceptRuleDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpAcceptRuleDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpAcceptRuleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
