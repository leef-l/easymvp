package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// Workflow 项目流程 API（手写，独立于 codegen 生成的 CRUD）

// WorkflowCreateProjectReq 创建项目请求
type WorkflowCreateProjectReq struct {
	g.Meta           `path:"/workflow/create-project" method:"post" tags:"项目流程" summary:"创建项目"`
	Name             string              `json:"name" v:"required" dc:"项目名称"`
	ProjectCategory  string              `json:"projectCategory" v:"required|max-length:50" dc:"项目分类"`
	Description      string              `json:"description" dc:"项目简介"`
	WorkDir          string              `json:"workDir" v:"max-length:500" dc:"代码工作目录（编码类项目必填，非编码类可留空由系统自动生成）"`
	ArchitectModelID snowflake.JsonInt64  `json:"architectModelID" v:"required" dc:"架构师AI模型ID"`
	EngineVersion    string               `json:"engineVersion" dc:"引擎版本: legacy(默认) / workflow_v2"`
}

// WorkflowCreateProjectRes 创建项目响应
type WorkflowCreateProjectRes struct {
	g.Meta         `mime:"application/json"`
	ProjectID      snowflake.JsonInt64 `json:"projectID"`
	ConversationID snowflake.JsonInt64 `json:"conversationID"`    // 架构师对话 ID
	WorkflowRunID  snowflake.JsonInt64 `json:"workflowRunID"`     // V2 工作流运行 ID（legacy 为 0）
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
	g.Meta             `mime:"application/json"`
	Status             string         `json:"status"`
	PauseReason        string         `json:"pauseReason,omitempty"`
	ActiveBatch        int            `json:"activeBatch"`
	TotalTasks         int            `json:"totalTasks"`
	StatusCounts       map[string]int `json:"statusCounts"`
	LastActiveAt       *gtime.Time    `json:"lastActiveAt,omitempty"`
	IsActuallyWorking  bool           `json:"isActuallyWorking"`
	ActiveRunningTasks int            `json:"activeRunningTasks"`
	StalledTaskCount   int            `json:"stalledTaskCount"`
	// V2 聚合字段
	EngineVersion   string `json:"engineVersion,omitempty"`
	WorkflowStatus  string `json:"workflowStatus,omitempty"`
	CurrentStage    string `json:"currentStage,omitempty"`
	ProgressPercent int    `json:"progressPercent,omitempty"`
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
	g.Meta          `path:"/workflow/role-presets" method:"get" tags:"项目流程" summary:"获取角色预设列表"`
	ProjectCategory string `json:"projectCategory" dc:"项目分类，为空则返回全部"`
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

// ==================== 审核相关 API ====================

// WorkflowReviewStatusReq 获取审核状态请求
type WorkflowReviewStatusReq struct {
	g.Meta    `path:"/workflow/review-status" method:"get" tags:"项目流程" summary:"获取审核状态"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowReviewStatusRes 获取审核状态响应
type WorkflowReviewStatusRes struct {
	g.Meta          `mime:"application/json"`
	PlanVersionID   snowflake.JsonInt64 `json:"planVersionID"`
	ReviewStatus    string              `json:"reviewStatus"`    // pending/approved/rejected
	StageRunID      snowflake.JsonInt64 `json:"stageRunID"`
	StageStatus     string              `json:"stageStatus"`     // pending/running/completed/failed
	StageTasks      []ReviewStageTask   `json:"stageTasks"`
	ErrorCount      int                 `json:"errorCount"`
	WarningCount    int                 `json:"warningCount"`
	BlueprintCount  int                 `json:"blueprintCount"`
}

// ReviewStageTask 审核阶段子任务
type ReviewStageTask struct {
	ID        snowflake.JsonInt64 `json:"id"`
	TaskType  string              `json:"taskType"`
	RoleType  string              `json:"roleType"`
	Status    string              `json:"status"`
	StartedAt *gtime.Time         `json:"startedAt,omitempty"`
	CompletedAt *gtime.Time       `json:"completedAt,omitempty"`
	ErrorMessage string           `json:"errorMessage,omitempty"`
}

// WorkflowReviewIssuesReq 获取审核问题列表请求
type WorkflowReviewIssuesReq struct {
	g.Meta    `path:"/workflow/review-issues" method:"get" tags:"项目流程" summary:"获取审核问题列表"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowReviewIssuesRes 获取审核问题列表响应
type WorkflowReviewIssuesRes struct {
	g.Meta `mime:"application/json"`
	Issues []ReviewIssueItem `json:"issues"`
}

// ReviewIssueItem 审核问题项
type ReviewIssueItem struct {
	ID          snowflake.JsonInt64 `json:"id"`
	Severity    string              `json:"severity"`
	IssueCode   string              `json:"issueCode"`
	SourceRole  string              `json:"sourceRole"`
	TaskName    string              `json:"taskName"`
	Message     string              `json:"message"`
	Suggestion  string              `json:"suggestion,omitempty"`
	Status      string              `json:"status"`
	CreatedAt   *gtime.Time         `json:"createdAt"`
}

// WorkflowManualApproveReq 手动审批请求（跳过 AI 审核，人工通过）
type WorkflowManualApproveReq struct {
	g.Meta    `path:"/workflow/manual-approve" method:"post" tags:"项目流程" summary:"手动审批通过"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowManualApproveRes 手动审批响应
type WorkflowManualApproveRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowManualRejectReq 手动驳回请求
type WorkflowManualRejectReq struct {
	g.Meta    `path:"/workflow/manual-reject" method:"post" tags:"项目流程" summary:"手动驳回"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	Reason    string              `json:"reason" v:"required" dc:"驳回原因"`
}

// WorkflowManualRejectRes 手动驳回响应
type WorkflowManualRejectRes struct {
	g.Meta `mime:"application/json"`
}
