package eventstream

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/event"
)

// Handler 定义事件流消息处理回调。
type Handler func(ctx context.Context, evt event.Event) error

// RuntimeSnapshot 对外暴露 consumer 运行态快照，避免上层直接依赖 event 包的内部细节。
type RuntimeSnapshot = event.StreamRuntimeSnapshot

type streamMessage struct {
	ID     string
	Fields map[string]string
}

type pendingEntry struct {
	ID     string
	IdleMS int64
}

// Consumer 负责消费 Redis Stream 并执行回调。
type Consumer struct {
	mu           sync.Mutex
	cfg          Config
	redis        RedisCommander
	handler      Handler
	groupReady   bool
	startedAt    time.Time
	lastReclaim  time.Time
	lastConsume  time.Time
	lastAck      time.Time
	lastPulse    time.Time
	reclaimHits  int64
	reclaimMsgs  int64
	pendingKnown bool
	lastPending  int64
	lagKnown     bool
	lastLag      int64
	lastError    string
}

// NewConsumer 创建事件流 consumer。
func NewConsumer(redis RedisCommander, cfg Config, handler Handler) *Consumer {
	return &Consumer{
		cfg:     cfg.Normalize(),
		redis:   redis,
		handler: handler,
	}
}

// Start 启动消费循环。
func (c *Consumer) Start(ctx context.Context) {
	if !c.cfg.Enabled || !c.cfg.ConsumerEnabled {
		return
	}
	if c.redis == nil {
		g.Log().Warningf(ctx, "[WorkflowEventConsumer] redis unavailable, consumer skipped")
		return
	}
	c.markStarted()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		c.PulseHeartbeat()

		if err := c.ProcessOnce(ctx); err != nil {
			g.Log().Warningf(ctx, "[WorkflowEventConsumer] process once failed: %v", err)
			time.Sleep(2 * time.Second)
		}
	}
}

// PulseHeartbeat 更新 worker 心跳。
func (c *Consumer) PulseHeartbeat() {
	c.mu.Lock()
	c.lastPulse = time.Now()
	c.mu.Unlock()
}

// IsStarted 返回 consumer 是否已经进入运行态。
func (c *Consumer) IsStarted() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.startedAt.IsZero()
}

// Snapshot 返回运行态快照，优先保留内存态，再尝试补充 Redis 运行数据。
func (c *Consumer) Snapshot(ctx context.Context) event.StreamRuntimeSnapshot {
	snap := c.snapshotBase()
	if !c.cfg.Enabled || !c.cfg.ConsumerEnabled {
		return snap
	}

	if c.redis == nil {
		snap.Degraded = true
		snap.LastError = "redis client unavailable"
		return snap
	}

	if err := c.refreshRedisSnapshot(ctx, &snap); err != nil {
		snap.Degraded = true
		snap.LastError = err.Error()
	}
	return snap
}

// ProcessOnce 执行单轮消费。
func (c *Consumer) ProcessOnce(ctx context.Context) error {
	c.PulseHeartbeat()
	if err := c.ensureGroup(ctx); err != nil {
		c.recordError(err)
		return err
	}
	if err := c.maybeReclaimPending(ctx); err != nil {
		c.recordError(err)
		return err
	}
	if err := c.readAndHandle(ctx); err != nil {
		c.recordError(err)
		return err
	}
	c.clearError()
	return nil
}

