package autonomy

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/frame/g"
)

// mapInt64 从 g.Map 安全提取 int64 值。
func mapInt64(m g.Map, key string) int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
	}
}

// mapString 从 g.Map 安全提取 string 值。
func mapString(m g.Map, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return ""
	}
}

// parseJSONMap 将 JSON 字符串解析为 map。空串或解析失败返回空 map（非 nil）。
func parseJSONMap(s string) map[string]interface{} {
	if s == "" || s == "null" {
		return make(map[string]interface{})
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		g.Log().Debugf(context.Background(), "[autonomy] parseJSONMap failed: %v, input: %.100s", err, s)
		return make(map[string]interface{})
	}
	if m == nil {
		return make(map[string]interface{})
	}
	return m
}
