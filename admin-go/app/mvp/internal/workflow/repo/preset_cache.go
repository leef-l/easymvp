package repo

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// presetCache 预设查询缓存，按 categoryCode 缓存默认预设列表。
// 缓存 TTL 5 分钟，自动过期。
type presetCache struct {
	mu      sync.RWMutex
	entries map[string]*presetCacheEntry
}

type presetCacheEntry struct {
	data      gdb.Result
	expiresAt time.Time
}

// modelPromptCache 模型 role_prompt 缓存，按 modelID 缓存。
type modelPromptCache struct {
	mu      sync.RWMutex
	entries map[int64]*promptCacheEntry
}

type promptCacheEntry struct {
	prompt    string
	expiresAt time.Time
}

const cacheTTL = 5 * time.Minute

var (
	globalPresetCache       = &presetCache{entries: make(map[string]*presetCacheEntry)}
	globalModelPromptCache  = &modelPromptCache{entries: make(map[int64]*promptCacheEntry)}
)

// getCachedPresets 从缓存获取默认预设列表，未命中时查库并缓存。
func getCachedPresets(ctx context.Context, categoryCode string, projectCategory string) (gdb.Result, error) {
	cacheKey := categoryCode
	if cacheKey == "" {
		cacheKey = "legacy:" + projectCategory
	}

	globalPresetCache.mu.RLock()
	if entry, ok := globalPresetCache.entries[cacheKey]; ok && time.Now().Before(entry.expiresAt) {
		result := entry.data
		globalPresetCache.mu.RUnlock()
		return result, nil
	}
	globalPresetCache.mu.RUnlock()

	// 查库
	result, err := ListRolePresets(ctx, RolePresetQuery{
		CategoryCode:    categoryCode,
		ProjectCategory: projectCategory,
		DefaultOnly:     true,
	})
	if err != nil {
		return nil, err
	}

	globalPresetCache.mu.Lock()
	globalPresetCache.entries[cacheKey] = &presetCacheEntry{
		data:      result,
		expiresAt: time.Now().Add(cacheTTL),
	}
	globalPresetCache.mu.Unlock()

	return result, nil
}

// getCachedModelPrompt 从缓存获取模型 role_prompt，未命中时查库并缓存。
func getCachedModelPrompt(ctx context.Context, modelID int64) string {
	if modelID == 0 {
		return ""
	}

	globalModelPromptCache.mu.RLock()
	if entry, ok := globalModelPromptCache.entries[modelID]; ok && time.Now().Before(entry.expiresAt) {
		prompt := entry.prompt
		globalModelPromptCache.mu.RUnlock()
		return prompt
	}
	globalModelPromptCache.mu.RUnlock()

	// 查库
	modelRec, err := g.DB().Model("ai_model").Ctx(ctx).
		Fields("role_prompt").
		Where("id", modelID).
		WhereNull("deleted_at").
		One()
	if err != nil {
		g.Log().Warningf(ctx, "[PresetCache] 加载模型 role_prompt 失败: model=%d err=%v", modelID, err)
		return ""
	}
	prompt := ""
	if !modelRec.IsEmpty() {
		prompt = modelRec["role_prompt"].String()
	}

	globalModelPromptCache.mu.Lock()
	globalModelPromptCache.entries[modelID] = &promptCacheEntry{
		prompt:    prompt,
		expiresAt: time.Now().Add(cacheTTL),
	}
	globalModelPromptCache.mu.Unlock()

	return prompt
}

// InvalidatePresetCache 清除预设缓存（用于预设变更后刷新）。
func InvalidatePresetCache() {
	globalPresetCache.mu.Lock()
	globalPresetCache.entries = make(map[string]*presetCacheEntry)
	globalPresetCache.mu.Unlock()
}

// InvalidateModelPromptCache 清除模型提示词缓存。
func InvalidateModelPromptCache() {
	globalModelPromptCache.mu.Lock()
	globalModelPromptCache.entries = make(map[int64]*promptCacheEntry)
	globalModelPromptCache.mu.Unlock()
}
