package engine

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow"
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
	if !ok || time.Now().After(entry.expiresAt) {
		c.mu.RUnlock()
		return "", false
	}
	val := entry.value
	c.mu.RUnlock()
	return val, true
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
	return GetConfigIntAny(ctx, []string{key}, []string{yamlPath}, defaultVal)
}

// GetConfigIntAny 按顺序尝试多个 DB key / YAML path，常用于新旧配置键并存的兼容读取。
func GetConfigIntAny(ctx context.Context, keys []string, yamlPaths []string, defaultVal int) int {
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		if cached, ok := cfgCache.get(key); ok {
			if v, e := strconv.Atoi(cached); e == nil {
				return v
			}
		}

		row, err := g.DB().Model("mvp_config").Ctx(ctx).
			Where("config_key", key).
			WhereNull("deleted_at").
			Fields("config_value").
			One()
		if err == nil && !row.IsEmpty() {
			val := row["config_value"].String()
			cfgCache.set(key, val)
			if v, e := strconv.Atoi(val); e == nil {
				return v
			}
		}
	}

	for _, yamlPath := range yamlPaths {
		yamlPath = strings.TrimSpace(yamlPath)
		if yamlPath == "" {
			continue
		}
		cfgVal := g.Cfg().MustGet(ctx, yamlPath)
		if cfgVal != nil && !cfgVal.IsEmpty() {
			return cfgVal.Int()
		}
	}

	return defaultVal
}

// CategoryFamily 项目分类族
type CategoryFamily string

const (
	CategoryFamilyCoding   CategoryFamily = "coding"   // 软件开发、游戏开发
	CategoryFamilyCreative CategoryFamily = "creative" // 小说创作、动漫创作、漫剧创作、大电影创作、动画创作
	CategoryFamilyAnalysis CategoryFamily = "analysis" // 数据分析、产品设计
)

// categoryFamilyFallback 硬编码兜底映射（仅在 CategoryResolver 不可用时使用）
var categoryFamilyFallback = map[string]CategoryFamily{
	// category_code
	"software_dev": CategoryFamilyCoding, "game_dev": CategoryFamilyCoding,
	"novel_writing": CategoryFamilyCreative, "animation_writing": CategoryFamilyCreative,
	"comic_drama_writing": CategoryFamilyCreative, "movie_writing": CategoryFamilyCreative,
	"animation_project": CategoryFamilyCreative,
	"data_analysis":     CategoryFamilyAnalysis, "product_design": CategoryFamilyAnalysis,
	// display_name（兼容旧调用）
	"软件开发": CategoryFamilyCoding, "游戏开发": CategoryFamilyCoding,
	"小说创作": CategoryFamilyCreative, "动漫创作": CategoryFamilyCreative,
	"漫剧创作": CategoryFamilyCreative, "大电影创作": CategoryFamilyCreative,
	"动画创作": CategoryFamilyCreative,
	"数据分析": CategoryFamilyAnalysis, "产品设计": CategoryFamilyAnalysis,
}

// globalResolver 全局分类解析器单例
var globalResolver = workflow.NewCategoryResolver()

// GetCategoryResolver 获取全局分类解析器实例。
func GetCategoryResolver() *workflow.CategoryResolver {
	return globalResolver
}

