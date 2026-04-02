// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Menu is the golang structure of table menu for DAO operations like Where/Data.
type Menu struct {
	g.Meta     `orm:"table:menu, do:true"`
	Id         any         // 菜单ID（Snowflake）
	ParentId   any         // 上级菜单ID，0 表示顶级菜单
	Title      any         // 菜单名称
	Type       any         // 类型:1=目录,2=菜单,3=按钮,4=外链,5=内链
	Path       any         // 前端路由路径
	Component  any         // 前端组件路径
	Permission any         // 权限标识（如 system:dept:list）
	Icon       any         // 菜单图标（图标名称）
	Sort       any         // 排序（升序）
	IsShow     any         // 是否显示:0=隐藏,1=显示
	IsCache    any         // 是否缓存:0=不缓存,1=缓存
	LinkUrl    any         // 外链/内链地址（type=4或5时有效）
	Status     any         // 状态:0=关闭,1=开启
	CreatedBy  any         // 创建人ID
	DeptId     any         // 所属部门ID
	CreatedAt  *gtime.Time // 创建时间
	UpdatedAt  *gtime.Time // 更新时间
	DeletedAt  *gtime.Time // 软删除时间，非 NULL 表示已删除
}
