package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// AIProviderRepo AI 供应商仓储。
type AIProviderRepo struct{}

func NewAIProviderRepo() *AIProviderRepo { return &AIProviderRepo{} }

func (r *AIProviderRepo) table() string { return "ai_provider" }

// CountEnabledWithBaseURL 统计已启用且配置了 base_url 的供应商数。
func (r *AIProviderRepo) CountEnabledWithBaseURL(ctx context.Context) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("status", 1).
		Where("base_url != ''").
		WhereNull("deleted_at").
		Count()
}
