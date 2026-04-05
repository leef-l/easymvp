package provider

import (
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

// create 根据类型创建 Provider
func (f *Factory) create(cfg Config) (Provider, error) {
	switch cfg.ProviderType {
	case TypeOpenAICompatible:
		return NewOpenAI(cfg), nil
	case TypeAnthropic:
		return NewAnthropic(cfg), nil

	// Anthropic 协议的 Coding Plan（腾讯云、百度等通过 Anthropic 兼容端点代理）
	case "tencent_coding", "baidu_coding":
		return NewAnthropic(cfg), nil

	// OpenAI 兼容协议（大多数国内/海外供应商）
	case "google", "deepseek", "qwen", "zhipu", "moonshot", "minimax",
		"baidu", "tencent", "bytedance", "01ai", "mistral", "groq", "cohere",
		"aliyun_coding":
		return NewOpenAI(cfg), nil

	default:
		// 未知类型默认尝试 OpenAI 兼容协议（大多数供应商都兼容）
		return NewOpenAI(cfg), nil
	}
}

// cacheKey 生成缓存 key
func cacheKey(cfg Config) string {
	return cfg.ProviderType + ":" + cfg.BaseURL + ":" + cfg.APIKey
}
