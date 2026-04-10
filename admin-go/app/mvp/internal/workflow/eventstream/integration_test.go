package eventstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gredis"

	"easymvp/app/mvp/internal/workflow/event"
)

func TestBridgeRecoversHealthyStatusAfterRedisRestored(t *testing.T) {
	redis := newStatefulStreamRedis("bridge-status-stream")
	bridge := NewBridge(redis, Config{
		Enabled:       true,
		StreamName:    "bridge-status-stream",
		RedisRequired: false,
	})

	if err := bridge.Publish(testStreamEvent("evt-healthy", "idem-healthy")); err != nil {
		t.Fatalf("healthy publish failed: %v", err)
	}
	if status := bridge.Status(); status.Degraded {
		t.Fatalf("expected healthy status after successful publish, got %+v", status)
	}

	redis.SetAvailable(false)
	if err := bridge.Publish(testStreamEvent("evt-degraded", "idem-degraded")); err != nil {
		t.Fatalf("optional redis publish should not hard-fail on outage, got %v", err)
	}
	if status := bridge.Status(); !status.Degraded {
		t.Fatalf("expected degraded status after outage, got %+v", status)
	}

	redis.SetAvailable(true)
	if err := bridge.Publish(testStreamEvent("evt-recovered", "idem-recovered")); err != nil {
		t.Fatalf("publish after redis recovery failed: %v", err)
	}
	if status := bridge.Status(); status.Degraded {
		t.Fatalf("expected healthy status after recovery, got %+v", status)
	}
}

func TestConsumerRecoversAfterRedisRestoredAndReclaimsPending(t *testing.T) {
	redis := newStatefulStreamRedis("consumer-reclaim-stream")
	redis.SeedPending(testStreamEvent("evt-pending", "idem-pending"), "stale-consumer", 2*time.Second)
	redis.SeedNew(testStreamEvent("evt-backlog", "idem-backlog"))

	consumer := NewConsumer(redis, Config{
		Enabled:            true,
		ConsumerEnabled:    true,
		StreamName:         "consumer-reclaim-stream",
		ConsumerGroup:      "consumer-reclaim-group",
		ConsumerName:       "consumer-reclaimer",
		ReclaimIdleSeconds: 1,
		ReclaimCount:       10,
		ReadCount:          10,
	}, func(ctx context.Context, evt event.Event) error {
		return nil
	})

	redis.SetAvailable(false)
	err := consumer.ProcessOnce(context.Background())
	if err == nil {
		t.Fatal("expected redis outage to fail consumer poll")
	}
	if snapshot := consumer.Snapshot(context.Background()); !snapshot.Degraded {
		t.Fatalf("expected degraded snapshot during outage, got %+v", snapshot)
	}

	redis.SetAvailable(true)
	if err := consumer.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("consumer poll after recovery failed: %v", err)
	}

	snapshot := consumer.Snapshot(context.Background())
	if snapshot.Degraded {
		t.Fatalf("expected healthy snapshot after recovery, got %+v", snapshot)
	}
	if !snapshot.PendingKnown || snapshot.Pending != 0 {
		t.Fatalf("expected pending backlog to be drained, got %+v", snapshot)
	}
	if snapshot.ReclaimAttempts == 0 || snapshot.ReclaimedMessages == 0 {
		t.Fatalf("expected reclaim metrics to be recorded, got %+v", snapshot)
	}
	if snapshot.LastAckAt.IsZero() {
		t.Fatalf("expected ack timestamp after recovery, got %+v", snapshot)
	}
	if redis.AckedCount() != 2 {
		t.Fatalf("expected both reclaimed and backlog messages acked, got %d", redis.AckedCount())
	}
}

