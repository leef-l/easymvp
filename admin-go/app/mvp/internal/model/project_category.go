package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// ProjectCategory DTO 模型

// ProjectCategoryCreateInput 创建项目分类配置表输入
type ProjectCategoryCreateInput struct {
	CategoryCode string `json:"categoryCode"`
	DisplayName string `json:"displayName"`
	FamilyCode string `json:"familyCode"`
	Description string `json:"description"`
	Status int `json:"status"`
	Sort int `json:"sort"`
}

// ProjectCategoryUpdateInput 更新项目分类配置表输入
type ProjectCategoryUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	CategoryCode string `json:"categoryCode"`
	DisplayName string `json:"displayName"`
	FamilyCode string `json:"familyCode"`
	Description string `json:"description"`
	Status int `json:"status"`
	Sort int `json:"sort"`
}

// ProjectCategoryDetailOutput 项目分类配置表详情输出
type ProjectCategoryDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	CategoryCode string `json:"categoryCode"`
	DisplayName string `json:"displayName"`
	FamilyCode string `json:"familyCode"`
	Description string `json:"description"`
	Status int `json:"status"`
	Sort int `json:"sort"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ProjectCategoryListOutput 项目分类配置表列表输出
type ProjectCategoryListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	CategoryCode string `json:"categoryCode"`
	DisplayName string `json:"displayName"`
	FamilyCode string `json:"familyCode"`
	Description string `json:"description"`
	Status int `json:"status"`
	Sort int `json:"sort"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ProjectCategoryListInput 项目分类配置表列表查询输入
type ProjectCategoryListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	DisplayName string `json:"displayName"`
}


// ProjectCategoryBatchUpdateInput 批量编辑项目分类配置表输入
type ProjectCategoryBatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *int                  `json:"status"`
}

