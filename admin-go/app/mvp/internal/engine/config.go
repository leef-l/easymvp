package engine

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
)

// GetConfigInt 读取整型配置，三级 fallback：数据库 mvp_config → config.yaml → 硬编码默认值
func GetConfigInt(ctx context.Context, key string, yamlPath string, defaultVal int) int {
	// 1. 数据库 mvp_config 表
	row, err := g.DB().Model("mvp_config").
		Where("config_key", key).
		Where("deleted_at IS NULL").
		Fields("config_value").
		One()
	if err == nil && !row.IsEmpty() {
		if v, e := strconv.Atoi(row["config_value"].String()); e == nil {
			return v
		}
	}

	// 2. config.yaml
	cfgVal := g.Cfg().MustGet(ctx, yamlPath)
	if cfgVal != nil && !cfgVal.IsEmpty() {
		return cfgVal.Int()
	}

	// 3. 硬编码默认值
	return defaultVal
}

// GetConfigString 读取字符串配置，三级 fallback
func GetConfigString(ctx context.Context, key string, yamlPath string, defaultVal string) string {
	// 1. 数据库
	row, err := g.DB().Model("mvp_config").
		Where("config_key", key).
		Where("deleted_at IS NULL").
		Fields("config_value").
		One()
	if err == nil && !row.IsEmpty() {
		return row["config_value"].String()
	}

	// 2. config.yaml
	cfgVal := g.Cfg().MustGet(ctx, yamlPath)
	if cfgVal != nil && !cfgVal.IsEmpty() {
		return cfgVal.String()
	}

	// 3. 默认值
	return defaultVal
}
