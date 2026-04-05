package autonomy

import (
	"context"
	"sort"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
)

// Planner 策略编排器。
//
// 职责：遍历已注册策略 → 按优先级排序 → 选出最优行动计划。
// 核心原则：冲突时取最保守方案。
type Planner struct {
	strategies []Strategy
}

// NewPlanner 创建策略编排器。
func NewPlanner() *Planner {
	return &Planner{}
}

// Register 注册策略。
func (p *Planner) Register(s Strategy) {
	p.strategies = append(p.strategies, s)
}

// IsEnabled 灰度检查：strategy_enabled 总开关。
func (p *Planner) IsEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.autonomy.strategy_enabled", "workflow.autonomy.strategyEnabled", 0) == 1
}

// isStrategyEnabled 检查单个策略的独立灰度开关。
// 开关键格式：workflow.autonomy.{strategy_name}_enabled
func (p *Planner) isStrategyEnabled(ctx context.Context, name string) bool {
	key := "workflow.autonomy." + name + "_enabled"
	camelKey := "workflow.autonomy." + name + "Enabled"
	return engine.GetConfigInt(ctx, key, camelKey, 1) == 1
}

// Plan 评估所有适用策略，返回最优行动计划。
//
// 流程：
//  1. 过滤 Applicable + 灰度启用的策略
//  2. 按 Priority 降序排列
//  3. 逐一 Evaluate，收集非 nil 结果
//  4. 按置信度排序，选置信度最高的作为主计划
//  5. 如果多个计划建议不同决策级别，取最保守的
//
// 返回 nil 表示无策略建议，调用方应走原 PolicyEngine 路径。
func (p *Planner) Plan(ctx context.Context, sit *Situation, req *DecisionRequest) *ActionPlan {
	if sit == nil || len(p.strategies) == 0 {
		return nil
	}

	// 1. 过滤适用策略
	var applicable []Strategy
	for _, s := range p.strategies {
		if !s.Applicable(sit, req.TriggerSource) {
			continue
		}
		if !p.isStrategyEnabled(ctx, s.Name()) {
			g.Log().Debugf(ctx, "[Planner] 策略 %s 灰度关闭，跳过", s.Name())
			continue
		}
		applicable = append(applicable, s)
	}
	if len(applicable) == 0 {
		return nil
	}

	// 2. 按优先级降序排序
	sort.Slice(applicable, func(i, j int) bool {
		return applicable[i].Priority() > applicable[j].Priority()
	})

	// 3. 逐一评估，收集结果
	var plans ActionPlanList
	for _, s := range applicable {
		plan := s.Evaluate(ctx, sit, req)
		if plan == nil {
			continue
		}
		if plan.Meta == nil {
			plan.Meta = &DecisionMeta{Confidence: 0.5, EvidenceSufficiency: 0.5, Reversibility: "full", BlastRadius: "task"}
		}
		plans = append(plans, plan)
		g.Log().Infof(ctx, "[Planner] 策略 %s 输出计划: action=%s level=%s confidence=%.2f",
			s.Name(), plan.ActionType, plan.DecisionLevel, plan.Meta.Confidence)
	}
	if len(plans) == 0 {
		return nil
	}

	// 4. 按置信度排序，选最高的
	sort.Sort(plans)
	best := plans[0]

	// 5. 安全合并：决策级别取所有计划中最保守的
	conservativeLevel := plans.MostConservativeLevel()
	if levelToInt(conservativeLevel) > levelToInt(best.DecisionLevel) {
		g.Log().Infof(ctx, "[Planner] 决策级别升级: %s → %s (安全合并)", best.DecisionLevel, conservativeLevel)
		best.DecisionLevel = conservativeLevel
	}

	// 6. DecisionMeta 校验
	if validation := best.Meta.Validate(best.DecisionLevel); validation != nil {
		if !validation.Allowed {
			g.Log().Warningf(ctx, "[Planner] DecisionMeta 校验拒绝: %s", validation.Reason)
			best.DecisionLevel = consts.DecisionLevelC
		} else if validation.UpgradeTo != "" && levelToInt(validation.UpgradeTo) > levelToInt(best.DecisionLevel) {
			g.Log().Infof(ctx, "[Planner] DecisionMeta 升级: %s → %s (%s)", best.DecisionLevel, validation.UpgradeTo, validation.Reason)
			best.DecisionLevel = validation.UpgradeTo
		}
	}

	return best
}
