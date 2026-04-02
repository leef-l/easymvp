// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpMessageDao is the data access object for the table mvp_message.
type MvpMessageDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MvpMessageColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MvpMessageColumns defines and stores column names for the table mvp_message.
type MvpMessageColumns struct {
	Id             string // 雪花ID
	ConversationId string // 对话ID
	Role           string // 消息角色：user/assistant/system
	Content        string // 消息内容
	ModelId        string // 使用的AI模型ID
	TokenUsage     string // token消耗：{prompt_tokens, completion_tokens}
	Status         string // 状态：pending/streaming/completed/failed
	CreatedBy      string // 创建人ID
	DeptId         string // 所属部门ID
	CreatedAt      string // 创建时间
	UpdatedAt      string // 更新时间
	DeletedAt      string // 软删除时间
}

// mvpMessageColumns holds the columns for the table mvp_message.
var mvpMessageColumns = MvpMessageColumns{
	Id:             "id",
	ConversationId: "conversation_id",
	Role:           "role",
	Content:        "content",
	ModelId:        "model_id",
	TokenUsage:     "token_usage",
	Status:         "status",
	CreatedBy:      "created_by",
	DeptId:         "dept_id",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewMvpMessageDao creates and returns a new DAO object for table data access.
func NewMvpMessageDao(handlers ...gdb.ModelHandler) *MvpMessageDao {
	return &MvpMessageDao{
		group:    "default",
		table:    "mvp_message",
		columns:  mvpMessageColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpMessageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpMessageDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpMessageDao) Columns() MvpMessageColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpMessageDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpMessageDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpMessageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
