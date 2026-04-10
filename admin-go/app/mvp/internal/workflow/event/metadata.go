package event

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"easymvp/utility/snowflake"
)

// EnsureMetadata 补齐事件元数据，供本地快路径和持久化事件流统一使用。
func (evt Event) EnsureMetadata() Event {
	if strings.TrimSpace(evt.EventID) == "" {
		evt.EventID = strconv.FormatInt(int64(snowflake.Generate()), 10)
	}
	if evt.CreatedAtUnix <= 0 {
		evt.CreatedAtUnix = time.Now().Unix()
	}
	if evt.Attempt <= 0 {
		evt.Attempt = resolveAttemptFromPayload(evt.Payload)
		if evt.Attempt <= 0 {
			evt.Attempt = 1
		}
	}
	if strings.TrimSpace(evt.IdempotencyKey) == "" {
		evt.IdempotencyKey = buildIdempotencyKey(evt)
	}
	return evt
}

func buildIdempotencyKey(evt Event) string {
	var (
		entityID int64
		taskID   int64
	)
	if evt.EntityID != nil {
		entityID = *evt.EntityID
	}
	taskID = resolveTaskID(evt)
	return fmt.Sprintf(
		"wf:%d:task:%d:entity:%d:type:%s:attempt:%d",
		evt.WorkflowRunID,
		taskID,
		entityID,
		strings.TrimSpace(evt.EventType),
		evt.Attempt,
	)
}

func resolveTaskID(evt Event) int64 {
	if evt.EntityType == EntityDomainTask && evt.EntityID != nil {
		return *evt.EntityID
	}
	if evt.Payload == nil {
		return 0
	}
	switch payload := evt.Payload.(type) {
	case map[string]interface{}:
		return parseTaskIDFromMap(payload)
	case map[string]string:
		if raw, ok := payload["task_id"]; ok {
			if id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64); err == nil {
				return id
			}
		}
	case string:
		var mapPayload map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &mapPayload); err == nil {
			return parseTaskIDFromMap(mapPayload)
		}
	}
	return 0
}

func parseTaskIDFromMap(payload map[string]interface{}) int64 {
	if payload == nil {
		return 0
	}
	raw, ok := payload["task_id"]
	if !ok {
		raw, ok = payload["taskId"]
		if !ok {
			return 0
		}
	}
	switch v := raw.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case json.Number:
		val, _ := v.Int64()
		return val
	case string:
		val, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return val
	default:
		return 0
	}
}

func resolveAttemptFromPayload(payload interface{}) int {
	if payload == nil {
		return 0
	}
	switch v := payload.(type) {
	case map[string]interface{}:
		return parseAttemptFromMap(v)
	case string:
		var mapPayload map[string]interface{}
		if err := json.Unmarshal([]byte(v), &mapPayload); err == nil {
			return parseAttemptFromMap(mapPayload)
		}
	}
	return 0
}

func parseAttemptFromMap(payload map[string]interface{}) int {
	if payload == nil {
		return 0
	}
	raw, ok := payload["attempt"]
	if !ok {
		return 0
	}
	switch v := raw.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		val, _ := v.Int64()
		return int(val)
	case string:
		val, _ := strconv.Atoi(strings.TrimSpace(v))
		return val
	default:
		return 0
	}
}
