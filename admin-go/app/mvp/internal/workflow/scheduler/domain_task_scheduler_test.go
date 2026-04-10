package scheduler

import (
	"context"
	"testing"
)

func TestReconcileWorkflowProgressReturnsFalseWhenNoUnfinishedTasks(t *testing.T) {
	t.Parallel()

	var (
		scheduleCalled bool
		checkCalled    bool
	)

	remaining := reconcileWorkflowProgress(
		context.Background(),
		123,
		func(ctx context.Context, workflowRunID int64) {
			scheduleCalled = true
			if workflowRunID != 123 {
				t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
			}
		},
		func(ctx context.Context, workflowRunID int64) {
			checkCalled = true
			if workflowRunID != 123 {
				t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
			}
		},
		func(ctx context.Context, workflowRunID int64) bool {
			if workflowRunID != 123 {
				t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
			}
			return false
		},
	)

	if remaining {
		t.Fatal("expected workflow to be fully reconciled")
	}
	if !scheduleCalled || !checkCalled {
		t.Fatalf("expected both callbacks to be called, schedule=%t check=%t", scheduleCalled, checkCalled)
	}
}

func TestReconcileWorkflowProgressReturnsTrueWhenWorkRemains(t *testing.T) {
	t.Parallel()

	remaining := reconcileWorkflowProgress(
		context.Background(),
		456,
		nil,
		nil,
		func(ctx context.Context, workflowRunID int64) bool {
			if workflowRunID != 456 {
				t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
			}
			return true
		},
	)

	if !remaining {
		t.Fatal("expected workflow to still have unfinished tasks")
	}
}
