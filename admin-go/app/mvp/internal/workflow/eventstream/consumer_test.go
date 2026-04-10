package eventstream

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	redis "github.com/redis/go-redis/v9"

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

func TestParseRedisStructResults(t *testing.T) {
	xreadRaw := gvar.New([]redis.XStream{
		{
			Stream: "stream-key",
			Messages: []redis.XMessage{
				{
					ID: "1716300000-0",
					Values: map[string]interface{}{
						"event_id":        "evt-1",
						"idempotency_key": "idem-1",
						"event_type":      event.EventTaskCompleted,
						"workflow_run_id": "101",
						"entity_type":     event.EntityDomainTask,
						"entity_id":       "202",
						"attempt":         "1",
						"created_at":      "1716300000",
						"payload_json":    `{"task_id":202}`,
					},
				},
			},
		},
	})
	messages, err := parseXReadGroupResult(xreadRaw)
	if err != nil {
		t.Fatalf("parse redis xreadgroup result failed: %v", err)
	}
	if len(messages) != 1 || messages[0].ID != "1716300000-0" {
		t.Fatalf("unexpected redis xreadgroup parse result: %+v", messages)
	}

	xclaimRaw := gvar.New([]redis.XMessage{
		{
			ID: "1716300000-1",
			Values: map[string]interface{}{
				"event_id":        "evt-2",
				"idempotency_key": "idem-2",
				"event_type":      event.EventTaskCompleted,
				"workflow_run_id": "102",
				"entity_type":     event.EntityDomainTask,
				"entity_id":       "203",
				"attempt":         "1",
				"created_at":      "1716300001",
				"payload_json":    `{"task_id":203}`,
			},
		},
	})
	claimed, err := parseXClaimResult(xclaimRaw)
	if err != nil {
		t.Fatalf("parse redis xclaim result failed: %v", err)
	}
	if len(claimed) != 1 || claimed[0].ID != "1716300000-1" {
		t.Fatalf("unexpected redis xclaim parse result: %+v", claimed)
	}

	pendingRaw := gvar.New([]redis.XPendingExt{
		{
			ID:         "1716300000-2",
			Consumer:   "consumer-a",
			Idle:       70 * time.Second,
			RetryCount: 1,
		},
	})
	entries := parsePendingEntries(pendingRaw)
	if len(entries) != 1 || entries[0].IdleMS != 70000 {
		t.Fatalf("unexpected redis pending parse result: %+v", entries)
	}

	groupRaw := gvar.New([]redis.XInfoGroup{
		{
			Name:      "group-a",
			Consumers: 1,
			Pending:   2,
			Lag:       3,
		},
	})
	group, ok := parseGroupInfo(groupRaw, "group-a")
	if !ok {
		t.Fatal("expected redis group info to be parsed")
	}
	if group.Pending != 2 || !group.LagKnown || group.Lag != 3 {
		t.Fatalf("unexpected redis group info parse result: %+v", group)
	}
}

func TestParseRedisStringifiedResults(t *testing.T) {
	pendingEntries := parsePendingEntries(gvar.New([]string{
		`["1716300000-0","consumer-a",70000,1]`,
	}))
	if len(pendingEntries) != 1 || pendingEntries[0].IdleMS != 70000 {
		t.Fatalf("unexpected stringified pending parse result: %+v", pendingEntries)
	}

	claimed, err := parseXClaimResult(gvar.New([]string{
		`["1716300000-1",["event_id","evt-2","idempotency_key","idem-2","event_type","task.completed","workflow_run_id","101","entity_type","domain_task","entity_id","202","attempt","1","created_at","1716300000","payload_json","{\"task_id\":202}"]]`,
	}))
	if err != nil {
		t.Fatalf("parse stringified xclaim result failed: %v", err)
	}
	if len(claimed) != 1 || claimed[0].ID != "1716300000-1" {
		t.Fatalf("unexpected stringified xclaim parse result: %+v", claimed)
	}

	group, ok := parseGroupInfo(gvar.New([]string{
		`map[consumers:2 entries-read:1 lag:0 last-delivered-id:1716300000-1 name:group-a pending:1]`,
	}), "group-a")
	if !ok {
		t.Fatal("expected stringified group info to be parsed")
	}
	if group.Pending != 1 || !group.LagKnown || group.Lag != 0 {
		t.Fatalf("unexpected stringified group info parse result: %+v", group)
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
