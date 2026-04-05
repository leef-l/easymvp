package autonomy

import (
	"context"
	"fmt"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
)

// BatchAdjustStrategy 批次并发度动态调整策略（B4）。
//
// 根据态势动态调整调度器的 max_concurrent 参数：
//   - 失败率高（≥30%）或有 critical 异常 → 降低并发（保守）
//   - P95 时长正常 + 失败率低（<10%）+ 资源充足 → 提升并发（激进）
//   - 其余情况维持现状
//
// 触发时机：task.failed、workflow.circuit_break、rework.completed。
type BatchAdjustStrategy struct{}

func NewBatchAdjustStrategy() *BatchAdjustStrategy {
	return &BatchAdjustStrategy{}
}

func (s *BatchAdjustStrategy) Name() string { return "batch_adjust" }

func (s *BatchAdjustStrategy) Priority() int { return 60 }

func (s *BatchAdjustStrategy) Applicable(sit *Situation, trigger string) bool {
	if sit == nil || sit.Health == nil || sit.Resource == nil {
		return false
	}
	switch trigger {
	case consts.TriggerTaskFailed,
		consts.TriggerCircuitBreak,
		consts.TriggerReworkCompleted:
		return true
	}
	return false
}

func (s *BatchAdjustStrategy) Evaluate(ctx context.Context, sit *Situation, req *DecisionRequest) *ActionPlan {
	if sit == nil || sit.Health == nil || sit.Resource == nil {
		return nil
	}

	h := sit.Health
	r := sit.Resource

	currentMax := r.MaxConcurrency
	if currentMax <= 0 {
		// 读取引擎配置默认值
		currentMax = engine.GetConfigInt(ctx, "scheduler.max_concurrent", "scheduler.maxConcurrent", 20)
	}

	direction, newMax, reason, confidence := s.calcAdjustment(sit, h, r, currentMax)
	if direction == "none" {
		return nil
	}

	level := consts.DecisionLevelA
	if direction == "decrease" && h.ConsecutiveFailures >= 5 {
		// 连续失败严重，升到 B 级让人确认是否真的要降并发
		level = consts.DecisionLevelB
		confidence = 0.7
	}

	return &ActionPlan{
		StrategyName:    s.Name(),
		Trigger:         req.TriggerSource,
		DecisionLevel:   level,
		ActionType:      consts.ActionTypeNotifyHuman, // 实际执行由 ActionDispatcher 映射到调参动作
		TargetID:        req.WorkflowRunID,
		Reasoning:       reason,
		RollbackAction:  consts.ActionTypePauseWorkflow,
		ExpectedOutcome: fmt.Sprintf("并发度%s至 %d，提升系统稳定性", directionLabel(direction), newMax),
		Parameters: map[string]interface{}{
			"adjust_direction": direction,        // "increase" / "decrease"
			"new_max_concurrent": newMax,
			"old_max_concurrent": currentMax,
		},
		Meta: &DecisionMeta{
			Confidence:          confidence,
			EvidenceSufficiency: 0.75,
			Reversibility:       "full",
			BlastRadius:         "workflow",
		},
	}
}

// calcAdjustment 计算并发度调整方向和目标值。
// 返回 direction（increase/decrease/none）、newMax、reason、confidence。
func (s *BatchAdjustStrategy) calcAdjustment(
	sit *Situation,
	h *HealthMetrics,
	r *ResourceMetrics,
	currentMax int,
) (string, int, string, float64) {

	// ---- 降并发条件 ----

	// 有 critical 异常信号
	if sit.HasCriticalAnomaly() {
		newMax := max2(1, currentMax/2)
		return "decrease", newMax,
			fmt.Sprintf("检测到 critical 异常信号，并发度减半（%d → %d）", currentMax, newMax),
			0.90
	}

	// 近期失败率 ≥ 30%
	if h.RecentFailureRate >= 0.30 {
		step := calcDecreaseStep(currentMax)
		newMax := max2(1, currentMax-step)
		return "decrease", newMax,
			fmt.Sprintf("近期失败率 %.0f%%（≥30%%），降低并发度（%d → %d）", h.RecentFailureRate*100, currentMax, newMax),
			0.80
	}

	// 连续失败 ≥ 3
	if h.ConsecutiveFailures >= 3 {
		step := calcDecreaseStep(currentMax)
		newMax := max2(1, currentMax-step)
		return "decrease", newMax,
			fmt.Sprintf("连续失败 %d 次，降低并发度（%d → %d）", h.ConsecutiveFailures, currentMax, newMax),
			0.75
	}

	// 时长趋势上升（任务变慢，资源可能过载）
	if sit.Trend != nil && sit.Trend.DurationTrend == "rising" && r.ResourceUtilization >= 0.85 {
		step := calcDecreaseStep(currentMax)
		newMax := max2(1, currentMax-step)
		return "decrease", newMax,
			fmt.Sprintf("任务耗时趋势上升且资源利用率 %.0f%%，适度降低并发度（%d → %d）", r.ResourceUtilization*100, currentMax, newMax),
			0.65
	}

	// ---- 升并发条件 ----

	// 失败率低 + 吞吐量趋势上升 + 资源宽松
	if h.RecentFailureRate < 0.10 &&
		sit.Trend != nil && sit.Trend.ThroughputTrend != "falling" &&
		r.ResourceUtilization < 0.60 &&
		h.ConsecutiveFailures == 0 {

		// 读取最大并发上限（避免无限升），默认上限 50
		hardLimit := 50
		if currentMax >= hardLimit {
			return "none", currentMax, "", 0
		}
		step := calcIncreaseStep(currentMax)
		newMax := min2(hardLimit, currentMax+step)
		return "increase", newMax,
			fmt.Sprintf("系统健康稳定（失败率 %.0f%%，资源利用率 %.0f%%），提升并发度（%d → %d）",
				h.RecentFailureRate*100, r.ResourceUtilization*100, currentMax, newMax),
			0.70
	}

	return "none", currentMax, "", 0
}

func calcDecreaseStep(current int) int {
	if current <= 5 {
		return 1
	}
	if current <= 10 {
		return 2
	}
	return current / 4 // 每次降 25%
}

func calcIncreaseStep(current int) int {
	if current <= 5 {
		return 2
	}
	if current <= 20 {
		return 3
	}
	return 5
}

func directionLabel(d string) string {
	if d == "increase" {
		return "提升"
	}
	return "降低"
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}