func (c *Consumer) ensureGroup(ctx context.Context) error {
	c.mu.Lock()
	if c.groupReady {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	_, err := c.redis.Do(ctx, "XGROUP", "CREATE", c.cfg.StreamName, c.cfg.ConsumerGroup, "$", "MKSTREAM")
	if err != nil && !isBusyGroupErr(err) {
		return err
	}

	c.mu.Lock()
	c.groupReady = true
	c.lastError = ""
	c.mu.Unlock()
	return nil
}

func (c *Consumer) maybeReclaimPending(ctx context.Context) error {
	c.mu.Lock()
	last := c.lastReclaim
	if time.Since(last) < 5*time.Second {
		c.mu.Unlock()
		return nil
	}
	c.lastReclaim = time.Now()
	c.mu.Unlock()

	result, err := c.redis.Do(
		ctx,
		"XPENDING",
		c.cfg.StreamName,
		c.cfg.ConsumerGroup,
		"-",
		"+",
		c.cfg.ReclaimCount,
	)
	if err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "NOGROUP") {
			c.mu.Lock()
			c.groupReady = false
			c.mu.Unlock()
			return c.ensureGroup(ctx)
		}
		return err
	}
	entries := parsePendingEntries(result)
	if len(entries) == 0 {
		return nil
	}

	minIdleMS := int64(c.cfg.ReclaimIdleSeconds) * 1000
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IdleMS >= minIdleMS {
			ids = append(ids, entry.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	c.mu.Lock()
	c.lastReclaim = time.Now()
	c.reclaimHits++
	c.mu.Unlock()

	args := []any{
		c.cfg.StreamName,
		c.cfg.ConsumerGroup,
		c.cfg.ConsumerName,
		minIdleMS,
	}
	for _, id := range ids {
		args = append(args, id)
	}
	claimed, err := c.redis.Do(ctx, "XCLAIM", args...)
	if err != nil {
		return err
	}
	messages, err := parseXClaimResult(claimed)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.reclaimMsgs += int64(len(messages))
	c.mu.Unlock()
	return c.handleMessages(ctx, messages)
}

func (c *Consumer) readAndHandle(ctx context.Context) error {
	result, err := c.redis.Do(
		ctx,
		"XREADGROUP",
		"GROUP", c.cfg.ConsumerGroup, c.cfg.ConsumerName,
		"COUNT", c.cfg.ReadCount,
		"BLOCK", c.cfg.BlockMS,
		"STREAMS", c.cfg.StreamName, ">",
	)
	if err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "NOGROUP") {
			c.mu.Lock()
			c.groupReady = false
			c.mu.Unlock()
			return c.ensureGroup(ctx)
		}
		return err
	}
	messages, err := parseXReadGroupResult(result)
	if err != nil {
		return err
	}
	return c.handleMessages(ctx, messages)
}

func (c *Consumer) handleMessages(ctx context.Context, messages []streamMessage) error {
	for _, msg := range messages {
		evt := decodeStreamEvent(msg.Fields)
		c.mu.Lock()
		c.lastConsume = time.Now()
		c.mu.Unlock()
		if c.handler != nil {
			if err := c.handler(ctx, evt); err != nil {
				return err
			}
		}
		if _, err := c.redis.Do(ctx, "XACK", c.cfg.StreamName, c.cfg.ConsumerGroup, msg.ID); err != nil {
			return err
		}
		c.mu.Lock()
		c.lastAck = time.Now()
		c.mu.Unlock()
	}
	return nil
}

func (c *Consumer) markStarted() {
	c.mu.Lock()
	if c.startedAt.IsZero() {
		c.startedAt = time.Now()
	}
	c.mu.Unlock()
}

func (c *Consumer) snapshotBase() event.StreamRuntimeSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	snap := event.StreamRuntimeSnapshot{
		Enabled:           c.cfg.Enabled,
		ConsumerEnabled:   c.cfg.ConsumerEnabled,
		ConsumerCreated:   true,
		ConsumerStarted:   !c.startedAt.IsZero(),
		GroupReady:        c.groupReady,
		StreamName:        c.cfg.StreamName,
		ConsumerGroup:     c.cfg.ConsumerGroup,
		ConsumerName:      c.cfg.ConsumerName,
		ReclaimAttempts:   c.reclaimHits,
		ReclaimedMessages: c.reclaimMsgs,
		LastConsumeAt:     c.lastConsume,
		LastAckAt:         c.lastAck,
		LastReclaimAt:     c.lastReclaim,
		WorkerHeartbeatAt: c.lastPulse,
		StartedAt:         c.startedAt,
		UpdatedAt:         time.Now(),
		LastError:         c.lastError,
		Degraded:          strings.TrimSpace(c.lastError) != "",
	}
	if c.pendingKnown {
		snap.Pending = c.lastPending
		snap.PendingKnown = true
	}
	if c.lagKnown {
		snap.Lag = c.lastLag
		snap.LagKnown = true
	}
	return snap
}

