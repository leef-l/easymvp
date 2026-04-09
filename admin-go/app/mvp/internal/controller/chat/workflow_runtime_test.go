package chat

import (
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/autonomy"
)

func TestShouldUseRuntimeSnapshot(t *testing.T) {
	t.Parallel()

	freshSnapshot := &projectRuntimeSnapshot{
		CreatedAt: gtime.NewFromTime(time.Now().Add(-30 * time.Second)),
		Situation: autonomy.Situation{
			WorkflowStatus: "executing",
			Progress:       &autonomy.ProgressMetrics{TotalTasks: 3},
		},
	}
	if !shouldUseRuntimeSnapshot(freshSnapshot, "running") {
		t.Fatal("expected fresh snapshot to be used")
	}

	completedSnapshot := &projectRuntimeSnapshot{
		CreatedAt: gtime.NewFromTime(time.Now().Add(-10 * time.Minute)),
		Situation: autonomy.Situation{
			WorkflowStatus: "completed",
			Progress:       &autonomy.ProgressMetrics{TotalTasks: 3},
		},
	}
	if !shouldUseRuntimeSnapshot(completedSnapshot, "completed") {
		t.Fatal("expected terminal snapshot with matching status to be used")
	}

	staleSnapshot := &projectRuntimeSnapshot{
		CreatedAt: gtime.NewFromTime(time.Now().Add(-10 * time.Minute)),
		Situation: autonomy.Situation{
			WorkflowStatus: "executing",
			Progress:       &autonomy.ProgressMetrics{TotalTasks: 3},
		},
	}
	if shouldUseRuntimeSnapshot(staleSnapshot, "running") {
		t.Fatal("did not expect stale snapshot to be used")
	}
}

func TestTaskStatFromProgress(t *testing.T) {
	t.Parallel()

	stat := taskStatFromProgress(&autonomy.ProgressMetrics{
		TotalTasks:     8,
		CompletedTasks: 3,
		FailedTasks:    2,
		RunningTasks:   1,
	})
	if stat.TotalTasks != 8 || stat.CompletedTasks != 3 || stat.FailedTasks != 2 || stat.RunningTasks != 1 {
		t.Fatalf("unexpected task stat: %+v", stat)
	}

	if zero := taskStatFromProgress(nil); zero != (projectRuntimeTaskStat{}) {
		t.Fatalf("expected zero task stat for nil progress, got %+v", zero)
	}
}

func TestLatestNonNilTime(t *testing.T) {
	t.Parallel()

	older := gtime.NewFromTime(time.Now().Add(-5 * time.Minute))
	newer := gtime.NewFromTime(time.Now().Add(-1 * time.Minute))

	got := latestNonNilTime(nil, older, newer)
	if got == nil || got.TimestampMilli() != newer.TimestampMilli() {
		t.Fatalf("latestNonNilTime() got %v, want %v", got, newer)
	}

	if latestNonNilTime(nil, nil) != nil {
		t.Fatal("expected nil when all items are nil")
	}
}
