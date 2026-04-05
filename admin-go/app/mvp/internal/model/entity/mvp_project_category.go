// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MvpProjectCategory is the golang structure for table mvp_project_category.
type MvpProjectCategory struct {
	Id           int64       `orm:"id"            description:"主键ID"`    // 主键ID
	CategoryCode string      `orm:"category_code" description:"稳定分类编码"`  // 稳定分类编码
	DisplayName  string      `orm:"display_name"  description:"展示名称"`    // 展示名称
	FamilyCode   string      `orm:"family_code"   description:"能力家族编码"`  // 能力家族编码
	Description  string      `orm:"description"   description:"分类说明"`    // 分类说明
	Status       int         `orm:"status"        description:"1启用 0停用"` // 1启用 0停用
	Sort         int         `orm:"sort"          description:"排序"`      // 排序
	CreatedBy    int64       `orm:"created_by"    description:"创建人"`     // 创建人
	DeptId       int64       `orm:"dept_id"       description:"部门ID"`    // 部门ID
	CreatedAt    *gtime.Time `orm:"created_at"    description:"创建时间"`    // 创建时间
	UpdatedAt    *gtime.Time `orm:"updated_at"    description:"更新时间"`    // 更新时间
	DeletedAt    *gtime.Time `orm:"deleted_at"    description:"软删除时间"`   // 软删除时间
}
