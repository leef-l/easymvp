// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// DomainTasksDao is the data access object for the table domain_tasks.
type DomainTasksDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  DomainTasksColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// DomainTasksColumns defines and stores column names for the table domain_tasks.
type DomainTasksColumns struct {
	Id                   string //
	ProjectId            string //
	SourceCompiledPlanId string //
	SourceCompiledTaskId string //
	SourceTaskKey        string //
	CompiledVersion      string //
	Name                 string //
	Phase                string //
	TaskKind             string //
	RoleType             string //
	BrainKind            string //
	RiskLevel            string //
	Status               string //
	ManualReviewRequired string //
	CreatedAt            string //
	UpdatedAt            string //
}

// domainTasksColumns holds the columns for the table domain_tasks.
var domainTasksColumns = DomainTasksColumns{
	Id:                   "id",
	ProjectId:            "project_id",
	SourceCompiledPlanId: "source_compiled_plan_id",
	SourceCompiledTaskId: "source_compiled_task_id",
	SourceTaskKey:        "source_task_key",
	CompiledVersion:      "compiled_version",
	Name:                 "name",
	Phase:                "phase",
	TaskKind:             "task_kind",
	RoleType:             "role_type",
	BrainKind:            "brain_kind",
	RiskLevel:            "risk_level",
	Status:               "status",
	ManualReviewRequired: "manual_review_required",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
}

// NewDomainTasksDao creates and returns a new DAO object for table data access.
func NewDomainTasksDao(handlers ...gdb.ModelHandler) *DomainTasksDao {
	return &DomainTasksDao{
		group:    "default",
		table:    "domain_tasks",
		columns:  domainTasksColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *DomainTasksDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *DomainTasksDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *DomainTasksDao) Columns() DomainTasksColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *DomainTasksDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *DomainTasksDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *DomainTasksDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
