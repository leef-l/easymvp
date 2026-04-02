package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Task DTO 模型

// TaskCreateInput 创建MVP任务表输入
type TaskCreateInput struct {
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ParentID snowflake.JsonInt64 `json:"parentID"`
	Name string `json:"name"`
	Description string `json:"description"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	Status string `json:"status"`
	Sort int `json:"sort"`
	BatchNo int `json:"batchNo"`
	AffectedResources string `json:"affectedResources"`
	DependsOn string `json:"dependsOn"`
	Result string `json:"result"`
	ContextSummary string `json:"contextSummary"`
	ErrorMessage string `json:"errorMessage"`
	StartedAt *gtime.Time `json:"startedAt"`
	CompletedAt *gtime.Time `json:"completedAt"`
}

// TaskUpdateInput 更新MVP任务表输入
type TaskUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ParentID snowflake.JsonInt64 `json:"parentID"`
	Name string `json:"name"`
	Description string `json:"description"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	Status string `json:"status"`
	Sort int `json:"sort"`
	BatchNo int `json:"batchNo"`
	AffectedResources string `json:"affectedResources"`
	DependsOn string `json:"dependsOn"`
	Result string `json:"result"`
	ContextSummary string `json:"contextSummary"`
	ErrorMessage string `json:"errorMessage"`
	StartedAt *gtime.Time `json:"startedAt"`
	CompletedAt *gtime.Time `json:"completedAt"`
}

// TaskDetailOutput MVP任务表详情输出
type TaskDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ProjectName string `json:"projectName"`
	ParentID snowflake.JsonInt64 `json:"parentID"`
	TaskName string `json:"taskName"`
	Name string `json:"name"`
	Description string `json:"description"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	Status string `json:"status"`
	Sort int `json:"sort"`
	BatchNo int `json:"batchNo"`
	AffectedResources string `json:"affectedResources"`
	DependsOn string `json:"dependsOn"`
	Result string `json:"result"`
	ContextSummary string `json:"contextSummary"`
	ErrorMessage string `json:"errorMessage"`
	StartedAt *gtime.Time `json:"startedAt"`
	CompletedAt *gtime.Time `json:"completedAt"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// TaskListOutput MVP任务表列表输出
type TaskListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ProjectName string `json:"projectName"`
	ParentID snowflake.JsonInt64 `json:"parentID"`
	TaskName string `json:"taskName"`
	Name string `json:"name"`
	Description string `json:"description"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	ConversationID snowflake.JsonInt64 `json:"conversationID"`
	Status string `json:"status"`
	Sort int `json:"sort"`
	BatchNo int `json:"batchNo"`
	AffectedResources string `json:"affectedResources"`
	DependsOn string `json:"dependsOn"`
	Result string `json:"result"`
	ContextSummary string `json:"contextSummary"`
	ErrorMessage string `json:"errorMessage"`
	StartedAt *gtime.Time `json:"startedAt"`
	CompletedAt *gtime.Time `json:"completedAt"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// TaskListInput MVP任务表列表查询输入
type TaskListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Name string `json:"name"`
}

// TaskTreeInput MVP任务表树形查询输入
type TaskTreeInput struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Name      string `json:"name"`
	ProjectID int64  `json:"projectID"` // 按项目筛选
	Status    string `json:"status"`    // 按状态筛选
	BatchNo   int    `json:"batchNo"`   // 按批次筛选
	RoleType  string `json:"roleType"`  // 按角色类型筛选
}

// TaskTreeOutput MVP任务表树形输出
type TaskTreeOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	ProjectID snowflake.JsonInt64 `json:"projectID"`
	ProjectName string `json:"projectName"`
	ParentID snowflake.JsonInt64 `json:"parentID"`
	TaskName string `json:"taskName"`
	Name string `json:"name"`
	Description string `json:"description"`
	RoleType string `json:"roleType"`
	RoleLevel string `json:"roleLevel"`
	ModelID snowflake.JsonInt64 `json:"modelID"`
	ConversationID snowflake.JsonInt64 `json:"conversationID"`
	Status string `json:"status"`
	Sort int `json:"sort"`
	BatchNo int `json:"batchNo"`
	AffectedResources string `json:"affectedResources"`
	DependsOn string `json:"dependsOn"`
	Result string `json:"result"`
	ContextSummary string `json:"contextSummary"`
	ErrorMessage string `json:"errorMessage"`
	StartedAt *gtime.Time `json:"startedAt"`
	CompletedAt *gtime.Time `json:"completedAt"`
	Children []*TaskTreeOutput `json:"children"`
}


// TaskBatchUpdateInput 批量编辑MVP任务表输入
type TaskBatchUpdateInput struct {
	IDs    []snowflake.JsonInt64 `json:"ids"`
	Status *int                  `json:"status"`
}

