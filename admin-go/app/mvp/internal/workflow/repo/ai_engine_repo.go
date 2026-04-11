package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// AIEngineRepo AI 执行引擎仓储。
type AIEngineRepo struct{}

func NewAIEngineRepo() *AIEngineRepo { return &AIEngineRepo{} }

func (r *AIEngineRepo) table() string { return "ai_engine" }

// CountEnabled 统计启用中的执行引擎数量。
func (r *AIEngineRepo) CountEnabled(ctx context.Context) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("status", 1).
		WhereNull("deleted_at").
		Count()
}

// AIEngineConfigRepo AI 引擎配置仓储。
type AIEngineConfigRepo struct{}

func NewAIEngineConfigRepo() *AIEngineConfigRepo { return &AIEngineConfigRepo{} }

func (r *AIEngineConfigRepo) table() string { return "ai_engine_config" }

// GetByCode 按引擎编码读取配置。
func (r *AIEngineConfigRepo) GetByCode(ctx context.Context, engineCode string, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
		Where("engine_code", engineCode).
		WhereNull("deleted_at")
	if len(fields) > 0 {
		model = model.Fields(strings.Join(fields, ","))
	}
	record, err := model.One()
	if err != nil || record.IsEmpty() {
		return nil, err
	}
	return record.Map(), nil
}

// SystemRoleAIEngineRepo 角色引擎授权仓储。
type SystemRoleAIEngineRepo struct{}

func NewSystemRoleAIEngineRepo() *SystemRoleAIEngineRepo { return &SystemRoleAIEngineRepo{} }

func (r *SystemRoleAIEngineRepo) table() string { return "system_role_ai_engine" }

// CountAll 统计全部角色引擎授权记录。
func (r *SystemRoleAIEngineRepo) CountAll(ctx context.Context) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).Count()
}
