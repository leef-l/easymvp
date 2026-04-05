package autonomy

import (
	"context"
	"fmt"

	"easymvp/app/mvp/internal/consts"
)

// ProactiveReplanStrategy 主动重规划策略。
//
// 在趋势恶化时主动触发重规划，不等人工干预：
//   - 失败率趋势上升 + 连续失败 ≥ 3：建议重规划
//   - 失败率趋势上升 + 批次进度 < 30%：早期阻断，建议重规划
//   - 已有多次重规划：升级为人工决策
type ProactiveReplanStrategy struct{}

func NewProactiveReplanStrategy() *ProactiveReplanStrategy {
	return &ProactiveReplanStrategy{}
}

func (s *ProactiveReplanStrategy) Name() string { return "proactive_replan" }

func (s *ProactiveReplanStrategy) Priority() int { return 70 }

func (s *ProactiveReplanStrategy) Applicable(sit *Situation, trigger string) bool {
	// 任何失败类触发 + 熔断触发都可能需要重规划
	switch trigger {
	case consts.TriggerTaskFailed,
		consts.TriggerTaskRetryExhausted,
		consts.TriggerAcceptFailed,
		consts.TriggerCircuitBreak:
		return true
	}
	return false
}

func (s *ProactiveReplanStrategy) Evaluate(ctx context.Context, sit *Situation, req *DecisionRequest) *ActionPlan {
	if sit.Health == nil || sit.Trend == nil {
		return nil
	}

	h := sit.Health
	t := sit.Trend

	// 失败率趋势未上升 → 不需要主动重规划
	if t.FailureRateTrend != "rising" {
		return nil
	}

	// 条件 1：连续失败 ≥ 3
	// 条件 2：批次早期（进度 < 30%）即出现上升趋势
	earlyBatch := sit.Progress != nil && sit.Progress.BatchProgress < 0.3
	consecutiveHigh := h.ConsecutiveFailures >= 3

	if !consecutiveHigh && !earlyBatch {
		return nil
	}

	// 已重规划多次 → 升级为人工
	level := consts.DecisionLevelB
	confidence := 0.75
	if h.ReplanCount >= 2 {
		level = consts.DecisionLevelC
		confidence = 0.6
	}

	reason := fmt.Sprintf("失败率趋势上升，连续失败 %d 次，已重规划 %d 次", h.ConsecutiveFailures, h.ReplanCount)
	if earlyBatch {
		reason = fmt.Sprintf("批次早期(进度 %.0f%%)即失败率上升，建议提前重规划", sit.Progress.BatchProgress*100)
	}

	return &ActionPlan{
		StrategyName:    "proactive_replan",
		Trigger:         req.TriggerSource,
		DecisionLevel:   level,
		ActionType:      consts.ActionTypeReplanWorkflow,
		TargetID:        req.WorkflowRunID,
		Reasoning:       reason,
		RollbackAction:  consts.ActionTypePauseWorkflow,
		ExpectedOutcome: "重规划后调整任务方案，改善失败率",
		Meta: &DecisionMeta{
			Confidence:          confidence,
			EvidenceSufficiency: 0.7,
			Reversibility:       "partial",
			BlastRadius:         "stage",
		},
	}
}