func (c *Consumer) refreshRedisSnapshot(ctx context.Context, snap *event.StreamRuntimeSnapshot) error {
	if c.redis == nil {
		return nil
	}

	result, err := c.redis.Do(ctx, "XINFO", "GROUPS", c.cfg.StreamName)
	if err != nil {
		return err
	}
	groupInfo, ok := parseGroupInfo(result, c.cfg.ConsumerGroup)
	if !ok {
		return nil
	}

	snap.Pending = groupInfo.Pending
	snap.PendingKnown = true
	snap.Lag = groupInfo.Lag
	snap.LagKnown = groupInfo.LagKnown
	snap.GroupReady = true

	c.mu.Lock()
	c.lastPending = groupInfo.Pending
	c.pendingKnown = true
	c.lastLag = groupInfo.Lag
	c.lagKnown = groupInfo.LagKnown
	c.groupReady = true
	c.mu.Unlock()

	return nil
}

func (c *Consumer) recordError(err error) {
	if err == nil {
		return
	}
	c.mu.Lock()
	c.lastError = err.Error()
	c.mu.Unlock()
}

func (c *Consumer) clearError() {
	c.mu.Lock()
	c.lastError = ""
	c.mu.Unlock()
}

func decodeStreamEvent(fields map[string]string) event.Event {
	evt := event.Event{
		EventID:        strings.TrimSpace(fields["event_id"]),
		IdempotencyKey: strings.TrimSpace(fields["idempotency_key"]),
		EventType:      strings.TrimSpace(fields["event_type"]),
		EntityType:     strings.TrimSpace(fields["entity_type"]),
		WorkflowRunID:  parseInt64(fields["workflow_run_id"]),
		Attempt:        int(parseInt64(fields["attempt"])),
		CreatedAtUnix:  parseInt64(fields["created_at"]),
	}
	if stageRunID := parseInt64(fields["stage_run_id"]); stageRunID > 0 {
		evt.StageRunID = &stageRunID
	}
	if entityID := parseInt64(fields["entity_id"]); entityID > 0 {
		evt.EntityID = &entityID
	}
	payload := strings.TrimSpace(fields["payload_json"])
	if payload != "" {
		var decoded interface{}
		if err := json.Unmarshal([]byte(payload), &decoded); err == nil {
			evt.Payload = decoded
		} else {
			evt.Payload = payload
		}
	}
	return evt.EnsureMetadata()
}

func parseXReadGroupResult(result interface{ Interface() any }) ([]streamMessage, error) {
	if result == nil {
		return nil, nil
	}
	if messages, ok := parseXReadGroupMapResult(result.Interface()); ok {
		return messages, nil
	}
	streams := toAnySlice(result.Interface())
	if len(streams) == 0 {
		return nil, nil
	}

	messages := make([]streamMessage, 0, len(streams))
	for _, streamItem := range streams {
		if streamMessages, ok := parseStreamMessagesFromStruct(streamItem); ok {
			messages = append(messages, streamMessages...)
			continue
		}
		pair := toAnySlice(streamItem)
		if len(pair) < 2 {
			continue
		}
		entryItems := toAnySlice(pair[1])
		for _, entry := range entryItems {
			msg, ok := parseStreamMessageEntry(entry)
			if ok {
				messages = append(messages, msg)
			}
		}
	}
	return messages, nil
}

func parseXClaimResult(result interface{ Interface() any }) ([]streamMessage, error) {
	if result == nil {
		return nil, nil
	}
	if stringItems, ok := toStringSlice(result.Interface()); ok {
		messages := make([]streamMessage, 0, len(stringItems))
		for _, item := range stringItems {
			entry, ok := parseJSONArrayEntry(item)
			if !ok {
				continue
			}
			msg, ok := parseStreamMessageEntry(entry)
			if ok {
				messages = append(messages, msg)
			}
		}
		return messages, nil
	}
	items := toAnySlice(result.Interface())
	if len(items) == 0 {
		return nil, nil
	}
	messages := make([]streamMessage, 0, len(items))
	for _, entry := range items {
		msg, ok := parseStreamMessageEntry(entry)
		if ok {
			messages = append(messages, msg)
		}
	}
	return messages, nil
}

func parseStreamMessageEntry(entry any) (streamMessage, bool) {
	if msg, ok := parseStreamMessageStruct(entry); ok {
		return msg, true
	}
	values := toAnySlice(entry)
	if len(values) < 2 {
		return streamMessage{}, false
	}
	messageID := toString(values[0])
	if strings.TrimSpace(messageID) == "" {
		return streamMessage{}, false
	}
	fields, ok := toStringMapFromPairs(values[1])
	if !ok {
		return streamMessage{}, false
	}
	return streamMessage{ID: messageID, Fields: fields}, true
}

