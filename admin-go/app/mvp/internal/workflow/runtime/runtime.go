// Package runtime 管理 WorkflowRun 级别的运行时上下文。
// 职责：创建/暂停/恢复/取消 context，runtime registry。
package runtime

import (
	"context"
	"sync"
)

// WorkflowRuntime 单个工作流运行的运行时实例。
type WorkflowRuntime struct {
	WorkflowRunID int64
	ProjectID     int64
	Ctx           context.Context
	Cancel        context.CancelFunc
}

// Manager 管理所有活跃的 WorkflowRuntime。
type Manager struct {
	mu       sync.RWMutex
	runtimes map[int64]*WorkflowRuntime // key: workflowRunID
}

// NewManager 创建 runtime 管理器。
func NewManager() *Manager {
	return &Manager{
		runtimes: make(map[int64]*WorkflowRuntime),
	}
}

// Create 为指定工作流创建运行时。如已存在则先取消旧的。
func (m *Manager) Create(workflowRunID, projectID int64) *WorkflowRuntime {
	m.mu.Lock()
	defer m.mu.Unlock()

	if old, ok := m.runtimes[workflowRunID]; ok {
		old.Cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	rt := &WorkflowRuntime{
		WorkflowRunID: workflowRunID,
		ProjectID:     projectID,
		Ctx:           ctx,
		Cancel:        cancel,
	}
	m.runtimes[workflowRunID] = rt
	return rt
}

// Get 获取指定工作流的运行时，不存在返回 nil。
func (m *Manager) Get(workflowRunID int64) *WorkflowRuntime {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.runtimes[workflowRunID]
}

// GetContext 获取指定工作流的 context，不存在返回 Background。
func (m *Manager) GetContext(workflowRunID int64) context.Context {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if rt, ok := m.runtimes[workflowRunID]; ok {
		return rt.Ctx
	}
	return context.Background()
}

// Cancel 取消指定工作流的运行时并移除。
func (m *Manager) Cancel(workflowRunID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rt, ok := m.runtimes[workflowRunID]; ok {
		rt.Cancel()
		delete(m.runtimes, workflowRunID)
	}
}

// CancelByProject 取消指定项目下所有工作流的运行时。
func (m *Manager) CancelByProject(projectID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, rt := range m.runtimes {
		if rt.ProjectID == projectID {
			rt.Cancel()
			delete(m.runtimes, id)
		}
	}
}

// ActiveCount 返回活跃运行时数量。
func (m *Manager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.runtimes)
}
