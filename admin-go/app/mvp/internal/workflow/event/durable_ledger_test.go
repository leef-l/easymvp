package event

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

func TestBeginDurableEventClaimCreatesNewClaim(t *testing.T) {
	originalInsert := insertDurableEventLedgerFn
	originalRevive := reviveDurableEventLedgerFn
	defer func() {
		insertDurableEventLedgerFn = originalInsert
		reviveDurableEventLedgerFn = originalRevive
	}()

	insertDurableEventLedgerFn = func(ctx context.Context, data g.Map) error {
		if data["scope"] != DurableEventClaimScopePublish {
			t.Fatalf("unexpected scope: %v", data["scope"])
		}
		if data["status"] != durableEventLedgerStatusHandling {
			t.Fatalf("unexpected status: %v", data["status"])
		}
		return nil
	}
	reviveDurableEventLedgerFn = func(ctx context.Context, scope, idempotencyKey string, evt Event, now time.Time) (int64, error) {
		t.Fatal("revive should not be called on fresh claim")
		return 0, nil
	}

	claim, shouldProcess, err := BeginDurableEventClaim(context.Background(), "", Event{
		WorkflowRunID: 1,
		EntityType:    EntityWorkflowRun,
		EventType:     EventSchedulerWakeup,
	})
	if err != nil {
		t.Fatalf("BeginDurableEventClaim failed: %v", err)
	}
	if !shouldProcess {
		t.Fatal("expected fresh claim to continue processing")
	}
	if claim == nil {
		t.Fatal("expected non-nil claim")
	}
}

func TestBeginDurableEventClaimRevivesFailedClaimOnDuplicate(t *testing.T) {
	originalInsert := insertDurableEventLedgerFn
	originalRevive := reviveDurableEventLedgerFn
	defer func() {
		insertDurableEventLedgerFn = originalInsert
		reviveDurableEventLedgerFn = originalRevive
	}()

	insertDurableEventLedgerFn = func(ctx context.Context, data g.Map) error {
		return errors.New("Duplicate entry 'workflow.publish:abc' for key")
	}
	reviveDurableEventLedgerFn = func(ctx context.Context, scope, idempotencyKey string, evt Event, now time.Time) (int64, error) {
		if scope != DurableEventClaimScopePublish {
			t.Fatalf("unexpected scope: %s", scope)
		}
		if strings.TrimSpace(idempotencyKey) == "" {
			t.Fatal("expected idempotency key")
		}
		return 1, nil
	}

	claim, shouldProcess, err := BeginDurableEventClaim(context.Background(), DurableEventClaimScopePublish, Event{
		WorkflowRunID: 2,
		EntityType:    EntityWorkflowRun,
		EventType:     EventSchedulerWakeup,
	})
	if err != nil {
		t.Fatalf("BeginDurableEventClaim failed: %v", err)
	}
	if !shouldProcess {
		t.Fatal("expected failed claim revive to continue processing")
	}
	if claim == nil {
		t.Fatal("expected revived claim")
	}
}

func TestBeginDurableEventClaimSkipsHandledDuplicate(t *testing.T) {
	originalInsert := insertDurableEventLedgerFn
	originalRevive := reviveDurableEventLedgerFn
	defer func() {
		insertDurableEventLedgerFn = originalInsert
		reviveDurableEventLedgerFn = originalRevive
	}()

	insertDurableEventLedgerFn = func(ctx context.Context, data g.Map) error {
		return errors.New("Duplicate entry 'workflow.publish:abc' for key")
	}
	reviveDurableEventLedgerFn = func(ctx context.Context, scope, idempotencyKey string, evt Event, now time.Time) (int64, error) {
		return 0, nil
	}

	claim, shouldProcess, err := BeginDurableEventClaim(context.Background(), DurableEventClaimScopePublish, Event{
		WorkflowRunID: 3,
		EntityType:    EntityWorkflowRun,
		EventType:     EventSchedulerWakeup,
	})
	if err != nil {
		t.Fatalf("BeginDurableEventClaim failed: %v", err)
	}
	if shouldProcess {
		t.Fatal("expected handled duplicate to skip processing")
	}
	if claim != nil {
		t.Fatalf("expected nil claim when duplicate already handled, got %#v", claim)
	}
}

func TestDurableEventClaimMarkHandledUpdatesLedger(t *testing.T) {
	originalUpdate := updateDurableEventLedgerFn
	defer func() {
		updateDurableEventLedgerFn = originalUpdate
	}()

	updateDurableEventLedgerFn = func(ctx context.Context, scope, idempotencyKey string, data g.Map) error {
		if scope != DurableEventClaimScopePublish {
			t.Fatalf("unexpected scope: %s", scope)
		}
		if idempotencyKey != "wf:1:event" {
			t.Fatalf("unexpected key: %s", idempotencyKey)
		}
		if data["status"] != durableEventLedgerStatusHandled {
			t.Fatalf("unexpected status: %v", data["status"])
		}
		if data["handled_at"] == nil {
			t.Fatal("expected handled_at timestamp")
		}
		return nil
	}

	err := (&durableEventClaim{scope: DurableEventClaimScopePublish, idempotencyKey: "wf:1:event"}).MarkHandled(context.Background())
	if err != nil {
		t.Fatalf("MarkHandled failed: %v", err)
	}
}

func TestDurableEventClaimMarkFailedTrimsError(t *testing.T) {
	originalUpdate := updateDurableEventLedgerFn
	defer func() {
		updateDurableEventLedgerFn = originalUpdate
	}()

	longErr := errors.New(strings.Repeat("x", 520))
	updateDurableEventLedgerFn = func(ctx context.Context, scope, idempotencyKey string, data g.Map) error {
		if data["status"] != durableEventLedgerStatusFailed {
			t.Fatalf("unexpected status: %v", data["status"])
		}
		lastError, ok := data["last_error"].(string)
		if !ok {
			t.Fatalf("expected string last_error, got %T", data["last_error"])
		}
		if len(lastError) != 500 {
			t.Fatalf("expected trimmed last_error length 500, got %d", len(lastError))
		}
		return nil
	}

	err := (&durableEventClaim{scope: DurableEventClaimScopePublish, idempotencyKey: "wf:2:event"}).MarkFailed(context.Background(), longErr)
	if err != nil {
		t.Fatalf("MarkFailed failed: %v", err)
	}
}
