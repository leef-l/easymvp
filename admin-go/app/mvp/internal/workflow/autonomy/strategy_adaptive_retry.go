package autonomy

import (
	"context"
	"fmt"

	"easymvp/app/mvp/internal/consts"
)

// AdaptiveRetryStrategy 自适应重试策略。
//
// 根据错误类型（瞬态/结构性/致命）动态调整重试行为：
//   - 瞬态错误（网络超时、API 限流）：立即重试，最多 3 次
//   - 结构性错误（方案缺陷、依赖缺失）：转入返工，不重试
//   - 致命错误（环境损坏）：暂停工作流，通知人工
type AdaptiveRetryStrategy struct{}

func NewAdaptiveRetryStrategy() *AdaptiveRetryStrategy {
	return &AdaptiveRetryStrategy{}
}

func (s *AdaptiveRetryStrategy) Name() string { return "adaptive_retry" }

func (s *AdaptiveRetryStrategy) Priority() int { return 90 }

func (s *AdaptiveRetryStrategy) Applicable(sit *Situation, trigger string) bool {
	switch trigger {
	case consts.TriggerTaskFailed, consts.TriggerTaskTimeout, consts.TriggerTaskRetryExhausted:
		return true
	}
	return false
}

func (s *AdaptiveRetryStrategy) Evaluate(ctx context.Context, sit *Situation, req *DecisionRequest) *ActionPlan {
	if sit.Health == nil {
		return nil
	}

	h := sit.Health
	riskLevel := s.classifyRisk(sit, req)

	switch riskLevel {
	case RiskTransient:
		return s.handleTransient(h, req)
	case RiskStructural:
		return s.handleStructural(h, req)
	case RiskFatal:
		return s.handleFatal(req)
	default:
		return s.handleTransient(h, req)
	}
}

// classifyRisk 根据态势信号推断错误类型。
func (s *AdaptiveRetryStrategy) classifyRisk(sit *Situation, req *DecisionRequest) string {
	// 存在 critical 异常信号 → 致命
	if sit.HasCriticalAnomaly() {
		return RiskFatal
	}

	h := sit.Health
	// 连续失败 ≥ 3 且失败率持续上升 → 结构性
	if h.ConsecutiveFailures >= 3 && sit.Trend != nil && sit.Trend.FailureRateTrend == "rising" {
		return RiskStructural
	}

	// 重试已用尽 → 结构性
	if req.TriggerSource == consts.TriggerTaskRetryExhausted {
		return RiskStructural
	}

	// 返工轮数 ≥ 2 → 结构性
	if h.ReworkRounds >= 2 {
		return RiskStructural
	}

	return RiskTransient
}

func (s *AdaptiveRetryStrategy) handleTransient(h *HealthMetrics, req *DecisionRequest) *ActionPlan {
	// 已重试多次，升级为 B 级需人工确认
	level := consts.DecisionLevelA
	confidence := 0.85
	if h.RetryCount >= 2 {
		level = consts.DecisionLevelB
		confidence = 0.65
	}

	return &ActionPlan{
		StrategyName:    "adaptive_retry",
		Trigger:         req.TriggerSource,
		DecisionLevel:   level,
		ActionType:      consts.ActionTypeRetryTask,
		TargetID:        req.DomainTaskID,
		Reasoning:       fmt.Sprintf("瞬态错误，建议重试（已重试 %d 次）", h.RetryCount),
		RollbackAction:  consts.ActionTypePauseWorkflow,
		ExpectedOutcome: "任务重试后成功完成",
		Meta: &DecisionMeta{
			Confidence:          confidence,
			EvidenceSufficiency: 0.7,
			Reversibility:       "full",
			BlastRadius:         "task",
		},
	}
}

func (s *AdaptiveRetryStrategy) handleStructural(h *HealthMetrics, req *DecisionRequest) *ActionPlan {
	level := consts.DecisionLevelB
	if h.ReworkRounds >= 2 {
		level = consts.DecisionLevelC
	}

	return &ActionPlan{
		StrategyName:    "adaptive_retry",
		Trigger:         req.TriggerSource,
		DecisionLevel:   level,
		ActionType:      consts.ActionTypeTriggerRework,
		TargetID:        req.DomainTaskID,
		Reasoning:       fmt.Sprintf("结构性错误：连续失败 %d 次，返工 %d 轮，建议返工修复", h.ConsecutiveFailures, h.ReworkRounds),
		RollbackAction:  consts.ActionTypePauseWorkflow,
		ExpectedOutcome: "通过返工修正方案缺陷",
		Meta: &DecisionMeta{
			Confidence:          0.7,
			EvidenceSufficiency: 0.65,
			Reversibility:       "partial",
			BlastRadius:         "batch",
		},
	}
}

func (s *AdaptiveRetryStrategy) handleFatal(req *DecisionRequest) *ActionPlan {
	return &ActionPlan{
		StrategyName:    "adaptive_retry",
		Trigger:         req.TriggerSource,
		DecisionLevel:   consts.DecisionLevelC,
		ActionType:      consts.ActionTypePauseWorkflow,
		TargetID:        req.WorkflowRunID,
		Reasoning:       "致命错误：检测到 critical 异常信号，建议暂停工作流并人工介入",
		RollbackAction:  "",
		ExpectedOutcome: "人工诊断并修复环境后恢复",
		Meta: &DecisionMeta{
			Confidence:          0.9,
			EvidenceSufficiency: 0.8,
			Reversibility:       "none",
			BlastRadius:         "workflow",
		},
	}
}
