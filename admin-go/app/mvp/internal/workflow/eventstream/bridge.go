package eventstream

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"easymvp/app/mvp/internal/workflow/event"
)

// Bridge 将 event.Publisher 的事件投递到 Redis Stream。
type Bridge struct {
	mu     sync.RWMutex
	cfg    Config
	redis  RedisCommander
	status event.StreamStatus
}

// NewBridge 创建事件流桥接器。
func NewBridge(redis RedisCommander, cfg Config) *Bridge {
	cfg = cfg.Normalize()
	status := event.StreamStatus{
		Enabled:    cfg.Enabled,
		StreamName: cfg.StreamName,
		UpdatedAt:  time.Now(),
	}
	if cfg.Enabled && redis == nil {
		status.Degraded = true
		status.LastError = "redis client unavailable"
	}
	return &Bridge{
		cfg:    cfg,
		redis:  redis,
		status: status,
	}
}

// Publish 将事件写入 Redis Stream。
func (b *Bridge) Publish(evt event.Event) error {
	if !b.cfg.Enabled {
		return nil
	}
	evt = evt.EnsureMetadata()
	if b.redis == nil {
		err := fmt.Errorf("redis client unavailable")
		b.markDegraded(err)
		if b.cfg.RedisRequired {
			return err
		}
		return nil
	}

	payloadJSON, err := marshalPayload(evt.Payload)
	if err != nil {
		b.markDegraded(err)
		if b.cfg.RedisRequired {
			return err
		}
		return nil
	}

	args := []any{
		b.cfg.StreamName,
		"*",
		"event_id", evt.EventID,
		"idempotency_key", evt.IdempotencyKey,
		"event_type", evt.EventType,
		"workflow_run_id", strconv.FormatInt(evt.WorkflowRunID, 10),
		"entity_type", evt.EntityType,
		"attempt", strconv.Itoa(evt.Attempt),
		"created_at", strconv.FormatInt(evt.CreatedAtUnix, 10),
		"payload_json", payloadJSON,
	}
	if evt.StageRunID != nil {
		args = append(args, "stage_run_id", strconv.FormatInt(*evt.StageRunID, 10))
	}
	if evt.EntityID != nil {
		args = append(args, "entity_id", strconv.FormatInt(*evt.EntityID, 10))
	}

	if _, err := b.redis.Do(context.Background(), "XADD", args...); err != nil {
		b.markDegraded(err)
		if b.cfg.RedisRequired {
			return err
		}
		return nil
	}

	b.markHealthy()
	return nil
}

// Status 返回桥接器状态。
func (b *Bridge) Status() event.StreamStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

func (b *Bridge) markHealthy() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status.Degraded = false
	b.status.LastError = ""
	b.status.UpdatedAt = time.Now()
}

func (b *Bridge) markDegraded(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status.Degraded = true
	if err != nil {
		b.status.LastError = err.Error()
	}
	b.status.UpdatedAt = time.Now()
}

func marshalPayload(payload interface{}) (string, error) {
	if payload == nil {
		return "{}", nil
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
