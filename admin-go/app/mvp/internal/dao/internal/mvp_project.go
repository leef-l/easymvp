// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpProjectDao is the data access object for the table mvp_project.
type MvpProjectDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MvpProjectColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MvpProjectColumns defines and stores column names for the table mvp_project.
type MvpProjectColumns struct {
	Id                  string // 雪花ID
	Name                string // 项目名称
	ProjectCategory     string // 项目分类
	CategoryCode        string // 项目分类编码
	Description         string // 项目简介
	Status              string // 项目状态: designing/reviewing/running/paused/completed
	PauseReason         string // 暂停原因
	GlobalContext       string // 项目全局上下文（架构师需求分析+方案设计的压缩摘要）
	ObjectiveJson       string // 项目目标约束(JSON): budget/deadline/risk_tolerance/autonomy_level
	ArchitectModelId    string // 架构师使用的AI模型ID
	WorkDir             string // 项目代码工作目录（Aider执行路径）
	ActiveBatchNo       string // 当前活跃批次号（调度器持久化，0=无活跃批次）
	EngineVersion       string // 执行引擎版本: legacy/workflow_v2
	ActiveWorkflowRunId string // 当前活跃工作流运行ID
	CreatedBy           string // 创建人ID
	DeptId              string // 所属部门ID
	CreatedAt           string // 创建时间
	UpdatedAt           string // 更新时间
	DeletedAt           string // 软删除时间
}

// mvpProjectColumns holds the columns for the table mvp_project.
var mvpProjectColumns = MvpProjectColumns{
	Id:                  "id",
	Name:                "name",
	ProjectCategory:     "project_category",
	CategoryCode:        "category_code",
	Description:         "description",
	Status:              "status",
	PauseReason:         "pause_reason",
	GlobalContext:       "global_context",
	ObjectiveJson:       "objective_json",
	ArchitectModelId:    "architect_model_id",
	WorkDir:             "work_dir",
	ActiveBatchNo:       "active_batch_no",
	EngineVersion:       "engine_version",
	ActiveWorkflowRunId: "active_workflow_run_id",
	CreatedBy:           "created_by",
	DeptId:              "dept_id",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
	DeletedAt:           "deleted_at",
}

// NewMvpProjectDao creates and returns a new DAO object for table data access.
func NewMvpProjectDao(handlers ...gdb.ModelHandler) *MvpProjectDao {
	return &MvpProjectDao{
		group:    "default",
		table:    "mvp_project",
		columns:  mvpProjectColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpProjectDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpProjectDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpProjectDao) Columns() MvpProjectColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpProjectDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpProjectDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpProjectDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
