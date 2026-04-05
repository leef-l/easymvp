// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpAcceptEvidenceDao is the data access object for the table mvp_accept_evidence.
type MvpAcceptEvidenceDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  MvpAcceptEvidenceColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// MvpAcceptEvidenceColumns defines and stores column names for the table mvp_accept_evidence.
type MvpAcceptEvidenceColumns struct {
	Id           string // 主键ID
	AcceptRunId  string // 验收运行ID
	EvidenceType string // task_output/file/log/diff/stage_output/handoff/summary
	SourceType   string // domain_task/stage_run/file/handoff_record/workflow_run
	SourceId     string // 来源对象ID
	ContentRef   string // 证据引用或JSON
	Summary      string // 证据摘要
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
	DeletedAt    string // 删除时间
}

// mvpAcceptEvidenceColumns holds the columns for the table mvp_accept_evidence.
var mvpAcceptEvidenceColumns = MvpAcceptEvidenceColumns{
	Id:           "id",
	AcceptRunId:  "accept_run_id",
	EvidenceType: "evidence_type",
	SourceType:   "source_type",
	SourceId:     "source_id",
	ContentRef:   "content_ref",
	Summary:      "summary",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewMvpAcceptEvidenceDao creates and returns a new DAO object for table data access.
func NewMvpAcceptEvidenceDao(handlers ...gdb.ModelHandler) *MvpAcceptEvidenceDao {
	return &MvpAcceptEvidenceDao{
		group:    "default",
		table:    "mvp_accept_evidence",
		columns:  mvpAcceptEvidenceColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpAcceptEvidenceDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpAcceptEvidenceDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpAcceptEvidenceDao) Columns() MvpAcceptEvidenceColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpAcceptEvidenceDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpAcceptEvidenceDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpAcceptEvidenceDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
