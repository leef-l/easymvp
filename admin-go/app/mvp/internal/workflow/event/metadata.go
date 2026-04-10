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
		entityID   int64
		taskID     int64
		stageRunID int64
	)
	entityID = resolveEntityID(evt)
	taskID = resolveTaskID(evt)
	stageRunID = resolveStageRunID(evt)
	return fmt.Sprintf(
		"wf:%d:task:%d:entity:%d:stage:%d:type:%s:attempt:%d",
		evt.WorkflowRunID,
		taskID,
		entityID,
		stageRunID,
		strings.TrimSpace(evt.EventType),
		evt.Attempt,
	)
}

func resolveEntityID(evt Event) int64 {
	if evt.EntityID != nil {
		return *evt.EntityID
	}
	return resolveInt64FromPayload(evt.Payload, "entity_id", "entityId")
}

func resolveTaskID(evt Event) int64 {
	if evt.EntityType == EntityDomainTask && evt.EntityID != nil {
		return *evt.EntityID
	}
	return resolveInt64FromPayload(evt.Payload,
		"task_id", "taskId",
		"failed_task_id", "failedTaskId",
		"domain_task_id", "domainTaskId",
		"source_task_id", "sourceTaskId",
	)
}

func resolveStageRunID(evt Event) int64 {
	if evt.StageRunID != nil {
		return *evt.StageRunID
	}
	return resolveInt64FromPayload(evt.Payload, "stage_run_id", "stageRunId")
}

func resolveInt64FromPayload(payload interface{}, keys ...string) int64 {
	if payload == nil || len(keys) == 0 {
		return 0
	}
	switch v := payload.(type) {
	case map[string]interface{}:
		return parseInt64FromInterfaceMap(v, keys...)
	case map[string]string:
		return parseInt64FromStringMap(v, keys...)
	case string:
		var mapPayload map[string]interface{}
		if err := json.Unmarshal([]byte(v), &mapPayload); err == nil {
			return parseInt64FromInterfaceMap(mapPayload, keys...)
		}
	}
	return 0
}

func parseInt64FromInterfaceMap(payload map[string]interface{}, keys ...string) int64 {
	if payload == nil {
		return 0
	}
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok {
			continue
		}
		if val := parseInt64Value(raw); val > 0 {
			return val
		}
	}
	return 0
}

func parseInt64FromStringMap(payload map[string]string, keys ...string) int64 {
	if payload == nil {
		return 0
	}
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok {
			continue
		}
		if val, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64); err == nil && val > 0 {
			return val
		}
	}
	return 0
}

func parseInt64Value(raw interface{}) int64 {
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
