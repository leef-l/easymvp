// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Menu is the golang structure for table menu.
type Menu struct {
	Id         uint64      `orm:"id"         description:"菜单ID（Snowflake）"`             // 菜单ID（Snowflake）
	ParentId   uint64      `orm:"parent_id"  description:"上级菜单ID，0 表示顶级菜单"`             // 上级菜单ID，0 表示顶级菜单
	Title      string      `orm:"title"      description:"菜单名称"`                        // 菜单名称
	Type       int         `orm:"type"       description:"类型:1=目录,2=菜单,3=按钮,4=外链,5=内链"` // 类型:1=目录,2=菜单,3=按钮,4=外链,5=内链
	Path       string      `orm:"path"       description:"前端路由路径"`                      // 前端路由路径
	Component  string      `orm:"component"  description:"前端组件路径"`                      // 前端组件路径
	Permission string      `orm:"permission" description:"权限标识（如 system:dept:list）"`    // 权限标识（如 system:dept:list）
	Icon       string      `orm:"icon"       description:"菜单图标（图标名称）"`                  // 菜单图标（图标名称）
	Sort       int         `orm:"sort"       description:"排序（升序）"`                      // 排序（升序）
	IsShow     int         `orm:"is_show"    description:"是否显示:0=隐藏,1=显示"`              // 是否显示:0=隐藏,1=显示
	IsCache    int         `orm:"is_cache"   description:"是否缓存:0=不缓存,1=缓存"`             // 是否缓存:0=不缓存,1=缓存
	LinkUrl    string      `orm:"link_url"   description:"外链/内链地址（type=4或5时有效）"`        // 外链/内链地址（type=4或5时有效）
	Status     int         `orm:"status"     description:"状态:0=关闭,1=开启"`                // 状态:0=关闭,1=开启
	CreatedBy  uint64      `orm:"created_by" description:"创建人ID"`                       // 创建人ID
	DeptId     uint64      `orm:"dept_id"    description:"所属部门ID"`                      // 所属部门ID
	CreatedAt  *gtime.Time `orm:"created_at" description:"创建时间"`                        // 创建时间
	UpdatedAt  *gtime.Time `orm:"updated_at" description:"更新时间"`                        // 更新时间
	DeletedAt  *gtime.Time `orm:"deleted_at" description:"软删除时间，非 NULL 表示已删除"`          // 软删除时间，非 NULL 表示已删除
}
