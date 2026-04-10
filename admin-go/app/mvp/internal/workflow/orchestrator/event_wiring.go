package orchestrator

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/autonomy"
	domainTask "easymvp/app/mvp/internal/workflow/domain/task"
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/eventstream"
)

var (
	workflowEventConsumer       *eventstream.Consumer
	recoveryEventGuard          = newRecoveryEventGuard(10 * time.Minute)
	reconcileWorkflowProgressFn = func(ctx context.Context, workflowRunID int64) bool {
		if workflowRunID == 0 || taskScheduler == nil {
			return false
		}
		return taskScheduler.ReconcileWorkflowProgress(ctx, workflowRunID)
	}
	beginRecoveryEventClaimFn = func(ctx context.Context, evt event.Event) (event.DurableEventClaim, bool, error) {
		return event.BeginDurableEventClaim(ctx, "workflow.recovery."+strings.TrimSpace(evt.EventType), evt)
	}
)

type recoveryEventDeduper struct {
	mu      sync.Mutex
	ttl     time.Duration
	lastGC  time.Time
	handled map[string]time.Time
}

func newRecoveryEventGuard(ttl time.Duration) *recoveryEventDeduper {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &recoveryEventDeduper{
		ttl:     ttl,
		handled: make(map[string]time.Time),
	}
}

func (d *recoveryEventDeduper) Allow(evt event.Event) bool {
	key := strings.TrimSpace(evt.IdempotencyKey)
	if key == "" {
		return true
	}

	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.lastGC.IsZero() || now.Sub(d.lastGC) >= d.ttl {
		for handledKey, handledAt := range d.handled {
			if now.Sub(handledAt) >= d.ttl {
				delete(d.handled, handledKey)
			}
		}
		d.lastGC = now
	}

	if handledAt, ok := d.handled[key]; ok && now.Sub(handledAt) < d.ttl {
		return false
	}

	d.handled[key] = now
	return true
}

func setupWorkflowEventing(ctx context.Context) {
	if eventPublisher == nil || eventBus == nil || taskScheduler == nil {
		return
	}

	taskScheduler.SetEventPublisher(eventPublisher)

	registerRecoveryEventHandlers()

	cfg := loadWorkflowEventStreamConfig(ctx)
	if !cfg.Enabled {
		eventPublisher.SetStreamSink(nil)
		workflowEventConsumer = nil
		return
	}

	redisClient := safeGetWorkflowEventRedis(ctx)
	var redisCommander eventstream.RedisCommander
	if redisClient != nil {
		redisCommander = redisClient
	}
	bridge := eventstream.NewBridge(redisCommander, cfg)
	eventPublisher.SetStreamSink(bridge)

	if cfg.ConsumerEnabled {
		workflowEventConsumer = eventstream.NewConsumer(redisCommander, cfg, handleWorkflowRecoveryEvent)
		return
	}
	workflowEventConsumer = nil
}

func registerRecoveryEventHandlers() {
	eventBus.Subscribe(event.EventSchedulerWakeup, func(evt event.Event) {
		if err := handleWorkflowRecoveryEvent(context.Background(), evt); err != nil {
			g.Log().Warningf(context.Background(), "[WorkflowEvent] handle scheduler wakeup failed: event=%s err=%v", evt.EventID, err)
		}
	})
	eventBus.Subscribe(event.EventTaskCompleted, func(evt event.Event) {
		if err := handleWorkflowRecoveryEvent(context.Background(), evt); err != nil {
			g.Log().Warningf(context.Background(), "[WorkflowEvent] handle task completed failed: event=%s err=%v", evt.EventID, err)
		}
	})
	eventBus.Subscribe(event.EventTaskRetryDue, func(evt event.Event) {
		if err := handleWorkflowRecoveryEvent(context.Background(), evt); err != nil {
			g.Log().Warningf(context.Background(), "[WorkflowEvent] handle task retry due failed: event=%s err=%v", evt.EventID, err)
		}
	})
	eventBus.Subscribe(event.EventTaskEscalateDue, func(evt event.Event) {
		if err := handleWorkflowRecoveryEvent(context.Background(), evt); err != nil {
			g.Log().Warningf(context.Background(), "[WorkflowEvent] handle task escalate due failed: event=%s err=%v", evt.EventID, err)
		}
	})
}

