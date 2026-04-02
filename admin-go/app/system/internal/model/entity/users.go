// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Users is the golang structure for table users.
type Users struct {
	Id        uint64      `orm:"id"         description:"用户ID（Snowflake）"`    // 用户ID（Snowflake）
	Username  string      `orm:"username"   description:"登录用户名"`              // 登录用户名
	Password  string      `orm:"password"   description:"密码（bcrypt 加密）"`      // 密码（bcrypt 加密）
	Nickname  string      `orm:"nickname"   description:"昵称/显示名"`             // 昵称/显示名
	Email     string      `orm:"email"      description:"邮箱地址"`               // 邮箱地址
	Avatar    string      `orm:"avatar"     description:"头像图片 URL"`           // 头像图片 URL
	Status    int         `orm:"status"     description:"状态:0=关闭,1=开启"`       // 状态:0=关闭,1=开启
	CreatedBy uint64      `orm:"created_by" description:"创建人ID"`              // 创建人ID
	DeptId    uint64      `orm:"dept_id"    description:"所属部门ID"`             // 所属部门ID
	CreatedAt *gtime.Time `orm:"created_at" description:"创建时间"`               // 创建时间
	UpdatedAt *gtime.Time `orm:"updated_at" description:"更新时间"`               // 更新时间
	DeletedAt *gtime.Time `orm:"deleted_at" description:"软删除时间，非 NULL 表示已删除"` // 软删除时间，非 NULL 表示已删除
}
