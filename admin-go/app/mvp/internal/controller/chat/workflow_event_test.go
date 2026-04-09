package chat

import (
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
)

func TestBuildTimelineEvent(t *testing.T) {
	t.Parallel()

	record := gdb.Record{
		"id":              gvar.New(1),
		"workflow_run_id": gvar.New(2),
		"stage_run_id":    gvar.New(3),
		"entity_type":     gvar.New("domain_task"),
		"entity_id":       gvar.New(4),
		"event_type":      gvar.New("workflow.force_stage"),
		"payload":         gvar.New(`{"stage_type":"rework","reason":"人工接管"}`),
		"created_at":      gvar.New("2026-04-08 15:27:30"),
	}

	item := buildTimelineEvent(record)
	if item.ID != 1 || item.WorkflowRunID != 2 {
		t.Fatalf("unexpected event ids: %+v", item)
	}
	if item.StageRunID == nil || *item.StageRunID != 3 {
		t.Fatalf("unexpected stage run id: %+v", item)
	}
	if item.EntityID == nil || *item.EntityID != 4 {
		t.Fatalf("unexpected entity id: %+v", item)
	}
	if item.Label != "工作流已人工切换到返工阶段：人工接管" {
		t.Fatalf("unexpected event label: %s", item.Label)
	}
	if item.CreatedAt == nil {
		t.Fatalf("expected created_at to be normalized: %+v", item)
	}
}

func TestJsonInt64SliceToInt64(t *testing.T) {
	t.Parallel()

	got := jsonInt64SliceToInt64([]int64{1, 0, -1, 3})
	if len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("unexpected slice conversion: %+v", got)
	}
	if empty := jsonInt64SliceToInt64(nil); len(empty) != 0 {
		t.Fatalf("expected nil input to normalize to empty slice, got %+v", empty)
	}
}