func loadWorkflowEventStreamConfig(ctx context.Context) eventstream.Config {
	var (
		enabled         = engine.GetConfigInt(ctx, "workflow.event_stream.enabled", "workflow.event_stream.enabled", 0) == 1
		streamName      = expandConfigEnv(engine.GetConfigString(ctx, "workflow.event_stream.stream_name", "workflow.event_stream.stream_name", ""))
		consumerGroup   = expandConfigEnv(engine.GetConfigString(ctx, "workflow.event_stream.consumer_group", "workflow.event_stream.consumer_group", ""))
		consumerName    = expandConfigEnv(engine.GetConfigString(ctx, "workflow.event_stream.consumer_name", "workflow.event_stream.consumer_name", ""))
		blockMS         = engine.GetConfigInt(ctx, "workflow.event_stream.block_ms", "workflow.event_stream.block_ms", 5000)
		reclaimIdle     = engine.GetConfigInt(ctx, "workflow.event_stream.reclaim_idle_seconds", "workflow.event_stream.reclaim_idle_seconds", 60)
		redisRequired   = engine.GetConfigInt(ctx, "workflow.event_stream.redis_required", "workflow.event_stream.redis_required", 0) == 1
		consumerEnabled = engine.GetConfigInt(ctx, "workflow.event_stream.consumer_enabled", "workflow.event_stream.consumer_enabled", 0) == 1
	)

	if strings.TrimSpace(consumerName) == "" {
		if hostName, err := os.Hostname(); err == nil {
			consumerName = hostName
		}
	}

	return eventstream.Config{
		Enabled:            enabled,
		StreamName:         streamName,
		ConsumerGroup:      consumerGroup,
		ConsumerName:       consumerName,
		BlockMS:            blockMS,
		ReclaimIdleSeconds: reclaimIdle,
		RedisRequired:      redisRequired,
		ConsumerEnabled:    consumerEnabled,
	}.Normalize()
}

func safeGetWorkflowEventRedis(ctx context.Context) (r *gredis.Redis) {
	defer func() {
		if rec := recover(); rec != nil {
			g.Log().Warningf(ctx, "[WorkflowEvent] redis unavailable, fallback to local fast path: %v", rec)
			r = nil
		}
	}()

	r = g.Redis()
	if r != nil {
		if _, err := r.Do(ctx, "PING"); err == nil {
			return r
		} else {
			g.Log().Warningf(ctx, "[WorkflowEvent] default redis ping failed, retry fallback: %v", err)
			if !shouldFallbackToDirectRedisErr(err) {
				return r
			}
		}
	}

	addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	pass := resolveRedisPass()
	client, err := gredis.New(&gredis.Config{
		Address: addr,
		Pass:    pass,
		Db:      0,
	})
	if err != nil {
		g.Log().Warningf(ctx, "[WorkflowEvent] redis direct client init failed addr=%s err=%v", addr, err)
		return nil
	}
	if _, err := client.Do(ctx, "PING"); err != nil {
		g.Log().Warningf(ctx, "[WorkflowEvent] redis direct client ping failed addr=%s err=%v", addr, err)
		return nil
	}
	return client
}

func handleWorkflowRecoveryEvent(ctx context.Context, evt event.Event) error {
	if !isRecoveryManagedEvent(evt.EventType) {
		return nil
	}
	if evt.WorkflowRunID == 0 {
		return nil
	}
	if !recoveryEventGuard.Allow(evt) {
		return nil
	}

	claim, shouldProcess, claimErr := beginRecoveryEventClaimFn(ctx, evt)
	if claimErr != nil {
		if event.IsMissingDurableEventLedgerErr(claimErr) {
			g.Log().Warningf(ctx, "[WorkflowEvent] durable recovery ledger migration missing, fallback to in-memory dedupe: %v", claimErr)
			return handleWorkflowRecoveryEventOnce(ctx, evt)
		}
		return claimErr
	}
	if !shouldProcess {
		return nil
	}

	handlerErr := handleWorkflowRecoveryEventOnce(ctx, evt)
	if handlerErr != nil {
		if claim != nil {
			if markErr := claim.MarkFailed(ctx, handlerErr); markErr != nil {
				g.Log().Warningf(ctx, "[WorkflowEvent] mark durable claim failed: event=%s err=%v", evt.EventType, markErr)
			}
		}
		return handlerErr
	}
	if claim != nil {
		if markErr := claim.MarkHandled(ctx); markErr != nil {
			g.Log().Warningf(ctx, "[WorkflowEvent] mark durable claim handled failed: event=%s err=%v", evt.EventType, markErr)
		}
	}
	return nil
}

