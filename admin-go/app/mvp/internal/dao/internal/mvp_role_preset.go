// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpRolePresetDao is the data access object for the table mvp_role_preset.
type MvpRolePresetDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  MvpRolePresetColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// MvpRolePresetColumns defines and stores column names for the table mvp_role_preset.
type MvpRolePresetColumns struct {
	Id              string // 雪花ID
	ProjectCategory string // 项目分类
	RoleType        string // 角色类型：architect/implementer/auditor/coordinator
	RoleLevel    string // 角色等级：lite/pro/max
	ModelId      string // AI模型ID
	SystemPrompt string // 默认系统提示词（角色设定）
	Status       string // 状态:0=禁用,1=启用
	Sort         string // 排序
	CreatedBy    string // 创建人ID
	DeptId       string // 所属部门ID
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
	DeletedAt    string // 软删除时间
}

// mvpRolePresetColumns holds the columns for the table mvp_role_preset.
var mvpRolePresetColumns = MvpRolePresetColumns{
	Id:              "id",
	ProjectCategory: "project_category",
	RoleType:        "role_type",
	RoleLevel:    "role_level",
	ModelId:      "model_id",
	SystemPrompt: "system_prompt",
	Status:       "status",
	Sort:         "sort",
	CreatedBy:    "created_by",
	DeptId:       "dept_id",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewMvpRolePresetDao creates and returns a new DAO object for table data access.
func NewMvpRolePresetDao(handlers ...gdb.ModelHandler) *MvpRolePresetDao {
	return &MvpRolePresetDao{
		group:    "default",
		table:    "mvp_role_preset",
		columns:  mvpRolePresetColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpRolePresetDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpRolePresetDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpRolePresetDao) Columns() MvpRolePresetColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpRolePresetDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpRolePresetDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpRolePresetDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
