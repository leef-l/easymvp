package eventstream

import (
	"context"
	"errors"
	"testing"

	"github.com/gogf/gf/v2/container/gvar"

	"easymvp/app/mvp/internal/workflow/event"
)

func TestParseXReadGroupResult(t *testing.T) {
	raw := gvar.New([]any{
		[]any{
			"stream-key",
			[]any{
				[]any{
					"1716300000-0",
					[]any{
						"event_id", "evt-1",
						"event_type", event.EventTaskFailed,
						"workflow_run_id", "101",
						"entity_type", event.EntityDomainTask,
						"entity_id", "202",
						"attempt", "2",
						"created_at", "1716300000",
						"payload_json", `{"task_id":202}`,
					},
				},
			},
		},
	})
	messages, err := parseXReadGroupResult(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].ID != "1716300000-0" {
		t.Fatalf("unexpected message id: %s", messages[0].ID)
	}
	evt := decodeStreamEvent(messages[0].Fields)
	if evt.WorkflowRunID != 101 {
		t.Fatalf("unexpected workflow run id: %d", evt.WorkflowRunID)
	}
	if evt.Attempt != 2 {
		t.Fatalf("unexpected attempt: %d", evt.Attempt)
	}
}

func TestParsePendingEntries(t *testing.T) {
	raw := gvar.New([]any{
		[]any{"1716300000-0", "consumer-a", int64(70000), int64(1)},
		[]any{"1716300000-1", "consumer-b", int64(1200), int64(2)},
	})
	entries := parsePendingEntries(raw)
	if len(entries) != 2 {
		t.Fatalf("expected 2 pending entries, got %d", len(entries))
	}
	if entries[0].IdleMS != 70000 {
		t.Fatalf("unexpected idle ms: %d", entries[0].IdleMS)
	}
}

func TestConsumerProcessOnceHandlesAndAcks(t *testing.T) {
	redis := &fakeRedis{
		resp: []fakeRedisResp{
			{Err: errors.New("BUSYGROUP Consumer Group name already exists")},
			{Value: []any{}}, // XPENDING empty
			{Value: []any{
				[]any{
					"stream-key",
					[]any{
						[]any{
							"1716300000-0",
							[]any{
								"event_id", "evt-2",
								"event_type", event.EventTaskCompleted,
								"workflow_run_id", "111",
								"entity_type", event.EntityDomainTask,
								"entity_id", "222",
								"attempt", "1",
								"created_at", "1716300000",
								"payload_json", `{"task_id":222}`,
							},
						},
					},
				},
			}},
			{Value: int64(1)}, // XACK
		},
	}

	handled := 0
	consumer := NewConsumer(redis, Config{
		Enabled:         true,
		ConsumerEnabled: true,
		StreamName:      "stream-key",
		ConsumerGroup:   "group-a",
		ConsumerName:    "consumer-a",
	}, func(ctx context.Context, evt event.Event) error {
		handled++
		if evt.EventType != event.EventTaskCompleted {
			t.Fatalf("unexpected event type: %s", evt.EventType)
		}
		return nil
	})

	if err := consumer.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("process once failed: %v", err)
	}
	if handled != 1 {
		t.Fatalf("expected handler called once, got %d", handled)
	}
	if len(redis.calls) < 4 {
		t.Fatalf("unexpected call count: %d", len(redis.calls))
	}
	if redis.calls[len(redis.calls)-1].Command != "XACK" {
		t.Fatalf("expected XACK as last call, got %s", redis.calls[len(redis.calls)-1].Command)
	}
}

func TestConsumerSnapshotReadsRuntimeFromRedis(t *testing.T) {
	redis := &fakeRedis{
		resp: []fakeRedisResp{
			{Value: []any{
				[]any{
					"name", "group-a",
					"consumers", int64(1),
					"pending", int64(3),
					"last-delivered-id", "1716300000-0",
					"entries-read", int64(10),
					"lag", int64(7),
				},
			}},
		},
	}
	consumer := NewConsumer(redis, Config{
		Enabled:         true,
		ConsumerEnabled: true,
		StreamName:      "stream-key",
		ConsumerGroup:   "group-a",
		ConsumerName:    "consumer-a",
	}, nil)

	snapshot := consumer.Snapshot(context.Background())
	if !snapshot.PendingKnown || snapshot.Pending != 3 {
		t.Fatalf("unexpected pending snapshot: %+v", snapshot)
	}
	if !snapshot.LagKnown || snapshot.Lag != 7 {
		t.Fatalf("unexpected lag snapshot: %+v", snapshot)
	}
	if snapshot.ConsumerCreated != true {
		t.Fatalf("expected consumer to be marked created: %+v", snapshot)
	}
}

func TestConsumerHeartbeatAndStartedState(t *testing.T) {
	consumer := NewConsumer(nil, Config{
		Enabled:         true,
		ConsumerEnabled: true,
		StreamName:      "stream-key",
		ConsumerGroup:   "group-a",
		ConsumerName:    "consumer-a",
	}, nil)

	if consumer.IsStarted() {
		t.Fatal("expected consumer to be stopped initially")
	}
	consumer.PulseHeartbeat()
	snapshot := consumer.Snapshot(context.Background())
	if snapshot.WorkerHeartbeatAt.IsZero() {
		t.Fatal("expected heartbeat to be recorded")
	}
	if snapshot.ConsumerStarted {
		t.Fatal("expected consumer to remain stopped")
	}
}
