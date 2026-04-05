package provider

import (
	"strings"
	"sync"
)

// Provider 类型常量
const (
	TypeOpenAICompatible = "openai_compatible"
	TypeAnthropic        = "anthropic"
	TypeGoogle           = "google"
)

// Factory Provider 工厂，管理 Provider 实例的创建和缓存
type Factory struct {
	mu    sync.RWMutex
	cache map[string]Provider // key: "providerType:baseURL:apiKey" 的 hash
}

// NewFactory 创建工厂实例
func NewFactory() *Factory {
	return &Factory{
		cache: make(map[string]Provider),
	}
}

// 全局默认工厂
var defaultFactory = NewFactory()

// GetProvider 根据配置获取或创建 Provider 实例
func GetProvider(cfg Config) (Provider, error) {
	return defaultFactory.Get(cfg)
}

// Get 根据配置获取或创建 Provider
func (f *Factory) Get(cfg Config) (Provider, error) {
	key := cacheKey(cfg)

	// 先尝试读缓存
	f.mu.RLock()
	if p, ok := f.cache[key]; ok {
		f.mu.RUnlock()
		return p, nil
	}
	f.mu.RUnlock()

	// 创建新实例
	p, err := f.create(cfg)
	if err != nil {
		return nil, err
	}

	// 写入缓存（double-check：防止并发时重复创建）
	f.mu.Lock()
	if existing, ok := f.cache[key]; ok {
		f.mu.Unlock()
		return existing, nil
	}
	f.cache[key] = p
	f.mu.Unlock()

	return p, nil
}

// ClearCache 清除缓存（配置变更后调用）
func (f *Factory) ClearCache() {
	f.mu.Lock()
	f.cache = make(map[string]Provider)
	f.mu.Unlock()
}

// detectProtocol 根据 base_url 特征自动识别协议类型，优先级高于 provider_type
// 规则：
//   - 包含 /anthropic 路径或 api.anthropic.com 域名 → Anthropic 协议
//   - 包含 generativelanguage.googleapis.com         → Google 原生（当前走 OpenAI 兼容端点）
//   - 其他一律 → OpenAI 兼容协议
func detectProtocol(cfg Config) string {
	url := strings.ToLower(strings.TrimSpace(cfg.BaseURL))
	if strings.Contains(url, "api.anthropic.com") || strings.Contains(url, "/anthropic/") {
		return TypeAnthropic
	}
	// provider_type 明确指定 anthropic 系时也走 Anthropic 协议
	pt := strings.ToLower(cfg.ProviderType)
	if pt == TypeAnthropic || pt == "tencent_coding" || pt == "baidu_coding" {
		return TypeAnthropic
	}
	return TypeOpenAICompatible
}

// create 根据 base_url + provider_type 共同判断协议，创建 Provider
func (f *Factory) create(cfg Config) (Provider, error) {
	switch detectProtocol(cfg) {
	case TypeAnthropic:
		return NewAnthropic(cfg), nil
	default:
		return NewOpenAI(cfg), nil
	}
}

// cacheKey 生成缓存 key
func cacheKey(cfg Config) string {
	return cfg.ProviderType + ":" + cfg.BaseURL + ":" + cfg.APIKey
}
