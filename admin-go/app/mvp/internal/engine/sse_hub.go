package engine

import (
	"sync"
)

// SSEHub 管理所有活跃的 SSE 连接
// key: messageID (AI 回复消息 ID)
// value: channel，每个 chunk 推送到这个 channel
type SSEHub struct {
	mu       sync.RWMutex
	channels map[int64][]*sseSubscriber
}

type sseSubscriber struct {
	ch     chan string
	closed bool
	mu     sync.Mutex
}

var hub = &SSEHub{
	channels: make(map[int64][]*sseSubscriber),
}

// GetHub 获取全局 SSE Hub
func GetHub() *SSEHub {
	return hub
}

// Subscribe 订阅某个消息的流式输出，返回 channel 和取消函数
func (h *SSEHub) Subscribe(messageID int64) (ch chan string, unsubscribe func()) {
	sub := &sseSubscriber{
		ch: make(chan string, 100),
	}

	h.mu.Lock()
	h.channels[messageID] = append(h.channels[messageID], sub)
	h.mu.Unlock()

	unsubscribe = func() {
		sub.mu.Lock()
		defer sub.mu.Unlock()
		if !sub.closed {
			sub.closed = true
			close(sub.ch)
		}

		h.mu.Lock()
		defer h.mu.Unlock()
		subs := h.channels[messageID]
		for i, s := range subs {
			if s == sub {
				h.channels[messageID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		if len(h.channels[messageID]) == 0 {
			delete(h.channels, messageID)
		}
	}

	return sub.ch, unsubscribe
}

// Publish 向某个消息的所有订阅者推送 chunk
func (h *SSEHub) Publish(messageID int64, data string) {
	h.mu.RLock()
	subs := make([]*sseSubscriber, len(h.channels[messageID]))
	copy(subs, h.channels[messageID])
	h.mu.RUnlock()

	for _, sub := range subs {
		sub.mu.Lock()
		if !sub.closed {
			select {
			case sub.ch <- data:
			default:
				// channel 满了就跳过，避免阻塞
			}
		}
		sub.mu.Unlock()
	}
}

// Done 通知某个消息的流式输出已完成，关闭所有订阅
func (h *SSEHub) Done(messageID int64) {
	h.mu.Lock()
	subs := h.channels[messageID]
	delete(h.channels, messageID)
	h.mu.Unlock()

	for _, sub := range subs {
		sub.mu.Lock()
		if !sub.closed {
			sub.closed = true
			close(sub.ch)
		}
		sub.mu.Unlock()
	}
}

// HasSubscribers 检查某个消息是否有活跃的订阅者
func (h *SSEHub) HasSubscribers(messageID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.channels[messageID]) > 0
}
