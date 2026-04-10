package orchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"easymvp/app/mvp/internal/workflow/event"
)

func TestHandleTaskCompletedEventTriggersReconcileOncePerIdempotencyKey(t *testing.T) {
	originalGuard := recoveryEventGuard
	originalReconcile := reconcileWorkflowProgressFn
	originalBeginClaim := beginRecoveryEventClaimFn
	defer func() {
		recoveryEventGuard = originalGuard
		reconcileWorkflowProgressFn = originalReconcile
		beginRecoveryEventClaimFn = originalBeginClaim
	}()

	recoveryEventGuard = newRecoveryEventGuard(time.Hour)
	beginRecoveryEventClaimFn = func(ctx context.Context, evt event.Event) (event.DurableEventClaim, bool, error) {
		return noopDurableClaim{}, true, nil
	}

	calls := 0
	reconcileWorkflowProgressFn = func(ctx context.Context, workflowRunID int64) bool {
		calls++
		if workflowRunID != 42 {
			t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
		}
		return true
	}

	evt := event.Event{
		WorkflowRunID:  42,
		EventType:      event.EventTaskCompleted,
		IdempotencyKey: "wf:42:task:100:task.completed",
	}

	if err := handleWorkflowRecoveryEvent(context.Background(), evt); err != nil {
		t.Fatalf("first handle failed: %v", err)
	}
	if err := handleWorkflowRecoveryEvent(context.Background(), evt); err != nil {
		t.Fatalf("second handle failed: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected reconcile to run once, got %d", calls)
	}
}

func TestHandleWorkflowRecoveryEventSkipsWhenDurableClaimAlreadyHandled(t *testing.T) {
	originalGuard := recoveryEventGuard
	originalReconcile := reconcileWorkflowProgressFn
	originalBeginClaim := beginRecoveryEventClaimFn
	defer func() {
		recoveryEventGuard = originalGuard
		reconcileWorkflowProgressFn = originalReconcile
		beginRecoveryEventClaimFn = originalBeginClaim
	}()

	recoveryEventGuard = newRecoveryEventGuard(time.Hour)
	beginRecoveryEventClaimFn = func(ctx context.Context, evt event.Event) (event.DurableEventClaim, bool, error) {
		return nil, false, nil
	}

	calls := 0
	reconcileWorkflowProgressFn = func(ctx context.Context, workflowRunID int64) bool {
		calls++
		return true
	}

	if err := handleWorkflowRecoveryEvent(context.Background(), event.Event{
		WorkflowRunID:  99,
		EventType:      event.EventTaskCompleted,
		IdempotencyKey: "wf:99:task:1:type:task.completed:attempt:1",
	}); err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected durable duplicate to skip reconcile, got %d calls", calls)
	}
}

func TestHandleWorkflowRecoveryEventFallsBackWhenLedgerMigrationMissing(t *testing.T) {
	originalGuard := recoveryEventGuard
	originalReconcile := reconcileWorkflowProgressFn
	originalBeginClaim := beginRecoveryEventClaimFn
	defer func() {
		recoveryEventGuard = originalGuard
		reconcileWorkflowProgressFn = originalReconcile
		beginRecoveryEventClaimFn = originalBeginClaim
	}()

	recoveryEventGuard = newRecoveryEventGuard(time.Hour)
	beginRecoveryEventClaimFn = func(ctx context.Context, evt event.Event) (event.DurableEventClaim, bool, error) {
		return nil, false, errors.New("Error 1146 (42S02): Table 'test.mvp_workflow_event_ledger' doesn't exist")
	}

	calls := 0
	reconcileWorkflowProgressFn = func(ctx context.Context, workflowRunID int64) bool {
		calls++
		if workflowRunID != 7 {
			t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
		}
		return true
	}

	if err := handleWorkflowRecoveryEvent(context.Background(), event.Event{
		WorkflowRunID:  7,
		EventType:      event.EventTaskCompleted,
		IdempotencyKey: "wf:7:task:1:type:task.completed:attempt:1",
	}); err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected fallback reconcile once, got %d", calls)
	}
}

type noopDurableClaim struct{}

func (noopDurableClaim) MarkHandled(ctx context.Context) error {
	return nil
}

func (noopDurableClaim) MarkFailed(ctx context.Context, handleErr error) error {
	return nil
}
