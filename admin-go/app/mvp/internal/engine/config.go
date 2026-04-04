package engine

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// configCache 配置缓存（内存 + TTL）
// 避免每次 GetConfigInt/GetConfigString 都查 DB
type configCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
}

type cacheEntry struct {
	value     string
	expiresAt time.Time
}

var cfgCache = &configCache{
	entries: make(map[string]*cacheEntry),
	ttl:     60 * time.Second, // 默认 60 秒 TTL
}

// ClearConfigCache 清除配置缓存（管理员修改配置后调用）
func ClearConfigCache() {
	cfgCache.mu.Lock()
	cfgCache.entries = make(map[string]*cacheEntry)
	cfgCache.mu.Unlock()
}

// getFromCache 从缓存读取，返回 (value, found)
func (c *configCache) get(key string) (string, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.value, true
}

// set 写入缓存
func (c *configCache) set(key, value string) {
	c.mu.Lock()
	c.entries[key] = &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// GetConfigInt 读取整型配置，三级 fallback：缓存 → 数据库 mvp_config → config.yaml → 硬编码默认值
func GetConfigInt(ctx context.Context, key string, yamlPath string, defaultVal int) int {
	// 0. 缓存
	if cached, ok := cfgCache.get(key); ok {
		if v, e := strconv.Atoi(cached); e == nil {
			return v
		}
	}

	// 1. 数据库 mvp_config 表
	row, err := g.DB().Model("mvp_config").
		Where("config_key", key).
		Where("deleted_at IS NULL").
		Fields("config_value").
		One()
	if err == nil && !row.IsEmpty() {
		val := row["config_value"].String()
		cfgCache.set(key, val) // 写缓存
		if v, e := strconv.Atoi(val); e == nil {
			return v
		}
	}

	// 2. config.yaml
	cfgVal := g.Cfg().MustGet(ctx, yamlPath)
	if cfgVal != nil && !cfgVal.IsEmpty() {
		return cfgVal.Int()
	}

	// 3. 硬编码默认值
	return defaultVal
}

// CategoryFamily 项目分类族
type CategoryFamily string

const (
	CategoryFamilyCoding   CategoryFamily = "coding"   // 软件开发、游戏开发
	CategoryFamilyCreative CategoryFamily = "creative" // 小说创作、动漫创作、漫剧创作、大电影创作、动画创作
	CategoryFamilyAnalysis CategoryFamily = "analysis" // 数据分析、产品设计
)

// categoryFamilyMap 项目分类 → 分类族映射
var categoryFamilyMap = map[string]CategoryFamily{
	"软件开发":  CategoryFamilyCoding,
	"游戏开发":  CategoryFamilyCoding,
	"小说创作":  CategoryFamilyCreative,
	"动漫创作":  CategoryFamilyCreative,
	"漫剧创作":  CategoryFamilyCreative,
	"大电影创作": CategoryFamilyCreative,
	"动画创作":  CategoryFamilyCreative,
	"数据分析":  CategoryFamilyAnalysis,
	"产品设计":  CategoryFamilyAnalysis,
}

// GetCategoryFamily 获取项目分类所属的分类族
func GetCategoryFamily(projectCategory string) CategoryFamily {
	if f, ok := categoryFamilyMap[projectCategory]; ok {
		return f
	}
	return CategoryFamilyCoding // 默认按编码类处理
}

// GetHeartbeatTimeout 根据分类族获取心跳超时时间（秒）
// coding=120s, creative=300s, analysis=180s
func GetHeartbeatTimeout(ctx context.Context, projectCategory string) int {
	family := GetCategoryFamily(projectCategory)
	switch family {
	case CategoryFamilyCreative:
		return GetConfigInt(ctx, "watchdog.heartbeat_timeout.creative", "engine.watchdog.heartbeatTimeout.creative", 300)
	case CategoryFamilyAnalysis:
		return GetConfigInt(ctx, "watchdog.heartbeat_timeout.analysis", "engine.watchdog.heartbeatTimeout.analysis", 180)
	default:
		return GetConfigInt(ctx, "watchdog.heartbeat_timeout.coding", "engine.watchdog.heartbeatTimeout.coding", 120)
	}
}

// GetConfigString 读取字符串配置，三级 fallback
func GetConfigString(ctx context.Context, key string, yamlPath string, defaultVal string) string {
	// 0. 缓存
	if cached, ok := cfgCache.get(key); ok {
		return cached
	}

	// 1. 数据库
	row, err := g.DB().Model("mvp_config").
		Where("config_key", key).
		Where("deleted_at IS NULL").
		Fields("config_value").
		One()
	if err == nil && !row.IsEmpty() {
		val := row["config_value"].String()
		cfgCache.set(key, val) // 写缓存
		return val
	}

	// 2. config.yaml
	cfgVal := g.Cfg().MustGet(ctx, yamlPath)
	if cfgVal != nil && !cfgVal.IsEmpty() {
		return cfgVal.String()
	}

	// 3. 默认值
	return defaultVal
}
