package engine

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SSEHub 管理所有活跃的 SSE 连接
// key: messageID (AI 回复消息 ID)
// value: channel，每个 chunk 推送到这个 channel
type SSEHub struct {
	mu       sync.RWMutex
	channels map[int64][]*sseSubscriber
}

type sseSubscriber struct {
	ch        chan string
	closed    bool
	mu        sync.Mutex
	createdAt time.Time
}

var hub = &SSEHub{
	channels: make(map[int64][]*sseSubscriber),
}

func init() {
	// 启动僵尸订阅清理协程：每 60 秒清理超过 10 分钟未关闭的订阅
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			hub.cleanupStaleSubscribers()
		}
	}()
}

// GetHub 获取全局 SSE Hub
func GetHub() *SSEHub {
	return hub
}

// cleanupStaleSubscribers 清理超时未关闭的僵尸订阅
func (h *SSEHub) cleanupStaleSubscribers() {
	threshold := time.Now().Add(-10 * time.Minute)

	h.mu.Lock()
	defer h.mu.Unlock()

	cleaned := 0
	for msgID, subs := range h.channels {
		remaining := subs[:0]
		for _, sub := range subs {
			sub.mu.Lock()
			if !sub.closed && sub.createdAt.Before(threshold) {
				sub.closed = true
				close(sub.ch)
				cleaned++
			} else if !sub.closed {
				remaining = append(remaining, sub)
			}
			sub.mu.Unlock()
		}
		if len(remaining) == 0 {
			delete(h.channels, msgID)
		} else {
			h.channels[msgID] = remaining
		}
	}

	if cleaned > 0 {
		g.Log().Infof(context.Background(), "[SSEHub] 清理了 %d 个僵尸订阅", cleaned)
	}
}

// Subscribe 订阅某个消息的流式输出，返回 channel 和取消函数
func (h *SSEHub) Subscribe(messageID int64) (ch chan string, unsubscribe func()) {
	sub := &sseSubscriber{
		ch:        make(chan string, 100),
		createdAt: time.Now(),
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
