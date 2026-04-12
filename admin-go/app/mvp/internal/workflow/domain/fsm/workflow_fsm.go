package fsm

// WorkflowStatus 工作流运行状态（强类型枚举）。
type WorkflowStatus string

const (
	WorkflowDesigning WorkflowStatus = "designing"
	WorkflowReviewing WorkflowStatus = "reviewing"
	WorkflowExecuting WorkflowStatus = "executing"
	WorkflowAccepting WorkflowStatus = "accepting"
	WorkflowReworking WorkflowStatus = "reworking"
	WorkflowPaused    WorkflowStatus = "paused"
	WorkflowCompleted WorkflowStatus = "completed"
	WorkflowFailed    WorkflowStatus = "failed"
	WorkflowCanceled  WorkflowStatus = "canceled"
)

// String 返回状态的字符串值，用于与现有字符串类型代码互操作。
func (s WorkflowStatus) String() string { return string(s) }

// ParseWorkflowStatus 将字符串解析为 WorkflowStatus，未知值返回空字符串状态。
func ParseWorkflowStatus(s string) WorkflowStatus {
	switch WorkflowStatus(s) {
	case WorkflowDesigning, WorkflowReviewing, WorkflowExecuting,
		WorkflowAccepting, WorkflowReworking, WorkflowPaused,
		WorkflowCompleted, WorkflowFailed, WorkflowCanceled:
		return WorkflowStatus(s)
	}
	return ""
}

// workflowTransitions 合法的工作流状态迁移矩阵。
var workflowTransitions = map[WorkflowStatus][]WorkflowStatus{
	WorkflowDesigning: {WorkflowReviewing, WorkflowPaused, WorkflowCanceled},
	WorkflowReviewing: {WorkflowExecuting, WorkflowDesigning, WorkflowPaused, WorkflowCanceled, WorkflowFailed},
	WorkflowExecuting: {WorkflowAccepting, WorkflowCompleted, WorkflowReworking, WorkflowPaused, WorkflowFailed, WorkflowCanceled},
	WorkflowAccepting: {WorkflowCompleted, WorkflowReworking, WorkflowFailed, WorkflowPaused, WorkflowCanceled},
	WorkflowReworking: {WorkflowExecuting, WorkflowAccepting, WorkflowReviewing, WorkflowPaused, WorkflowCanceled},
	WorkflowPaused:    {WorkflowDesigning, WorkflowReviewing, WorkflowExecuting, WorkflowAccepting, WorkflowReworking, WorkflowCanceled},
	WorkflowCompleted: {},
	WorkflowFailed:    {WorkflowDesigning, WorkflowReworking, WorkflowCanceled},
	WorkflowCanceled:  {},
}

// WorkflowInitial 返回工作流初始状态。
func WorkflowInitial() WorkflowStatus { return WorkflowDesigning }

// IsValidWorkflowTransition 检查工作流状态迁移是否合法。
func IsValidWorkflowTransition(from, to WorkflowStatus) bool {
	targets, ok := workflowTransitions[from]
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

// WorkflowTargets 返回给定状态的合法目标状态列表。
func WorkflowTargets(from WorkflowStatus) []WorkflowStatus {
	targets := workflowTransitions[from]
	if targets == nil {
		return []WorkflowStatus{}
	}
	result := make([]WorkflowStatus, len(targets))
	copy(result, targets)
	return result
}

// IsTerminal 检查工作流状态是否为终态（completed/failed/canceled）。
func IsTerminal(s WorkflowStatus) bool {
	return s == WorkflowCompleted || s == WorkflowFailed || s == WorkflowCanceled
}

// StageCodeToWorkflowStatus 将阶段代码映射为对应的 workflow 状态，
// 与 orchestrator/transition.go 的 StageTypeToWorkflowStatus 对齐。
func StageCodeToWorkflowStatus(stageCode string) WorkflowStatus {
	switch stageCode {
	case "design":
		return WorkflowDesigning
	case "review":
		return WorkflowReviewing
	case "execute":
		return WorkflowExecuting
	case "accept":
		return WorkflowAccepting
	case "rework":
		return WorkflowReworking
	case "complete":
		return WorkflowCompleted
	default:
		return ""
	}
}
