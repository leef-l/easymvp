// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpConfig is the golang structure for table mvp_config.
type MvpConfig struct {
	Id          uint64      `orm:"id"           description:"自增ID"`                                 // 自增ID
	ConfigKey   string      `orm:"config_key"   description:"配置键（唯一）"`                              // 配置键（唯一）
	ConfigValue string      `orm:"config_value" description:"配置值"`                                  // 配置值
	ConfigType  string      `orm:"config_type"  description:"值类型：string/int/float/bool/json"`       // 值类型：string/int/float/bool/json
	Category    string      `orm:"category"     description:"分类：engine/watchdog/scheduler/general"` // 分类：engine/watchdog/scheduler/general
	Description string      `orm:"description"  description:"配置说明"`                                 // 配置说明
	CreatedBy   uint64      `orm:"created_by"   description:"创建人"`                                  // 创建人
	DeptId      uint64      `orm:"dept_id"      description:"部门ID"`                                 // 部门ID
	CreatedAt   *gtime.Time `orm:"created_at"   description:"创建时间"`                                 // 创建时间
	UpdatedAt   *gtime.Time `orm:"updated_at"   description:"更新时间"`                                 // 更新时间
	DeletedAt   *gtime.Time `orm:"deleted_at"   description:"软删除时间"`                                // 软删除时间
}
