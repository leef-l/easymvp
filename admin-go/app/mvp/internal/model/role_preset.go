package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// RolePreset DTO 模型

// RolePresetCreateInput 创建角色预设模板输入
type RolePresetCreateInput struct {
	ProjectCategory string `json:"projectCategory"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	SystemPrompt string `json:"systemPrompt"`
	Status int `json:"status"`
	Sort int `json:"sort"`
}

// RolePresetUpdateInput 更新角色预设模板输入
type RolePresetUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectCategory string `json:"projectCategory"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	SystemPrompt string `json:"systemPrompt"`
	Status int `json:"status"`
	Sort int `json:"sort"`
}

// RolePresetDetailOutput 角色预设模板详情输出
type RolePresetDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectCategory string `json:"projectCategory"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	ModelName string `json:"modelName"`
	SystemPrompt string `json:"systemPrompt"`
	Status int `json:"status"`
	Sort int `json:"sort"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// RolePresetListOutput 角色预设模板列表输出
type RolePresetListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectCategory string `json:"projectCategory"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	ModelName string `json:"modelName"`
	SystemPrompt string `json:"systemPrompt"`
	Status int `json:"status"`
	Sort int `json:"sort"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// RolePresetListInput 角色预设模板列表查询输入
type RolePresetListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Status *int `json:"status"`
}


// RolePresetBatchUpdateInput 批量编辑角色预设模板输入
type RolePresetBatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *int                  `json:"status"`
}

