package service

import (
	"sync"
	"time"
)

// ProgressEvent 表示 MACCS 闭环工作流的实时进度事件。
type ProgressEvent struct {
	Type      string `json:"type"`               // phase.started/phase.completed/task.started/task.progress/task.completed/task.failed/review.round/acceptance.layer/delivery.ready
	ProjectID string `json:"project_id"`
	Phase     string `json:"phase,omitempty"`    // requirement/design/review/execution/acceptance/delivery/retrospective
	TaskID    string `json:"task_id,omitempty"`
	RunID     string `json:"run_id,omitempty"`
	Detail    string `json:"detail,omitempty"`
	Progress  int    `json:"progress,omitempty"` // 0-100
	Timestamp string `json:"timestamp"`
}

// progressBus 是 MACCS 进度事件的内存 pub/sub 总线。
// 每个项目维护一组订阅 channel，Publish 向所有订阅者广播。
type progressBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan ProgressEvent
}

// ProgressBus 全局单例。
var (
	globalProgressBus     *progressBus
	globalProgressBusOnce sync.Once
)

// ProgressBusInstance 返回全局 ProgressBus 单例。
func ProgressBusInstance() *progressBus {
	globalProgressBusOnce.Do(func() {
		globalProgressBus = &progressBus{
			subscribers: make(map[string][]chan ProgressEvent),
		}
	})
	return globalProgressBus
}

// Subscribe 为指定项目创建一个缓冲订阅 channel 并返回它。
// 调用方负责在不再需要时调用 Unsubscribe。
func (b *progressBus) Subscribe(projectID string) <-chan ProgressEvent {
	ch := make(chan ProgressEvent, 64)
	b.mu.Lock()
	b.subscribers[projectID] = append(b.subscribers[projectID], ch)
	b.mu.Unlock()
	return ch
}

// Unsubscribe 从订阅列表中移除 channel 并关闭它。
func (b *progressBus) Unsubscribe(projectID string, ch <-chan ProgressEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	list := b.subscribers[projectID]
	for i, sub := range list {
		if sub == ch {
			b.subscribers[projectID] = append(list[:i], list[i+1:]...)
			close(sub)
			break
		}
	}
	if len(b.subscribers[projectID]) == 0 {
		delete(b.subscribers, projectID)
	}
}

// Publish 向指定项目的所有订阅者广播事件。
// 若订阅者 channel 已满则跳过（非阻塞），避免慢客户端拖慢发布方。
// 若 event.Timestamp 为空，自动填写当前 UTC 时间（RFC3339）。
func (b *progressBus) Publish(projectID string, event ProgressEvent) {
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if event.ProjectID == "" {
		event.ProjectID = projectID
	}

	b.mu.RLock()
	list := make([]chan ProgressEvent, len(b.subscribers[projectID]))
	copy(list, b.subscribers[projectID])
	b.mu.RUnlock()

	for _, ch := range list {
		select {
		case ch <- event:
		default:
			// 订阅者消费太慢，丢弃本次事件避免阻塞。
		}
	}
}
