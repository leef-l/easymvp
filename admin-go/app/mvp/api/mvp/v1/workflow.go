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
	ProjectCategory  string              `json:"projectCategory" v:"max-length:50" dc:"项目分类（展示名，兼容旧版）"`
	CategoryCode     string              `json:"categoryCode" v:"max-length:64" dc:"项目分类编码（优先使用）"`
	Description      string              `json:"description" dc:"项目简介"`
	WorkDir          string              `json:"workDir" v:"max-length:500" dc:"代码工作目录（编码类项目必填，非编码类可留空由系统自动生成）"`
	ArchitectModelID snowflake.JsonInt64 `json:"architectModelID" dc:"架构师AI模型ID（可选，选择后会生成项目级架构师配置）"`
	EngineVersion    string              `json:"engineVersion" dc:"引擎版本: legacy(默认) / workflow_v2"`
	SelectedRoles    []SelectedRoleItem  `json:"selectedRoles" dc:"用户选择的角色预设ID列表；为空则不生成项目级角色配置，运行时直接走分类默认预设"`
}

// SelectedRoleItem 用户选择的角色预设
type SelectedRoleItem struct {
	PresetID snowflake.JsonInt64 `json:"presetID" dc:"角色预设ID"`
}

// WorkflowCreateProjectRes 创建项目响应
type WorkflowCreateProjectRes struct {
	g.Meta         `mime:"application/json"`
	ProjectID      snowflake.JsonInt64 `json:"projectID"`
	ConversationID snowflake.JsonInt64 `json:"conversationID"` // 架构师对话 ID
	WorkflowRunID  snowflake.JsonInt64 `json:"workflowRunID"`  // V2 工作流运行 ID（legacy 为 0）
}

