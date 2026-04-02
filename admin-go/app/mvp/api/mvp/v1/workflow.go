package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// Workflow 项目流程 API（手写，独立于 codegen 生成的 CRUD）

// WorkflowCreateProjectReq 创建项目请求
type WorkflowCreateProjectReq struct {
	g.Meta           `path:"/workflow/create-project" method:"post" tags:"项目流程" summary:"创建项目"`
	Name             string              `json:"name" v:"required" dc:"项目名称"`
	Description      string              `json:"description" dc:"项目简介"`
	ArchitectModelID snowflake.JsonInt64  `json:"architectModelID" v:"required" dc:"架构师AI模型ID"`
}

// WorkflowCreateProjectRes 创建项目响应
type WorkflowCreateProjectRes struct {
	g.Meta         `mime:"application/json"`
	ProjectID      snowflake.JsonInt64 `json:"projectID"`
	ConversationID snowflake.JsonInt64 `json:"conversationID"` // 架构师对话 ID
}

// WorkflowConfirmPlanReq 确认实施方案请求
type WorkflowConfirmPlanReq struct {
	g.Meta    `path:"/workflow/confirm-plan" method:"post" tags:"项目流程" summary:"确认实施方案"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowConfirmPlanRes 确认实施方案响应
type WorkflowConfirmPlanRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowPauseReq 暂停项目请求
type WorkflowPauseReq struct {
	g.Meta      `path:"/workflow/pause" method:"post" tags:"项目流程" summary:"暂停项目"`
	ProjectID   snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	PauseReason string              `json:"pauseReason" v:"required" dc:"暂停原因"`
}

// WorkflowPauseRes 暂停项目响应
type WorkflowPauseRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowResumeReq 恢复项目请求
type WorkflowResumeReq struct {
	g.Meta    `path:"/workflow/resume" method:"post" tags:"项目流程" summary:"恢复项目"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowResumeRes 恢复项目响应
type WorkflowResumeRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowRetryTaskReq 重新执行失败任务请求
type WorkflowRetryTaskReq struct {
	g.Meta    `path:"/workflow/retry-task" method:"post" tags:"项目流程" summary:"重新执行失败任务"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	TaskID    snowflake.JsonInt64 `json:"taskID" v:"required" dc:"任务ID"`
}

// WorkflowRetryTaskRes 重新执行失败任务响应
type WorkflowRetryTaskRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowProjectStatusReq 获取项目状态请求
type WorkflowProjectStatusReq struct {
	g.Meta    `path:"/workflow/project-status" method:"get" tags:"项目流程" summary:"获取项目状态"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowProjectStatusRes 获取项目状态响应
type WorkflowProjectStatusRes struct {
	g.Meta       `mime:"application/json"`
	Status       string            `json:"status"`
	PauseReason  string            `json:"pauseReason,omitempty"` // 暂停原因
	TotalTasks   int               `json:"totalTasks"`
	StatusCounts map[string]int    `json:"statusCounts"` // 各状态任务数量
}
