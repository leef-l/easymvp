package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Config DTO 模型

// ConfigCreateInput 创建MVP配置表输入
type ConfigCreateInput struct {
	ConfigKey string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	ConfigType string `json:"configType"`
	Category string `json:"category"`
	Description string `json:"description"`
}

// ConfigUpdateInput 更新MVP配置表输入
type ConfigUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ConfigKey string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	ConfigType string `json:"configType"`
	Category string `json:"category"`
	Description string `json:"description"`
}

// ConfigDetailOutput MVP配置表详情输出
type ConfigDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ConfigKey string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	ConfigType string `json:"configType"`
	Category string `json:"category"`
	Description string `json:"description"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ConfigListOutput MVP配置表列表输出
type ConfigListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ConfigKey string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	ConfigType string `json:"configType"`
	Category string `json:"category"`
	Description string `json:"description"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ConfigListInput MVP配置表列表查询输入
type ConfigListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}


