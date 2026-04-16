package autonomy

import "easymvp/app/mvp/internal/consts"

type scopedBudgetMetrics struct {
	retryCount   int
	reworkRounds int
	scope        string
}

func resolveScopedBudgetMetrics(sit *Situation, req *DecisionRequest) scopedBudgetMetrics {
	metrics := scopedBudgetMetrics{scope: "workflow"}
	if sit == nil || sit.Health == nil {
		return metrics
	}

	metrics.retryCount = sit.Health.RetryCount
	metrics.reworkRounds = sit.Health.ReworkRounds

	if shouldUseTaskScopedBudget(req) && sit.Health.FocusedTaskID == req.DomainTaskID {
		metrics.retryCount = sit.Health.TaskRetryCount
		metrics.reworkRounds = sit.Health.TaskReworkRounds
		metrics.scope = "task"
	}

	return metrics
}

func shouldUseTaskScopedBudget(req *DecisionRequest) bool {
	if req == nil || req.DomainTaskID == 0 {
		return false
	}
	switch req.TriggerSource {
	case consts.TriggerTaskFailed,
		consts.TriggerTaskTimeout,
		consts.TriggerTaskRetryExhausted,
		consts.TriggerAcceptFailed:
		return true
	default:
		return false
	}
}