func TestConsumerRealRedisReclaimsPendingAfterRecovery(t *testing.T) {
	addr := strings.TrimSpace(os.Getenv("EASYMVP_TEST_REDIS_ADDR"))
	if addr == "" {
		addr = strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	}
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	pass := strings.TrimSpace(os.Getenv("EASYMVP_TEST_REDIS_PASS"))
	if pass == "" {
		pass = strings.TrimSpace(os.Getenv("REDIS_PASS"))
	}
	if pass == "" {
		t.Skip("set EASYMVP_TEST_REDIS_PASS or REDIS_PASS to run real redis workflow event integration")
	}

	goodClient, err := gredis.New(&gredis.Config{Address: addr, Pass: pass, Db: 0})
	if err != nil {
		t.Fatalf("create redis client failed: %v", err)
	}
	badClient, err := gredis.New(&gredis.Config{Address: addr, Pass: pass + "-wrong", Db: 0})
	if err != nil {
		t.Fatalf("create bad redis client failed: %v", err)
	}

	ctx := context.Background()
	streamName := fmt.Sprintf("easymvp:test:workflow:%d", time.Now().UnixNano())
	groupName := streamName + ":group"
	cfg := Config{
		Enabled:            true,
		ConsumerEnabled:    true,
		StreamName:         streamName,
		ConsumerGroup:      groupName,
		ConsumerName:       "reclaimer",
		ReclaimIdleSeconds: 1,
		ReclaimCount:       10,
		ReadCount:          10,
		RedisRequired:      false,
	}
	switcher := &switchableRedisCommander{delegate: badClient}
	bridge := NewBridge(switcher, cfg)

	if err := bridge.Publish(testStreamEvent("evt-outage", "idem-outage")); err != nil {
		t.Fatalf("optional bridge publish should not hard-fail during outage, got %v", err)
	}
	if status := bridge.Status(); !status.Degraded {
		t.Fatalf("expected degraded bridge status during outage, got %+v", status)
	}

	switcher.Set(goodClient)
	if err := bridge.Publish(testStreamEvent("evt-reclaim", "idem-reclaim")); err != nil {
		t.Fatalf("bridge publish after redis recovery failed: %v", err)
	}
	if status := bridge.Status(); status.Degraded {
		t.Fatalf("expected bridge to recover healthy status, got %+v", status)
	}

	t.Cleanup(func() {
		_, _ = goodClient.Do(ctx, "XGROUP", "DESTROY", streamName, groupName)
		_, _ = goodClient.Do(ctx, "DEL", streamName)
	})

	if _, err := goodClient.Do(ctx, "XGROUP", "CREATE", streamName, groupName, "0", "MKSTREAM"); err != nil &&
		!strings.Contains(strings.ToUpper(err.Error()), "BUSYGROUP") {
		t.Fatalf("create redis group failed: %v", err)
	}
	if _, err := goodClient.Do(ctx,
		"XREADGROUP",
		"GROUP", groupName, "stale-consumer",
		"COUNT", 1,
		"STREAMS", streamName, ">",
	); err != nil {
		t.Fatalf("seed pending message via stale consumer failed: %v", err)
	}

	time.Sleep(1100 * time.Millisecond)

	handled := 0
	consumer := NewConsumer(switcher, cfg, func(ctx context.Context, evt event.Event) error {
		handled++
		return nil
	})
	if err := consumer.ProcessOnce(ctx); err != nil {
		t.Fatalf("consumer reclaim poll failed: %v", err)
	}
	if handled != 1 {
		t.Fatalf("expected exactly one reclaimed message to be handled, got %d", handled)
	}

	snapshot := consumer.Snapshot(ctx)
	if snapshot.Degraded {
		t.Fatalf("expected healthy snapshot after redis recovery, got %+v", snapshot)
	}
	if !snapshot.PendingKnown || snapshot.Pending != 0 {
		t.Fatalf("expected redis pending backlog to be drained, got %+v", snapshot)
	}
	if snapshot.ReclaimAttempts == 0 || snapshot.ReclaimedMessages == 0 {
		t.Fatalf("expected reclaim metrics after real redis recovery, got %+v", snapshot)
	}
}

type switchableRedisCommander struct {
	mu       sync.RWMutex
	delegate RedisCommander
}

func (s *switchableRedisCommander) Set(delegate RedisCommander) {
	s.mu.Lock()
	s.delegate = delegate
	s.mu.Unlock()
}

