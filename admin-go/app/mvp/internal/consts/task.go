package consts

// 任务状态
const (
	TaskStatusDraft         = "draft"          // 草稿（架构师拆分后未确认）
	TaskStatusPending       = "pending"        // 待执行
	TaskStatusRunning       = "running"        // 执行中
	TaskStatusCompleted     = "completed"      // 已完成
	TaskStatusFailed        = "failed"         // 执行失败
	TaskStatusEscalated     = "escalated"      // 已升级给架构师
	TaskStatusAuditing      = "auditing"       // 审计中
	TaskStatusBugFound      = "bug_found"      // 发现 Bug
	TaskStatusBugDispatched = "bug_dispatched"  // Bug 已派发修复
	TaskStatusSubmitError   = "submit_error"   // 提交异常（git 冲突等）
)

// 角色类型
const (
	RoleTypeArchitect    = "architect"    // 架构师
	RoleTypeImplementer  = "implementer"  // 实现者
	RoleTypeAuditor      = "auditor"      // 审计员
	RoleTypeOperator     = "operator"     // 运维/恢复
	RoleTypeCoordinator  = "coordinator"  // 协调者
)

// 角色等级
const (
	RoleLevelLite = "lite" // 轻量级（小任务）
	RoleLevelPro  = "pro"  // 专业级（中等任务）
	RoleLevelMax  = "max"  // 最高级（复杂任务）
)

// AllTaskStatuses 所有有效的任务状态
var AllTaskStatuses = []string{
	TaskStatusDraft,
	TaskStatusPending,
	TaskStatusRunning,
	TaskStatusCompleted,
	TaskStatusFailed,
	TaskStatusEscalated,
	TaskStatusAuditing,
	TaskStatusBugFound,
	TaskStatusBugDispatched,
	TaskStatusSubmitError,
}

// AllRoleTypes 所有角色类型
var AllRoleTypes = []string{
	RoleTypeArchitect,
	RoleTypeImplementer,
	RoleTypeAuditor,
	RoleTypeOperator,
	RoleTypeCoordinator,
}

// 任务记录类型（task_kind）
const (
	TaskKindImplement       = "implement"        // 原始实施任务
	TaskKindAudit           = "audit"             // 审计任务
	TaskKindBugAnalysis     = "bug_analysis"      // Bug 分析任务
	TaskKindFailureAnalysis = "failure_analysis"  // 失败分析任务
)

// AllRoleLevels 所有角色等级
var AllRoleLevels = []string{
	RoleLevelLite,
	RoleLevelPro,
	RoleLevelMax,
}

// 引擎版本
const (
	EngineVersionLegacy     = "legacy"
	EngineVersionWorkflowV2 = "workflow_v2"
)

// 工作流运行状态（阶段化语义）
const (
	WorkflowRunStatusDesigning = "designing"
	WorkflowRunStatusReviewing = "reviewing"
	WorkflowRunStatusExecuting = "executing"
	WorkflowRunStatusAccepting = "accepting"
	WorkflowRunStatusReworking = "reworking"
	WorkflowRunStatusPaused    = "paused"
	WorkflowRunStatusCompleted = "completed"
	WorkflowRunStatusFailed    = "failed"
	WorkflowRunStatusCanceled  = "canceled"
)

// 阶段类型
const (
	StageTypeDesign   = "design"
	StageTypeReview   = "review"
	StageTypeExecute  = "execute"
	StageTypeAccept   = "accept"
	StageTypeRework   = "rework"
	StageTypeComplete = "complete"
)

// 阶段状态
const (
	StageStatusPending   = "pending"
	StageStatusRunning   = "running"
	StageStatusCompleted = "completed"
	StageStatusFailed    = "failed"
	StageStatusSkipped   = "skipped"
)

// 计划版本状态
const (
	PlanVersionStatusDraft      = "draft"
	PlanVersionStatusActive     = "active"
	PlanVersionStatusSuperseded = "superseded"
)

// 计划版本审核状态
const (
	PlanReviewStatusPending  = "pending"
	PlanReviewStatusApproved = "approved"
	PlanReviewStatusRejected = "rejected"
)

// 蓝图状态
const (
	BlueprintStatusDraft      = "draft"
	BlueprintStatusConfirmed  = "confirmed"
	BlueprintStatusSuperseded = "superseded"
)
