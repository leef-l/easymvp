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
	WorkDir          string              `json:"workDir" v:"required|max-length:500" dc:"代码工作目录（Aider执行路径）"`
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

// WorkflowSkipTaskReq 跳过失败任务请求
type WorkflowSkipTaskReq struct {
	g.Meta    `path:"/workflow/skip-task" method:"post" tags:"项目流程" summary:"跳过失败任务"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	TaskID    snowflake.JsonInt64 `json:"taskID" v:"required" dc:"任务ID"`
	Reason    string              `json:"reason" v:"required" dc:"跳过原因"`
}

// WorkflowSkipTaskRes 跳过失败任务响应
type WorkflowSkipTaskRes struct {
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
	ActiveBatch  int               `json:"activeBatch"`           // 当前活跃批次号
	TotalTasks   int               `json:"totalTasks"`
	StatusCounts map[string]int    `json:"statusCounts"` // 各状态任务数量
}

// WorkflowParseTasksReq 手动解析架构师回复中的任务清单
type WorkflowParseTasksReq struct {
	g.Meta    `path:"/workflow/parse-tasks" method:"post" tags:"项目流程" summary:"手动解析任务"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	DryRun    bool                `json:"dryRun" dc:"仅检查不创建"`
}

// WorkflowParseTasksRes 手动解析任务响应
type WorkflowParseTasksRes struct {
	g.Meta    `mime:"application/json"`
	HasTasks  bool `json:"hasTasks"`  // AI回复中是否包含任务清单
	TaskCount int  `json:"taskCount"` // 解析出的任务数量
}

// WorkflowRolePresetsReq 获取角色预设列表请求
type WorkflowRolePresetsReq struct {
	g.Meta `path:"/workflow/role-presets" method:"get" tags:"项目流程" summary:"获取角色预设列表"`
}

// WorkflowRolePresetsRes 获取角色预设列表响应
type WorkflowRolePresetsRes struct {
	g.Meta  `mime:"application/json"`
	List    []RolePresetItem `json:"list"`
}

// RolePresetItem 角色预设项
type RolePresetItem struct {
	RoleType     string             `json:"roleType"`
	RoleLevel    string             `json:"roleLevel"`
	ModelID      snowflake.JsonInt64 `json:"modelID"`
	ModelName    string             `json:"modelName"`
	SystemPrompt string             `json:"systemPrompt"`
}

// SystemCheckReq 系统配置检测请求
type SystemCheckReq struct {
	g.Meta `path:"/workflow/system-check" method:"get" tags:"项目流程" summary:"系统配置检测"`
}

// SystemCheckRes 系统配置检测响应
type SystemCheckRes struct {
	g.Meta  `mime:"application/json"`
	Items   []SystemCheckItem `json:"items"`
	AllPass bool              `json:"allPass"`
}

// SystemCheckItem 单项检测结果
type SystemCheckItem struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Status  string `json:"status"`  // ok / warning / error
	Message string `json:"message"`
	Link    string `json:"link"`    // 前端跳转路径
}
