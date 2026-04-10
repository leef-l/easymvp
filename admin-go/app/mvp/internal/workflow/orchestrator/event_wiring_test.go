package orchestrator

import (
	"context"
	"errors"
	"os"
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

func TestShouldFallbackToDirectRedisErr(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "auth error", err: errors.New("NOAUTH Authentication required"), want: true},
		{name: "nil redis object", err: errors.New("the Redis object is nil"), want: true},
		{name: "client unavailable", err: errors.New("redis client unavailable"), want: true},
		{name: "generic timeout", err: errors.New("i/o timeout"), want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldFallbackToDirectRedisErr(tc.err); got != tc.want {
				t.Fatalf("shouldFallbackToDirectRedisErr(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestResolveRedisPass(t *testing.T) {
	t.Parallel()

	oldPass := os.Getenv("REDIS_PASS")
	oldPassword := os.Getenv("REDIS_PASSWORD")
	defer func() {
		_ = os.Setenv("REDIS_PASS", oldPass)
		_ = os.Setenv("REDIS_PASSWORD", oldPassword)
	}()

	_ = os.Unsetenv("REDIS_PASS")
	_ = os.Unsetenv("REDIS_PASSWORD")
	if got := resolveRedisPass(); got != "" {
		t.Fatalf("resolveRedisPass() without env = %q, want empty", got)
	}

	_ = os.Setenv("REDIS_PASSWORD", "password-only")
	if got := resolveRedisPass(); got != "password-only" {
		t.Fatalf("resolveRedisPass() with REDIS_PASSWORD = %q, want %q", got, "password-only")
	}

	_ = os.Setenv("REDIS_PASS", "pass-first")
	if got := resolveRedisPass(); got != "pass-first" {
		t.Fatalf("resolveRedisPass() with REDIS_PASS = %q, want %q", got, "pass-first")
	}
}
