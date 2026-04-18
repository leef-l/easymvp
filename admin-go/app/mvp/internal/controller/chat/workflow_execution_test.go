package chat

import (
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
)

func TestResolveExecutionStageRunPrefersCurrentExecuteStage(t *testing.T) {
	t.Parallel()

	wfRun := gdb.Record{
		"id":                   gvar.New(101),
		"current_stage":        gvar.New("execute"),
		"current_stage_run_id": gvar.New(202),
	}
	current := gdb.Record{
		"id":         gvar.New(202),
		"stage_type": gvar.New("execute"),
		"status":     gvar.New("running"),
	}
	latest := gdb.Record{
		"id":         gvar.New(201),
		"stage_type": gvar.New("execute"),
		"status":     gvar.New("failed"),
	}

	got := selectExecutionStageRunRecord(wfRun, current, latest)
	if got["id"].Int64() != 202 {
		t.Fatalf("expected current execute stage run, got %+v", got)
	}
}

func TestResolveExecutionStageRunReturnsNilWhenCurrentStageIsNotExecute(t *testing.T) {
	t.Parallel()

	wfRun := gdb.Record{
		"id":                   gvar.New(101),
		"current_stage":        gvar.New("rework"),
		"current_stage_run_id": gvar.New(303),
	}
	current := gdb.Record{
		"id":         gvar.New(303),
		"stage_type": gvar.New("rework"),
		"status":     gvar.New("running"),
	}
	latest := gdb.Record{
		"id":         gvar.New(201),
		"stage_type": gvar.New("execute"),
		"status":     gvar.New("completed"),
	}

	got := selectExecutionStageRunRecord(wfRun, current, latest)
	if got != nil {
		t.Fatalf("expected nil when workflow is not in execute, got %+v", got)
	}
}
