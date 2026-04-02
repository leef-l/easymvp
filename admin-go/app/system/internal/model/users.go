package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Users DTO 模型

// UsersCreateInput 创建用户表输入
type UsersCreateInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Email string `json:"email"`
	Avatar string `json:"avatar"`
	Status int `json:"status"`
	DeptID snowflake.JsonInt64 `json:"deptId"`
	RoleIDs []snowflake.JsonInt64 `json:"roleIds"`
}

// UsersUpdateInput 更新用户表输入
type UsersUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Email string `json:"email"`
	Avatar string `json:"avatar"`
	Status int `json:"status"`
	DeptID snowflake.JsonInt64 `json:"deptId"`
	RoleIDs []snowflake.JsonInt64 `json:"roleIds"`
}

// UsersDetailOutput 用户表详情输出
type UsersDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email string `json:"email"`
	Avatar string `json:"avatar"`
	Status int `json:"status"`
	DeptID snowflake.JsonInt64 `json:"deptId"`
	DeptTitle string `json:"deptTitle"`
	RoleIDs []snowflake.JsonInt64 `json:"roleIds"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// UsersListOutput 用户表列表输出
type UsersListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email string `json:"email"`
	Avatar string `json:"avatar"`
	Status int `json:"status"`
	DeptID snowflake.JsonInt64 `json:"deptId"`
	DeptTitle string `json:"deptTitle"`
	RoleTitles []string `json:"roleTitles"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// UsersListInput 用户表列表查询输入
type UsersListInput struct {
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	DeptId   snowflake.JsonInt64 `json:"deptId"`
	Status   int `json:"status"`
}

// UsersResetPasswordInput 重置用户密码输入
type UsersResetPasswordInput struct {
	ID       snowflake.JsonInt64 `json:"id"`
	Password string              `json:"password"`
}

