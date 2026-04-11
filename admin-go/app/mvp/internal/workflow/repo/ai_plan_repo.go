package repo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// AIPlanRepo AI 套餐仓储。
type AIPlanRepo struct{}

func NewAIPlanRepo() *AIPlanRepo { return &AIPlanRepo{} }

func (r *AIPlanRepo) table() string { return "ai_plan" }

// CountEnabledWithAPIKey 统计已启用且配置了 api_key 的套餐数。
func (r *AIPlanRepo) CountEnabledWithAPIKey(ctx context.Context) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("status", 1).
		Where("api_key != ''").
		WhereNull("deleted_at").
		Count()
}
