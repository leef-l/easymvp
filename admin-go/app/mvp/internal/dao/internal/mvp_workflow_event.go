// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MvpWorkflowEventDao is the data access object for the table mvp_workflow_event.
type MvpWorkflowEventDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  MvpWorkflowEventColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// MvpWorkflowEventColumns defines and stores column names for the table mvp_workflow_event.
type MvpWorkflowEventColumns struct {
	Id             string // 雪花ID
	EventId        string // 事件元ID
	WorkflowRunId  string // 所属工作流运行ID
	StageRunId     string // 关联阶段运行ID
	EntityType     string // 实体类型: workflow_run/stage_run/plan_version/domain_task/review_issue
	EntityId       string // 实体ID
	EventType      string // 事件类型: workflow.created/stage.started/task.completed等
	Attempt        string // 事件尝试次数
	IdempotencyKey string // 幂等键
	Payload        string // 事件载荷(JSON)
	CreatedAt      string // 创建时间
}

// mvpWorkflowEventColumns holds the columns for the table mvp_workflow_event.
var mvpWorkflowEventColumns = MvpWorkflowEventColumns{
	Id:             "id",
	EventId:        "event_id",
	WorkflowRunId:  "workflow_run_id",
	StageRunId:     "stage_run_id",
	EntityType:     "entity_type",
	EntityId:       "entity_id",
	EventType:      "event_type",
	Attempt:        "attempt",
	IdempotencyKey: "idempotency_key",
	Payload:        "payload",
	CreatedAt:      "created_at",
}

// NewMvpWorkflowEventDao creates and returns a new DAO object for table data access.
func NewMvpWorkflowEventDao(handlers ...gdb.ModelHandler) *MvpWorkflowEventDao {
	return &MvpWorkflowEventDao{
		group:    "default",
		table:    "mvp_workflow_event",
		columns:  mvpWorkflowEventColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MvpWorkflowEventDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MvpWorkflowEventDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MvpWorkflowEventDao) Columns() MvpWorkflowEventColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MvpWorkflowEventDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MvpWorkflowEventDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *MvpWorkflowEventDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