func parsePendingEntries(result interface{ Interface() any }) []pendingEntry {
	if result == nil {
		return nil
	}
	if stringItems, ok := toStringSlice(result.Interface()); ok {
		entries := make([]pendingEntry, 0, len(stringItems))
		for _, item := range stringItems {
			entry, ok := parseJSONArrayEntry(item)
			if !ok {
				continue
			}
			pending, ok := parsePendingEntrySlice(entry)
			if ok {
				entries = append(entries, pending)
			}
		}
		return entries
	}
	items := toAnySlice(result.Interface())
	if len(items) == 0 {
		return nil
	}
	entries := make([]pendingEntry, 0, len(items))
	for _, item := range items {
		if entry, ok := parsePendingEntryStruct(item); ok {
			entries = append(entries, entry)
			continue
		}
		entry, ok := parsePendingEntrySlice(toAnySlice(item))
		if ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

type groupInfo struct {
	Pending  int64
	Lag      int64
	LagKnown bool
}

func parseGroupInfo(result interface{ Interface() any }, consumerGroup string) (groupInfo, bool) {
	if stringItems, ok := toStringSlice(result.Interface()); ok {
		for _, item := range stringItems {
			kv, ok := parseBracketMapString(item)
			if !ok {
				continue
			}
			if strings.TrimSpace(kv["name"]) != strings.TrimSpace(consumerGroup) {
				continue
			}

			info := groupInfo{}
			info.Pending = parseAnyInt64(kv["pending"])
			if raw, ok := kv["lag"]; ok {
				info.Lag = parseAnyInt64(raw)
				info.LagKnown = true
			}
			return info, true
		}
	}

	items := toAnySlice(result.Interface())
	if len(items) == 0 {
		return groupInfo{}, false
	}

	for _, item := range items {
		if info, ok := parseGroupInfoStruct(item, consumerGroup); ok {
			return info, true
		}
		kv, ok := toStringMapFromPairs(item)
		if !ok {
			continue
		}
		if strings.TrimSpace(kv["name"]) != strings.TrimSpace(consumerGroup) {
			continue
		}

		info := groupInfo{}
		info.Pending = parseAnyInt64(kv["pending"])
		if raw, ok := kv["lag"]; ok {
			info.Lag = parseAnyInt64(raw)
			info.LagKnown = true
		}
		return info, true
	}

	return groupInfo{}, false
}

func isBusyGroupErr(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToUpper(err.Error())
	return strings.Contains(message, "BUSYGROUP")
}

func parseInt64(raw string) int64 {
	val, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return val
}

func parseAnyInt64(raw string) int64 {
	val, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return val
}

func toAnySlice(value any) []any {
	switch typed := value.(type) {
	case []any:
		return typed
	case nil:
		return nil
	default:
		rv := reflect.ValueOf(value)
		if !rv.IsValid() || rv.Kind() != reflect.Slice {
			return nil
		}
		out := make([]any, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out = append(out, rv.Index(i).Interface())
		}
		return out
	}
}

func parseStreamMessagesFromStruct(value any) ([]streamMessage, bool) {
	streamName, ok := structFieldValue(value, "Stream")
	if !ok || strings.TrimSpace(toString(streamName)) == "" {
		return nil, false
	}
	rawMessages, ok := structFieldValue(value, "Messages")
	if !ok {
		return nil, false
	}
	items := toAnySlice(rawMessages)
	if len(items) == 0 {
		return nil, true
	}
	messages := make([]streamMessage, 0, len(items))
	for _, item := range items {
		msg, ok := parseStreamMessageEntry(item)
		if ok {
			messages = append(messages, msg)
		}
	}
	return messages, true
}

func parseXReadGroupMapResult(value any) ([]streamMessage, bool) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, false
	}
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Map {
		return nil, false
	}

	messages := make([]streamMessage, 0, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		entryItems := toAnySlice(iter.Value().Interface())
		for _, entry := range entryItems {
			msg, ok := parseStreamMessageEntry(entry)
			if ok {
				messages = append(messages, msg)
			}
		}
	}
	return messages, true
}

func parseStreamMessageStruct(value any) (streamMessage, bool) {
	rawID, ok := structFieldValue(value, "ID")
	if !ok {
		return streamMessage{}, false
	}
	messageID := strings.TrimSpace(toString(rawID))
	if messageID == "" {
		return streamMessage{}, false
	}
	rawValues, ok := structFieldValue(value, "Values")
	if !ok {
		return streamMessage{}, false
	}
	fields, ok := toStringMap(rawValues)
	if !ok {
		return streamMessage{}, false
	}
	return streamMessage{ID: messageID, Fields: fields}, true
}