// GetCategoryFamily 获取项目分类所属的分类族。
// 同时接受 category_code（如 software_dev）和 display_name（如 软件开发）。
func GetCategoryFamily(projectCategory string) CategoryFamily {
	// 先通过 CategoryResolver 查询（走缓存，基本零成本）
	ctx := context.Background()
	info, err := globalResolver.ResolveByCode(ctx, projectCategory)
	if err == nil && info != nil {
		return CategoryFamily(info.FamilyCode)
	}
	info, err = globalResolver.ResolveByDisplayName(ctx, projectCategory)
	if err == nil && info != nil {
		return CategoryFamily(info.FamilyCode)
	}

	// 兜底硬编码
	if f, ok := categoryFamilyFallback[projectCategory]; ok {
		return f
	}
	return CategoryFamilyCoding
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

// GetWatchdogHeartbeatTimeoutSeconds 读取 V2 watchdog 的 lease 超时阈值（秒）。
// 优先读取 watchdog.heartbeat_timeout_seconds；未配置时回退到旧语义：
// check_interval * max_stale_count。
func GetWatchdogHeartbeatTimeoutSeconds(ctx context.Context) int {
	explicitTimeout := GetConfigIntAny(ctx,
		[]string{"watchdog.heartbeat_timeout_seconds"},
		[]string{"watchdog.heartbeat_timeout_seconds", "engine.watchdog.heartbeatTimeoutSeconds"},
		0,
	)
	checkInterval := GetConfigInt(ctx, "watchdog.check_interval", "engine.watchdog.checkInterval", 120)
	maxStaleCount := GetConfigInt(ctx, "watchdog.max_stale_count", "engine.watchdog.maxStaleCount", 3)
	return ResolveWatchdogHeartbeatTimeoutSeconds(explicitTimeout, checkInterval, maxStaleCount)
}

// ResolveWatchdogHeartbeatTimeoutSeconds 计算 watchdog lease 超时阈值（秒）。
func ResolveWatchdogHeartbeatTimeoutSeconds(explicitTimeout, checkInterval, maxStaleCount int) int {
	if explicitTimeout > 0 {
		return explicitTimeout
	}
	if checkInterval < 1 {
		checkInterval = 1
	}
	if maxStaleCount < 1 {
		maxStaleCount = 1
	}
	return checkInterval * maxStaleCount
}

// GetReviewTimeout 获取方案审核超时时间（秒）
func GetReviewTimeout(ctx context.Context) int {
	return GetConfigInt(ctx, "review.timeout_seconds", "engine.review.timeoutSeconds", 300)
}

// GetReviewAutoFixBatch 审核预检时是否自动修正 batch_no 不合理的问题
func GetReviewAutoFixBatch(ctx context.Context) bool {
	return GetConfigInt(ctx, "review.auto_fix_batch", "engine.review.autoFixBatch", 1) == 1
}

// GetSchedulerMaxConcurrency 读取调度最大并发，兼容新旧配置键。
// 默认值固定为 1，优先保证低配服务器可安全落地。
func GetSchedulerMaxConcurrency(ctx context.Context) int {
	maxConcurrency := GetConfigIntAny(ctx,
		[]string{"workflow.scheduler.max_concurrency", "scheduler.max_concurrent"},
		[]string{"workflow.scheduler.maxConcurrency", "engine.scheduler.maxConcurrent"},
		1,
	)
	if maxConcurrency < 1 {
		return 1
	}
	return maxConcurrency
}

// IsFeatureEnabledForProjectType 判断某功能是否对指定项目类型启用。
// configKey 指向一个 JSON 数组配置（如 accept.llm_judge_project_types），值为 ["software_dev","game_dev"]。
// 空字符串、"*"、"[]" 均视为"对所有项目类型启用"。
func IsFeatureEnabledForProjectType(ctx context.Context, configKey, yamlPath, projectType string) bool {
	raw := GetConfigString(ctx, configKey, yamlPath, "")
	raw = strings.TrimSpace(raw)
	// 空值或通配符 → 全部启用
	if raw == "" || raw == "*" || raw == "[]" {
		return true
	}
	var types []string
	if err := json.Unmarshal([]byte(raw), &types); err != nil {
		// 解析失败 → 兜底全部启用（不因配置错误阻塞业务）
		return true
	}
	if len(types) == 0 {
		return true
	}
	for _, t := range types {
		if t == projectType || t == "*" {
			return true
		}
	}
	return false
}

// GetConfigString 读取字符串配置，三级 fallback
func GetConfigString(ctx context.Context, key string, yamlPath string, defaultVal string) string {
	// 0. 缓存
	if cached, ok := cfgCache.get(key); ok {
		return cached
	}

	// 1. 数据库
	row, err := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
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
