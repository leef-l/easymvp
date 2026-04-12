// Package consts 自治中台相关常量。
package consts

// ==================== 决策等级 ====================

const (
	DecisionLevelA = "A" // 系统自动执行
	DecisionLevelB = "B" // 系统建议，人确认
	DecisionLevelC = "C" // 必须人工决定
)

// ==================== 决策动作类型 ====================

const (
	ActionTypeRetryTask       = "retry_task"
	ActionTypeTriggerRework   = "trigger_rework"
	ActionTypeRerunAccept     = "rerun_accept"
	ActionTypePauseWorkflow   = "pause_workflow"
	ActionTypeSwitchExecutor  = "switch_executor"
	ActionTypeReplanWorkflow  = "replan_workflow"
	ActionTypeNotifyHuman     = "notify_human"
	ActionTypeApproveComplete = "approve_complete"
)

// ==================== 触发源（输入事件）====================

const (
	TriggerTaskFailed         = "task.failed"
	TriggerTaskTimeout        = "task.timeout"
	TriggerTaskRetryExhausted = "task.retry_exhausted"
	TriggerAcceptPassed       = "accept.passed"
	TriggerAcceptFailed       = "accept.failed"
	TriggerAcceptManualReview = "accept.manual_review"
	TriggerCircuitBreak       = "workflow.circuit_break"
	TriggerReworkCompleted    = "rework.completed"
	TriggerReplanSuggested    = "replan.suggested"
	TriggerHumanOverride      = "human.override"
	TriggerReviewCompleted    = "review.completed"
	TriggerExecuteCompleted   = "execute.completed"
)

// ==================== 决策动作状态 ====================

const (
	ActionStatusPending      = "pending"
	ActionStatusAutoExecuted = "auto_executed"
	ActionStatusWaitingHuman = "waiting_human"
	ActionStatusApproved     = "approved"
	ActionStatusRejected     = "rejected"
	ActionStatusFailed       = "failed"
)

// ==================== 风险闸门类型 ====================

const (
	GateTypePermission = "permission"
	GateTypeQuality    = "quality"
	GateTypeCost       = "cost"
	GateTypeRuntime    = "runtime"
)

// ==================== 人工节点类型 ====================

const (
	CheckpointManualReview = "manual_review"
	CheckpointApproval     = "approval"
	CheckpointEscalation   = "escalation"
)

// ==================== 人工节点状态 ====================

const (
	CheckpointStatusOpen     = "open"
	CheckpointStatusHandled  = "handled"
	CheckpointStatusExpired  = "expired"
	CheckpointStatusCanceled = "canceled"
)

// ==================== 人工处理动作 ====================

const (
	HandleActionApprove  = "approve"
	HandleActionReject   = "reject"
	HandleActionRetry    = "retry"
	HandleActionRework   = "rework"
	HandleActionOverride = "override"
)
