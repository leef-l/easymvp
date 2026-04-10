package event

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

const (
	DurableEventClaimScopePublish = "workflow.publish"

	durableEventLedgerTable          = "mvp_workflow_event_ledger"
	durableEventLedgerStatusHandling = "handling"
	durableEventLedgerStatusHandled  = "handled"
	durableEventLedgerStatusFailed   = "failed"
)

// DurableEventClaim 表示一次 durable 幂等处理 claim。
type DurableEventClaim interface {
	MarkHandled(ctx context.Context) error
	MarkFailed(ctx context.Context, handleErr error) error
}

type durableEventClaim struct {
	scope          string
	idempotencyKey string
}

type noopDurableEventClaim struct{}

var insertDurableEventLedgerFn = func(ctx context.Context, data g.Map) error {
	_, err := g.DB().Model(durableEventLedgerTable).Ctx(ctx).Insert(data)
	return err
}

var reviveDurableEventLedgerFn = func(ctx context.Context, scope, idempotencyKey string, evt Event, now time.Time) (int64, error) {
	result, err := g.DB().Model(durableEventLedgerTable).Ctx(ctx).
		Where("scope", scope).
		Where("idempotency_key", idempotencyKey).
		Where("status", durableEventLedgerStatusFailed).
		Data(g.Map{
			"event_id":        evt.EventID,
			"workflow_run_id": evt.WorkflowRunID,
			"event_type":      evt.EventType,
			"attempt":         evt.Attempt,
			"status":          durableEventLedgerStatusHandling,
			"last_error":      nil,
			"updated_at":      now,
		}).
		Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

var updateDurableEventLedgerFn = func(ctx context.Context, scope, idempotencyKey string, data g.Map) error {
	_, err := g.DB().Model(durableEventLedgerTable).Ctx(ctx).
		Where("scope", scope).
		Where("idempotency_key", idempotencyKey).
		Data(data).
		Update()
	return err
}

// BeginDurableEventClaim 为同一 scope + idempotency_key 建立一次可恢复 claim。
// 返回 false 表示该事件已被其他进程/先前流程处理，无需重复执行。
func BeginDurableEventClaim(ctx context.Context, scope string, evt Event) (DurableEventClaim, bool, error) {
	evt = evt.EnsureMetadata()
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = DurableEventClaimScopePublish
	}

	now := time.Now()
	err := insertDurableEventLedgerFn(ctx, g.Map{
		"id":              snowflake.Generate(),
		"scope":           scope,
		"event_id":        evt.EventID,
		"idempotency_key": evt.IdempotencyKey,
		"workflow_run_id": evt.WorkflowRunID,
		"event_type":      evt.EventType,
		"attempt":         evt.Attempt,
		"status":          durableEventLedgerStatusHandling,
		"created_at":      now,
		"updated_at":      now,
	})
	if err == nil {
		return &durableEventClaim{scope: scope, idempotencyKey: evt.IdempotencyKey}, true, nil
	}
	if IsMissingDurableEventLedgerErr(err) {
		g.Log().Warningf(ctx, "[WorkflowEventLedger] durable ledger migration missing, fallback to non-durable path: %v", err)
		return noopDurableEventClaim{}, true, nil
	}
	if !isDuplicateKeyErr(err) {
		return nil, false, err
	}

	rows, upErr := reviveDurableEventLedgerFn(ctx, scope, evt.IdempotencyKey, evt, now)
	if upErr != nil {
		return nil, false, upErr
	}
	if rows == 0 {
		return nil, false, nil
	}

	return &durableEventClaim{scope: scope, idempotencyKey: evt.IdempotencyKey}, true, nil
}

func (noopDurableEventClaim) MarkHandled(ctx context.Context) error { return nil }

func (noopDurableEventClaim) MarkFailed(ctx context.Context, handleErr error) error { return nil }

func (c *durableEventClaim) MarkHandled(ctx context.Context) error {
	if c == nil || strings.TrimSpace(c.scope) == "" || strings.TrimSpace(c.idempotencyKey) == "" {
		return nil
	}
	return updateDurableEventLedgerFn(ctx, c.scope, c.idempotencyKey, g.Map{
		"status":     durableEventLedgerStatusHandled,
		"last_error": nil,
		"handled_at": time.Now(),
		"updated_at": time.Now(),
	})
}

func (c *durableEventClaim) MarkFailed(ctx context.Context, handleErr error) error {
	if c == nil || strings.TrimSpace(c.scope) == "" || strings.TrimSpace(c.idempotencyKey) == "" {
		return nil
	}
	return updateDurableEventLedgerFn(ctx, c.scope, c.idempotencyKey, g.Map{
		"status":     durableEventLedgerStatusFailed,
		"last_error": trimLedgerError(handleErr),
		"updated_at": time.Now(),
	})
}

func trimLedgerError(handleErr error) string {
	if handleErr == nil {
		return ""
	}
	message := strings.TrimSpace(handleErr.Error())
	if len(message) <= 500 {
		return message
	}
	return message[:500]
}

func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	lowerErr := strings.ToLower(err.Error())
	return strings.Contains(lowerErr, "duplicate") && strings.Contains(lowerErr, "key")
}

// IsMissingDurableEventLedgerErr 判断 durable ledger migration 是否尚未执行。
func IsMissingDurableEventLedgerErr(err error) bool {
	if err == nil {
		return false
	}
	lowerErr := strings.ToLower(err.Error())
	return strings.Contains(lowerErr, durableEventLedgerTable) &&
		(strings.Contains(lowerErr, "doesn't exist") ||
			strings.Contains(lowerErr, "does not exist") ||
			strings.Contains(lowerErr, "1146"))
}
