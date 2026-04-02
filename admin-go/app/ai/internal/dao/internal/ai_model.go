// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AiModelDao is the data access object for the table ai_model.
type AiModelDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AiModelColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AiModelColumns defines and stores column names for the table ai_model.
type AiModelColumns struct {
	Id             string // 雪花ID
	PlanId         string // 套餐ID
	ProviderId     string // 供应商ID（冗余便于查询）
	Name           string // 模型显示名称
	ModelCode      string // 模型代码（API调用用）
	Capability     string // 能力：chat/reasoning/coding
	MaxTokens      string // 最大输出token
	ContextWindow  string // 上下文窗口大小
	SupportsStream string // 是否支持流式输出:0=否,1=是
	RolePrompt     string // 默认角色提示词
	Status         string // 状态:0=禁用,1=启用
	Sort           string // 排序
	CreatedBy      string // 创建人ID
	DeptId         string // 所属部门ID
	CreatedAt      string // 创建时间
	UpdatedAt      string // 更新时间
	DeletedAt      string // 软删除时间
}

// aiModelColumns holds the columns for the table ai_model.
var aiModelColumns = AiModelColumns{
	Id:             "id",
	PlanId:         "plan_id",
	ProviderId:     "provider_id",
	Name:           "name",
	ModelCode:      "model_code",
	Capability:     "capability",
	MaxTokens:      "max_tokens",
	ContextWindow:  "context_window",
	SupportsStream: "supports_stream",
	RolePrompt:     "role_prompt",
	Status:         "status",
	Sort:           "sort",
	CreatedBy:      "created_by",
	DeptId:         "dept_id",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewAiModelDao creates and returns a new DAO object for table data access.
func NewAiModelDao(handlers ...gdb.ModelHandler) *AiModelDao {
	return &AiModelDao{
		group:    "default",
		table:    "ai_model",
		columns:  aiModelColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AiModelDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AiModelDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AiModelDao) Columns() AiModelColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AiModelDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AiModelDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AiModelDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
