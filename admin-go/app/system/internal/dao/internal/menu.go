// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MenuDao is the data access object for the table system_menu.
type MenuDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MenuColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MenuColumns defines and stores column names for the table system_menu.
type MenuColumns struct {
	Id         string // 菜单ID（Snowflake）
	ParentId   string // 上级菜单ID，0 表示顶级菜单
	Title      string // 菜单名称
	Type       string // 类型:1=目录,2=菜单,3=按钮,4=外链,5=内链
	Path       string // 前端路由路径
	Component  string // 前端组件路径
	Permission string // 权限标识（如 system:dept:list）
	Icon       string // 菜单图标（图标名称）
	Sort       string // 排序（升序）
	IsShow     string // 是否显示:0=隐藏,1=显示
	IsCache    string // 是否缓存:0=不缓存,1=缓存
	LinkUrl    string // 外链/内链地址（type=4或5时有效）
	Status     string // 状态:0=关闭,1=开启
	CreatedBy  string // 创建人ID
	DeptId     string // 所属部门ID
	CreatedAt  string // 创建时间
	UpdatedAt  string // 更新时间
	DeletedAt  string // 软删除时间，非 NULL 表示已删除
}

// menuColumns holds the columns for the table system_menu.
var menuColumns = MenuColumns{
	Id:         "id",
	ParentId:   "parent_id",
	Title:      "title",
	Type:       "type",
	Path:       "path",
	Component:  "component",
	Permission: "permission",
	Icon:       "icon",
	Sort:       "sort",
	IsShow:     "is_show",
	IsCache:    "is_cache",
	LinkUrl:    "link_url",
	Status:     "status",
	CreatedBy:  "created_by",
	DeptId:     "dept_id",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
	DeletedAt:  "deleted_at",
}

// NewMenuDao creates and returns a new DAO object for table data access.
func NewMenuDao(handlers ...gdb.ModelHandler) *MenuDao {
	return &MenuDao{
		group:    "default",
		table:    "system_menu",
		columns:  menuColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MenuDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MenuDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MenuDao) Columns() MenuColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MenuDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MenuDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MenuDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
