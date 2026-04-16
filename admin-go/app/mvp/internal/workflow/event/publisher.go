package event

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

var beginPublishEventClaimFn = func(ctx context.Context, evt Event) (DurableEventClaim, bool, error) {
	return BeginDurableEventClaim(ctx, DurableEventClaimScopePublish, evt)
}

var insertWorkflowEventRecordFn = func(ctx context.Context, data g.Map) error {
	return repo.NewWorkflowEventRepo().Insert(ctx, data)
}

var persistWorkflowEventFn = func(ctx context.Context, evt Event, recordID int64, createdAt time.Time) error {
	data := g.Map{
		"id":              recordID,
		"event_id":        evt.EventID,
		"workflow_run_id": evt.WorkflowRunID,
		"stage_run_id":    evt.StageRunID,
		"entity_type":     evt.EntityType,
		"entity_id":       evt.EntityID,
		"event_type":      evt.EventType,
		"idempotency_key": evt.IdempotencyKey,
		"attempt":         evt.Attempt,
		"payload":         evt.Payload,
		"created_at":      createdAt,
	}
	err := insertWorkflowEventRecordFn(ctx, data)
	if err == nil || !isMissingWorkflowEventMetadataErr(err) {
		return err
	}

	delete(data, "event_id")
	delete(data, "idempotency_key")
	delete(data, "attempt")
	legacyErr := insertWorkflowEventRecordFn(ctx, data)
	if legacyErr == nil {
		g.Log().Warningf(ctx, "[EventPublisher] workflow event durable metadata migration missing, fallback to legacy insert")
	}
	return legacyErr
}

// PersistRecord 持久化单条事件记录，但不广播到内存 bus/stream。
// 供工作流外围的审计/验证/交付路径复用统一的事件元数据写入逻辑。
func PersistRecord(ctx context.Context, evt Event) error {
	evt = evt.EnsureMetadata()
	return persistWorkflowEventFn(ctx, evt, int64(snowflake.Generate()), time.Now())
}

func isMissingWorkflowEventMetadataErr(err error) bool {
	if err == nil {
		return false
	}
	lowerErr := strings.ToLower(err.Error())
	return strings.Contains(lowerErr, "unknown column") &&
		(strings.Contains(lowerErr, "event_id") || strings.Contains(lowerErr, "idempotency_key") || strings.Contains(lowerErr, "attempt"))
}

// Publisher 统一事件发布器，负责写入 DB + 推送到 Bus。
type Publisher struct {
	bus        *Bus
	streamSink StreamSink
}

// PublisherOption 发布器构造参数。
type PublisherOption func(p *Publisher)

// WithStreamSink 注入可选的持久化事件流投递端。
func WithStreamSink(sink StreamSink) PublisherOption {
	return func(p *Publisher) {
		p.streamSink = sink
	}
}

// NewPublisher 创建事件发布器。
func NewPublisher(bus *Bus, opts ...PublisherOption) *Publisher {
	p := &Publisher{bus: bus}
	for _, opt := range opts {
		if opt != nil {
			opt(p)
		}
	}
	return p
}

// SetStreamSink 运行时更新持久化事件流投递端。
func (p *Publisher) SetStreamSink(sink StreamSink) {
	p.streamSink = sink
}

// StreamStatus 返回持久化事件流状态（如未启用返回 disabled）。
func (p *Publisher) StreamStatus() StreamStatus {
	if p.streamSink == nil {
		return StreamStatus{}
	}
	return p.streamSink.Status()
}

// Emit 发布一个事件：持久化到 mvp_workflow_event 表，同时推送到内存 Bus。
func (p *Publisher) Emit(ctx context.Context, evt Event) (err error) {
	now := time.Now()
	evt = evt.EnsureMetadata()

	var (
		claim         DurableEventClaim
		shouldProcess = true
	)
	claim, shouldProcess, claimErr := beginPublishEventClaimFn(ctx, evt)
	if claimErr != nil {
		if IsMissingDurableEventLedgerErr(claimErr) {
			g.Log().Warningf(ctx, "[EventPublisher] durable event ledger migration missing, fallback to non-durable publish: %v", claimErr)
			claim = nil
			shouldProcess = true
		} else {
			return claimErr
		}
	}
	if !shouldProcess {
		return nil
	}
	defer func() {
		if claim == nil {
			return
		}
		if err != nil {
			if markErr := claim.MarkFailed(ctx, err); markErr != nil {
				g.Log().Warningf(ctx, "[EventPublisher] mark durable claim failed: event=%s err=%v", evt.EventType, markErr)
			}
			return
		}
		if markErr := claim.MarkHandled(ctx); markErr != nil {
			g.Log().Warningf(ctx, "[EventPublisher] mark durable claim handled failed: event=%s err=%v", evt.EventType, markErr)
		}
	}()

	eventID := snowflake.Generate()

	// 持久化到数据库
	err = persistWorkflowEventFn(ctx, evt, int64(eventID), now)
	if err != nil {
		g.Log().Errorf(ctx, "[EventPublisher] 持久化事件失败: %v", err)
		// 持久化失败不阻塞内存推送
	}

	// 推送到内存 Bus
	p.bus.Publish(evt)

	// 可选持久化事件流（降级为 warning，不阻塞主流程）
	if p.streamSink != nil {
		if sinkErr := p.streamSink.Publish(evt); sinkErr != nil {
			g.Log().Warningf(ctx, "[EventPublisher] stream sink publish failed: event=%s err=%v", evt.EventType, sinkErr)
		}
	}
	return err
}
