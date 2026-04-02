package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Dept DTO 模型

// DeptCreateInput 创建部门表输入
type DeptCreateInput struct {
	ParentID snowflake.JsonInt64 `json:"parentID"`
	Title string `json:"title"`
	Username string `json:"username"`
	Email string `json:"email"`
	Sort int `json:"sort"`
	Status int `json:"status"`
}

// DeptUpdateInput 更新部门表输入
type DeptUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ParentID snowflake.JsonInt64 `json:"parentID"`
	Title string `json:"title"`
	Username string `json:"username"`
	Email string `json:"email"`
	Sort int `json:"sort"`
	Status int `json:"status"`
}

// DeptDetailOutput 部门表详情输出
type DeptDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ParentID snowflake.JsonInt64 `json:"parentID"`
	DeptTitle string `json:"deptTitle"`
	Title string `json:"title"`
	Username string `json:"username"`
	Email string `json:"email"`
	Sort int `json:"sort"`
	Status int `json:"status"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// DeptListOutput 部门表列表输出
type DeptListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ParentID snowflake.JsonInt64 `json:"parentID"`
	DeptTitle string `json:"deptTitle"`
	Title string `json:"title"`
	Username string `json:"username"`
	Email string `json:"email"`
	Sort int `json:"sort"`
	Status int `json:"status"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// DeptListInput 部门表列表查询输入
type DeptListInput struct {
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
	Status int `json:"status"`
}

// DeptTreeOutput 部门表树形输出
type DeptTreeOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ParentID snowflake.JsonInt64 `json:"parentID"`
	DeptTitle string `json:"deptTitle"`
	Title string `json:"title"`
	Username string `json:"username"`
	Email string `json:"email"`
	Sort int `json:"sort"`
	Status int `json:"status"`
	Children []*DeptTreeOutput `json:"children"`
}

