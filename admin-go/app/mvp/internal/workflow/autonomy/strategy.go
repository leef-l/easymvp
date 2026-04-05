package autonomy

import "context"

// Strategy 策略函数接口。
// 每个策略负责一个决策域（重试/执行器选择/并发度/成本/重规划/质量门），
// 由 Planner 统一编排。
type Strategy interface {
	// Name 策略唯一标识，对应灰度开关前缀。
	Name() string

	// Evaluate 评估当前态势，输出行动计划。
	// 返回 nil 表示该策略无建议。
	Evaluate(ctx context.Context, sit *Situation, req *DecisionRequest) *ActionPlan

	// Applicable 判断该策略是否适用于当前态势和触发源。
	Applicable(sit *Situation, trigger string) bool

	// Priority 优先级，数值越大越优先被评估。
	Priority() int
}

// ActionPlan 策略评估输出的行动计划。
type ActionPlan struct {
	StrategyName    string                 `json:"strategyName"`
	Trigger         string                 `json:"trigger"`
	DecisionLevel   string                 `json:"decisionLevel"`   // A / B / C
	ActionType      string                 `json:"actionType"`      // consts.ActionType*
	TargetID        int64                  `json:"targetId,omitempty"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
	Meta            *DecisionMeta          `json:"meta"`
	Reasoning       string                 `json:"reasoning"`
	RollbackAction  string                 `json:"rollbackAction,omitempty"`
	ExpectedOutcome string                 `json:"expectedOutcome,omitempty"`
}

// ActionPlanList 多策略输出列表，提供排序和合并能力。
type ActionPlanList []*ActionPlan

func (l ActionPlanList) Len() int      { return len(l) }
func (l ActionPlanList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// Less 排序规则：置信度高的优先；置信度相同时，决策级别更保守（C>B>A）的优先。
func (l ActionPlanList) Less(i, j int) bool {
	ci, cj := l[i].Meta.Confidence, l[j].Meta.Confidence
	if ci != cj {
		return ci > cj
	}
	return levelToInt(l[i].DecisionLevel) > levelToInt(l[j].DecisionLevel)
}

// MostConservativeLevel 从多个计划中取最保守的决策级别。
func (l ActionPlanList) MostConservativeLevel() string {
	best := "A"
	for _, p := range l {
		if levelToInt(p.DecisionLevel) > levelToInt(best) {
			best = p.DecisionLevel
		}
	}
	return best
}
