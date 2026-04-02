package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Plan DTO 模型

// PlanCreateInput 创建AI套餐表输入
type PlanCreateInput struct {
	ProviderID snowflake.JsonInt64 `json:"providerID"`
	Name string `json:"name"`
	Code string `json:"code"`
	ApiKey string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	Status int `json:"status"`
	Sort int `json:"sort"`
}

// PlanUpdateInput 更新AI套餐表输入
type PlanUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProviderID snowflake.JsonInt64 `json:"providerID"`
	Name string `json:"name"`
	Code string `json:"code"`
	ApiKey string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	Status int `json:"status"`
	Sort int `json:"sort"`
}

// PlanDetailOutput AI套餐表详情输出
type PlanDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProviderID snowflake.JsonInt64 `json:"providerID"`
	ProviderName string `json:"providerName"`
	Name string `json:"name"`
	Code string `json:"code"`
	ApiKey string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	Status int `json:"status"`
	Sort int `json:"sort"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// PlanListOutput AI套餐表列表输出
type PlanListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProviderID snowflake.JsonInt64 `json:"providerID"`
	ProviderName string `json:"providerName"`
	Name string `json:"name"`
	Code string `json:"code"`
	ApiKey string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	Status int `json:"status"`
	Sort int `json:"sort"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// PlanListInput AI套餐表列表查询输入
type PlanListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Status *int `json:"status"`
	Name string `json:"name"`
}


// PlanBatchUpdateInput 批量编辑AI套餐表输入
type PlanBatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *int                  `json:"status"`
}

