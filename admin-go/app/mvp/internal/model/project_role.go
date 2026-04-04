package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// ProjectRole DTO 模型

// ProjectRoleCreateInput 创建项目角色配置表输入
type ProjectRoleCreateInput struct {
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ProjectCategory string `json:"projectCategory"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	SystemPrompt string `json:"systemPrompt"`
	ExecutionMode string `json:"executionMode"`
	Status int `json:"status"`
}

// ProjectRoleUpdateInput 更新项目角色配置表输入
type ProjectRoleUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ProjectCategory string `json:"projectCategory"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	SystemPrompt string `json:"systemPrompt"`
	ExecutionMode string `json:"executionMode"`
	Status int `json:"status"`
}

// ProjectRoleDetailOutput 项目角色配置表详情输出
type ProjectRoleDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ProjectCategory string `json:"projectCategory"`
	ProjectName string `json:"projectName"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	ModelName string `json:"modelName"`
	SystemPrompt string `json:"systemPrompt"`
	ExecutionMode string `json:"executionMode"`
	Status int `json:"status"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ProjectRoleListOutput 项目角色配置表列表输出
type ProjectRoleListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ProjectCategory string `json:"projectCategory"`
	ProjectName string `json:"projectName"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	ModelName string `json:"modelName"`
	SystemPrompt string `json:"systemPrompt"`
	ExecutionMode string `json:"executionMode"`
	Status int `json:"status"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ProjectRoleListInput 项目角色配置表列表查询输入
type ProjectRoleListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Status *int `json:"status"`
}


// ProjectRoleBatchUpdateInput 批量编辑项目角色配置表输入
type ProjectRoleBatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *int                  `json:"status"`
}

