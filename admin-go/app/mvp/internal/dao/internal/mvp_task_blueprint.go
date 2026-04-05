// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpTaskBlueprintDao is the data access object for the table mvp_task_blueprint.
type MvpTaskBlueprintDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  MvpTaskBlueprintColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// MvpTaskBlueprintColumns defines and stores column names for the table mvp_task_blueprint.
type MvpTaskBlueprintColumns struct {
	Id                    string // 雪花ID
	PlanVersionId         string // 所属计划版本ID
	ParentBlueprintId     string // 父蓝图ID(支持层级)
	Name                  string // 任务名称
	Description           string // 任务描述
	RoleType              string // 角色类型: architect/implementer/auditor/coordinator
	RoleLevel             string // 角色等级: lite/pro/max
	BatchNo               string // 批次号
	Sort                  string // 排序
	AffectedResources     string // 影响资源列表(JSON)
	DependsOnBlueprintIds string // 依赖蓝图ID列表(JSON)
	BlueprintStatus       string // 蓝图状态: draft/confirmed/superseded
	CreatedAt             string // 创建时间
	UpdatedAt             string // 更新时间
	DeletedAt             string // 软删除时间
	CreatedBy             string // 创建人ID
	DeptId                string // 部门ID
}

// mvpTaskBlueprintColumns holds the columns for the table mvp_task_blueprint.
var mvpTaskBlueprintColumns = MvpTaskBlueprintColumns{
	Id:                    "id",
	PlanVersionId:         "plan_version_id",
	ParentBlueprintId:     "parent_blueprint_id",
	Name:                  "name",
	Description:           "description",
	RoleType:              "role_type",
	RoleLevel:             "role_level",
	BatchNo:               "batch_no",
	Sort:                  "sort",
	AffectedResources:     "affected_resources",
	DependsOnBlueprintIds: "depends_on_blueprint_ids",
	BlueprintStatus:       "blueprint_status",
	CreatedAt:             "created_at",
	UpdatedAt:             "updated_at",
	DeletedAt:             "deleted_at",
	CreatedBy:             "created_by",
	DeptId:                "dept_id",
}

// NewMvpTaskBlueprintDao creates and returns a new DAO object for table data access.
func NewMvpTaskBlueprintDao(handlers ...gdb.ModelHandler) *MvpTaskBlueprintDao {
	return &MvpTaskBlueprintDao{
		group:    "default",
		table:    "mvp_task_blueprint",
		columns:  mvpTaskBlueprintColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpTaskBlueprintDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpTaskBlueprintDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpTaskBlueprintDao) Columns() MvpTaskBlueprintColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpTaskBlueprintDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpTaskBlueprintDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpTaskBlueprintDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
