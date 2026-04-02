// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// DeptDao is the data access object for the table system_dept.
type DeptDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  DeptColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// DeptColumns defines and stores column names for the table system_dept.
type DeptColumns struct {
	Id        string // 部门ID（Snowflake）
	ParentId  string // 上级部门ID，0 表示顶级部门
	Title     string // 部门名称
	Username  string // 部门负责人姓名
	Email     string // 负责人邮箱
	Sort      string // 排序（升序）
	Status    string // 状态:0=关闭,1=开启
	CreatedBy string // 创建人ID
	DeptId    string // 所属部门ID
	CreatedAt string // 创建时间
	UpdatedAt string // 更新时间
	DeletedAt string // 软删除时间，非 NULL 表示已删除
}

// deptColumns holds the columns for the table system_dept.
var deptColumns = DeptColumns{
	Id:        "id",
	ParentId:  "parent_id",
	Title:     "title",
	Username:  "username",
	Email:     "email",
	Sort:      "sort",
	Status:    "status",
	CreatedBy: "created_by",
	DeptId:    "dept_id",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewDeptDao creates and returns a new DAO object for table data access.
func NewDeptDao(handlers ...gdb.ModelHandler) *DeptDao {
	return &DeptDao{
		group:    "default",
		table:    "system_dept",
		columns:  deptColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *DeptDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *DeptDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *DeptDao) Columns() DeptColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *DeptDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *DeptDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *DeptDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
