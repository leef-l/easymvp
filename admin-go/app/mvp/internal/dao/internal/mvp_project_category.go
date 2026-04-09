// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpProjectCategoryDao is the data access object for the table mvp_project_category.
type MvpProjectCategoryDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  MvpProjectCategoryColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// MvpProjectCategoryColumns defines and stores column names for the table mvp_project_category.
type MvpProjectCategoryColumns struct {
	Id                      string // 主键ID
	CategoryCode            string // 稳定分类编码
	DisplayName             string // 展示名称
	FamilyCode              string // 能力家族编码
	Description             string // 分类说明
	VerificationProfileJson string // 分类默认验证配置(JSON)
	VerificationGateJson    string // 分类验证放行规则(JSON)
	Status                  string // 1启用 0停用
	Sort                    string // 排序
	CreatedBy               string // 创建人
	DeptId                  string // 部门ID
	CreatedAt               string // 创建时间
	UpdatedAt               string // 更新时间
	DeletedAt               string // 软删除时间
}

// mvpProjectCategoryColumns holds the columns for the table mvp_project_category.
var mvpProjectCategoryColumns = MvpProjectCategoryColumns{
	Id:                      "id",
	CategoryCode:            "category_code",
	DisplayName:             "display_name",
	FamilyCode:              "family_code",
	Description:             "description",
	VerificationProfileJson: "verification_profile_json",
	VerificationGateJson:    "verification_gate_json",
	Status:                  "status",
	Sort:                    "sort",
	CreatedBy:               "created_by",
	DeptId:                  "dept_id",
	CreatedAt:               "created_at",
	UpdatedAt:               "updated_at",
	DeletedAt:               "deleted_at",
}

// NewMvpProjectCategoryDao creates and returns a new DAO object for table data access.
func NewMvpProjectCategoryDao(handlers ...gdb.ModelHandler) *MvpProjectCategoryDao {
	return &MvpProjectCategoryDao{
		group:    "default",
		table:    "mvp_project_category",
		columns:  mvpProjectCategoryColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpProjectCategoryDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpProjectCategoryDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpProjectCategoryDao) Columns() MvpProjectCategoryColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpProjectCategoryDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpProjectCategoryDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpProjectCategoryDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
