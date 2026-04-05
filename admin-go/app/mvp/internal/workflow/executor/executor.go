// Package executor 统一执行器抽象层。
// 所有任务执行器（aider/chat/openhands/claude_code 等）实现 Executor 接口，
// 通过 Registry 注册后由调度链路统一调用，新执行器接入不再改调度器主逻辑。
package executor

import (
	"context"
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/database/gdb"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workspace"
)

// Request 执行器统一请求。
type Request struct {
	ProjectID     int64
	WorkflowRunID int64
	TaskID        int64
	TaskRecord    gdb.Record         // mvp_domain_task 完整记录
	ModelInfo     *engine.ModelInfo   // 解析后的模型信息
	Workspace     *workspace.TaskWorkspace // 可能为 nil（chat 模式不需要隔离）
}

// Result 执行器统一结果。
type Result struct {
	Success bool
	Output  string
	Error   error
}

// Executor 执行器接口。
// 每种执行模式（aider/chat/openhands 等）实现此接口。
type Executor interface {
	// Name 返回执行器名称（与 execution_mode 字段一致）。
	Name() string
	// Execute 执行任务，阻塞直到完成或失败。
	Execute(ctx context.Context, req *Request) *Result
	// NeedsWorkspace 是否需要工作空间隔离。
	NeedsWorkspace() bool
}

// Registry 执行器注册表。
type Registry struct {
	mu        sync.RWMutex
	executors map[string]Executor
}

// NewRegistry 创建执行器注册表。
func NewRegistry() *Registry {
	return &Registry{executors: make(map[string]Executor)}
}

// Register 注册执行器。
func (r *Registry) Register(e Executor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executors[e.Name()] = e
}

// Get 获取执行器，未注册返回 nil。
func (r *Registry) Get(mode string) Executor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.executors[mode]
}

// MustGet 获取执行器，未注册返回 error。
func (r *Registry) MustGet(mode string) (Executor, error) {
	e := r.Get(mode)
	if e == nil {
		return nil, fmt.Errorf("执行模式 %q 尚未注册，请配置为已接入的执行器", mode)
	}
	return e, nil
}

// Modes 返回所有已注册的执行模式名称。
func (r *Registry) Modes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	modes := make([]string, 0, len(r.executors))
	for k := range r.executors {
		modes = append(modes, k)
	}
	return modes
}
