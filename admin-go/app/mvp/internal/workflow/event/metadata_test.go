package event

import (
	"strings"
	"testing"
)

func TestEnsureMetadataFillsDefaults(t *testing.T) {
	evt := Event{
		WorkflowRunID: 11,
		EntityType:    EntityDomainTask,
		EntityID:      ptrInt64(22),
		EventType:     EventTaskFailed,
		Payload: map[string]interface{}{
			"attempt": 3,
		},
	}

	evt = evt.EnsureMetadata()

	if strings.TrimSpace(evt.EventID) == "" {
		t.Fatal("expected event_id to be filled")
	}
	if evt.CreatedAtUnix <= 0 {
		t.Fatal("expected created_at_unix to be filled")
	}
	if evt.Attempt != 3 {
		t.Fatalf("expected attempt=3, got %d", evt.Attempt)
	}
	if !strings.Contains(evt.IdempotencyKey, "wf:11:task:22") {
		t.Fatalf("unexpected idempotency key: %s", evt.IdempotencyKey)
	}
}

func TestEnsureMetadataFallsBackAttemptToOne(t *testing.T) {
	evt := Event{
		WorkflowRunID: 1,
		EntityType:    EntityWorkflowRun,
		EventType:     EventSchedulerWakeup,
	}

	evt = evt.EnsureMetadata()
	if evt.Attempt != 1 {
		t.Fatalf("expected attempt=1, got %d", evt.Attempt)
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}
