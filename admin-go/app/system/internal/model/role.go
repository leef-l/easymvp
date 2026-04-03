package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Role DTO 模型

// RoleCreateInput 创建角色表输入
type RoleCreateInput struct {
	ParentID        snowflake.JsonInt64 `json:"parentID"`
	Title           string              `json:"title"`
	DataScope       int                 `json:"dataScope"`
	DefaultAiEngine string              `json:"defaultAiEngine"`
	Sort            int                 `json:"sort"`
	Status          int                 `json:"status"`
	IsAdmin         int                 `json:"isAdmin"`
}

// RoleUpdateInput 更新角色表输入
type RoleUpdateInput struct {
	ID              snowflake.JsonInt64 `json:"id"`
	ParentID        snowflake.JsonInt64 `json:"parentID"`
	Title           string              `json:"title"`
	DataScope       int                 `json:"dataScope"`
	DefaultAiEngine string              `json:"defaultAiEngine"`
	Sort            int                 `json:"sort"`
	Status          int                 `json:"status"`
	IsAdmin         int                 `json:"isAdmin"`
}

// RoleDetailOutput 角色表详情输出
type RoleDetailOutput struct {
	ID              snowflake.JsonInt64 `json:"id"`
	ParentID        snowflake.JsonInt64 `json:"parentID"`
	RoleTitle       string              `json:"roleTitle"`
	Title           string              `json:"title"`
	DataScope       int                 `json:"dataScope"`
	DefaultAiEngine string              `json:"defaultAiEngine"`
	Sort            int                 `json:"sort"`
	Status          int                 `json:"status"`
	IsAdmin         int                 `json:"isAdmin"`
	CreatedAt       *gtime.Time         `json:"createdAt"`
	UpdatedAt       *gtime.Time         `json:"updatedAt"`
}

// RoleListOutput 角色表列表输出
type RoleListOutput struct {
	ID              snowflake.JsonInt64 `json:"id"`
	ParentID        snowflake.JsonInt64 `json:"parentID"`
	RoleTitle       string              `json:"roleTitle"`
	Title           string              `json:"title"`
	DataScope       int                 `json:"dataScope"`
	DefaultAiEngine string              `json:"defaultAiEngine"`
	Sort            int                 `json:"sort"`
	Status          int                 `json:"status"`
	IsAdmin         int                 `json:"isAdmin"`
	CreatedAt       *gtime.Time         `json:"createdAt"`
	UpdatedAt       *gtime.Time         `json:"updatedAt"`
}

// RoleListInput 角色表列表查询输入
type RoleListInput struct {
	PageNum   int `json:"pageNum"`
	PageSize  int `json:"pageSize"`
	DataScope int `json:"dataScope"`
	Status    int `json:"status"`
}

// RoleTreeOutput 角色表树形输出
type RoleTreeOutput struct {
	ID              snowflake.JsonInt64 `json:"id"`
	ParentID        snowflake.JsonInt64 `json:"parentID"`
	RoleTitle       string              `json:"roleTitle"`
	Title           string              `json:"title"`
	DataScope       int                 `json:"dataScope"`
	DefaultAiEngine string              `json:"defaultAiEngine"`
	Sort            int                 `json:"sort"`
	Status          int                 `json:"status"`
	IsAdmin         int                 `json:"isAdmin"`
	Children        []*RoleTreeOutput   `json:"children"`
}

// RoleGrantMenuInput 角色授权菜单输入
type RoleGrantMenuInput struct {
	ID      snowflake.JsonInt64
	MenuIDs []snowflake.JsonInt64
}

// RoleGrantDeptInput 角色授权数据权限输入
type RoleGrantDeptInput struct {
	ID        snowflake.JsonInt64
	DataScope int
	DeptIDs   []snowflake.JsonInt64
}

// RoleGrantAiEngineInput 角色授权AI执行引擎输入
type RoleGrantAiEngineInput struct {
	ID          snowflake.JsonInt64
	EngineCodes []string
}
