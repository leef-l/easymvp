package service

import (
	"testing"

	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func TestNormalizeExecutionViewLimit(t *testing.T) {
	t.Parallel()

	if got := normalizeExecutionViewLimit(0, 10, 50); got != 10 {
		t.Fatalf("default limit mismatch: got %d want %d", got, 10)
	}
	if got := normalizeExecutionViewLimit(99, 10, 50); got != 50 {
		t.Fatalf("max limit mismatch: got %d want %d", got, 50)
	}
	if got := normalizeExecutionViewLimit(7, 10, 50); got != 7 {
		t.Fatalf("explicit limit mismatch: got %d want %d", got, 7)
	}
}

func TestPrioritizeExecutionBinding(t *testing.T) {
	t.Parallel()

	bindings := []entity.BrainRunBindings{
		{Id: "runbind_1"},
		{Id: "runbind_2"},
		{Id: "runbind_3"},
	}

	selected := entity.BrainRunBindings{Id: "runbind_2"}
	got := prioritizeExecutionBinding(bindings, selected, 3)

	if len(got) != 3 {
		t.Fatalf("unexpected binding count: got %d want %d", len(got), 3)
	}
	if got[0].Id != selected.Id {
		t.Fatalf("selected binding not promoted: got %s want %s", got[0].Id, selected.Id)
	}
	if got[1].Id != "runbind_1" || got[2].Id != "runbind_3" {
		t.Fatalf("unexpected binding order: %#v", got)
	}

	got = prioritizeExecutionBinding(bindings, entity.BrainRunBindings{Id: "runbind_9"}, 2)
	if len(got) != 2 {
		t.Fatalf("unexpected binding count for missing selected: got %d want %d", len(got), 2)
	}
	if got[0].Id != "runbind_9" || got[1].Id != "runbind_1" {
		t.Fatalf("unexpected prioritized order for missing selected: %#v", got)
	}
}
