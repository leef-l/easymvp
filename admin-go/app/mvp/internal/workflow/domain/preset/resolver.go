package preset

import (
	"container/list"
	"context"
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
)

// RolePreset 角色预设记录。
type RolePreset struct {
	ID              int64
	ProjectCategory string
	RoleType        string
	RoleLevel       string
	ModelID         int64
	SystemPrompt    string
	ExecutionMode   string
}

// cacheKey 生成 LRU 缓存键。
func cacheKey(category, roleType, roleLevel string) string {
	return fmt.Sprintf("%s:%s:%s", category, roleType, roleLevel)
}

// lruEntry LRU 链表节点存储的内容。
type lruEntry struct {
	key   string
	value *RolePreset
}

// lruCache 简单的 LRU 缓存，基于 container/list + map 实现。
type lruCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

func newLRUCache(capacity int) *lruCache {
	return &lruCache{
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		order:    list.New(),
	}
}

func (c *lruCache) get(key string) (*RolePreset, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		return el.Value.(*lruEntry).value, true
	}
	return nil, false
}

func (c *lruCache) add(key string, val *RolePreset) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		el.Value.(*lruEntry).value = val
		return
	}
	if c.order.Len() >= c.capacity {
		back := c.order.Back()
		if back != nil {
			c.order.Remove(back)
			delete(c.items, back.Value.(*lruEntry).key)
		}
	}
	el := c.order.PushFront(&lruEntry{key: key, value: val})
	c.items[key] = el
}

func (c *lruCache) remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.order.Remove(el)
		delete(c.items, key)
	}
}

func (c *lruCache) purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element, c.capacity)
	c.order.Init()
}

// Resolver LRU 缓存的角色预设解析器。
type Resolver struct {
	cache *lruCache
}

// NewResolver 创建预设解析器，capacity 为 LRU 缓存容量。
func NewResolver(capacity int) (*Resolver, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("LRU 缓存容量必须大于 0")
	}
	return &Resolver{cache: newLRUCache(capacity)}, nil
}

// Resolve 解析角色预设。先查 LRU 缓存，未命中则查库。
// 不做多级 fallback，找不到直接返回 nil, nil。
func (r *Resolver) Resolve(ctx context.Context, category, roleType, roleLevel string) (*RolePreset, error) {
	key := cacheKey(category, roleType, roleLevel)

	// 缓存命中
	if val, ok := r.cache.get(key); ok {
		return val, nil
	}

	// 查库（单层查询，不做 fallback）
	var preset RolePreset
	err := g.DB().Ctx(ctx).Model("mvp_role_preset").
		Where("project_category", category).
		Where("role_type", roleType).
		Where("role_level", roleLevel).
		Where("is_default", 1).
		Where("status", 1).
		Where("deleted_at IS NULL").
		Fields("id, project_category, role_type, role_level, model_id, system_prompt, execution_mode").
		Scan(&preset)
	if err != nil {
		return nil, fmt.Errorf("查询角色预设失败: %w", err)
	}

	// 未找到
	if preset.ID == 0 {
		return nil, nil
	}

	// 写入缓存
	r.cache.add(key, &preset)
	return &preset, nil
}

// Invalidate 使缓存中指定条目失效。
func (r *Resolver) Invalidate(category, roleType, roleLevel string) {
	r.cache.remove(cacheKey(category, roleType, roleLevel))
}

// InvalidateAll 清空所有缓存。
func (r *Resolver) InvalidateAll() {
	r.cache.purge()
}
