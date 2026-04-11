package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowEventLedgerRepo Workflow 事件账本仓储。
type WorkflowEventLedgerRepo struct{}

func NewWorkflowEventLedgerRepo() *WorkflowEventLedgerRepo { return &WorkflowEventLedgerRepo{} }

func (r *WorkflowEventLedgerRepo) table() string { return "mvp_workflow_event_ledger" }

// InspectDurableColumns 检查 durable ledger 必需列是否存在。
func (r *WorkflowEventLedgerRepo) InspectDurableColumns(ctx context.Context) error {
	_, err := g.DB().Ctx(ctx).Model(r.table()).
		Fields("scope,event_id,idempotency_key,status").
		Limit(1).
		All()
	return err
}
