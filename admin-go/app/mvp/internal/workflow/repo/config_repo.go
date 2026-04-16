package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/configrepo"
)

// ConfigRepo 对外仍归于 repo 命名空间，底层实现收敛到独立叶子包以避免 import cycle。
type ConfigRepo struct {
	*configrepo.ConfigRepo
}

func NewConfigRepo() *ConfigRepo {
	return &ConfigRepo{ConfigRepo: configrepo.NewConfigRepo()}
}

// GetByKey 按配置键查询。
func (r *ConfigRepo) GetByKey(ctx context.Context, key string, fields ...string) (g.Map, error) {
	model := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", key).
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

// CountByKey 统计配置键是否存在。
func (r *ConfigRepo) CountByKey(ctx context.Context, key string) (int, error) {
	return g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Count()
}

// CountByKeys 统计给定配置键集合已落库的数量。
func (r *ConfigRepo) CountByKeys(ctx context.Context, keys []string) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	return g.DB().Model("mvp_config").Ctx(ctx).
		WhereIn("config_key", keys).
		WhereNull("deleted_at").
		Count()
}

// UpdateValueByKey 更新已有配置项的值。
func (r *ConfigRepo) UpdateValueByKey(ctx context.Context, key, value string) (int64, error) {
	result, err := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Data(g.Map{
			"config_value": value,
		}).
		Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}
