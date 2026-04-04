package event

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// Publisher 统一事件发布器，负责写入 DB + 推送到 Bus。
type Publisher struct {
	bus *Bus
}

// NewPublisher 创建事件发布器。
func NewPublisher(bus *Bus) *Publisher {
	return &Publisher{bus: bus}
}

// Emit 发布一个事件：持久化到 mvp_workflow_event 表，同时推送到内存 Bus。
func (p *Publisher) Emit(ctx context.Context, evt Event) error {
	eventID := snowflake.Generate()
	now := time.Now()

	// 持久化到数据库
	_, err := g.DB().Model("mvp_workflow_event").Ctx(ctx).Insert(g.Map{
		"id":              eventID,
		"workflow_run_id": evt.WorkflowRunID,
		"stage_run_id":    evt.StageRunID,
		"entity_type":     evt.EntityType,
		"entity_id":       evt.EntityID,
		"event_type":      evt.EventType,
		"payload":         evt.Payload,
		"created_at":      now,
	})
	if err != nil {
		g.Log().Errorf(ctx, "[EventPublisher] 持久化事件失败: %v", err)
		// 持久化失败不阻塞内存推送
	}

	// 推送到内存 Bus
	p.bus.Publish(evt)
	return err
}
