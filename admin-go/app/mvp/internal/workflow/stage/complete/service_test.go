package complete

import (
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
)

func TestNormalizeDBUTCGTimeConvertsUTCToLocal(t *testing.T) {
	t.Parallel()

	raw := gtime.NewFromTime(time.Date(2026, 4, 9, 7, 26, 49, 0, time.Local))
	got := normalizeDBUTCGTime(raw)
	if got == nil {
		t.Fatal("expected normalized time")
	}

	want := time.Date(2026, 4, 9, 7, 26, 49, 0, time.UTC).In(time.Local)
	if got.TimestampMilli() != want.UnixMilli() {
		t.Fatalf("normalized timestamp=%d want=%d", got.TimestampMilli(), want.UnixMilli())
	}
}

func TestApplyTaskMetricRowsIncludesSkippedCount(t *testing.T) {
	t.Parallel()

	rows := gdb.Result{
		{"status": gvar.New("completed"), "cnt": gvar.New(2)},
		{"status": gvar.New("failed"), "cnt": gvar.New(1)},
	}

	summary := &CompletionSummary{}
	applyTaskMetricRows(summary, rows, 2)

	if summary.TotalTasks != 3 {
		t.Fatalf("totalTasks=%d want=3", summary.TotalTasks)
	}
	if summary.CompletedTasks != 2 {
		t.Fatalf("completedTasks=%d want=2", summary.CompletedTasks)
	}
	if summary.FailedTasks != 1 {
		t.Fatalf("failedTasks=%d want=1", summary.FailedTasks)
	}
	if summary.SkippedTasks != 2 {
		t.Fatalf("skippedTasks=%d want=2", summary.SkippedTasks)
	}
	if summary.SuccessRate != 0.6667 {
		t.Fatalf("successRate=%v want=0.6667", summary.SuccessRate)
	}
}
