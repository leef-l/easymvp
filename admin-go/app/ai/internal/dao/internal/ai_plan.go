// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AiPlanDao is the data access object for the table ai_plan.
type AiPlanDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AiPlanColumns      // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AiPlanColumns defines and stores column names for the table ai_plan.
type AiPlanColumns struct {
	Id         string // 雪花ID
	ProviderId string // 供应商ID
	Name       string // 套餐名称
	Code       string // 套餐代码
	ApiKey     string // API Key（加密存储）
	ApiSecret  string // API Secret（部分供应商需要）
	Status     string // 状态:0=禁用,1=启用
	Sort       string // 排序
	CreatedBy  string // 创建人ID
	DeptId     string // 所属部门ID
	CreatedAt  string // 创建时间
	UpdatedAt  string // 更新时间
	DeletedAt  string // 软删除时间
}

// aiPlanColumns holds the columns for the table ai_plan.
var aiPlanColumns = AiPlanColumns{
	Id:         "id",
	ProviderId: "provider_id",
	Name:       "name",
	Code:       "code",
	ApiKey:     "api_key",
	ApiSecret:  "api_secret",
	Status:     "status",
	Sort:       "sort",
	CreatedBy:  "created_by",
	DeptId:     "dept_id",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	DeletedAt:  "deleted_at",
}

// NewAiPlanDao creates and returns a new DAO object for table data access.
func NewAiPlanDao(handlers ...gdb.ModelHandler) *AiPlanDao {
	return &AiPlanDao{
		group:    "default",
		table:    "ai_plan",
		columns:  aiPlanColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AiPlanDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AiPlanDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AiPlanDao) Columns() AiPlanColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AiPlanDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AiPlanDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AiPlanDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