func (s *switchableRedisCommander) Do(ctx context.Context, command string, args ...any) (*gvar.Var, error) {
	s.mu.RLock()
	delegate := s.delegate
	s.mu.RUnlock()
	if delegate == nil {
		return nil, errors.New("redis client unavailable")
	}
	return delegate.Do(ctx, command, args...)
}

type statefulStreamRedis struct {
	mu         sync.Mutex
	available  bool
	stream     string
	groupReady bool
	groupName  string
	seq        int
	messages   []*statefulStreamMessage
}

type statefulStreamMessage struct {
	id           string
	fields       map[string]string
	pending      bool
	acked        bool
	consumer     string
	lastDelivery time.Time
}

func newStatefulStreamRedis(stream string) *statefulStreamRedis {
	return &statefulStreamRedis{
		available: true,
		stream:    stream,
	}
}

func (r *statefulStreamRedis) SetAvailable(available bool) {
	r.mu.Lock()
	r.available = available
	r.mu.Unlock()
}

func (r *statefulStreamRedis) SeedPending(evt event.Event, consumer string, idle time.Duration) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	message := r.addMessageLocked(evt)
	message.pending = true
	message.consumer = consumer
	message.lastDelivery = time.Now().Add(-idle)
	return message.id
}

func (r *statefulStreamRedis) SeedNew(evt event.Event) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.addMessageLocked(evt).id
}

func (r *statefulStreamRedis) AckedCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, message := range r.messages {
		if message.acked {
			count++
		}
	}
	return count
}

func (r *statefulStreamRedis) Do(ctx context.Context, command string, args ...any) (*gvar.Var, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.available {
		return nil, errors.New("redis unavailable")
	}
	switch strings.ToUpper(strings.TrimSpace(command)) {
	case "XADD":
		evt, err := eventFromXAddArgs(args)
		if err != nil {
			return nil, err
		}
		message := r.addMessageLocked(evt)
		return gvar.New(message.id), nil

	case "XGROUP":
		if len(args) >= 3 {
			r.groupName = fmt.Sprint(args[2])
		}
		if r.groupReady {
			return nil, errors.New("BUSYGROUP Consumer Group name already exists")
		}
		r.groupReady = true
		return gvar.New("OK"), nil

	case "XPENDING":
		result := make([]any, 0)
		for _, message := range r.messages {
			if message.acked || !message.pending {
				continue
			}
			result = append(result, []any{
				message.id,
				message.consumer,
				int64(time.Since(message.lastDelivery) / time.Millisecond),
				int64(1),
			})
		}
		return gvar.New(result), nil

	case "XCLAIM":
		if len(args) < 5 {
			return nil, fmt.Errorf("invalid XCLAIM args")
		}
		consumer := fmt.Sprint(args[2])
		minIdleMS, _ := strconv.ParseInt(fmt.Sprint(args[3]), 10, 64)
		now := time.Now()
		result := make([]any, 0)
		for _, rawID := range args[4:] {
			id := fmt.Sprint(rawID)
			message := r.findMessageLocked(id)
			if message == nil || message.acked || !message.pending {
				continue
			}
			if time.Since(message.lastDelivery) < time.Duration(minIdleMS)*time.Millisecond {
				continue
			}
			message.consumer = consumer
			message.lastDelivery = now
			result = append(result, []any{message.id, pairsFromFields(message.fields)})
		}
		return gvar.New(result), nil

	case "XREADGROUP":
		if len(args) < 8 {
			return nil, fmt.Errorf("invalid XREADGROUP args")
		}
		consumer := fmt.Sprint(args[2])
		now := time.Now()
		entries := make([]any, 0)
		for _, message := range r.messages {
			if message.acked || message.pending {
				continue
			}
			message.pending = true
			message.consumer = consumer
			message.lastDelivery = now
			entries = append(entries, []any{message.id, pairsFromFields(message.fields)})
		}
		if len(entries) == 0 {
			return gvar.New([]any{}), nil
		}
		return gvar.New([]any{
			[]any{r.stream, entries},
		}), nil

	case "XACK":
		if len(args) < 3 {
			return nil, fmt.Errorf("invalid XACK args")
		}
		acked := int64(0)
		for _, rawID := range args[2:] {
			if message := r.findMessageLocked(fmt.Sprint(rawID)); message != nil && !message.acked {
				message.acked = true
				message.pending = false
				acked++
			}
		}
		return gvar.New(acked), nil

	case "XINFO":
		if len(args) >= 2 && strings.EqualFold(fmt.Sprint(args[0]), "GROUPS") {
			if !r.groupReady {
				return gvar.New([]any{}), nil
			}
			pending := int64(0)
			lag := int64(0)
			for _, message := range r.messages {
				if message.acked {
					continue
				}
				if message.pending {
					pending++
					continue
				}
				lag++
			}
			return gvar.New([]any{
				[]any{
					"name", r.groupName,
					"consumers", int64(1),
					"pending", pending,
					"last-delivered-id", strconv.Itoa(r.seq) + "-0",
					"lag", lag,
				},
			}), nil
		}
	}
	return nil, fmt.Errorf("unsupported redis command: %s", command)
}

