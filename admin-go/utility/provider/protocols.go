package provider

import (
	"encoding/json"
	"strings"
)

func isTencentCodingProvider(providerType string, baseURL string) bool {
	providerType = strings.ToLower(strings.TrimSpace(providerType))
	baseURL = strings.ToLower(strings.TrimSpace(baseURL))
	return providerType == "tencent_coding" || strings.Contains(baseURL, "api.lkeap.cloud.tencent.com/coding/")
}

func normalizeProtocolValue(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "openai", TypeOpenAICompatible:
		return TypeOpenAICompatible, true
	case TypeAnthropic:
		return TypeAnthropic, true
	case TypeGoogle:
		return TypeGoogle, true
	default:
		return "", false
	}
}

// NormalizeProtocols 规范化供应商支持的协议列表，并补入默认类型。
func NormalizeProtocols(providerType string, supported []string) []string {
	seen := make(map[string]struct{})
	list := make([]string, 0, len(supported)+1)
	appendProtocol := func(v string) {
		normalized, ok := normalizeProtocolValue(v)
		if !ok {
			return
		}
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		list = append(list, normalized)
	}
	appendProtocol(providerType)
	for _, item := range supported {
		appendProtocol(item)
	}
	return list
}

// DecodeSupportedProtocols 从 JSON 文本解析协议列表，失败时回退为默认类型。
func DecodeSupportedProtocols(raw string, providerType string) []string {
	raw = strings.TrimSpace(raw)
	if raw != "" && raw != "null" {
		var out []string
		if err := json.Unmarshal([]byte(raw), &out); err == nil {
			return NormalizeProtocols(providerType, out)
		}
	}
	return NormalizeProtocols(providerType, nil)
}

// SupportsProtocol 判断供应商是否支持目标协议。
func SupportsProtocol(providerType string, supported []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, item := range NormalizeProtocols(providerType, supported) {
		if item == target {
			return true
		}
	}
	return false
}

// ResolveBaseURLForProtocol 按目标协议解析规范化后的 Base URL。
func ResolveBaseURLForProtocol(cfg Config, protocol string) string {
	raw := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	target, ok := normalizeProtocolValue(protocol)
	if !ok {
		target = ResolveProtocol(cfg)
	}

	if raw == "" {
		if isTencentCodingProvider(cfg.ProviderType, raw) {
			switch target {
			case TypeAnthropic:
				return "https://api.lkeap.cloud.tencent.com/coding/anthropic"
			case TypeOpenAICompatible:
				return "https://api.lkeap.cloud.tencent.com/coding/v3"
			}
		}
		return ""
	}

	if isTencentCodingProvider(cfg.ProviderType, raw) {
		switch target {
		case TypeAnthropic:
			return "https://api.lkeap.cloud.tencent.com/coding/anthropic"
		case TypeOpenAICompatible:
			return "https://api.lkeap.cloud.tencent.com/coding/v3"
		}
	}

	if target == TypeAnthropic {
		lowerRaw := strings.ToLower(raw)
		if strings.HasSuffix(lowerRaw, "/v1") {
			raw = strings.TrimSuffix(raw, "/v1")
		}
	}
	return strings.TrimRight(raw, "/")
}

// ResolveBaseURL 按当前配置推导默认协议对应的 Base URL。
func ResolveBaseURL(cfg Config) string {
	return ResolveBaseURLForProtocol(cfg, "")
}
