// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// EvidenceItemsDao is the data access object for the table evidence_items.
type EvidenceItemsDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  EvidenceItemsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// EvidenceItemsColumns defines and stores column names for the table evidence_items.
type EvidenceItemsColumns struct {
	Id           string //
	ProjectId    string //
	RunId        string //
	Surface      string //
	Journey      string //
	EvidenceType string //
	FilePath     string //
	ContentHash  string //
	FileSize     string //
	CapturedAt   string //
	CreatedAt    string //
}

// evidenceItemsColumns holds the columns for the table evidence_items.
var evidenceItemsColumns = EvidenceItemsColumns{
	Id:           "id",
	ProjectId:    "project_id",
	RunId:        "run_id",
	Surface:      "surface",
	Journey:      "journey",
	EvidenceType: "evidence_type",
	FilePath:     "file_path",
	ContentHash:  "content_hash",
	FileSize:     "file_size",
	CapturedAt:   "captured_at",
	CreatedAt:    "created_at",
}

// NewEvidenceItemsDao creates and returns a new DAO object for table data access.
func NewEvidenceItemsDao(handlers ...gdb.ModelHandler) *EvidenceItemsDao {
	return &EvidenceItemsDao{
		group:    "default",
		table:    "evidence_items",
		columns:  evidenceItemsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *EvidenceItemsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *EvidenceItemsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *EvidenceItemsDao) Columns() EvidenceItemsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *EvidenceItemsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *EvidenceItemsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *EvidenceItemsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