func (r *statefulStreamRedis) addMessageLocked(evt event.Event) *statefulStreamMessage {
	r.seq++
	message := &statefulStreamMessage{
		id:           fmt.Sprintf("%d-0", r.seq),
		fields:       fieldsFromEvent(evt),
		lastDelivery: time.Now(),
	}
	r.messages = append(r.messages, message)
	return message
}

func (r *statefulStreamRedis) findMessageLocked(id string) *statefulStreamMessage {
	for _, message := range r.messages {
		if message.id == id {
			return message
		}
	}
	return nil
}

func testStreamEvent(eventID, idempotencyKey string) event.Event {
	taskID := int64(99)
	return event.Event{
		EventID:        eventID,
		IdempotencyKey: idempotencyKey,
		EventType:      event.EventTaskCompleted,
		WorkflowRunID:  42,
		EntityType:     event.EntityDomainTask,
		EntityID:       &taskID,
		Attempt:        1,
		Payload: map[string]interface{}{
			"task_id": taskID,
		},
	}
}

func eventFromXAddArgs(args []any) (event.Event, error) {
	if len(args) < 3 {
		return event.Event{}, fmt.Errorf("invalid XADD args")
	}
	fields := make(map[string]string, (len(args)-2)/2)
	for i := 2; i+1 < len(args); i += 2 {
		fields[fmt.Sprint(args[i])] = fmt.Sprint(args[i+1])
	}
	evt := decodeStreamEvent(fields)
	if evt.EventID == "" {
		return event.Event{}, fmt.Errorf("missing event_id in XADD payload")
	}
	return evt, nil
}

func fieldsFromEvent(evt event.Event) map[string]string {
	evt = evt.EnsureMetadata()
	fields := map[string]string{
		"event_id":        evt.EventID,
		"idempotency_key": evt.IdempotencyKey,
		"event_type":      evt.EventType,
		"workflow_run_id": strconv.FormatInt(evt.WorkflowRunID, 10),
		"entity_type":     evt.EntityType,
		"attempt":         strconv.Itoa(evt.Attempt),
		"created_at":      strconv.FormatInt(evt.CreatedAtUnix, 10),
	}
	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		fields["payload_json"] = "{}"
	} else {
		fields["payload_json"] = string(payload)
	}
	if evt.EntityID != nil {
		fields["entity_id"] = strconv.FormatInt(*evt.EntityID, 10)
	}
	return fields
}

func pairsFromFields(fields map[string]string) []any {
	keys := []string{
		"event_id",
		"idempotency_key",
		"event_type",
		"workflow_run_id",
		"entity_type",
		"entity_id",
		"attempt",
		"created_at",
		"payload_json",
	}
	result := make([]any, 0, len(keys)*2)
	for _, key := range keys {
		if value, ok := fields[key]; ok {
			result = append(result, key, value)
		}
	}
	return result
}
