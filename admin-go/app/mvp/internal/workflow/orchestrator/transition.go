package orchestrator

// StageType 阶段类型常量。
const (
	StageDesign   = "design"
	StageReview   = "review"
	StageExecute  = "execute"
	StageRework   = "rework"
	StageComplete = "complete"
)

// WorkflowStatus 工作流状态常量。
const (
	WorkflowPending   = "pending"
	WorkflowRunning   = "running"
	WorkflowPaused    = "paused"
	WorkflowCompleted = "completed"
	WorkflowCanceled  = "canceled"
)

// StageStatus 阶段状态常量。
const (
	StagePending   = "pending"
	StageRunning   = "running"
	StageCompleted = "completed"
	StageFailed    = "failed"
	StageSkipped   = "skipped"
)

// validWorkflowTransitions 合法的工作流状态迁移。
var validWorkflowTransitions = map[string][]string{
	WorkflowPending:   {WorkflowRunning, WorkflowCanceled},
	WorkflowRunning:   {WorkflowPaused, WorkflowCompleted, WorkflowCanceled},
	WorkflowPaused:    {WorkflowRunning, WorkflowCanceled},
	WorkflowCompleted: {},
	WorkflowCanceled:  {},
}

// validStageTransitions 合法的阶段状态迁移。
var validStageTransitions = map[string][]string{
	StagePending:   {StageRunning, StageSkipped},
	StageRunning:   {StageCompleted, StageFailed},
	StageCompleted: {},
	StageFailed:    {StagePending}, // 可重试
	StageSkipped:   {},
}

// stageOrder 阶段执行顺序。
var stageOrder = []string{
	StageDesign,
	StageReview,
	StageExecute,
	StageRework,
	StageComplete,
}

// IsValidWorkflowTransition 检查工作流状态迁移是否合法。
func IsValidWorkflowTransition(from, to string) bool {
	targets, ok := validWorkflowTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// IsValidStageTransition 检查阶段状态迁移是否合法。
func IsValidStageTransition(from, to string) bool {
	targets, ok := validStageTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// NextStage 返回当前阶段的下一个阶段，如果已是最后一个返回空。
func NextStage(current string) string {
	for i, s := range stageOrder {
		if s == current && i+1 < len(stageOrder) {
			return stageOrder[i+1]
		}
	}
	return ""
}
