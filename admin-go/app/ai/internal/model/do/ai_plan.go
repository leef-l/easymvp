// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AiPlan is the golang structure of table ai_plan for DAO operations like Where/Data.
type AiPlan struct {
	g.Meta     `orm:"table:ai_plan, do:true"`
	Id         any         // 雪花ID
	ProviderId any         // 供应商ID
	Name       any         // 套餐名称
	Code       any         // 套餐代码
	ApiKey     any         // API Key（加密存储）
	ApiSecret  any         // API Secret（部分供应商需要）
	Status     any         // 状态:0=禁用,1=启用
	Sort       any         // 排序
	CreatedBy  any         // 创建人ID
	DeptId     any         // 所属部门ID
	CreatedAt  *gtime.Time // 创建时间
	UpdatedAt  *gtime.Time // 更新时间
	DeletedAt  *gtime.Time // 软删除时间
}