func parsePendingEntryStruct(value any) (pendingEntry, bool) {
	rawID, ok := structFieldValue(value, "ID")
	if !ok {
		return pendingEntry{}, false
	}
	id := strings.TrimSpace(toString(rawID))
	if id == "" {
		return pendingEntry{}, false
	}
	rawIdle, ok := structFieldValue(value, "Idle")
	if !ok {
		return pendingEntry{}, false
	}

	var idleMS int64
	switch typed := rawIdle.(type) {
	case time.Duration:
		idleMS = typed.Milliseconds()
	default:
		idleMS = parseAnyInt64(toString(rawIdle))
	}

	return pendingEntry{ID: id, IdleMS: idleMS}, true
}

func parsePendingEntrySlice(values []any) (pendingEntry, bool) {
	if len(values) < 3 {
		return pendingEntry{}, false
	}
	id := strings.TrimSpace(toString(values[0]))
	if id == "" {
		return pendingEntry{}, false
	}
	return pendingEntry{
		ID:     id,
		IdleMS: parseInt64(toString(values[2])),
	}, true
}

func parseGroupInfoStruct(value any, consumerGroup string) (groupInfo, bool) {
	rawName, ok := structFieldValue(value, "Name")
	if !ok || strings.TrimSpace(toString(rawName)) != strings.TrimSpace(consumerGroup) {
		return groupInfo{}, false
	}

	info := groupInfo{}
	if rawPending, ok := structFieldValue(value, "Pending"); ok {
		info.Pending = parseAnyInt64(toString(rawPending))
	}
	if rawLag, ok := structFieldValue(value, "Lag"); ok {
		info.Lag = parseAnyInt64(toString(rawLag))
		info.LagKnown = true
	}
	return info, true
}

func structFieldValue(value any, fieldName string) (any, bool) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, false
	}
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, false
	}
	field := rv.FieldByName(fieldName)
	if !field.IsValid() || !field.CanInterface() {
		return nil, false
	}
	return field.Interface(), true
}

func toStringSlice(value any) ([]string, bool) {
	switch typed := value.(type) {
	case []string:
		return typed, true
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() || rv.Kind() != reflect.Slice || rv.Type().Elem().Kind() != reflect.String {
		return nil, false
	}
	out := make([]string, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out = append(out, toString(rv.Index(i).Interface()))
	}
	return out, true
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return fmt.Sprint(value)
	}
}

func parseJSONArrayEntry(raw string) ([]any, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}
	var entry []any
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		return nil, false
	}
	return entry, true
}

func parseBracketMapString(raw string) (map[string]string, bool) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "map[") || !strings.HasSuffix(raw, "]") {
		return nil, false
	}
	body := strings.TrimSuffix(strings.TrimPrefix(raw, "map["), "]")
	if strings.TrimSpace(body) == "" {
		return map[string]string{}, true
	}

	fields := make(map[string]string)
	for _, token := range strings.Fields(body) {
		parts := strings.SplitN(token, ":", 2)
		if len(parts) != 2 {
			continue
		}
		fields[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return fields, true
}

func toStringMap(value any) (map[string]string, bool) {
	switch typed := value.(type) {
	case map[string]string:
		return typed, true
	case map[string]interface{}:
		fields := make(map[string]string, len(typed))
		for key, raw := range typed {
			fields[key] = toString(raw)
		}
		return fields, true
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, false
	}
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Map {
		return nil, false
	}

	fields := make(map[string]string, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		key := strings.TrimSpace(toString(iter.Key().Interface()))
		if key == "" {
			continue
		}
		fields[key] = toString(iter.Value().Interface())
	}
	return fields, true
}

func toStringMapFromPairs(value any) (map[string]string, bool) {
	entries := toAnySlice(value)
	if len(entries) == 0 || len(entries)%2 != 0 {
		return nil, false
	}
	fields := make(map[string]string, len(entries)/2)
	for idx := 0; idx < len(entries); idx += 2 {
		key := strings.TrimSpace(toString(entries[idx]))
		if key == "" {
			continue
		}
		fields[key] = toString(entries[idx+1])
	}
	return fields, true
}
