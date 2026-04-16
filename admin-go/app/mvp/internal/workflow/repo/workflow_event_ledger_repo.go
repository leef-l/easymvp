package repo

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowEventLedgerRepo Workflow 事件账本仓储。
type WorkflowEventLedgerRepo struct{}

func NewWorkflowEventLedgerRepo() *WorkflowEventLedgerRepo { return &WorkflowEventLedgerRepo{} }

func (r *WorkflowEventLedgerRepo) table() string { return "mvp_workflow_event_ledger" }

// Insert 创建 durable ledger 记录。
func (r *WorkflowEventLedgerRepo) Insert(ctx context.Context, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).Insert(data)
	return err
}

// ReviveFailedClaim 将失败记录恢复为 handling。
func (r *WorkflowEventLedgerRepo) ReviveFailedClaim(ctx context.Context, scope, idempotencyKey string, data g.Map) (int64, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("scope", scope).
		Where("idempotency_key", idempotencyKey).
		Where("status", "failed")
	result, err := model.Data(data).Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// UpdateByScopeKey 按 scope 和幂等键更新账本记录。
func (r *WorkflowEventLedgerRepo) UpdateByScopeKey(ctx context.Context, scope, idempotencyKey string, data g.Map) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("scope", scope).
		Where("idempotency_key", idempotencyKey).
		Data(data).
		Update()
	return err
}

// MarkHandled 标记事件已处理。
func (r *WorkflowEventLedgerRepo) MarkHandled(ctx context.Context, scope, idempotencyKey string, handledAt time.Time) error {
	return r.UpdateByScopeKey(ctx, scope, idempotencyKey, g.Map{
		"status":     "handled",
		"last_error": nil,
		"handled_at": handledAt,
		"updated_at": handledAt,
	})
}

// MarkFailed 标记事件处理失败。
func (r *WorkflowEventLedgerRepo) MarkFailed(ctx context.Context, scope, idempotencyKey string, handleErr error, updatedAt time.Time) error {
	return r.UpdateByScopeKey(ctx, scope, idempotencyKey, g.Map{
		"status":     "failed",
		"last_error": TrimWorkflowEventLedgerError(handleErr),
		"updated_at": updatedAt,
	})
}

// TrimWorkflowEventLedgerError 压缩 ledger 中存储的错误文本。
func TrimWorkflowEventLedgerError(handleErr error) string {
	if handleErr == nil {
		return ""
	}
	message := strings.TrimSpace(handleErr.Error())
	if len(message) <= 500 {
		return message
	}
	return message[:500]
}

// InspectDurableColumns 检查 durable ledger 必需列是否存在。
func (r *WorkflowEventLedgerRepo) InspectDurableColumns(ctx context.Context) error {
	_, err := g.DB().Ctx(ctx).Model(r.table()).
		Fields("scope,event_id,idempotency_key,status").
		Limit(1).
		All()
	return err
}
