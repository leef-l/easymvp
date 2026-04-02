package model

import "easymvp/utility/snowflake"

// Auth 模型

// AuthLoginInput 登录输入
type AuthLoginInput struct {
	Username string
	Password string
}

// AuthLoginOutput 登录输出
type AuthLoginOutput struct {
	Token    string              `json:"token"`
	UserID   snowflake.JsonInt64 `json:"userId"`
	Username string              `json:"username"`
	Nickname string              `json:"nickname"`
	Avatar   string              `json:"avatar"`
}

// AuthInfoOutput 当前用户信息输出
type AuthInfoOutput struct {
	UserID   snowflake.JsonInt64 `json:"userId"`
	Username string              `json:"username"`
	Nickname string              `json:"nickname"`
	Email    string              `json:"email"`
	Avatar   string              `json:"avatar"`
	DeptID   snowflake.JsonInt64 `json:"deptId"`
	Status   int                 `json:"status"`
	Roles    []string            `json:"roles"`
	Perms    []string            `json:"perms"`
}

// AuthChangePasswordInput 修改密码输入
type AuthChangePasswordInput struct {
	UserID      snowflake.JsonInt64
	OldPassword string
	NewPassword string
}

// AuthMenuOutput 用户菜单输出（Vben Admin 路由格式）
type AuthMenuOutput struct {
	ID        snowflake.JsonInt64  `json:"id"`
	ParentID  snowflake.JsonInt64  `json:"parentId"`
	Title     string               `json:"title"`
	Type      int                  `json:"type"`
	Path      string               `json:"path"`
	Component string               `json:"component"`
	Permission string              `json:"permission"`
	Icon      string               `json:"icon"`
	Sort      int                  `json:"sort"`
	IsShow    int                  `json:"isShow"`
	IsCache   int                  `json:"isCache"`
	LinkURL   string               `json:"linkUrl"`
	Status    int                  `json:"status"`
	Children  []*AuthMenuOutput    `json:"children"`
}
