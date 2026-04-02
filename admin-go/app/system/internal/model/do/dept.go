// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Dept is the golang structure of table dept for DAO operations like Where/Data.
type Dept struct {
	g.Meta    `orm:"table:dept, do:true"`
	Id        any         // 部门ID（Snowflake）
	ParentId  any         // 上级部门ID，0 表示顶级部门
	Title     any         // 部门名称
	Username  any         // 部门负责人姓名
	Email     any         // 负责人邮箱
	Sort      any         // 排序（升序）
	Status    any         // 状态:0=关闭,1=开启
	CreatedBy any         // 创建人ID
	DeptId    any         // 所属部门ID
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 软删除时间，非 NULL 表示已删除
}
