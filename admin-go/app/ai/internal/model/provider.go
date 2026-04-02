package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Provider DTO 模型

// ProviderCreateInput 创建AI供应商表输入
type ProviderCreateInput struct {
	Name string `json:"name"`
	Code string `json:"code"`
	ProviderType string `json:"providerType"`
	BaseURL string `json:"baseURL"`
	Icon string `json:"icon"`
	Status int `json:"status"`
	Sort int `json:"sort"`
}

// ProviderUpdateInput 更新AI供应商表输入
type ProviderUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	ProviderType string `json:"providerType"`
	BaseURL string `json:"baseURL"`
	Icon string `json:"icon"`
	Status int `json:"status"`
	Sort int `json:"sort"`
}

// ProviderDetailOutput AI供应商表详情输出
type ProviderDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	ProviderType string `json:"providerType"`
	BaseURL string `json:"baseURL"`
	Icon string `json:"icon"`
	Status int `json:"status"`
	Sort int `json:"sort"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ProviderListOutput AI供应商表列表输出
type ProviderListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	ProviderType string `json:"providerType"`
	BaseURL string `json:"baseURL"`
	Icon string `json:"icon"`
	Status int `json:"status"`
	Sort int `json:"sort"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ProviderListInput AI供应商表列表查询输入
type ProviderListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Status *int `json:"status"`
	Name string `json:"name"`
}


// ProviderBatchUpdateInput 批量编辑AI供应商表输入
type ProviderBatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *int                  `json:"status"`
}

