// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// AiPlan is the golang structure for table ai_plan.
type AiPlan struct {
	Id         uint64      `orm:"id"          description:"雪花ID"`                // 雪花ID
	ProviderId uint64      `orm:"provider_id" description:"供应商ID"`               // 供应商ID
	Name       string      `orm:"name"        description:"套餐名称"`                // 套餐名称
	Code       string      `orm:"code"        description:"套餐代码"`                // 套餐代码
	ApiKey     string      `orm:"api_key"     description:"API Key（加密存储）"`       // API Key（加密存储）
	ApiSecret  string      `orm:"api_secret"  description:"API Secret（部分供应商需要）"` // API Secret（部分供应商需要）
	Status     int         `orm:"status"      description:"状态:0=禁用,1=启用"`        // 状态:0=禁用,1=启用
	Sort       int         `orm:"sort"        description:"排序"`                  // 排序
	CreatedBy  uint64      `orm:"created_by"  description:"创建人ID"`               // 创建人ID
	DeptId     uint64      `orm:"dept_id"     description:"所属部门ID"`              // 所属部门ID
	CreatedAt  *gtime.Time `orm:"created_at"  description:"创建时间"`                // 创建时间
	UpdatedAt  *gtime.Time `orm:"updated_at"  description:"更新时间"`                // 更新时间
	DeletedAt  *gtime.Time `orm:"deleted_at"  description:"软删除时间"`               // 软删除时间
}
