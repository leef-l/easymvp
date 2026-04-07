// Package event 提供统一的工作流事件发布与订阅。
// 职责：事件发布、SSE 桥接、事件持久化。
package event

import (
	"sync"
)

// Event 工作流事件。
type Event struct {
	WorkflowRunID int64       `json:"workflow_run_id"`
	StageRunID    *int64      `json:"stage_run_id,omitempty"`
	EntityType    string      `json:"entity_type"`
	EntityID      *int64      `json:"entity_id,omitempty"`
	EventType     string      `json:"event_type"`
	Payload       interface{} `json:"payload,omitempty"`
}

// Handler 事件处理函数。
type Handler func(evt Event)

// Bus 事件总线。
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler // key: event_type pattern
}

// NewBus 创建事件总线。
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe 订阅指定事件类型。pattern 支持精确匹配或 "*" 全匹配。
func (b *Bus) Subscribe(pattern string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[pattern] = append(b.handlers[pattern], handler)
}

// Publish 发布事件，同步通知所有匹配的处理函数。
// 先在锁内复制 handler 列表，再在锁外调用，避免 handler 内部 Subscribe 导致死锁。
func (b *Bus) Publish(evt Event) {
	b.mu.RLock()
	var matched []Handler
	if handlers, ok := b.handlers[evt.EventType]; ok {
		matched = append(matched, handlers...)
	}
	if handlers, ok := b.handlers["*"]; ok {
		matched = append(matched, handlers...)
	}
	b.mu.RUnlock()

	for _, h := range matched {
		h(evt)
	}
}
