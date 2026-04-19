// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowCompiledTasksDao is the data access object for the table workflow_compiled_tasks.
type WorkflowCompiledTasksDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  WorkflowCompiledTasksColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// WorkflowCompiledTasksColumns defines and stores column names for the table workflow_compiled_tasks.
type WorkflowCompiledTasksColumns struct {
	Id                       string //
	CompiledPlanId           string //
	TaskKey                  string //
	Name                     string //
	Description              string //
	Phase                    string //
	TaskKind                 string //
	RoleType                 string //
	BrainKind                string //
	RiskLevel                string //
	AffectedResourcesJson    string //
	DeliveryContractJson     string //
	VerificationContractJson string //
	ManualReviewRequired     string //
	DependsOnTaskKeysJson    string //
	Status                   string //
	CreatedAt                string //
	UpdatedAt                string //
}

// workflowCompiledTasksColumns holds the columns for the table workflow_compiled_tasks.
var workflowCompiledTasksColumns = WorkflowCompiledTasksColumns{
	Id:                       "id",
	CompiledPlanId:           "compiled_plan_id",
	TaskKey:                  "task_key",
	Name:                     "name",
	Description:              "description",
	Phase:                    "phase",
	TaskKind:                 "task_kind",
	RoleType:                 "role_type",
	BrainKind:                "brain_kind",
	RiskLevel:                "risk_level",
	AffectedResourcesJson:    "affected_resources_json",
	DeliveryContractJson:     "delivery_contract_json",
	VerificationContractJson: "verification_contract_json",
	ManualReviewRequired:     "manual_review_required",
	DependsOnTaskKeysJson:    "depends_on_task_keys_json",
	Status:                   "status",
	CreatedAt:                "created_at",
	UpdatedAt:                "updated_at",
}

// NewWorkflowCompiledTasksDao creates and returns a new DAO object for table data access.
func NewWorkflowCompiledTasksDao(handlers ...gdb.ModelHandler) *WorkflowCompiledTasksDao {
	return &WorkflowCompiledTasksDao{
		group:    "default",
		table:    "workflow_compiled_tasks",
		columns:  workflowCompiledTasksColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *WorkflowCompiledTasksDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *WorkflowCompiledTasksDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *WorkflowCompiledTasksDao) Columns() WorkflowCompiledTasksColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *WorkflowCompiledTasksDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *WorkflowCompiledTasksDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *WorkflowCompiledTasksDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
