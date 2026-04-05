package autonomy

import (
	"context"
	"fmt"

	"easymvp/app/mvp/internal/consts"
)

// QualityGateStrategy 审计严格度动态调整策略（B7）。
//
// 根据当前工程质量态势动态决定是否调整审计严格度：
//   - 高失败率 / 多轮返工 → 收紧（更严格审计）
//   - 持续稳定，零失败，完成率高 → 放松（跳过部分低风险审计）
//
// 触发时机：accept.passed、rework.completed、task.failed。
type QualityGateStrategy struct{}

func NewQualityGateStrategy() *QualityGateStrategy {
	return &QualityGateStrategy{}
}

func (s *QualityGateStrategy) Name() string { return "quality_gate" }

func (s *QualityGateStrategy) Priority() int { return 50 }

func (s *QualityGateStrategy) Applicable(sit *Situation, trigger string) bool {
	if sit == nil || sit.Health == nil {
		return false
	}
	switch trigger {
	case consts.TriggerAcceptPassed,
		consts.TriggerReworkCompleted,
		consts.TriggerTaskFailed:
		return true
	}
	return false
}

func (s *QualityGateStrategy) Evaluate(ctx context.Context, sit *Situation, req *DecisionRequest) *ActionPlan {
	if sit == nil || sit.Health == nil || sit.Progress == nil {
		return nil
	}

	h := sit.Health
	prog := sit.Progress

	mode, reason, confidence := s.calcMode(sit, h, prog)
	if mode == "" {
		return nil
	}

	level := consts.DecisionLevelA
	actionType := consts.ActionTypeNotifyHuman // 实际由 ActionDispatcher 映射到审计策略调参

	// 放松审计需要人工确认（B 级），避免自动跳过审计导致质量问题
	if mode == "relax" {
		level = consts.DecisionLevelB
	}

	// 收紧在高失败率下可以自动执行，但返工多时建议人工确认
	if mode == "tighten" && h.ReworkRounds >= 3 {
		level = consts.DecisionLevelB
		confidence = 0.75
	}

	return &ActionPlan{
		StrategyName:    s.Name(),
		Trigger:         req.TriggerSource,
		DecisionLevel:   level,
		ActionType:      actionType,
		TargetID:        req.WorkflowRunID,
		Reasoning:       reason,
		RollbackAction:  "", // 审计策略调整可随时回滚
		ExpectedOutcome: qualityOutcome(mode),
		Parameters: map[string]interface{}{
			"quality_gate_mode": mode, // "tighten" / "relax"
			"workflow_run_id":   req.WorkflowRunID,
		},
		Meta: &DecisionMeta{
			Confidence:          confidence,
			EvidenceSufficiency: 0.70,
			Reversibility:       "full",
			BlastRadius:         "workflow",
		},
	}
}

// calcMode 计算审计严格度调整方向。
// 返回 mode（tighten/relax/""）、reason、confidence。
func (s *QualityGateStrategy) calcMode(
	sit *Situation,
	h *HealthMetrics,
	prog *ProgressMetrics,
) (string, string, float64) {

	// ---- 收紧条件（优先级高于放松）----

	// 返工轮数多（质量问题持续）
	if h.ReworkRounds >= 2 {
		return "tighten",
			fmt.Sprintf("已返工 %d 轮，质量持续不达标，收紧审计严格度", h.ReworkRounds),
			0.85
	}

	// 高失败率 + critical 异常
	if sit.HasCriticalAnomaly() && h.RecentFailureRate >= 0.20 {
		return "tighten",
			fmt.Sprintf("存在 critical 异常且失败率 %.0f%%，收紧审计以尽早拦截质量问题", h.RecentFailureRate*100),
			0.88
	}

	// 近期失败率高（≥25%）
	if h.RecentFailureRate >= 0.25 {
		return "tighten",
			fmt.Sprintf("近期失败率 %.0f%%（≥25%%），收紧审计严格度", h.RecentFailureRate*100),
			0.80
	}

	// 升级次数多（说明问题被反复上报架构师）
	if h.EscalationCount >= 3 {
		return "tighten",
			fmt.Sprintf("已升级 %d 次，审计应更严格以减少后期返工", h.EscalationCount),
			0.75
	}

	// ---- 放松条件 ----

	// 项目进展顺利：完成率≥70%、连续零失败、零返工、无异常
	if prog.CompletionRate >= 0.70 &&
		h.ConsecutiveFailures == 0 &&
		h.ReworkRounds == 0 &&
		!sit.HasCriticalAnomaly() &&
		h.RecentFailureRate < 0.05 {

		return "relax",
			fmt.Sprintf("项目稳定（完成率 %.0f%%，近期失败率 %.0f%%，零返工），可适度放松低风险任务的审计",
				prog.CompletionRate*100, h.RecentFailureRate*100),
			0.65
	}

	// 吞吐量趋势上升 + 失败率极低
	if sit.Trend != nil &&
		sit.Trend.ThroughputTrend == "rising" &&
		h.RecentFailureRate < 0.08 &&
		h.ReworkRounds == 0 {

		return "relax",
			fmt.Sprintf("吞吐量上升且失败率仅 %.0f%%，适度放松审计以提升速度", h.RecentFailureRate*100),
			0.60
	}

	return "", "", 0
}

func qualityOutcome(mode string) string {
	if mode == "tighten" {
		return "通过更严格审计提前拦截质量问题，减少后期返工轮次"
	}
	return "适度放松低风险任务的审计门槛，提升整体吞吐量"
}
