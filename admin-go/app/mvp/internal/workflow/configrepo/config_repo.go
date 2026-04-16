package configrepo

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ConfigRepo 配置仓储。
// 保持在独立叶子包中，避免 workflow/repo 与 configstore 形成 import cycle。
type ConfigRepo struct{}

func NewConfigRepo() *ConfigRepo { return &ConfigRepo{} }

func (r *ConfigRepo) table() string { return "mvp_config" }

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
