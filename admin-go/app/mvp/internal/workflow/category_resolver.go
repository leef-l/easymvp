// Package workflow 提供工作流领域的公共能力。
package workflow

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"easymvp/app/mvp/internal/workflow/repo"
)

// CategoryInfo 分类解析结果。
type CategoryInfo struct {
	CategoryCode string // 稳定编码，如 software_dev
	DisplayName  string // 展示名，如 软件开发
	FamilyCode   string // 能力家族，如 coding
}

// CategoryResolver 统一项目分类解析器。
// 所有需要读取项目分类的模块都通过此入口，不直接摸 project_category 字段。
type CategoryResolver struct {
	categoryRepo *repo.ProjectCategoryRepo

	// 内存缓存：display_name → CategoryInfo
	mu          sync.RWMutex
	codeCache   map[string]*CategoryInfo // category_code → info
	nameCache   map[string]*CategoryInfo // display_name → info
	cacheLoaded bool
}

// NewCategoryResolver 创建分类解析器。
func NewCategoryResolver() *CategoryResolver {
	return &CategoryResolver{
		categoryRepo: repo.NewProjectCategoryRepo(),
		codeCache:    make(map[string]*CategoryInfo),
		nameCache:    make(map[string]*CategoryInfo),
	}
}

// ResolveByProject 根据项目 ID 解析分类。
// 优先读 category_code，若为空则根据 project_category 兼容映射并异步回填。
func (r *CategoryResolver) ResolveByProject(ctx context.Context, projectID int64) (*CategoryInfo, error) {
	// 查询项目的 category_code 和 project_category
	record, err := g.DB().Model("mvp_project").Ctx(ctx).
		Fields("category_code", "project_category").
		Where("id", projectID).
		WhereNull("deleted_at").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return r.fallbackDefault(), nil
	}

	categoryCode := record["category_code"].String()
	projectCategory := record["project_category"].String()

	// 优先走 category_code
	if categoryCode != "" {
		info, err := r.ResolveByCode(ctx, categoryCode)
		if err == nil && info != nil {
			return info, nil
		}
	}

	// 兼容：用 display_name（中文分类名）映射
	if projectCategory != "" {
		info, err := r.resolveByDisplayName(ctx, projectCategory)
		if err == nil && info != nil {
			// 异步回填 category_code
			go func() {
				defer func() {
					if r := recover(); r != nil {
						g.Log().Errorf(context.Background(), "[CategoryResolver] 回填 category_code panic: %v", r)
					}
				}()
				if _, bfErr := g.DB().Model("mvp_project").Ctx(context.Background()).
					Where("id", projectID).
					Where("category_code IS NULL OR category_code = ''").
					Data(g.Map{"category_code": info.CategoryCode}).
					Update(); bfErr != nil {
					g.Log().Warningf(context.Background(), "[CategoryResolver] 回填 category_code 失败: projectID=%d err=%v", projectID, bfErr)
				}
			}()
			return info, nil
		}
	}

	return r.fallbackDefault(), nil
}

// ResolveByCode 根据 category_code 解析分类。
func (r *CategoryResolver) ResolveByCode(ctx context.Context, categoryCode string) (*CategoryInfo, error) {
	if err := r.ensureCache(ctx); err != nil {
		// 缓存加载失败，直接查库
		return r.resolveByCodeDirect(ctx, categoryCode)
	}

	r.mu.RLock()
	info, ok := r.codeCache[categoryCode]
	r.mu.RUnlock()
	if ok {
		return info, nil
	}

	// 缓存未命中，查库
	return r.resolveByCodeDirect(ctx, categoryCode)
}

// ResolveByDisplayName 根据展示名称解析分类（兼容旧代码）。
func (r *CategoryResolver) ResolveByDisplayName(ctx context.Context, displayName string) (*CategoryInfo, error) {
	return r.resolveByDisplayName(ctx, displayName)
}

// GetFamilyCode 快捷方法：根据 category_code 获取 family_code。
func (r *CategoryResolver) GetFamilyCode(ctx context.Context, categoryCode string) string {
	info, err := r.ResolveByCode(ctx, categoryCode)
	if err != nil || info == nil {
		return "coding" // 默认回退
	}
	return info.FamilyCode
}

// ListAll 列出所有启用的分类。
func (r *CategoryResolver) ListAll(ctx context.Context) ([]*CategoryInfo, error) {
	if err := r.ensureCache(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*CategoryInfo, 0, len(r.codeCache))
	for _, info := range r.codeCache {
		result = append(result, info)
	}
	return result, nil
}

// InvalidateCache 清除缓存，下次访问时重新加载。
func (r *CategoryResolver) InvalidateCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cacheLoaded = false
	r.codeCache = make(map[string]*CategoryInfo)
	r.nameCache = make(map[string]*CategoryInfo)
}

// ── 内部方法 ──

func (r *CategoryResolver) ensureCache(ctx context.Context) error {
	r.mu.RLock()
	if r.cacheLoaded {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cacheLoaded {
		return nil
	}

	records, err := r.categoryRepo.ListAll(ctx)
	if err != nil {
		return err
	}

	for _, rec := range records {
		info := &CategoryInfo{
			CategoryCode: gconv.String(rec["category_code"]),
			DisplayName:  gconv.String(rec["display_name"]),
			FamilyCode:   gconv.String(rec["family_code"]),
		}
		r.codeCache[info.CategoryCode] = info
		r.nameCache[info.DisplayName] = info
	}
	r.cacheLoaded = true
	return nil
}

func (r *CategoryResolver) resolveByDisplayName(ctx context.Context, displayName string) (*CategoryInfo, error) {
	if err := r.ensureCache(ctx); err == nil {
		r.mu.RLock()
		info, ok := r.nameCache[displayName]
		r.mu.RUnlock()
		if ok {
			return info, nil
		}
	}

	// 缓存未命中，查库
	rec, err := r.categoryRepo.GetByDisplayName(ctx, displayName)
	if err != nil || rec == nil {
		return nil, err
	}
	return &CategoryInfo{
		CategoryCode: gconv.String(rec["category_code"]),
		DisplayName:  gconv.String(rec["display_name"]),
		FamilyCode:   gconv.String(rec["family_code"]),
	}, nil
}

func (r *CategoryResolver) resolveByCodeDirect(ctx context.Context, categoryCode string) (*CategoryInfo, error) {
	rec, err := r.categoryRepo.GetByCode(ctx, categoryCode)
	if err != nil || rec == nil {
		return nil, err
	}
	return &CategoryInfo{
		CategoryCode: gconv.String(rec["category_code"]),
		DisplayName:  gconv.String(rec["display_name"]),
		FamilyCode:   gconv.String(rec["family_code"]),
	}, nil
}

func (r *CategoryResolver) fallbackDefault() *CategoryInfo {
	return &CategoryInfo{
		CategoryCode: "software_dev",
		DisplayName:  "软件开发",
		FamilyCode:   "coding",
	}
}
