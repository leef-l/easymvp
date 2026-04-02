package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Conversation DTO 模型

// ConversationCreateInput 创建MVP对话表输入
type ConversationCreateInput struct {
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	TaskID snowflake.JsonInt64 `json:"taskID"`
	Title string `json:"title"`
	RoleType string `json:"roleType"`
	Status string `json:"status"`
}

// ConversationUpdateInput 更新MVP对话表输入
type ConversationUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	TaskID snowflake.JsonInt64 `json:"taskID"`
	Title string `json:"title"`
	RoleType string `json:"roleType"`
	Status string `json:"status"`
}

// ConversationDetailOutput MVP对话表详情输出
type ConversationDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ProjectName string `json:"projectName"`
	TaskID snowflake.JsonInt64 `json:"taskID"`
	TaskName string `json:"taskName"`
	Title string `json:"title"`
	RoleType string `json:"roleType"`
	Status string `json:"status"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ConversationListOutput MVP对话表列表输出
type ConversationListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ProjectName string `json:"projectName"`
	TaskID snowflake.JsonInt64 `json:"taskID"`
	TaskName string `json:"taskName"`
	Title string `json:"title"`
	RoleType string `json:"roleType"`
	Status string `json:"status"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// ConversationListInput MVP对话表列表查询输入
type ConversationListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Title     string `json:"title"`
	ProjectID int64  `json:"projectID"` // 按项目筛选
	RoleType  string `json:"roleType"`  // 按角色类型筛选
}


// ConversationBatchUpdateInput 批量编辑MVP对话表输入
type ConversationBatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *int                  `json:"status"`
}

