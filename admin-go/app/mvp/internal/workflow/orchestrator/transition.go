package orchestrator

// StageType 阶段类型常量。
const (
	StageDesign   = "design"
	StageReview   = "review"
	StageExecute  = "execute"
	StageRework   = "rework"
	StageComplete = "complete"
)

// WorkflowStatus 工作流状态常量（阶段化语义）。
const (
	WorkflowDesigning = "designing"
	WorkflowReviewing = "reviewing"
	WorkflowExecuting = "executing"
	WorkflowReworking = "reworking"
	WorkflowPaused    = "paused"
	WorkflowCompleted = "completed"
	WorkflowFailed    = "failed"
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
	WorkflowDesigning: {WorkflowReviewing, WorkflowPaused, WorkflowCanceled},
	WorkflowReviewing: {WorkflowExecuting, WorkflowDesigning, WorkflowPaused, WorkflowCanceled}, // 驳回可回 designing
	WorkflowExecuting: {WorkflowCompleted, WorkflowReworking, WorkflowPaused, WorkflowFailed, WorkflowCanceled},
	WorkflowReworking: {WorkflowReviewing, WorkflowPaused, WorkflowCanceled},
	WorkflowPaused:    {WorkflowDesigning, WorkflowReviewing, WorkflowExecuting, WorkflowReworking, WorkflowCanceled}, // 恢复到暂停前的阶段
	WorkflowCompleted: {},
	WorkflowFailed:    {WorkflowReworking, WorkflowCanceled}, // 失败后可返工或取消
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

// stageOrder 主链阶段执行顺序。
// rework 不在主链中——它是旁路分支，仅由 triggerReworkStage 在失败/升级场景中单独触发。
var stageOrder = []string{
	StageDesign,
	StageReview,
	StageExecute,
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

// StageTypeToWorkflowStatus 将阶段类型映射为对应的 workflow 状态。
func StageTypeToWorkflowStatus(stageType string) string {
	switch stageType {
	case StageDesign:
		return WorkflowDesigning
	case StageReview:
		return WorkflowReviewing
	case StageExecute:
		return WorkflowExecuting
	case StageRework:
		return WorkflowReworking
	case StageComplete:
		return WorkflowCompleted
	default:
		return ""
	}
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
