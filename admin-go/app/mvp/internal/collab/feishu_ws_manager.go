package collab

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
)

// wsManager 全局 WS 客户端管理器（单例）。
var wsManager = &FeishuWSManager{}

// FeishuWSManager 管理 WebSocket 客户端的生命周期。
type FeishuWSManager struct {
	mu     sync.Mutex
	client *FeishuWSClient
	cancel context.CancelFunc
}

// GetWSManager 返回全局 WS 管理器单例。
func GetWSManager() *FeishuWSManager {
	return wsManager
}

// StartWS 启动 WebSocket 长连接（如已运行则先停止旧连接）。
func (m *FeishuWSManager) StartWS(appID, appSecret string, onEvent func(ctx context.Context, header, event map[string]interface{})) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 停掉旧连接
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	if m.client != nil {
		m.client.Stop()
		m.client = nil
	}

	client := NewFeishuWSClient(appID, appSecret)
	client.OnEvent = onEvent

	ctx, cancel := context.WithCancel(context.Background())
	m.client = client
	m.cancel = cancel

	go func() {
		client.Start(ctx)
		g.Log().Info(ctx, "[FeishuWSManager] 长连接已退出")
	}()
}

// StopWS 停止 WebSocket 长连接。
func (m *FeishuWSManager) StopWS() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	if m.client != nil {
		m.client.Stop()
		m.client = nil
	}
}

// IsRunning 返回当前是否有活跃的 WS 连接。
func (m *FeishuWSManager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.client != nil && m.client.IsRunning()
}
