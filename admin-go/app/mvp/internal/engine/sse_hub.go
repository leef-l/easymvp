package engine

import (
	"sync"
)

// SSEHub 管理所有活跃的 SSE 连接
// key: messageID (AI 回复消息 ID)
// value: channel，每个 chunk 推送到这个 channel
type SSEHub struct {
	mu       sync.RWMutex
	channels map[int64][]chan string // 一个 messageID 可能有多个订阅者
}

var hub = &SSEHub{
	channels: make(map[int64][]chan string),
}

// GetHub 获取全局 SSE Hub
func GetHub() *SSEHub {
	return hub
}

// Subscribe 订阅某个消息的流式输出，返回 channel 和取消函数
func (h *SSEHub) Subscribe(messageID int64) (ch chan string, unsubscribe func()) {
	ch = make(chan string, 100) // 带缓冲防止阻塞

	h.mu.Lock()
	h.channels[messageID] = append(h.channels[messageID], ch)
	h.mu.Unlock()

	unsubscribe = func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		subs := h.channels[messageID]
		for i, sub := range subs {
			if sub == ch {
				h.channels[messageID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		if len(h.channels[messageID]) == 0 {
			delete(h.channels, messageID)
		}
		close(ch)
	}

	return ch, unsubscribe
}

// Publish 向某个消息的所有订阅者推送 chunk
func (h *SSEHub) Publish(messageID int64, data string) {
	h.mu.RLock()
	subs := h.channels[messageID]
	h.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- data:
		default:
			// channel 满了就跳过，避免阻塞
		}
	}
}

// Done 通知某个消息的流式输出已完成，关闭所有订阅
func (h *SSEHub) Done(messageID int64) {
	h.mu.Lock()
	subs := h.channels[messageID]
	delete(h.channels, messageID)
	h.mu.Unlock()

	for _, ch := range subs {
		close(ch)
	}
}

// HasSubscribers 检查某个消息是否有活跃的订阅者
func (h *SSEHub) HasSubscribers(messageID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.channels[messageID]) > 0
}
