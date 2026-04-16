package autonomy

import (
	"context"
	"testing"

	"easymvp/app/mvp/internal/consts"
)

func TestResolveScopedBudgetMetricsUsesTaskScope(t *testing.T) {
	t.Parallel()

	metrics := resolveScopedBudgetMetrics(&Situation{
		Health: &HealthMetrics{
			RetryCount:       5,
			TaskRetryCount:   1,
			ReworkRounds:     4,
			TaskReworkRounds: 0,
			FocusedTaskID:    42,
		},
	}, &DecisionRequest{
		DomainTaskID:  42,
		TriggerSource: consts.TriggerTaskFailed,
	})

	if metrics.scope != "task" {
		t.Fatalf("expected task scope, got %q", metrics.scope)
	}
	if metrics.retryCount != 1 || metrics.reworkRounds != 0 {
		t.Fatalf("unexpected scoped metrics: %+v", metrics)
	}
}

func TestResolveScopedBudgetMetricsFallsBackToWorkflowScope(t *testing.T) {
	t.Parallel()

	metrics := resolveScopedBudgetMetrics(&Situation{
		Health: &HealthMetrics{
			RetryCount:       5,
			TaskRetryCount:   1,
			ReworkRounds:     4,
			TaskReworkRounds: 0,
			FocusedTaskID:    0,
		},
	}, &DecisionRequest{
		DomainTaskID:  42,
		TriggerSource: consts.TriggerTaskFailed,
	})

	if metrics.scope != "workflow" {
		t.Fatalf("expected workflow scope fallback, got %q", metrics.scope)
	}
	if metrics.retryCount != 5 || metrics.reworkRounds != 4 {
		t.Fatalf("unexpected fallback metrics: %+v", metrics)
	}
}

func TestObjectiveCheckUsesTaskScopedReworkBudget(t *testing.T) {
	t.Parallel()

	svc := NewObjectiveService()
	admission, err := svc.Check(context.Background(), &Situation{
		Health: &HealthMetrics{
			ReworkRounds:     3,
			TaskReworkRounds: 0,
			FocusedTaskID:    100,
		},
	}, &ProjectObjective{MaxAutoReworks: 2}, &DecisionRequest{
		DomainTaskID:  100,
		TriggerSource: consts.TriggerTaskFailed,
	})
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if !admission.Allowed {
		t.Fatalf("expected task-scoped first failure to stay allowed, got %+v", admission)
	}
}

func TestAdaptiveRetryUsesTaskScopedBudgetForNewTask(t *testing.T) {
	t.Parallel()

	plan := NewAdaptiveRetryStrategy().Evaluate(context.Background(), &Situation{
		Health: &HealthMetrics{
			ConsecutiveFailures: 1,
			RetryCount:          4,
			TaskRetryCount:      0,
			ReworkRounds:        3,
			TaskReworkRounds:    0,
			FocusedTaskID:       7,
		},
		Trend: &TrendMetrics{FailureRateTrend: "stable"},
	}, &DecisionRequest{
		DomainTaskID:  7,
		TriggerSource: consts.TriggerTaskFailed,
	})

	if plan == nil {
		t.Fatal("expected adaptive retry plan")
	}
	if plan.ActionType != consts.ActionTypeRetryTask {
		t.Fatalf("expected retry action for new task, got %+v", plan)
	}
}
