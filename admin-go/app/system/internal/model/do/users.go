// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Users is the golang structure of table users for DAO operations like Where/Data.
type Users struct {
	g.Meta    `orm:"table:users, do:true"`
	Id        any         // 用户ID（Snowflake）
	Username  any         // 登录用户名
	Password  any         // 密码（bcrypt 加密）
	Nickname  any         // 昵称/显示名
	Email     any         // 邮箱地址
	Avatar    any         // 头像图片 URL
	Status    any         // 状态:0=关闭,1=开启
	CreatedBy any         // 创建人ID
	DeptId    any         // 所属部门ID
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 软删除时间，非 NULL 表示已删除
}
