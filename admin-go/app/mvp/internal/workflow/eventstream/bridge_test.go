package eventstream

import (
	"context"
	"errors"
	"testing"

	"github.com/gogf/gf/v2/container/gvar"

	"easymvp/app/mvp/internal/workflow/event"
)

type fakeRedisCall struct {
	Command string
	Args    []any
}

type fakeRedisResp struct {
	Value any
	Err   error
}

type fakeRedis struct {
	calls []fakeRedisCall
	resp  []fakeRedisResp
}

func (f *fakeRedis) Do(ctx context.Context, command string, args ...any) (*gvar.Var, error) {
	f.calls = append(f.calls, fakeRedisCall{
		Command: command,
		Args:    append([]any(nil), args...),
	})
	if len(f.resp) == 0 {
		return gvar.New(nil), nil
	}
	out := f.resp[0]
	f.resp = f.resp[1:]
	return gvar.New(out.Value), out.Err
}

func TestBridgePublishWritesXAdd(t *testing.T) {
	redis := &fakeRedis{
		resp: []fakeRedisResp{
			{Value: "1716300000-0"},
		},
	}
	bridge := NewBridge(redis, Config{Enabled: true, StreamName: "test-stream"})

	err := bridge.Publish(event.Event{
		WorkflowRunID: 3,
		EntityType:    event.EntityDomainTask,
		EntityID:      ptr(55),
		EventType:     event.EventTaskFailed,
		Payload: map[string]interface{}{
			"task_id": 55,
			"attempt": 2,
		},
	})
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if len(redis.calls) != 1 {
		t.Fatalf("expected 1 redis call, got %d", len(redis.calls))
	}
	if redis.calls[0].Command != "XADD" {
		t.Fatalf("expected XADD, got %s", redis.calls[0].Command)
	}
	status := bridge.Status()
	if status.Degraded {
		t.Fatalf("expected healthy status, got %+v", status)
	}
}

func TestBridgePublishDegradesWhenRedisUnavailable(t *testing.T) {
	bridge := NewBridge(nil, Config{
		Enabled:       true,
		RedisRequired: false,
	})
	if err := bridge.Publish(event.Event{EventType: event.EventSchedulerWakeup}); err != nil {
		t.Fatalf("expected nil err on degraded optional redis, got %v", err)
	}
	status := bridge.Status()
	if !status.Degraded {
		t.Fatalf("expected degraded status, got %+v", status)
	}
}

func TestBridgePublishReturnsErrorWhenRedisRequired(t *testing.T) {
	redis := &fakeRedis{
		resp: []fakeRedisResp{
			{Err: errors.New("redis down")},
		},
	}
	bridge := NewBridge(redis, Config{
		Enabled:       true,
		RedisRequired: true,
	})
	if err := bridge.Publish(event.Event{EventType: event.EventSchedulerWakeup}); err == nil {
		t.Fatal("expected error when redis is required")
	}
}

func ptr(v int64) *int64 {
	return &v
}
