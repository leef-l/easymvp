package model

import (
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// TaskLog DTO 模型

// TaskLogCreateInput 创建任务日志表输入
type TaskLogCreateInput struct {
	TaskID snowflake.JsonInt64 `json:"taskID"`
	Action string `json:"action"`
	FromStatus string `json:"fromStatus"`
	ToStatus string `json:"toStatus"`
	Message string `json:"message"`
	Operator string `json:"operator"`
}

// TaskLogUpdateInput 更新任务日志表输入
type TaskLogUpdateInput struct {
	ID snowflake.JsonInt64 `json:"id"`
	TaskID snowflake.JsonInt64 `json:"taskID"`
	Action string `json:"action"`
	FromStatus string `json:"fromStatus"`
	ToStatus string `json:"toStatus"`
	Message string `json:"message"`
	Operator string `json:"operator"`
}

// TaskLogDetailOutput 任务日志表详情输出
type TaskLogDetailOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	TaskID snowflake.JsonInt64 `json:"taskID"`
	TaskName string `json:"taskName"`
	Action string `json:"action"`
	FromStatus string `json:"fromStatus"`
	ToStatus string `json:"toStatus"`
	Message string `json:"message"`
	Operator string `json:"operator"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// TaskLogListOutput 任务日志表列表输出
type TaskLogListOutput struct {
	ID snowflake.JsonInt64 `json:"id"`
	TaskID snowflake.JsonInt64 `json:"taskID"`
	TaskName string `json:"taskName"`
	Action string `json:"action"`
	FromStatus string `json:"fromStatus"`
	ToStatus string `json:"toStatus"`
	Message string `json:"message"`
	Operator string `json:"operator"`
	CreatedAt *gtime.Time `json:"createdAt"`
	UpdatedAt *gtime.Time `json:"updatedAt"`
}

// TaskLogListInput 任务日志表列表查询输入
type TaskLogListInput struct {
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	OrderDir  string `json:"orderDir"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}


