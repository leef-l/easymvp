package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type mockStreamSink struct {
	status StreamStatus
	err    error
}

func (m *mockStreamSink) Publish(event Event) error {
	return m.err
}

func (m *mockStreamSink) Status() StreamStatus {
	return m.status
}

type mockDurableClaim struct {
	handled int
	failed  int
}

func (m *mockDurableClaim) MarkHandled(ctx context.Context) error {
	m.handled++
	return nil
}

func (m *mockDurableClaim) MarkFailed(ctx context.Context, handleErr error) error {
	m.failed++
	return nil
}

func TestNewPublisherWithStreamSink(t *testing.T) {
	bus := NewBus()
	sink := &mockStreamSink{
		status: StreamStatus{
			Enabled:    true,
			StreamName: "easymvp:workflow:task-events",
		},
	}
	p := NewPublisher(bus, WithStreamSink(sink))

	status := p.StreamStatus()
	if !status.Enabled {
		t.Fatal("expected stream status enabled")
	}
	if status.StreamName != "easymvp:workflow:task-events" {
		t.Fatalf("unexpected stream name: %s", status.StreamName)
	}
}

func TestSetStreamSink(t *testing.T) {
	bus := NewBus()
	p := NewPublisher(bus)
	if p.StreamStatus().Enabled {
		t.Fatal("expected disabled status by default")
	}

	sink := &mockStreamSink{
		status: StreamStatus{
			Enabled:   true,
			Degraded:  true,
			LastError: errors.New("redis down").Error(),
		},
	}
	p.SetStreamSink(sink)

	status := p.StreamStatus()
	if !status.Enabled || !status.Degraded {
		t.Fatalf("unexpected stream status: %+v", status)
	}
}

func TestPublisherEmitSkipsDuplicateDurableClaim(t *testing.T) {
	originalBeginClaim := beginPublishEventClaimFn
	defer func() {
		beginPublishEventClaimFn = originalBeginClaim
	}()

	bus := NewBus()
	p := NewPublisher(bus)
	published := 0
	bus.Subscribe("*", func(evt Event) {
		published++
	})

	beginPublishEventClaimFn = func(ctx context.Context, evt Event) (DurableEventClaim, bool, error) {
		return nil, false, nil
	}

	if err := p.Emit(context.Background(), Event{
		WorkflowRunID: 1,
		EntityType:    EntityWorkflowRun,
		EventType:     EventSchedulerWakeup,
	}); err != nil {
		t.Fatalf("emit failed: %v", err)
	}
	if published != 0 {
		t.Fatalf("expected duplicate event to be skipped, got published=%d", published)
	}
}

func TestPublisherEmitMarksDurableClaimHandledOnSuccess(t *testing.T) {
	originalBeginClaim := beginPublishEventClaimFn
	originalPersistEvent := persistWorkflowEventFn
	defer func() {
		beginPublishEventClaimFn = originalBeginClaim
		persistWorkflowEventFn = originalPersistEvent
	}()

	bus := NewBus()
	p := NewPublisher(bus)
	published := 0
	claim := &mockDurableClaim{}
	bus.Subscribe("*", func(evt Event) {
		published++
	})

	beginPublishEventClaimFn = func(ctx context.Context, evt Event) (DurableEventClaim, bool, error) {
		return claim, true, nil
	}
	persistWorkflowEventFn = func(ctx context.Context, evt Event, recordID int64, createdAt time.Time) error {
		return nil
	}

	if err := p.Emit(context.Background(), Event{
		WorkflowRunID: 2,
		EntityType:    EntityWorkflowRun,
		EventType:     EventSchedulerWakeup,
	}); err != nil {
		t.Fatalf("emit failed: %v", err)
	}
	if published != 1 {
		t.Fatalf("expected event to be published once, got %d", published)
	}
	if claim.handled != 1 || claim.failed != 0 {
		t.Fatalf("unexpected claim state: handled=%d failed=%d", claim.handled, claim.failed)
	}
}

func TestPublisherEmitFallsBackWhenDurableLedgerMigrationMissing(t *testing.T) {
	originalBeginClaim := beginPublishEventClaimFn
	originalPersistEvent := persistWorkflowEventFn
	defer func() {
		beginPublishEventClaimFn = originalBeginClaim
		persistWorkflowEventFn = originalPersistEvent
	}()

	bus := NewBus()
	p := NewPublisher(bus)
	published := 0
	bus.Subscribe("*", func(evt Event) {
		published++
	})

	beginPublishEventClaimFn = func(ctx context.Context, evt Event) (DurableEventClaim, bool, error) {
		return nil, false, errors.New("Error 1146 (42S02): Table 'test.mvp_workflow_event_ledger' doesn't exist")
	}
	persistWorkflowEventFn = func(ctx context.Context, evt Event, recordID int64, createdAt time.Time) error {
		return nil
	}

	if err := p.Emit(context.Background(), Event{
		WorkflowRunID: 3,
		EntityType:    EntityWorkflowRun,
		EventType:     EventSchedulerWakeup,
	}); err != nil {
		t.Fatalf("emit failed: %v", err)
	}
	if published != 1 {
		t.Fatalf("expected event to be published once, got %d", published)
	}
}

func TestPersistWorkflowEventFallsBackToLegacyInsertWhenMetadataColumnsMissing(t *testing.T) {
	originalInsert := insertWorkflowEventRecordFn
	defer func() {
		insertWorkflowEventRecordFn = originalInsert
	}()

	call := 0
	insertWorkflowEventRecordFn = func(ctx context.Context, data g.Map) error {
		call++
		switch call {
		case 1:
			if _, ok := data["event_id"]; !ok {
				t.Fatal("expected first insert to include event_id")
			}
			if _, ok := data["idempotency_key"]; !ok {
				t.Fatal("expected first insert to include idempotency_key")
			}
			if _, ok := data["attempt"]; !ok {
				t.Fatal("expected first insert to include attempt")
			}
			return errors.New("Error 1054 (42S22): Unknown column 'event_id' in 'field list'")
		case 2:
			if _, ok := data["event_id"]; ok {
				t.Fatal("expected legacy fallback insert to omit event_id")
			}
			if _, ok := data["idempotency_key"]; ok {
				t.Fatal("expected legacy fallback insert to omit idempotency_key")
			}
			if _, ok := data["attempt"]; ok {
				t.Fatal("expected legacy fallback insert to omit attempt")
			}
			return nil
		default:
			t.Fatalf("unexpected insert call: %d", call)
			return nil
		}
	}

	err := persistWorkflowEventFn(context.Background(), Event{
		EventID:        "evt-1",
		WorkflowRunID:  7,
		EntityType:     EntityWorkflowRun,
		EventType:      EventSchedulerWakeup,
		IdempotencyKey: "wf:7:type:scheduler.wakeup",
		Attempt:        2,
	}, 123, time.Now())
	if err != nil {
		t.Fatalf("persistWorkflowEventFn failed: %v", err)
	}
	if call != 2 {
		t.Fatalf("expected 2 insert attempts, got %d", call)
	}
}
