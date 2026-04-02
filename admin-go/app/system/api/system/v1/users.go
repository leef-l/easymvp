package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/system/internal/model"
	"easymvp/utility/snowflake"
)

// Users API

// UsersCreateReq 创建用户表请求
type UsersCreateReq struct {
	g.Meta `path:"/users/create" method:"post" tags:"用户表" summary:"创建用户表"`
	Username string `json:"username" v:"required#登录用户名不能为空" dc:"登录用户名"`
	Password string `json:"password" v:"required#密码不能为空" dc:"密码"`
	Nickname string `json:"nickname"  dc:"昵称/显示名"`
	Email string `json:"email"  dc:"邮箱地址"`
	Avatar string `json:"avatar"  dc:"头像图片 URL"`
	Status int `json:"status"  dc:"状态"`
	DeptID snowflake.JsonInt64 `json:"deptId" dc:"所属部门ID"`
	RoleIDs []snowflake.JsonInt64 `json:"roleIds" dc:"角色ID列表"`
}

// UsersCreateRes 创建用户表响应
type UsersCreateRes struct {
	g.Meta `mime:"application/json"`
}

// UsersUpdateReq 更新用户表请求
type UsersUpdateReq struct {
	g.Meta `path:"/users/update" method:"put" tags:"用户表" summary:"更新用户表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"用户表ID"`
	Username string `json:"username" dc:"登录用户名"`
	Password string `json:"password" dc:"密码"`
	Nickname string `json:"nickname" dc:"昵称/显示名"`
	Email string `json:"email" dc:"邮箱地址"`
	Avatar string `json:"avatar" dc:"头像图片 URL"`
	Status int `json:"status" dc:"状态"`
	DeptID snowflake.JsonInt64 `json:"deptId" dc:"所属部门ID"`
	RoleIDs []snowflake.JsonInt64 `json:"roleIds" dc:"角色ID列表"`
}

// UsersUpdateRes 更新用户表响应
type UsersUpdateRes struct {
	g.Meta `mime:"application/json"`
}

// UsersDeleteReq 删除用户表请求
type UsersDeleteReq struct {
	g.Meta `path:"/users/delete" method:"delete" tags:"用户表" summary:"删除用户表"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"用户表ID"`
}

// UsersDeleteRes 删除用户表响应
type UsersDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// UsersDetailReq 获取用户表详情请求
type UsersDetailReq struct {
	g.Meta `path:"/users/detail" method:"get" tags:"用户表" summary:"获取用户表详情"`
	ID     snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"用户表ID"`
}

// UsersDetailRes 获取用户表详情响应
type UsersDetailRes struct {
	g.Meta `mime:"application/json"`
	*model.UsersDetailOutput
}

// UsersListReq 获取用户表列表请求
type UsersListReq struct {
	g.Meta   `path:"/users/list" method:"get" tags:"用户表" summary:"获取用户表列表"`
	PageNum  int `json:"pageNum" d:"1" dc:"页码"`
	PageSize int `json:"pageSize" d:"10" dc:"每页数量"`
	Username string `json:"username" dc:"登录用户名"`
	Nickname string `json:"nickname" dc:"昵称"`
	Email    string `json:"email" dc:"邮箱"`
	DeptId   snowflake.JsonInt64 `json:"deptId" dc:"部门ID"`
	Status   int `json:"status" dc:"状态"`
}

// UsersListRes 获取用户表列表响应
type UsersListRes struct {
	g.Meta `mime:"application/json"`
	List   []*model.UsersListOutput `json:"list" dc:"列表数据"`
	Total  int                               `json:"total" dc:"总数"`
}

// UsersResetPasswordReq 重置用户密码请求
type UsersResetPasswordReq struct {
	g.Meta   `path:"/users/reset-password" method:"put" tags:"用户表" summary:"重置用户密码"`
	ID       snowflake.JsonInt64 `json:"id" v:"required#ID不能为空" dc:"用户ID"`
	Password string              `json:"password" v:"required#新密码不能为空" dc:"新密码"`
}

// UsersResetPasswordRes 重置用户密码响应
type UsersResetPasswordRes struct {
	g.Meta `mime:"application/json"`
}

