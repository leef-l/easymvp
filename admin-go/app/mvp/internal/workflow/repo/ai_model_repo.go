package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// AIModelRepo AI 模型仓储。
type AIModelRepo struct{}

func NewAIModelRepo() *AIModelRepo { return &AIModelRepo{} }

func (r *AIModelRepo) table() string { return "ai_model" }

// GetByID 按模型 ID 查询。
func (r *AIModelRepo) GetByID(ctx context.Context, modelID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("id", modelID).
		Where("deleted_at IS NULL")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetWithPlanByID 查询模型及其套餐配置。
func (r *AIModelRepo) GetWithPlanByID(ctx context.Context, modelID int64, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()+" m").Ctx(ctx).
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		Where("m.id", modelID).
		Where("m.deleted_at IS NULL")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetProviderConfigByID 查询模型及 provider/plan 调用配置。
func (r *AIModelRepo) GetProviderConfigByID(ctx context.Context, modelID int64) (g.Map, error) {
	record, err := g.DB().Model(r.table()+" m").Ctx(ctx).
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.model_code, pv.provider_type, pv.supported_protocols, pv.base_url, p.api_key, p.api_secret").
		Where("m.id", modelID).
		Where("m.deleted_at IS NULL").
		One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// GetRolePromptByID 查询模型角色提示词。
func (r *AIModelRepo) GetRolePromptByID(ctx context.Context, modelID int64) (string, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Fields("role_prompt").
		Where("id", modelID).
		WhereNull("deleted_at").
		One()
	if err != nil || record.IsEmpty() {
		return "", err
	}
	return record["role_prompt"].String(), nil
}

// CountEnabledByCapability 统计指定 capability 的启用模型数。
func (r *AIModelRepo) CountEnabledByCapability(ctx context.Context, capability string) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("capability", capability).
		Where("status", 1).
		WhereNull("deleted_at").
		Count()
}

// CountEnabledByCapabilities 统计指定 capability 集合的启用模型数。
func (r *AIModelRepo) CountEnabledByCapabilities(ctx context.Context, capabilities []string) (int, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("status", 1).
		WhereNull("deleted_at")
	if len(capabilities) > 0 {
		model = model.WhereIn("capability", capabilities)
	}
	return model.Count()
}

// GetFirstEnabledWithCredentials 查询首个启用且具备 API 凭据的模型。
func (r *AIModelRepo) GetFirstEnabledWithCredentials(ctx context.Context, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()+" m").Ctx(ctx).
		LeftJoin("ai_plan p", "p.id = m.plan_id AND p.deleted_at IS NULL").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id AND pv.deleted_at IS NULL AND pv.status = 1").
		Where("m.deleted_at IS NULL").
		Where("m.status", 1).
		Where("p.api_key != ''").
		OrderAsc("m.sort")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}