func isRecoveryManagedEvent(eventType string) bool {
	switch eventType {
	case event.EventSchedulerWakeup, event.EventTaskCompleted, event.EventTaskRetryDue, event.EventTaskEscalateDue:
		return true
	default:
		return false
	}
}

func handleWorkflowRecoveryEventOnce(ctx context.Context, evt event.Event) error {
	switch evt.EventType {
	case event.EventSchedulerWakeup:
		return handleSchedulerWakeupEvent(ctx, evt)
	case event.EventTaskCompleted:
		return handleTaskCompletedEvent(ctx, evt)
	case event.EventTaskRetryDue:
		return handleTaskRetryDueEvent(ctx, evt)
	case event.EventTaskEscalateDue:
		return handleTaskEscalateDueEvent(ctx, evt)
	default:
		return nil
	}
}

func handleTaskCompletedEvent(ctx context.Context, evt event.Event) error {
	reconcileWorkflowProgressFn(ctx, evt.WorkflowRunID)
	return nil
}

func handleSchedulerWakeupEvent(ctx context.Context, evt event.Event) error {
	if evt.WorkflowRunID == 0 || taskScheduler == nil {
		return nil
	}
	taskScheduler.Wakeup(ctx, evt.WorkflowRunID)
	return nil
}

func handleTaskRetryDueEvent(ctx context.Context, evt event.Event) error {
	var (
		payload = normalizeEventPayload(evt.Payload)
		taskID  = resolveTaskIDFromEvent(evt, payload)
		attempt = evt.Attempt
	)
	if taskID == 0 {
		return fmt.Errorf("task.retry_due 缺少 task_id")
	}
	if attempt <= 0 {
		attempt = int(int64FromPayload(payload, "attempt"))
	}
	if attempt <= 0 {
		attempt = 1
	}

	result, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).
		Where("status", domainTask.StatusFailed).
		Update(g.Map{
			"status":           domainTask.StatusPending,
			"retry_count":      attempt,
			"result":           nil,
			"error_message":    nil,
			"started_at":       nil,
			"completed_at":     nil,
			"heartbeat_at":     nil,
			"locked_resources": nil,
			"updated_at":       gtime.Now(),
		})
	if err != nil {
		return fmt.Errorf("更新任务为 pending 失败: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil
	}

	reason := stringFromPayload(payload, "reason", "error")
	publishWorkflowEvent(ctx, event.Event{
		WorkflowRunID: evt.WorkflowRunID,
		EntityType:    event.EntityDomainTask,
		EntityID:      &taskID,
		EventType:     event.EventTaskRetried,
		Attempt:       attempt,
		Payload: g.Map{
			"task_id": taskID,
			"attempt": attempt,
			"reason":  reason,
		},
	})
	emitSchedulerWakeup(ctx, evt.WorkflowRunID, "task.retry_due", map[string]interface{}{
		"task_id": taskID,
		"attempt": attempt,
		"reason":  reason,
	})
	return nil
}

func handleTaskEscalateDueEvent(ctx context.Context, evt event.Event) error {
	payload := normalizeEventPayload(evt.Payload)
	taskID := resolveTaskIDFromEvent(evt, payload)
	if taskID == 0 {
		return fmt.Errorf("task.escalate_due 缺少 task_id")
	}

	result, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).
		Where("status", domainTask.StatusFailed).
		Update(g.Map{
			"status":     domainTask.StatusEscalated,
			"updated_at": gtime.Now(),
		})
	if err != nil {
		return fmt.Errorf("更新任务为 escalated 失败: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil
	}

	reason := stringFromPayload(payload, "reason", "error")
	publishWorkflowEvent(ctx, event.Event{
		WorkflowRunID: evt.WorkflowRunID,
		EntityType:    event.EntityDomainTask,
		EntityID:      &taskID,
		EventType:     event.EventTaskEscalated,
		Attempt:       evt.Attempt,
		Payload: g.Map{
			"task_id": taskID,
			"reason":  reason,
		},
	})

	if decisionCenter != nil && decisionCenter.IsEnabled(ctx) {
		projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", evt.WorkflowRunID).Value("project_id") //nolint: errcheck
		resp := decisionCenter.Decide(ctx, &autonomy.DecisionRequest{
			WorkflowRunID:  evt.WorkflowRunID,
			ProjectID:      projectID.Int64(),
			DomainTaskID:   taskID,
			TriggerSource:  consts.TriggerTaskRetryExhausted,
			TriggerContext: map[string]interface{}{"task_id": taskID, "reason": reason},
		})
		if resp.Handled {
			return nil
		}
	}

	return triggerReworkStage(ctx, evt.WorkflowRunID, taskID)
}

