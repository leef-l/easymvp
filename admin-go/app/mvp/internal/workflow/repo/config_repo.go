package repo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ConfigRepo 配置仓储。
type ConfigRepo struct{}

func NewConfigRepo() *ConfigRepo { return &ConfigRepo{} }

func (r *ConfigRepo) table() string { return "mvp_config" }

// GetByKey 按配置键查询。
func (r *ConfigRepo) GetByKey(ctx context.Context, key string, fields ...string) (g.Map, error) {
	model := g.DB().Model(r.table()).Ctx(ctx).
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

// GetValueByKey 按配置键读取配置值。
func (r *ConfigRepo) GetValueByKey(ctx context.Context, key string) (string, error) {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Fields("config_value").
		One()
	if err != nil || record.IsEmpty() {
		return "", err
	}
	return record["config_value"].String(), nil
}

// CountByKey 统计配置键是否存在。
func (r *ConfigRepo) CountByKey(ctx context.Context, key string) (int, error) {
	return g.DB().Model(r.table()).Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Count()
}

// CountByKeys 统计给定配置键集合已落库的数量。
func (r *ConfigRepo) CountByKeys(ctx context.Context, keys []string) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	return g.DB().Model(r.table()).Ctx(ctx).
		WhereIn("config_key", keys).
		WhereNull("deleted_at").
		Count()
}

// UpsertByKey 按配置键写入配置。
func (r *ConfigRepo) UpsertByKey(ctx context.Context, key string, data g.Map) error {
	record, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Fields("id").
		One()
	if err != nil {
		return err
	}

	now := gtime.Now()
	updateData := g.Map{"updated_at": now}
	for k, v := range data {
		updateData[k] = v
	}

	if !record.IsEmpty() {
		delete(updateData, "config_key")
		delete(updateData, "created_at")
		_, err = g.DB().Model(r.table()).Ctx(ctx).
			Where("id", record["id"].Int64()).
			Data(updateData).
			Update()
		return err
	}

	insertData := g.Map{
		"config_key": key,
		"created_at": now,
		"updated_at": now,
	}
	for k, v := range data {
		insertData[k] = v
	}
	_, err = g.DB().Model(r.table()).Ctx(ctx).Insert(insertData)
	return err
}

// UpdateValueByKey 更新已有配置项的值。
func (r *ConfigRepo) UpdateValueByKey(ctx context.Context, key, value string) (int64, error) {
	result, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Data(g.Map{
			"config_value": value,
			"updated_at":   gtime.Now(),
		}).
		Update()
	if err != nil {
		return 0, err
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// SoftDeleteByKey 软删除配置。
func (r *ConfigRepo) SoftDeleteByKey(ctx context.Context, key string) error {
	_, err := g.DB().Model(r.table()).Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Data(g.Map{
			"deleted_at": gtime.Now(),
			"updated_at": gtime.Now(),
		}).
		Update()
	return err
}
