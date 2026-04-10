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

func TestEnsureMetadataAddsStageRunIDToFallbackIdempotencyKey(t *testing.T) {
	evtA := Event{
		WorkflowRunID: 5,
		StageRunID:    ptrInt64(101),
		EntityType:    EntityWorkflowRun,
		EventType:     "verification.repair_requested",
	}
	evtB := Event{
		WorkflowRunID: 5,
		StageRunID:    ptrInt64(102),
		EntityType:    EntityWorkflowRun,
		EventType:     "verification.repair_requested",
	}

	evtA = evtA.EnsureMetadata()
	evtB = evtB.EnsureMetadata()

	if evtA.IdempotencyKey == evtB.IdempotencyKey {
		t.Fatalf("expected stage-scoped idempotency keys to differ: %s", evtA.IdempotencyKey)
	}
	if !strings.Contains(evtA.IdempotencyKey, ":stage:101:") {
		t.Fatalf("expected stage segment in idempotency key: %s", evtA.IdempotencyKey)
	}
}

func TestEnsureMetadataResolvesFailedTaskIDFromPayload(t *testing.T) {
	evt := Event{
		WorkflowRunID: 7,
		StageRunID:    ptrInt64(200),
		EntityType:    "verification_issue",
		EventType:     "verification.repair_requested",
		Payload: map[string]interface{}{
			"failed_task_id": 998,
		},
	}

	evt = evt.EnsureMetadata()
	if !strings.Contains(evt.IdempotencyKey, "wf:7:task:998:") {
		t.Fatalf("unexpected idempotency key: %s", evt.IdempotencyKey)
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}
