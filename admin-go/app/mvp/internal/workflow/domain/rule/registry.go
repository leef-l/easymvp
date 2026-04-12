package rule

import "sync"

// Registry 规则检查器注册表。
//
// 使用方式：
//
//	reg := rule.NewRegistry()
//	checker := reg.Get("precheck")
//	results := checker.Check(ctx, &rule.CheckContext{...})
//
// 预留扩展点：
//   - 未来接入 CEL 时，通过 Register("precheck", celChecker) 替换内置检查器
//   - 支持多 category：precheck / accept_guard / autonomy
type Registry struct {
	mu       sync.RWMutex
	checkers map[string]Checker // category → checker
}

// NewRegistry 创建并初始化规则注册表，自动注册内置检查器。
func NewRegistry() *Registry {
	r := &Registry{
		checkers: make(map[string]Checker),
	}
	// 注册内置检查器（precheck category）
	r.Register("precheck", NewBuiltinChecker())
	return r
}

// Register 注册或替换指定 category 的检查器。
// 线程安全，可在运行时动态替换（如热更新 CEL 规则）。
func (r *Registry) Register(category string, checker Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers[category] = checker
}

// Get 获取指定 category 的检查器，不存在时返回 nil。
func (r *Registry) Get(category string) Checker {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.checkers[category]
}

// DefaultRegistry 全局默认注册表（单例），供 precheck 层通过 feature flag 切入使用。
var DefaultRegistry = NewRegistry()
