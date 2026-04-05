package autonomy

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
)

// RiskGate 风险闸门：加载适用规则，评估触发表达式，命中即阻断。
type RiskGate struct {
	gateRepo *repo.RiskGateRuleRepo
}

// NewRiskGate 创建风险闸门。
func NewRiskGate(gateRepo *repo.RiskGateRuleRepo) *RiskGate {
	return &RiskGate{gateRepo: gateRepo}
}

// Check 检查当前决策请求是否被闸门阻断。
func (rg *RiskGate) Check(ctx context.Context, req *DecisionRequest, family, categoryCode string) *GateCheckResult {
	rules, err := rg.gateRepo.ListEnabled(ctx, family, categoryCode)
	if err != nil {
		g.Log().Warningf(ctx, "[RiskGate] 加载闸门规则失败: err=%v", err)
		return &GateCheckResult{Blocked: false}
	}
	if len(rules) == 0 {
		return &GateCheckResult{Blocked: false}
	}

	result := &GateCheckResult{}
	for _, row := range rules {
		gate := mapToRiskGateRule(row)
		if rg.evaluate(ctx, gate, req) {
			result.Blocked = true
			result.BlockedGates = append(result.BlockedGates, BlockedGate{
				GateID:         gate.ID,
				GateCode:       gate.GateCode,
				GateType:       gate.GateType,
				BlockAction:    gate.BlockAction,
				FallbackAction: gate.FallbackAction,
				Reason:         fmt.Sprintf("闸门 [%s] 命中: %s", gate.GateCode, gate.GateName),
			})
		}
	}

	return result
}

// evaluate 评估单条闸门规则是否命中。
// trigger_expression 是一个简单的条件 map：
//
//	{"trigger_source": "task.failed", "min_retry_count": 5}
//	{"gate_type": "cost", "max_rework_rounds": 3}
//
// 目前支持基本的字符串和数值匹配，后续可扩展为 CEL 表达式引擎。
func (rg *RiskGate) evaluate(ctx context.Context, gate *RiskGateRule, req *DecisionRequest) bool {
	expr := gate.TriggerExpression
	if len(expr) == 0 {
		return false // 空表达式不命中
	}

	// 检查 trigger_source 匹配
	if ts, ok := expr["trigger_source"]; ok {
		expected := fmt.Sprintf("%v", ts)
		if expected != "" && expected != req.TriggerSource {
			return false
		}
	}

	triggerCtx := req.TriggerContext
	if triggerCtx == nil {
		triggerCtx = make(map[string]interface{})
	}

	// 检查数值阈值条件
	thresholdChecks := map[string]string{
		"min_retry_count":    "retry_count",
		"min_failure_count":  "failure_count",
		"max_rework_rounds":  "rework_rounds",
		"max_accept_rounds":  "accept_rounds",
	}
	for exprKey, ctxKey := range thresholdChecks {
		if threshold, ok := expr[exprKey]; ok {
			thresholdVal := toFloat64(threshold)
			actualVal := toFloat64(triggerCtx[ctxKey])
			if strings.HasPrefix(exprKey, "min_") {
				if actualVal < thresholdVal {
					return false
				}
			} else { // max_
				if actualVal < thresholdVal {
					return false // 未达到最大阈值，不命中
				}
			}
		}
	}

	return true
}

// mapToRiskGateRule 将数据库行映射为 RiskGateRule。
func mapToRiskGateRule(row g.Map) *RiskGateRule {
	gate := &RiskGateRule{
		ID:                  mapInt64(row, "id"),
		GateCode:            mapString(row, "gate_code"),
		GateName:            mapString(row, "gate_name"),
		GateType:            mapString(row, "gate_type"),
		ProjectFamily:       mapString(row, "project_family"),
		ProjectCategoryCode: mapString(row, "project_category_code"),
		BlockAction:         mapString(row, "block_action"),
		FallbackAction:      mapString(row, "fallback_action"),
		Enabled:             mapInt64(row, "enabled") == 1,
		Priority:            int(mapInt64(row, "priority")),
	}
	if expr := mapString(row, "trigger_expression"); expr != "" {
		gate.TriggerExpression = parseJSONMap(expr)
	}
	return gate
}

// toFloat64 安全转换为 float64。
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	default:
		return 0
	}
}
