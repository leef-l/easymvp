// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpConfig is the golang structure of table mvp_config for DAO operations like Where/Data.
type MvpConfig struct {
	g.Meta      `orm:"table:mvp_config, do:true"`
	Id          any         // 自增ID
	ConfigKey   any         // 配置键（唯一）
	ConfigValue any         // 配置值
	ConfigType  any         // 值类型：string/int/float/bool/json
	Category    any         // 分类：engine/watchdog/scheduler/general
	Description any         // 配置说明
	CreatedBy   any         // 创建人
	DeptId      any         // 部门ID
	CreatedAt   *gtime.Time // 创建时间
	UpdatedAt   *gtime.Time // 更新时间
	DeletedAt   *gtime.Time // 软删除时间
}
