// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpConversationDao is the data access object for the table mvp_conversation.
type MvpConversationDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  MvpConversationColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// MvpConversationColumns defines and stores column names for the table mvp_conversation.
type MvpConversationColumns struct {
	Id        string // 雪花ID
	ProjectId string // 项目ID
	TaskId    string // 关联任务ID，NULL=项目级对话
	Title     string // 对话标题
	RoleType  string // 对话角色类型
	Status    string // 状态：active/archived
	CreatedBy string // 创建人ID
	DeptId    string // 所属部门ID
	CreatedAt string // 创建时间
	UpdatedAt string // 更新时间
	DeletedAt string // 软删除时间
}

// mvpConversationColumns holds the columns for the table mvp_conversation.
var mvpConversationColumns = MvpConversationColumns{
	Id:        "id",
	ProjectId: "project_id",
	TaskId:    "task_id",
	Title:     "title",
	RoleType:  "role_type",
	Status:    "status",
	CreatedBy: "created_by",
	DeptId:    "dept_id",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewMvpConversationDao creates and returns a new DAO object for table data access.
func NewMvpConversationDao(handlers ...gdb.ModelHandler) *MvpConversationDao {
	return &MvpConversationDao{
		group:    "default",
		table:    "mvp_conversation",
		columns:  mvpConversationColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpConversationDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpConversationDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpConversationDao) Columns() MvpConversationColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpConversationDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpConversationDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpConversationDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
