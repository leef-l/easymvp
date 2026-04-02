// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Role is the golang structure of table role for DAO operations like Where/Data.
type Role struct {
	g.Meta    `orm:"table:role, do:true"`
	Id        any         // 角色ID（Snowflake）
	ParentId  any         // 上级角色ID，0 表示顶级角色
	Title     any         // 角色名称
	DataScope any         // 数据范围:1=全部,2=本部门及以下,3=本部门,4=仅本人,5=自定义
	Sort      any         // 排序（升序）
	Status    any         // 状态:0=关闭,1=开启
	IsAdmin   any         // 是否超级管理员:0=否,1=是
	CreatedBy any         // 创建人ID
	DeptId    any         // 所属部门ID
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 软删除时间，非 NULL 表示已删除
}
