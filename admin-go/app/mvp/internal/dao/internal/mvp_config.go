// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpConfigDao is the data access object for the table mvp_config.
type MvpConfigDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MvpConfigColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MvpConfigColumns defines and stores column names for the table mvp_config.
type MvpConfigColumns struct {
	Id          string // 自增ID
	ConfigKey   string // 配置键（唯一）
	ConfigValue string // 配置值
	ConfigType  string // 值类型：string/int/float/bool/json
	Category    string // 分类：engine/watchdog/scheduler/general
	Description string // 配置说明
	CreatedBy   string // 创建人
	DeptId      string // 部门ID
	CreatedAt   string // 创建时间
	UpdatedAt   string // 更新时间
	DeletedAt   string // 软删除时间
}

// mvpConfigColumns holds the columns for the table mvp_config.
var mvpConfigColumns = MvpConfigColumns{
	Id:          "id",
	ConfigKey:   "config_key",
	ConfigValue: "config_value",
	ConfigType:  "config_type",
	Category:    "category",
	Description: "description",
	CreatedBy:   "created_by",
	DeptId:      "dept_id",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
	DeletedAt:   "deleted_at",
}

// NewMvpConfigDao creates and returns a new DAO object for table data access.
func NewMvpConfigDao(handlers ...gdb.ModelHandler) *MvpConfigDao {
	return &MvpConfigDao{
		group:    "default",
		table:    "mvp_config",
		columns:  mvpConfigColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpConfigDao) Columns() MvpConfigColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpConfigDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
