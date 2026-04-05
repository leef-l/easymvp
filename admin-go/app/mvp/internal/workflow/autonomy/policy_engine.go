package autonomy

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/workflow/repo"
)

// PolicyEngine 策略引擎：根据触发源和项目作用域匹配策略规则，输出决策等级与动作类型。
type PolicyEngine struct {
	ruleRepo *repo.PolicyRuleRepo
}

// NewPolicyEngine 创建策略引擎。
func NewPolicyEngine(ruleRepo *repo.PolicyRuleRepo) *PolicyEngine {
	return &PolicyEngine{ruleRepo: ruleRepo}
}

// Match 匹配策略规则。
// 匹配逻辑：trigger_source 过滤 → project_category_code 精确匹配 → family 匹配 → 全局 → priority ASC 取首条。
func (pe *PolicyEngine) Match(ctx context.Context, triggerSource, family, categoryCode string) *PolicyMatch {
	rules, err := pe.ruleRepo.ListByTriggerAndScope(ctx, triggerSource, family, categoryCode)
	if err != nil {
		g.Log().Warningf(ctx, "[PolicyEngine] 查询策略规则失败: trigger=%s err=%v", triggerSource, err)
		return nil
	}
	if len(rules) == 0 {
		return nil
	}

	// 已按 priority ASC 排序，取首条
	row := rules[0]
	rule := mapToPolicyRule(row)

	decisionLevel := rule.DecisionLevel
	actionType := pe.resolveActionType(rule)
	autoExecutable := decisionLevel == consts.DecisionLevelA

	return &PolicyMatch{
		Rule:           rule,
		DecisionLevel:  decisionLevel,
		ActionType:     actionType,
		AutoExecutable: autoExecutable,
	}
}

// resolveActionType 从规则中解析动作类型。
// 优先使用 config_json.action_type，回退到按 decision_type 推导默认值。
func (pe *PolicyEngine) resolveActionType(rule *PolicyRule) string {
	if rule.ConfigJSON != nil {
		if at, ok := rule.ConfigJSON["action_type"]; ok {
			if s, ok := at.(string); ok && s != "" {
				return s
			}
		}
	}
	// 按触发源推导默认动作
	switch rule.TriggerSource {
	case consts.TriggerTaskFailed:
		return consts.ActionTypeRetryTask
	case consts.TriggerTaskRetryExhausted:
		return consts.ActionTypeTriggerRework
	case consts.TriggerCircuitBreak:
		return consts.ActionTypePauseWorkflow
	case consts.TriggerAcceptPassed:
		return consts.ActionTypeApproveComplete
	case consts.TriggerAcceptFailed:
		return consts.ActionTypeTriggerRework
	case consts.TriggerAcceptManualReview:
		return consts.ActionTypeNotifyHuman
	case consts.TriggerReworkCompleted:
		return consts.ActionTypeRerunAccept
	case consts.TriggerReplanSuggested:
		return consts.ActionTypeReplanWorkflow
	default:
		return consts.ActionTypeNotifyHuman
	}
}

// mapToPolicyRule 将数据库行映射为 PolicyRule 结构体。
func mapToPolicyRule(row g.Map) *PolicyRule {
	rule := &PolicyRule{
		ID:                  mapInt64(row, "id"),
		RuleCode:            mapString(row, "rule_code"),
		RuleName:            mapString(row, "rule_name"),
		DecisionType:        mapString(row, "decision_type"),
		DecisionLevel:       mapString(row, "decision_level"),
		TriggerSource:       mapString(row, "trigger_source"),
		ProjectFamily:       mapString(row, "project_family"),
		ProjectCategoryCode: mapString(row, "project_category_code"),
		Enabled:             mapInt64(row, "enabled") == 1,
		Priority:            int(mapInt64(row, "priority")),
	}
	if cfg := mapString(row, "config_json"); cfg != "" {
		rule.ConfigJSON = parseJSONMap(cfg)
	}
	return rule
}
