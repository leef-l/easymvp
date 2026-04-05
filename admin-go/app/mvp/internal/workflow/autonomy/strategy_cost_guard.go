package autonomy

import (
	"context"
	"fmt"

	"easymvp/app/mvp/internal/consts"
)

// CostGuardStrategy 成本守卫策略。
//
// 在每次决策前检查 Token 消耗是否接近或超出预算：
//   - 已用 ≥ 90% 预算：暂停，通知人工
//   - 已用 ≥ 70% 预算：降级为 B 级，需人工确认后续动作
//   - 预算为 0：不限制（策略不适用）
type CostGuardStrategy struct{}

func NewCostGuardStrategy() *CostGuardStrategy {
	return &CostGuardStrategy{}
}

func (s *CostGuardStrategy) Name() string { return "cost_guard" }

func (s *CostGuardStrategy) Priority() int { return 100 } // 最高优先级：成本超限优先于一切

func (s *CostGuardStrategy) Applicable(sit *Situation, trigger string) bool {
	// 有资源指标且存在 Token 预算时才适用
	if sit.Resource == nil {
		return false
	}
	// EstimatedTokensLeft == 0 且 TokensConsumed == 0 → 无预算配置
	if sit.Resource.TokensConsumed == 0 && sit.Resource.EstimatedTokensLeft == 0 {
		return false
	}
	totalBudget := sit.Resource.TokensConsumed + sit.Resource.EstimatedTokensLeft
	return totalBudget > 0
}

func (s *CostGuardStrategy) Evaluate(ctx context.Context, sit *Situation, req *DecisionRequest) *ActionPlan {
	r := sit.Resource
	totalBudget := r.TokensConsumed + r.EstimatedTokensLeft
	if totalBudget <= 0 {
		return nil
	}

	usageRate := float64(r.TokensConsumed) / float64(totalBudget)

	if usageRate >= 0.9 {
		return &ActionPlan{
			StrategyName:    "cost_guard",
			Trigger:         req.TriggerSource,
			DecisionLevel:   consts.DecisionLevelC,
			ActionType:      consts.ActionTypePauseWorkflow,
			TargetID:        req.WorkflowRunID,
			Reasoning:       fmt.Sprintf("Token 预算已用 %.0f%%（%d/%d），建议暂停", usageRate*100, r.TokensConsumed, totalBudget),
			RollbackAction:  "",
			ExpectedOutcome: "人工决定是否追加预算或终止项目",
			Meta: &DecisionMeta{
				Confidence:          0.95,
				EvidenceSufficiency: 0.9,
				Reversibility:       "full",
				BlastRadius:         "workflow",
			},
		}
	}

	if usageRate >= 0.7 {
		return &ActionPlan{
			StrategyName:    "cost_guard",
			Trigger:         req.TriggerSource,
			DecisionLevel:   consts.DecisionLevelB,
			ActionType:      consts.ActionTypeNotifyHuman,
			TargetID:        req.WorkflowRunID,
			Reasoning:       fmt.Sprintf("Token 预算已用 %.0f%%（%d/%d），接近上限", usageRate*100, r.TokensConsumed, totalBudget),
			ExpectedOutcome: "人工确认后继续或调整预算",
			Meta: &DecisionMeta{
				Confidence:          0.85,
				EvidenceSufficiency: 0.85,
				Reversibility:       "full",
				BlastRadius:         "workflow",
			},
		}
	}

	return nil // 预算充足，无需干预
}