func emitSchedulerWakeup(ctx context.Context, workflowRunID int64, reason string, payload map[string]interface{}) {
	if workflowRunID == 0 {
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	if _, ok := payload["reason"]; !ok && strings.TrimSpace(reason) != "" {
		payload["reason"] = reason
	}
	publishWorkflowEvent(ctx, event.Event{
		WorkflowRunID: workflowRunID,
		EntityType:    event.EntityWorkflowRun,
		EventType:     event.EventSchedulerWakeup,
		Payload:       payload,
	})
}

func emitTaskRetryDue(ctx context.Context, workflowRunID, taskID int64, attempt, maxRetries int, reason string) {
	if workflowRunID == 0 || taskID == 0 {
		return
	}
	taskIDCopy := taskID
	publishWorkflowEvent(ctx, event.Event{
		WorkflowRunID: workflowRunID,
		EntityType:    event.EntityDomainTask,
		EntityID:      &taskIDCopy,
		EventType:     event.EventTaskRetryDue,
		Attempt:       attempt,
		Payload: g.Map{
			"task_id":     taskID,
			"attempt":     attempt,
			"max_retries": maxRetries,
			"reason":      strings.TrimSpace(reason),
		},
	})
}

func emitTaskEscalateDue(ctx context.Context, workflowRunID, taskID int64, attempt int, reason string) {
	if workflowRunID == 0 || taskID == 0 {
		return
	}
	taskIDCopy := taskID
	publishWorkflowEvent(ctx, event.Event{
		WorkflowRunID: workflowRunID,
		EntityType:    event.EntityDomainTask,
		EntityID:      &taskIDCopy,
		EventType:     event.EventTaskEscalateDue,
		Attempt:       attempt,
		Payload: g.Map{
			"task_id": taskID,
			"attempt": attempt,
			"reason":  strings.TrimSpace(reason),
		},
	})
}

func publishWorkflowEvent(ctx context.Context, evt event.Event) {
	if eventPublisher == nil {
		return
	}
	if err := eventPublisher.Emit(ctx, evt); err != nil {
		g.Log().Warningf(ctx, "[WorkflowEvent] publish failed but local fast path already executed: type=%s wfRun=%d err=%v",
			evt.EventType, evt.WorkflowRunID, err)
	}
}

func normalizeEventPayload(payload interface{}) map[string]interface{} {
	switch v := payload.(type) {
	case nil:
		return map[string]interface{}{}
	case map[string]interface{}:
		return v
	default:
		return map[string]interface{}{}
	}
}

func resolveTaskIDFromEvent(evt event.Event, payload map[string]interface{}) int64 {
	if evt.EntityType == event.EntityDomainTask && evt.EntityID != nil {
		return *evt.EntityID
	}
	if payload == nil {
		return 0
	}
	return int64FromPayload(payload, "task_id", "taskId")
}

func int64FromPayload(payload map[string]interface{}, keys ...string) int64 {
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok {
			continue
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
		case string:
			if parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func stringFromPayload(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok {
			continue
		}
		switch v := raw.(type) {
		case string:
			return strings.TrimSpace(v)
		default:
			return strings.TrimSpace(fmt.Sprint(v))
		}
	}
	return ""
}

func expandConfigEnv(raw string) string {
	return strings.TrimSpace(os.ExpandEnv(strings.TrimSpace(raw)))
}

func shouldFallbackToDirectRedisErr(err error) bool {
	if err == nil {
		return false
	}
	lowerErr := strings.ToLower(err.Error())
	return strings.Contains(lowerErr, "noauth") ||
		strings.Contains(lowerErr, "authentication") ||
		strings.Contains(lowerErr, "redis object is nil") ||
		strings.Contains(lowerErr, "redis client unavailable")
}

func resolveRedisPass() string {
	if pass := strings.TrimSpace(os.Getenv("REDIS_PASS")); pass != "" {
		return pass
	}
	return strings.TrimSpace(os.Getenv("REDIS_PASSWORD"))
}
