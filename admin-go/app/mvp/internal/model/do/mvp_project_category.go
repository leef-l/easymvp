// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpProjectCategory is the golang structure of table mvp_project_category for DAO operations like Where/Data.
type MvpProjectCategory struct {
	g.Meta       `orm:"table:mvp_project_category, do:true"`
	Id           any         // 主键ID
	CategoryCode any         // 稳定分类编码
	DisplayName  any         // 展示名称
	FamilyCode   any         // 能力家族编码
	Description  any         // 分类说明
	Status       any         // 1启用 0停用
	Sort         any         // 排序
	CreatedBy    any         // 创建人
	DeptId       any         // 部门ID
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
	DeletedAt    *gtime.Time // 软删除时间
}
