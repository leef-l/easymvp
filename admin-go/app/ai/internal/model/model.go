package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Model DTO 模型

// ModelCreateInput 创建AI模型表输入
type ModelCreateInput struct {
	PlanID snowflake.JsonInt64 `json:"planID"`
	ProviderID snowflake.JsonInt64 `json:"providerID"`
	Name string `json:"name"`
	ModelCode string `json:"modelCode"`
	Capability string `json:"capability"`
	MaxTokens int `json:"maxTokens"`
	ContextWindow int `json:"contextWindow"`
	SupportsStream int `json:"supportsStream"`
	RolePrompt string `json:"rolePrompt"`
	Status int `json:"status"`
	Sort int `json:"sort"`
}

// ModelUpdateInput 更新AI模型表输入
type ModelUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	PlanID snowflake.JsonInt64 `json:"planID"`
	ProviderID snowflake.JsonInt64 `json:"providerID"`
	Name string `json:"name"`
	ModelCode string `json:"modelCode"`
	Capability string `json:"capability"`
	MaxTokens int `json:"maxTokens"`
	ContextWindow int `json:"contextWindow"`
	SupportsStream int `json:"supportsStream"`
	RolePrompt string `json:"rolePrompt"`
	Status int `json:"status"`
	Sort int `json:"sort"`
}

// ModelDetailOutput AI模型表详情输出
type ModelDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	PlanID snowflake.JsonInt64 `json:"planID"`
	PlanName string `json:"planName"`
	ProviderID snowflake.JsonInt64 `json:"providerID"`
	ProviderName string `json:"providerName"`
	Name string `json:"name"`
	ModelCode string `json:"modelCode"`
	Capability string `json:"capability"`
	MaxTokens int `json:"maxTokens"`
	ContextWindow int `json:"contextWindow"`
	SupportsStream int `json:"supportsStream"`
	RolePrompt string `json:"rolePrompt"`
	Status int `json:"status"`
	Sort int `json:"sort"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ModelListOutput AI模型表列表输出
type ModelListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	PlanID snowflake.JsonInt64 `json:"planID"`
	PlanName string `json:"planName"`
	ProviderID snowflake.JsonInt64 `json:"providerID"`
	ProviderName string `json:"providerName"`
	Name string `json:"name"`
	ModelCode string `json:"modelCode"`
	Capability string `json:"capability"`
	MaxTokens int `json:"maxTokens"`
	ContextWindow int `json:"contextWindow"`
	SupportsStream int `json:"supportsStream"`
	RolePrompt string `json:"rolePrompt"`
	Status int `json:"status"`
	Sort int `json:"sort"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ModelListInput AI模型表列表查询输入
type ModelListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	SupportsStream *int `json:"supportsStream"`
	Status *int `json:"status"`
	Name string `json:"name"`
	Capability string `json:"capability"`
}


// ModelBatchUpdateInput 批量编辑AI模型表输入
type ModelBatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *int                  `json:"status"`
}

