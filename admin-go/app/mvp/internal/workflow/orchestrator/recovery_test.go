package orchestrator

import (
	"context"
	"testing"

	"easymvp/app/mvp/internal/consts"
)

func TestRecoverWorkflowRunSkipsSchedulerRestartWhenExecuteAlreadyCompleted(t *testing.T) {
	originalCreateRuntime := createRecoveredRuntime
	originalPrepareScheduler := prepareRecoveredTaskScheduler
	originalHasUnfinished := hasRecoveredWorkflowUnfinishedTasks
	originalStartScheduler := startRecoveredTaskScheduler
	originalReconcile := reconcileWorkflowProgressFn
	defer func() {
		createRecoveredRuntime = originalCreateRuntime
		prepareRecoveredTaskScheduler = originalPrepareScheduler
		hasRecoveredWorkflowUnfinishedTasks = originalHasUnfinished
		startRecoveredTaskScheduler = originalStartScheduler
		reconcileWorkflowProgressFn = originalReconcile
	}()

	var (
		createdProjectID int64
		prepareCalls     int
		startCalls       int
		reconcileCalls   int
		hasUnfinishedSeq = []bool{false, false}
	)

	createRecoveredRuntime = func(workflowRunID, projectID int64) {
		if workflowRunID != 101 {
			t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
		}
		createdProjectID = projectID
	}
	prepareRecoveredTaskScheduler = func(ctx context.Context, workflowRunID int64, stage string, stageRunID int64) error {
		prepareCalls++
		if workflowRunID != 101 || stage != consts.StageTypeExecute || stageRunID != 202 {
			t.Fatalf("unexpected prepare args: workflow=%d stage=%s stageRunID=%d", workflowRunID, stage, stageRunID)
		}
		return nil
	}
	hasRecoveredWorkflowUnfinishedTasks = func(ctx context.Context, workflowRunID int64) bool {
		if workflowRunID != 101 {
			t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
		}
		current := hasUnfinishedSeq[0]
		hasUnfinishedSeq = hasUnfinishedSeq[1:]
		return current
	}
	startRecoveredTaskScheduler = func(ctx context.Context, workflowRunID int64) error {
		startCalls++
		return nil
	}
	reconcileWorkflowProgressFn = func(ctx context.Context, workflowRunID int64) bool {
		reconcileCalls++
		if workflowRunID != 101 {
			t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
		}
		return false
	}

	restarted, failed := recoverWorkflowRun(context.Background(), recoverableWorkflowRun{
		WorkflowRunID: 101,
		ProjectID:     303,
		Status:        consts.WorkflowRunStatusExecuting,
		Stage:         consts.StageTypeExecute,
		StageRunID:    202,
	})

	if failed {
		t.Fatal("expected recovery to succeed")
	}
	if restarted {
		t.Fatal("expected scheduler restart to be skipped")
	}
	if createdProjectID != 303 {
		t.Fatalf("unexpected created projectID: %d", createdProjectID)
	}
	if prepareCalls != 1 {
		t.Fatalf("expected prepare once, got %d", prepareCalls)
	}
	if startCalls != 0 {
		t.Fatalf("expected no scheduler restart, got %d", startCalls)
	}
	if reconcileCalls != 1 {
		t.Fatalf("expected reconcile once, got %d", reconcileCalls)
	}
}

func TestRecoverWorkflowRunRestartsSchedulerAndReconcilesExecuteStage(t *testing.T) {
	originalCreateRuntime := createRecoveredRuntime
	originalPrepareScheduler := prepareRecoveredTaskScheduler
	originalHasUnfinished := hasRecoveredWorkflowUnfinishedTasks
	originalStartScheduler := startRecoveredTaskScheduler
	originalReconcile := reconcileWorkflowProgressFn
	defer func() {
		createRecoveredRuntime = originalCreateRuntime
		prepareRecoveredTaskScheduler = originalPrepareScheduler
		hasRecoveredWorkflowUnfinishedTasks = originalHasUnfinished
		startRecoveredTaskScheduler = originalStartScheduler
		reconcileWorkflowProgressFn = originalReconcile
	}()

	var (
		prepareCalls   int
		startCalls     int
		reconcileCalls int
	)

	createRecoveredRuntime = func(workflowRunID, projectID int64) {}
	prepareRecoveredTaskScheduler = func(ctx context.Context, workflowRunID int64, stage string, stageRunID int64) error {
		prepareCalls++
		return nil
	}
	hasRecoveredWorkflowUnfinishedTasks = func(ctx context.Context, workflowRunID int64) bool {
		return true
	}
	startRecoveredTaskScheduler = func(ctx context.Context, workflowRunID int64) error {
		startCalls++
		if workflowRunID != 404 {
			t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
		}
		return nil
	}
	reconcileWorkflowProgressFn = func(ctx context.Context, workflowRunID int64) bool {
		reconcileCalls++
		if workflowRunID != 404 {
			t.Fatalf("unexpected workflowRunID: %d", workflowRunID)
		}
		return true
	}

	restarted, failed := recoverWorkflowRun(context.Background(), recoverableWorkflowRun{
		WorkflowRunID: 404,
		ProjectID:     505,
		Status:        consts.WorkflowRunStatusExecuting,
		Stage:         consts.StageTypeExecute,
		StageRunID:    606,
	})

	if failed {
		t.Fatal("expected recovery to succeed")
	}
	if !restarted {
		t.Fatal("expected scheduler restart")
	}
	if prepareCalls != 1 {
		t.Fatalf("expected prepare once, got %d", prepareCalls)
	}
	if startCalls != 1 {
		t.Fatalf("expected start once, got %d", startCalls)
	}
	if reconcileCalls != 1 {
		t.Fatalf("expected post-start reconcile once, got %d", reconcileCalls)
	}
}
