package event

// 工作流事件类型常量。
const (
	// Workflow 级
	EventWorkflowCreated       = "workflow.created"
	EventWorkflowStatusChanged = "workflow.status_changed"
	EventWorkflowPaused        = "workflow.paused"
	EventWorkflowResumed       = "workflow.resumed"
	EventWorkflowCanceled      = "workflow.canceled"
	EventWorkflowCompleted     = "workflow.completed"

	// Stage 级
	EventStageCreated   = "stage.created"
	EventStageStarted   = "stage.started"
	EventStageCompleted = "stage.completed"
	EventStageFailed    = "stage.failed"
	EventStageSkipped   = "stage.skipped"

	// Plan 级
	EventPlanVersionCreated    = "plan_version.created"
	EventPlanVersionSubmitted  = "plan_version.submitted"
	EventPlanVersionApproved   = "plan_version.approved"
	EventPlanVersionRejected   = "plan_version.rejected"
	EventPlanVersionSuperseded = "plan_version.superseded"

	// Review 级
	EventReviewIssueCreated  = "review.issue_created"
	EventReviewIssueResolved = "review.issue_resolved"
	EventReviewDecisionReady = "review.decision_ready"

	// Task 级
	EventTaskCreated   = "task.created"
	EventTaskStarted   = "task.started"
	EventTaskCompleted = "task.completed"
	EventTaskFailed    = "task.failed"
	EventTaskEscalated = "task.escalated"
	EventTaskRetried   = "task.retried"
	EventTaskSkipped   = "task.skipped"

	// Resource 级
	EventResourceLocked     = "resource.locked"
	EventResourceReleased   = "resource.released"
	EventResourceLockLeaked = "resource.lock_leaked"

	// Autonomy 级
	EventAutonomyDecisionCreated   = "autonomy.decision_created"
	EventAutonomyActionExecuted    = "autonomy.action_executed"
	EventAutonomyActionFailed      = "autonomy.action_failed"
	EventAutonomyGateBlocked       = "autonomy.gate_blocked"
	EventAutonomyCheckpointOpened  = "autonomy.checkpoint_opened"
	EventAutonomyCheckpointHandled = "autonomy.checkpoint_handled"
)

// EntityType 实体类型常量。
const (
	EntityWorkflowRun = "workflow_run"
	EntityStageRun    = "stage_run"
	EntityPlanVersion = "plan_version"
	EntityDomainTask  = "domain_task"
	EntityReviewIssue     = "review_issue"
	EntityBlueprint       = "task_blueprint"
	EntityDecisionAction  = "decision_action"
	EntityHumanCheckpoint = "human_checkpoint"
)
