package scheduler

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
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

func TestHasUnfinishedReturnsTrueWhenBlockingPendingDeliveryExists(t *testing.T) {
	prev := listReadyPendingSyncByWorkflowFn
	t.Cleanup(func() {
		listReadyPendingSyncByWorkflowFn = prev
	})

	listReadyPendingSyncByWorkflowFn = func(ctx context.Context, workflowRunID int64) ([]g.Map, error) {
		if workflowRunID != 88 {
			t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
		}
		return []g.Map{{"task_id": int64(1)}}, nil
	}

	s := &DomainTaskScheduler{}
	if !s.HasUnfinished(context.Background(), 88) {
		t.Fatal("pending delivery should block workflow completion")
	}
}

func TestCheckAllDoneSkipsCompletionWhenBlockingPendingDeliveryExists(t *testing.T) {
	prev := listReadyPendingSyncByWorkflowFn
	t.Cleanup(func() {
		listReadyPendingSyncByWorkflowFn = prev
	})

	listReadyPendingSyncByWorkflowFn = func(context.Context, int64) ([]g.Map, error) {
		return []g.Map{{"task_id": int64(2)}}, nil
	}

	called := false
	s := &DomainTaskScheduler{
		onAllDone: func(ctx context.Context, workflowRunID int64) {
			called = true
		},
	}
	s.checkAllDone(context.Background(), 77)
	if called {
		t.Fatal("blocking pending delivery should prevent completion callback")
	}
}

func TestHasBlockingPendingDeliveriesTreatsUnknownColumnAsDisabled(t *testing.T) {
	prev := listReadyPendingSyncByWorkflowFn
	t.Cleanup(func() {
		listReadyPendingSyncByWorkflowFn = prev
	})

	listReadyPendingSyncByWorkflowFn = func(context.Context, int64) ([]g.Map, error) {
		return nil, assertSchedulerError("Error 1054 (42S22): Unknown column 'sync_status' in 'field list'")
	}

	s := &DomainTaskScheduler{}
	if s.hasBlockingPendingDeliveries(context.Background(), 66) {
		t.Fatal("legacy schema should not block scheduler")
	}
}

func TestHasBlockingPendingDeliveriesTreatsOtherErrorsAsBlocking(t *testing.T) {
	prev := listReadyPendingSyncByWorkflowFn
	t.Cleanup(func() {
		listReadyPendingSyncByWorkflowFn = prev
	})

	listReadyPendingSyncByWorkflowFn = func(context.Context, int64) ([]g.Map, error) {
		return nil, assertSchedulerError("dial tcp timeout")
	}

	s := &DomainTaskScheduler{}
	if !s.hasBlockingPendingDeliveries(context.Background(), 66) {
		t.Fatal("query failure should conservatively block scheduler")
	}
}

type schedulerTestError struct {
	msg string
}

func (e *schedulerTestError) Error() string {
	return e.msg
}

func assertSchedulerError(msg string) error {
	return &schedulerTestError{msg: msg}
}
