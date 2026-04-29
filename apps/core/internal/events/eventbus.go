package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// WorkflowEventType defines the event types that drive the closed-loop workflow.
type WorkflowEventType string

const (
	// PlanReviewRejected triggers auto-redesign.
	PlanReviewRejected WorkflowEventType = "plan_review_rejected"
	// PlanReviewApproved triggers auto-compile after review passes.
	PlanReviewApproved WorkflowEventType = "plan_review_approved"
	// RepairDraftReady triggers auto-reexecution when repair plan is ready.
	RepairDraftReady WorkflowEventType = "repair_draft_ready"
	// RunTerminal triggers auto-acceptance adjudication.
	RunTerminal WorkflowEventType = "run_terminal"
	// AcceptanceFailed triggers repair draft creation.
	AcceptanceFailed WorkflowEventType = "acceptance_failed"
	// AcceptancePassed triggers project completion.
	AcceptancePassed WorkflowEventType = "acceptance_passed"
	// BrowserCheckCompleted carries browser verification results.
	BrowserCheckCompleted WorkflowEventType = "browser_check_completed"
	// VerifierCheckCompleted carries verifier verification results.
	VerifierCheckCompleted WorkflowEventType = "verifier_check_completed"
)

// WorkflowEvent is the unit of work in the event-driven architecture.
type WorkflowEvent struct {
	ID         string                 `json:"id"`
	ProjectID  string                 `json:"project_id"`
	EventType  WorkflowEventType      `json:"event_type"`
	Payload    map[string]interface{} `json:"payload"`
	Status     string                 `json:"status"`
	RetryCount int                    `json:"retry_count"`
	ErrorMsg   string                 `json:"error_message,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Handler is a callback function for workflow events.
type Handler func(ctx context.Context, evt *WorkflowEvent) error

// EventBus is a lightweight in-memory pub/sub bus backed by SQLite for durability.
type EventBus struct {
	mu         sync.RWMutex
	handlers   map[WorkflowEventType][]Handler
	buffer     chan *WorkflowEvent
	bufferSize int
}

// NewEventBus creates an event bus with a buffered channel.
func NewEventBus(bufferSize int) *EventBus {
	if bufferSize <= 0 {
		bufferSize = 256
	}
	return &EventBus{
		handlers:   make(map[WorkflowEventType][]Handler),
		buffer:     make(chan *WorkflowEvent, bufferSize),
		bufferSize: bufferSize,
	}
}

// Subscribe registers a handler for a specific event type.
func (eb *EventBus) Subscribe(eventType WorkflowEventType, h Handler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], h)
}

// Publish emits a workflow event. It is non-blocking: if the buffer is full,
// the event is dropped and a warning is logged. Events are also persisted to
// the workflow_events table for durability and replay.
func (eb *EventBus) Publish(ctx context.Context, evt *WorkflowEvent) {
	if evt.ID == "" {
		evt.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	if evt.CreatedAt.IsZero() {
		evt.CreatedAt = time.Now().UTC()
	}
	if evt.Status == "" {
		evt.Status = "pending"
	}

	// Persist to DB (best-effort, do not block publish on DB errors).
	go eb.persistEvent(ctx, evt)

	// Emit to in-memory subscribers.
	select {
	case eb.buffer <- evt:
		// Handed off to dispatcher goroutine.
	default:
		g.Log().Warningf(ctx, "[eventbus] buffer full, event %s dropped", evt.ID)
	}
}

// persistEvent saves the event to the workflow_events table.
func (eb *EventBus) persistEvent(ctx context.Context, evt *WorkflowEvent) {
	payloadJSON, _ := json.Marshal(evt.Payload)
	_, err := g.DB().Exec(ctx,
		`INSERT INTO workflow_events (id, project_id, event_type, payload_json, status, retry_count, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   status=excluded.status,
		   retry_count=excluded.retry_count,
		   error_message=excluded.error_message`,
		evt.ID, evt.ProjectID, string(evt.EventType), string(payloadJSON),
		evt.Status, evt.RetryCount, evt.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		g.Log().Warningf(ctx, "[eventbus] persist event %s failed: %v", evt.ID, err)
	}
}

// StartDispatcher starts a goroutine that consumes the event buffer and
// dispatches events to registered handlers.
func (eb *EventBus) StartDispatcher(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-eb.buffer:
				eb.dispatch(ctx, evt)
			}
		}
	}()
}

func (eb *EventBus) dispatch(ctx context.Context, evt *WorkflowEvent) {
	eb.mu.RLock()
	handlers := eb.handlers[evt.EventType]
	eb.mu.RUnlock()

	if len(handlers) == 0 {
		g.Log().Debugf(ctx, "[eventbus] no handlers for event type %s", evt.EventType)
		return
	}

	for _, h := range handlers {
		if err := h(ctx, evt); err != nil {
			g.Log().Warningf(ctx, "[eventbus] handler error for %s: %v", evt.EventType, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Global singleton (lazy init)
// ---------------------------------------------------------------------------

var (
	globalBus     *EventBus
	globalBusOnce sync.Once
)

// Bus returns the global event bus instance.
func Bus() *EventBus {
	globalBusOnce.Do(func() {
		globalBus = NewEventBus(256)
	})
	return globalBus
}

// Publish is a convenience wrapper around Bus().Publish.
func Publish(ctx context.Context, evt *WorkflowEvent) {
	Bus().Publish(ctx, evt)
}
