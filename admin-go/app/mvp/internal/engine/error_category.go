package engine

import "strings"

type taskFailureCategory string

const (
	taskFailurePlanning    taskFailureCategory = "planning_error"
	taskFailureExecution   taskFailureCategory = "execution_error"
	taskFailurePolicyGuard taskFailureCategory = "policy_guard_error"
)

func (c taskFailureCategory) pauseReason() string {
	switch c {
	case taskFailurePlanning:
		return "任务定义存在问题，已转交架构师重新拆分"
	case taskFailurePolicyGuard:
		return "任务命中了安全/边界守卫，已转交架构师重新拆分"
	default:
		return ""
	}
}

func classifyTaskConfigError(err error) taskFailureCategory {
	if err == nil {
		return taskFailureExecution
	}

	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(err.Error(), "未配置"),
		strings.Contains(err.Error(), "不存在"),
		strings.Contains(lower, "not configured"),
		strings.Contains(lower, "not found"):
		return taskFailurePlanning
	default:
		return taskFailureExecution
	}
}