// WorkflowConfirmPlanReq 确认实施方案请求
type WorkflowConfirmPlanReq struct {
	g.Meta    `path:"/workflow/confirm-plan" method:"post" tags:"项目流程" summary:"确认实施方案"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowConfirmPlanRes 确认实施方案响应
type WorkflowConfirmPlanRes struct {
	g.Meta       `mime:"application/json"`
	Submitted    bool              `json:"submitted"`              // 是否已成功提交审核
	ReviewPassed bool              `json:"reviewPassed"`           // 审核是否通过
	ReviewStatus string            `json:"reviewStatus,omitempty"` // pending/approved/rejected
	StageStatus  string            `json:"stageStatus,omitempty"`  // pending/running/completed/failed
	Message      string            `json:"message,omitempty"`      // 结果说明
	RejectReason string            `json:"rejectReason,omitempty"` // 驳回原因摘要
	Issues       []ReviewIssueItem `json:"issues,omitempty"`       // 审核问题列表
	ErrorCount   int               `json:"errorCount"`             // error 级别问题数
	WarningCount int               `json:"warningCount"`           // warning 级别问题数
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

// WorkflowResetToDesignReq 回到设计阶段请求（暂停状态下可用，清理已有方案和任务）
type WorkflowResetToDesignReq struct {
	g.Meta    `path:"/workflow/reset-to-design" method:"post" tags:"项目流程" summary:"回到设计阶段"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowResetToDesignRes 回到设计阶段响应
type WorkflowResetToDesignRes struct {
	g.Meta  `mime:"application/json"`
	Message string `json:"message"`
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
	HasTasks  bool   `json:"hasTasks"`          // AI回复中是否包含任务清单
	TaskCount int    `json:"taskCount"`         // 解析出的任务数量
	Message   string `json:"message,omitempty"` // 提示信息（如异步提取中）
}

// WorkflowRolePresetsReq 获取角色预设列表请求
type WorkflowRolePresetsReq struct {
	g.Meta          `path:"/workflow/role-presets" method:"get" tags:"项目流程" summary:"获取角色预设列表"`
	ProjectCategory string `json:"projectCategory" dc:"项目分类展示名（兼容旧版），为空则返回全部"`
	CategoryCode    string `json:"categoryCode" dc:"项目分类编码（优先使用）"`
	All             bool   `json:"all" dc:"是否返回全部预设（含扩展），默认只返回默认模板"`
}

// WorkflowRolePresetsRes 获取角色预设列表响应
type WorkflowRolePresetsRes struct {
	g.Meta `mime:"application/json"`
	List   []RolePresetItem `json:"list"`
}

// RolePresetItem 角色预设项
type RolePresetItem struct {
	ID            snowflake.JsonInt64 `json:"id"`
	RoleType      string              `json:"roleType"`
	RoleLevel     string              `json:"roleLevel"`
	ModelID       snowflake.JsonInt64 `json:"modelID"`
	ModelName     string              `json:"modelName"`
	ExecutionMode string              `json:"executionMode"`
	SystemPrompt  string              `json:"systemPrompt"`
	IsDefault     bool                `json:"isDefault"`
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
	Status  string `json:"status"` // ok / warning / error
	Message string `json:"message"`
	Link    string `json:"link"` // 前端跳转路径
}

// ==================== 项目分类 API ====================

// WorkflowCategoriesReq 获取项目分类列表请求
type WorkflowCategoriesReq struct {
	g.Meta `path:"/workflow/categories" method:"get" tags:"项目流程" summary:"获取项目分类列表"`
}

// WorkflowCategoriesRes 获取项目分类列表响应
type WorkflowCategoriesRes struct {
	g.Meta `mime:"application/json"`
	List   []CategoryItem `json:"list"`
}

// CategoryItem 分类列表项
type CategoryItem struct {
	CategoryCode string `json:"categoryCode"` // 稳定编码
	DisplayName  string `json:"displayName"`  // 展示名称
	FamilyCode   string `json:"familyCode"`   // 能力家族
	Description  string `json:"description"`  // 说明
}

// ==================== 审核相关 API ====================

// WorkflowReviewStatusReq 获取审核状态请求
type WorkflowReviewStatusReq struct {
	g.Meta    `path:"/workflow/review-status" method:"get" tags:"项目流程" summary:"获取审核状态"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowReviewStatusRes 获取审核状态响应
type WorkflowReviewStatusRes struct {
	g.Meta         `mime:"application/json"`
	PlanVersionID  snowflake.JsonInt64 `json:"planVersionID"`
	ReviewStatus   string              `json:"reviewStatus"` // pending/approved/rejected
	StageRunID     snowflake.JsonInt64 `json:"stageRunID"`
	StageStatus    string              `json:"stageStatus"` // pending/running/completed/failed
	StageTasks     []ReviewStageTask   `json:"stageTasks"`
	ErrorCount     int                 `json:"errorCount"`
	WarningCount   int                 `json:"warningCount"`
	BlueprintCount int                 `json:"blueprintCount"`
}

// ReviewStageTask 审核阶段子任务
type ReviewStageTask struct {
	ID           snowflake.JsonInt64 `json:"id"`
	TaskType     string              `json:"taskType"`
	RoleType     string              `json:"roleType"`
	Status       string              `json:"status"`
	StartedAt    *gtime.Time         `json:"startedAt,omitempty"`
	CompletedAt  *gtime.Time         `json:"completedAt,omitempty"`
	ErrorMessage string              `json:"errorMessage,omitempty"`
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
	ID         snowflake.JsonInt64 `json:"id"`
	Severity   string              `json:"severity"`
	IssueCode  string              `json:"issueCode"`
	SourceRole string              `json:"sourceRole"`
	TaskName   string              `json:"taskName"`
	Message    string              `json:"message"`
	Suggestion string              `json:"suggestion,omitempty"`
	Status     string              `json:"status"`
	CreatedAt  *gtime.Time         `json:"createdAt"`
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

// ==================== Timeline / Rework / Stage History ====================

// WorkflowTimelineReq 工作流时间线请求
type WorkflowTimelineReq struct {
	g.Meta    `path:"/workflow/timeline" method:"get" tags:"项目流程" summary:"工作流事件时间线"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	Limit     int                 `json:"limit" dc:"返回条数，默认50"`
}

// WorkflowTimelineRes 工作流时间线响应
type WorkflowTimelineRes struct {
	g.Meta `mime:"application/json"`
	Events []TimelineEvent `json:"events"`
}

// TimelineEvent 时间线事件
type TimelineEvent struct {
	ID            snowflake.JsonInt64  `json:"id"`
	WorkflowRunID snowflake.JsonInt64  `json:"workflowRunID"`
	StageRunID    *snowflake.JsonInt64 `json:"stageRunID,omitempty"`
	EntityType    string               `json:"entityType"`
	EntityID      *snowflake.JsonInt64 `json:"entityID,omitempty"`
	EventType     string               `json:"eventType"`
	Label         string               `json:"label"`
	Payload       string               `json:"payload,omitempty"`
	CreatedAt     *gtime.Time          `json:"createdAt"`
}

// WorkflowReworkStatusReq 返工状态请求
type WorkflowReworkStatusReq struct {
	g.Meta    `path:"/workflow/rework-status" method:"get" tags:"项目流程" summary:"返工阶段状态"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowReworkStatusRes 返工状态响应
type WorkflowReworkStatusRes struct {
	g.Meta       `mime:"application/json"`
	HasRework    bool              `json:"hasRework"`
	ReworkRounds int               `json:"reworkRounds"`
	CurrentStage *ReworkStageInfo  `json:"currentStage,omitempty"`
	History      []ReworkRoundInfo `json:"history"`
}

// ReworkStageInfo 返工阶段信息
type ReworkStageInfo struct {
	StageRunID snowflake.JsonInt64 `json:"stageRunID"`
	Status     string              `json:"status"`
	StartedAt  *gtime.Time         `json:"startedAt,omitempty"`
}

// ReworkRoundInfo 返工轮次信息
type ReworkRoundInfo struct {
	Round          int                  `json:"round"`
	FailedTaskID   snowflake.JsonInt64  `json:"failedTaskID"`
	FailedTaskName string               `json:"failedTaskName"`
	FailedReason   string               `json:"failedReason"`
	AnalysisTaskID *snowflake.JsonInt64 `json:"analysisTaskID,omitempty"`
	AnalysisResult string               `json:"analysisResult,omitempty"`
	HandoffType    string               `json:"handoffType"`
	CreatedAt      *gtime.Time          `json:"createdAt"`
}

// WorkflowStageHistoryReq 阶段历史请求
type WorkflowStageHistoryReq struct {
	g.Meta    `path:"/workflow/stage-history" method:"get" tags:"项目流程" summary:"工作流阶段历史"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowStageHistoryRes 阶段历史响应
type WorkflowStageHistoryRes struct {
	g.Meta `mime:"application/json"`
	Stages []StageHistoryItem `json:"stages"`
}

// StageHistoryItem 阶段历史项
type StageHistoryItem struct {
	ID         snowflake.JsonInt64 `json:"id"`
	StageType  string              `json:"stageType"`
	StageNo    int                 `json:"stageNo"`
	Status     string              `json:"status"`
	StartedAt  *gtime.Time         `json:"startedAt,omitempty"`
	FinishedAt *gtime.Time         `json:"finishedAt,omitempty"`
	Error      string              `json:"error,omitempty"`
}

// ==================== 完成总结 ====================

// WorkflowCompletionSummaryReq 获取完成总结请求
type WorkflowCompletionSummaryReq struct {
	g.Meta    `path:"/workflow/completion-summary" method:"get" tags:"项目流程" summary:"获取完成总结"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowCompletionSummaryRes 获取完成总结响应
type WorkflowCompletionSummaryRes struct {
	g.Meta          `mime:"application/json"`
	WorkflowRunID   snowflake.JsonInt64 `json:"workflowRunID"`
	ProjectID       snowflake.JsonInt64 `json:"projectID"`
	TotalTasks      int                 `json:"totalTasks"`
	CompletedTasks  int                 `json:"completedTasks"`
	FailedTasks     int                 `json:"failedTasks"`
	EscalatedTasks  int                 `json:"escalatedTasks"`
	SkippedTasks    int                 `json:"skippedTasks"`
	SuccessRate     float64             `json:"successRate"`
	TotalDuration   string              `json:"totalDuration"`
	AvgTaskDuration string              `json:"avgTaskDuration"`
	StageDurations  map[string]string   `json:"stageDurations"`
	ReworkRounds    int                 `json:"reworkRounds"`
	HandoffCount    int                 `json:"handoffCount"`
	StartedAt       string              `json:"startedAt"`
	FinishedAt      string              `json:"finishedAt"`
}

// ==================== 执行控制台 ====================

// WorkflowExecutionStatusReq 执行阶段实时状态请求
type WorkflowExecutionStatusReq struct {
	g.Meta    `path:"/workflow/execution-status" method:"get" tags:"项目流程" summary:"执行阶段实时状态"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowExecutionStatusRes 执行阶段实时状态响应
type WorkflowExecutionStatusRes struct {
	g.Meta         `mime:"application/json"`
	WorkflowRunID  snowflake.JsonInt64 `json:"workflowRunID"`
	StageRunID     snowflake.JsonInt64 `json:"stageRunID"`
	StageStatus    string              `json:"stageStatus"`
	ActiveBatch    int                 `json:"activeBatch"`
	TotalTasks     int                 `json:"totalTasks"`
	CompletedTasks int                 `json:"completedTasks"`
	RunningTasks   int                 `json:"runningTasks"`
	FailedTasks    int                 `json:"failedTasks"`
	PendingTasks   int                 `json:"pendingTasks"`
	EscalatedTasks int                 `json:"escalatedTasks"`
	Tasks          []DomainTaskItem    `json:"tasks"`
	ResourceLocks  []ResourceLockItem  `json:"resourceLocks"`
}

// DomainTaskItem 领域任务详情
type DomainTaskItem struct {
	ID                snowflake.JsonInt64 `json:"id"`
	Name              string              `json:"name"`
	Description       string              `json:"description,omitempty"`
	Status            string              `json:"status"`
	RoleType          string              `json:"roleType"`
	RoleLevel         string              `json:"roleLevel"`
	BatchNo           int                 `json:"batchNo"`
	Sort              int                 `json:"sort"`
	ExecutionMode     string              `json:"executionMode"`
	AffectedResources []string            `json:"affectedResources"`
	StartedAt         *gtime.Time         `json:"startedAt,omitempty"`
	CompletedAt       *gtime.Time         `json:"completedAt,omitempty"`
	ErrorMessage      string              `json:"errorMessage,omitempty"`
	Result            string              `json:"result,omitempty"`
	RetryCount        int                 `json:"retryCount"`
}

// ResourceLockItem 资源锁详情
type ResourceLockItem struct {
	Resource string              `json:"resource"`
	TaskID   snowflake.JsonInt64 `json:"taskID"`
	TaskName string              `json:"taskName"`
}

// WorkflowDomainTasksReq 领域任务列表请求
type WorkflowDomainTasksReq struct {
	g.Meta    `path:"/workflow/domain-tasks" method:"get" tags:"项目流程" summary:"领域任务列表"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	Status    string              `json:"status" dc:"状态筛选"`
	BatchNo   int                 `json:"batchNo" dc:"批次筛选"`
}

// WorkflowDomainTasksRes 领域任务列表响应
type WorkflowDomainTasksRes struct {
	g.Meta `mime:"application/json"`
	Tasks  []DomainTaskItem `json:"tasks"`
	Total  int              `json:"total"`
}

// WorkflowResourceLocksReq 资源锁列表请求
type WorkflowResourceLocksReq struct {
	g.Meta    `path:"/workflow/resource-locks" method:"get" tags:"项目流程" summary:"资源锁列表"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowResourceLocksRes 资源锁列表响应
type WorkflowResourceLocksRes struct {
	g.Meta `mime:"application/json"`
	Locks  []ResourceLockItem `json:"locks"`
}

// ==================== 验收控制台 API ====================

// WorkflowAcceptStatusReq 验收状态总览请求
type WorkflowAcceptStatusReq struct {
	g.Meta    `path:"/workflow/accept-status" method:"get" tags:"项目流程" summary:"验收状态总览"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowAcceptStatusRes 验收状态总览响应
type WorkflowAcceptStatusRes struct {
	g.Meta        `mime:"application/json"`
	AcceptRunID   snowflake.JsonInt64 `json:"acceptRunID"`
	WorkflowRunID snowflake.JsonInt64 `json:"workflowRunID"`
	AcceptRound   int                 `json:"acceptRound"`
	Status        string              `json:"status"`
	Decision      string              `json:"decision"`
	Score         float64             `json:"score"`
	Summary       string              `json:"summary"`
	RulesSnapshot string              `json:"rulesSnapshot,omitempty"`
	StartedAt     *gtime.Time         `json:"startedAt,omitempty"`
	FinishedAt    *gtime.Time         `json:"finishedAt,omitempty"`
	BlockerCount  int                 `json:"blockerCount"`
	ErrorCount    int                 `json:"errorCount"`
	WarnCount     int                 `json:"warnCount"`
	InfoCount     int                 `json:"infoCount"`
	EvidenceCount int                 `json:"evidenceCount"`
}

// WorkflowAcceptIssuesReq 验收问题列表请求
type WorkflowAcceptIssuesReq struct {
	g.Meta    `path:"/workflow/accept-issues" method:"get" tags:"项目流程" summary:"验收问题列表"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	Severity  string              `json:"severity" dc:"按严重级别过滤(blocker/error/warn/info)"`
}

// AcceptIssueItem 验收问题条目
type AcceptIssueItem struct {
	ID              snowflake.JsonInt64 `json:"id"`
	IssueType       string              `json:"issueType"`
	RuleCode        string              `json:"ruleCode"`
	Severity        string              `json:"severity"`
	Title           string              `json:"title"`
	Detail          string              `json:"detail"`
	ExpectedValue   string              `json:"expectedValue"`
	ActualValue     string              `json:"actualValue"`
	SuggestedAction string              `json:"suggestedAction"`
	DomainTaskID    snowflake.JsonInt64 `json:"domainTaskID,omitempty"`
	ResourceRef     string              `json:"resourceRef,omitempty"`
	Status          string              `json:"status"`
	CreatedAt       *gtime.Time         `json:"createdAt"`
}

// WorkflowAcceptIssuesRes 验收问题列表响应
type WorkflowAcceptIssuesRes struct {
	g.Meta `mime:"application/json"`
	Issues []AcceptIssueItem `json:"issues"`
}

// WorkflowAcceptEvidenceReq 验收证据列表请求
type WorkflowAcceptEvidenceReq struct {
	g.Meta    `path:"/workflow/accept-evidence" method:"get" tags:"项目流程" summary:"验收证据列表"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// AcceptEvidenceItem 验收证据条目
type AcceptEvidenceItem struct {
	ID           snowflake.JsonInt64 `json:"id"`
	EvidenceType string              `json:"evidenceType"`
	SourceType   string              `json:"sourceType"`
	SourceID     snowflake.JsonInt64 `json:"sourceID,omitempty"`
	ContentRef   string              `json:"contentRef,omitempty"`
	Summary      string              `json:"summary"`
	CreatedAt    *gtime.Time         `json:"createdAt"`
}

// WorkflowAcceptEvidenceRes 验收证据列表响应
type WorkflowAcceptEvidenceRes struct {
	g.Meta   `mime:"application/json"`
	Evidence []AcceptEvidenceItem `json:"evidence"`
}

// WorkflowAcceptApproveReq 人工放行请求
type WorkflowAcceptApproveReq struct {
	g.Meta    `path:"/workflow/accept-approve" method:"post" tags:"项目流程" summary:"人工放行验收"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	Reason    string              `json:"reason" dc:"放行原因"`
}

// WorkflowAcceptApproveRes 人工放行响应
type WorkflowAcceptApproveRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowAcceptRejectReq 驳回验收请求
type WorkflowAcceptRejectReq struct {
	g.Meta    `path:"/workflow/accept-reject" method:"post" tags:"项目流程" summary:"驳回验收"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	Reason    string              `json:"reason" v:"required" dc:"驳回原因"`
}

// WorkflowAcceptRejectRes 驳回验收响应
type WorkflowAcceptRejectRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowAcceptRerunReq 重新验收请求
type WorkflowAcceptRerunReq struct {
	g.Meta    `path:"/workflow/accept-rerun" method:"post" tags:"项目流程" summary:"重新验收"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowAcceptRerunRes 重新验收响应
type WorkflowAcceptRerunRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowAcceptReworkReq 驳回并返工请求
type WorkflowAcceptReworkReq struct {
	g.Meta    `path:"/workflow/accept-rework" method:"post" tags:"项目流程" summary:"驳回并返工"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	Reason    string              `json:"reason" v:"required" dc:"返工原因"`
}

// WorkflowAcceptReworkRes 驳回并返工响应
type WorkflowAcceptReworkRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 自治管理 API ====================

// WorkflowAutonomyDecisionsReq 自治决策列表请求
type WorkflowAutonomyDecisionsReq struct {
	g.Meta       `path:"/workflow/autonomy-decisions" method:"get" tags:"项目流程" summary:"自治决策列表"`
	ProjectID    snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	DecisionType string              `json:"decisionType" dc:"决策类型过滤"`
}

// AutonomyDecisionItem 自治决策条目
type AutonomyDecisionItem struct {
	ID             snowflake.JsonInt64 `json:"id"`
	DecisionType   string              `json:"decisionType"`
	TriggerSource  string              `json:"triggerSource"`
	TriggerContext string              `json:"triggerContext,omitempty"`
	Recommendation string              `json:"recommendation"`
	DecisionMode   string              `json:"decisionMode"`
	HumanAction    string              `json:"humanAction"`
	ExecutedAt     *gtime.Time         `json:"executedAt,omitempty"`
	Result         string              `json:"result,omitempty"`
	CreatedAt      *gtime.Time         `json:"createdAt"`
}

// WorkflowAutonomyDecisionsRes 自治决策列表响应
type WorkflowAutonomyDecisionsRes struct {
	g.Meta    `mime:"application/json"`
	Decisions []AutonomyDecisionItem `json:"decisions"`
}

// WorkflowApproveDecisionReq 批准自治决策请求
type WorkflowApproveDecisionReq struct {
	g.Meta     `path:"/workflow/approve-decision" method:"post" tags:"项目流程" summary:"批准自治决策"`
	ProjectID  snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	DecisionID snowflake.JsonInt64 `json:"decisionID" v:"required" dc:"决策ID"`
}

// WorkflowApproveDecisionRes 批准自治决策响应
type WorkflowApproveDecisionRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowRejectDecisionReq 拒绝自治决策请求
type WorkflowRejectDecisionReq struct {
	g.Meta     `path:"/workflow/reject-decision" method:"post" tags:"项目流程" summary:"拒绝自治决策"`
	ProjectID  snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	DecisionID snowflake.JsonInt64 `json:"decisionID" v:"required" dc:"决策ID"`
}

// WorkflowRejectDecisionRes 拒绝自治决策响应
type WorkflowRejectDecisionRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowTriggerReplanReq 手动触发重规划请求
type WorkflowTriggerReplanReq struct {
	g.Meta    `path:"/workflow/trigger-replan" method:"post" tags:"项目流程" summary:"手动触发重规划"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// WorkflowTriggerReplanRes 手动触发重规划响应
type WorkflowTriggerReplanRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowProjectReportsReq 项目报告列表请求
type WorkflowProjectReportsReq struct {
	g.Meta     `path:"/workflow/project-reports" method:"get" tags:"项目流程" summary:"项目报告列表"`
	ProjectID  snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	ReportType string              `json:"reportType" dc:"报告类型过滤(stage/daily/weekly/summary)"`
}

// ProjectReportItem 项目报告条目
type ProjectReportItem struct {
	ID         snowflake.JsonInt64 `json:"id"`
	ReportType string              `json:"reportType"`
	StageType  string              `json:"stageType,omitempty"`
	Title      string              `json:"title"`
	Content    string              `json:"content"`
	Metrics    string              `json:"metrics,omitempty"`
	CreatedAt  *gtime.Time         `json:"createdAt"`
}

// WorkflowProjectReportsRes 项目报告列表响应
type WorkflowProjectReportsRes struct {
	g.Meta  `mime:"application/json"`
	Reports []ProjectReportItem `json:"reports"`
}

// WorkflowTriggerReportReq 手动生成报告请求
type WorkflowTriggerReportReq struct {
	g.Meta    `path:"/workflow/trigger-report" method:"post" tags:"项目流程" summary:"手动生成报告"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	StageType string              `json:"stageType" dc:"阶段类型（不填则生成总结报告）"`
}

// WorkflowTriggerReportRes 手动生成报告响应
type WorkflowTriggerReportRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 自治模式配置 ====================

// WorkflowAutonomyModeReq 查询自治模式
type WorkflowAutonomyModeReq struct {
	g.Meta `path:"/workflow/autonomy-mode" method:"get" tags:"项目流程" summary:"查询自治模式"`
}

// WorkflowAutonomyModeRes 查询自治模式响应
type WorkflowAutonomyModeRes struct {
	g.Meta `mime:"application/json"`
	Mode   string `json:"mode" dc:"当前模式：suggest 或 auto"`
}

// WorkflowSetAutonomyModeReq 设置自治模式
type WorkflowSetAutonomyModeReq struct {
	g.Meta `path:"/workflow/set-autonomy-mode" method:"post" tags:"项目流程" summary:"设置自治模式"`
	Mode   string `json:"mode" v:"required|in:suggest,auto" dc:"suggest=建议型 auto=全自动"`
}

// WorkflowSetAutonomyModeRes 设置自治模式响应
type WorkflowSetAutonomyModeRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 项目列表批量状态 ====================

// ProjectRuntimeStat 单个项目的运行时统计。
type ProjectRuntimeStat struct {
	ProjectID      snowflake.JsonInt64 `json:"projectID"`
	CurrentStage   string              `json:"currentStage"`
	TotalTasks     int                 `json:"totalTasks"`
	CompletedTasks int                 `json:"completedTasks"`
	FailedTasks    int                 `json:"failedTasks"`
	RunningTasks   int                 `json:"runningTasks"`
}

// WorkflowBatchProjectStatsReq 批量查询项目运行时统计
type WorkflowBatchProjectStatsReq struct {
	g.Meta     `path:"/workflow/batch-project-stats" method:"post" tags:"项目流程" summary:"批量查询项目运行时统计"`
	ProjectIDs []snowflake.JsonInt64 `json:"projectIDs" v:"required" dc:"项目ID列表（最多50个）"`
}

// WorkflowBatchProjectStatsRes 批量查询项目运行时统计响应
type WorkflowBatchProjectStatsRes struct {
	g.Meta `mime:"application/json"`
	Stats  []ProjectRuntimeStat `json:"stats"`
}

// ==================== 自治中台 API ====================

// WorkflowAutonomyCheckpointsReq 查询项目待处理的人工节点
type WorkflowAutonomyCheckpointsReq struct {
	g.Meta    `path:"/workflow/autonomy-checkpoints" method:"get" tags:"自治中台" summary:"查询待处理人工节点"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// CheckpointDTO 人工节点 DTO（camelCase 输出给前端）
type CheckpointDTO struct {
	ID               snowflake.JsonInt64 `json:"id"`
	WorkflowRunID    snowflake.JsonInt64 `json:"workflowRunId"`
	ProjectID        snowflake.JsonInt64 `json:"projectId"`
	DecisionActionID snowflake.JsonInt64 `json:"decisionActionId"`
	CheckpointType   string              `json:"checkpointType"`
	Title            string              `json:"title"`
	Description      string              `json:"description,omitempty"`
	Status           string              `json:"status"`
	AssignedTo       snowflake.JsonInt64 `json:"assignedTo,omitempty"`
	HandledBy        snowflake.JsonInt64 `json:"handledBy,omitempty"`
	HandleAction     string              `json:"handleAction,omitempty"`
	HandleReason     string              `json:"handleReason,omitempty"`
	HandledAt        *gtime.Time         `json:"handledAt,omitempty"`
	ExpiresAt        *gtime.Time         `json:"expiresAt,omitempty"`
	CreatedAt        *gtime.Time         `json:"createdAt"`
}

// DecisionActionDTO 决策动作 DTO（camelCase 输出给前端）
type DecisionActionDTO struct {
	ID             snowflake.JsonInt64 `json:"id"`
	WorkflowRunID  snowflake.JsonInt64 `json:"workflowRunId"`
	ProjectID      snowflake.JsonInt64 `json:"projectId"`
	StageRunID     snowflake.JsonInt64 `json:"stageRunId,omitempty"`
	DomainTaskID   snowflake.JsonInt64 `json:"domainTaskId,omitempty"`
	DecisionType   string              `json:"decisionType"`
	DecisionLevel  string              `json:"decisionLevel"`
	TriggerSource  string              `json:"triggerSource"`
	TriggerContext string              `json:"triggerContext,omitempty"`
	MatchedRuleID  snowflake.JsonInt64 `json:"matchedRuleId,omitempty"`
	MatchedGateIDs string              `json:"matchedGateIds,omitempty"`
	ActionType     string              `json:"actionType"`
	Recommendation string              `json:"recommendation,omitempty"`
	FinalAction    string              `json:"finalAction,omitempty"`
	ActionStatus   string              `json:"actionStatus"`
	AutoExecutable int                 `json:"autoExecutable"`
	HumanRequired  int                 `json:"humanRequired"`
	ExecutedAt     *gtime.Time         `json:"executedAt,omitempty"`
	Result         string              `json:"result,omitempty"`
	CreatedAt      *gtime.Time         `json:"createdAt"`
}

// WorkflowAutonomyCheckpointsRes 查询待处理人工节点响应
type WorkflowAutonomyCheckpointsRes struct {
	g.Meta      `mime:"application/json"`
	Checkpoints []CheckpointDTO     `json:"checkpoints"`
	Actions     []DecisionActionDTO `json:"actions"`
}

// WorkflowAutonomyApproveReq 审批通过决策动作
type WorkflowAutonomyApproveReq struct {
	g.Meta   `path:"/workflow/autonomy-approve" method:"post" tags:"自治中台" summary:"审批通过"`
	ActionID snowflake.JsonInt64 `json:"actionID" v:"required" dc:"决策动作ID"`
}

// WorkflowAutonomyApproveRes 审批通过响应
type WorkflowAutonomyApproveRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowAutonomyRejectReq 驳回决策动作
type WorkflowAutonomyRejectReq struct {
	g.Meta   `path:"/workflow/autonomy-reject" method:"post" tags:"自治中台" summary:"驳回决策动作"`
	ActionID snowflake.JsonInt64 `json:"actionID" v:"required" dc:"决策动作ID"`
	Reason   string              `json:"reason" v:"required" dc:"驳回理由"`
}

// WorkflowAutonomyRejectRes 驳回决策动作响应
type WorkflowAutonomyRejectRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 自治中台：控制面查询 ====================

// WorkflowAutonomyActionsReq 查询项目全量决策记录
type WorkflowAutonomyActionsReq struct {
	g.Meta       `path:"/workflow/autonomy-actions" method:"get" tags:"自治中台" summary:"查询全量决策记录"`
	ProjectID    snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	ActionStatus string              `json:"actionStatus" dc:"状态过滤(可选)"`
	DecisionType string              `json:"decisionType" dc:"决策类型过滤(可选)"`
}

// WorkflowAutonomyActionsRes 全量决策记录响应
type WorkflowAutonomyActionsRes struct {
	g.Meta  `mime:"application/json"`
	Actions []DecisionActionDTO `json:"actions"`
}

// WorkflowAutonomyGateRulesReq 查询项目适用的风险闸门规则
type WorkflowAutonomyGateRulesReq struct {
	g.Meta    `path:"/workflow/autonomy-gate-rules" method:"get" tags:"自治中台" summary:"查询风险闸门规则"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// RiskGateRuleDTO 风险闸门规则 DTO
type RiskGateRuleDTO struct {
	ID                  snowflake.JsonInt64 `json:"id"`
	GateCode            string              `json:"gateCode"`
	GateName            string              `json:"gateName"`
	GateType            string              `json:"gateType"`
	ProjectFamily       string              `json:"projectFamily,omitempty"`
	ProjectCategoryCode string              `json:"projectCategoryCode,omitempty"`
	TriggerExpression   string              `json:"triggerExpression,omitempty"`
	BlockAction         string              `json:"blockAction"`
	FallbackAction      string              `json:"fallbackAction,omitempty"`
	Enabled             int                 `json:"enabled"`
	Priority            int                 `json:"priority"`
	CreatedAt           *gtime.Time         `json:"createdAt"`
}

// WorkflowAutonomyGateRulesRes 风险闸门规则响应
type WorkflowAutonomyGateRulesRes struct {
	g.Meta `mime:"application/json"`
	Rules  []RiskGateRuleDTO `json:"rules"`
}

// WorkflowAutonomyPolicyRulesReq 查询项目适用的策略规则
type WorkflowAutonomyPolicyRulesReq struct {
	g.Meta    `path:"/workflow/autonomy-policy-rules" method:"get" tags:"自治中台" summary:"查询策略规则"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

// PolicyRuleDTO 策略规则 DTO
type PolicyRuleDTO struct {
	ID                  snowflake.JsonInt64 `json:"id"`
	RuleCode            string              `json:"ruleCode"`
	RuleName            string              `json:"ruleName"`
	DecisionType        string              `json:"decisionType"`
	DecisionLevel       string              `json:"decisionLevel"`
	TriggerSource       string              `json:"triggerSource"`
	ProjectFamily       string              `json:"projectFamily,omitempty"`
	ProjectCategoryCode string              `json:"projectCategoryCode,omitempty"`
	ConfigJSON          string              `json:"configJson,omitempty"`
	Enabled             int                 `json:"enabled"`
	Priority            int                 `json:"priority"`
	CreatedAt           *gtime.Time         `json:"createdAt"`
}

// WorkflowAutonomyPolicyRulesRes 策略规则响应
type WorkflowAutonomyPolicyRulesRes struct {
	g.Meta `mime:"application/json"`
	Rules  []PolicyRuleDTO `json:"rules"`
}
