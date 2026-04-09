package workspace

import (
	"context"
	"fmt"
	"strings"
)

type RiskDeliveryPolicyReport struct {
	Policies map[string]RiskDeliveryPolicy
	Warnings []string
}

func InspectRiskDeliveryPolicies(ctx context.Context) RiskDeliveryPolicyReport {
	return InspectRiskDeliveryPoliciesFromMap(GetRiskDeliveryPolicies(ctx))
}

func InspectRiskDeliveryPoliciesFromMap(policies map[string]RiskDeliveryPolicy) RiskDeliveryPolicyReport {
	report := RiskDeliveryPolicyReport{
		Policies: policies,
		Warnings: make([]string, 0, 3),
	}

	for _, level := range []string{RiskLevelLow, RiskLevelMedium, RiskLevelHigh} {
		policy, ok := policies[level]
		if !ok {
			report.Warnings = append(report.Warnings, level+" 风险未配置")
			continue
		}
		switch level {
		case RiskLevelLow:
			if policy.DeliveryMode != DeliveryModePatch || policy.SyncStrategy != SyncStrategyAutoApply {
				report.Warnings = append(report.Warnings, "low 风险未配置为 patch + auto_apply")
			}
		case RiskLevelMedium:
			if policy.SyncStrategy == SyncStrategyAutoApply {
				report.Warnings = append(report.Warnings, "medium 风险仍允许 auto_apply")
			}
		case RiskLevelHigh:
			if policy.SyncStrategy == SyncStrategyAutoApply || policy.DeliveryMode == DeliveryModePatch {
				report.Warnings = append(report.Warnings, "high 风险未进入 PR/人工交付路径")
			}
		}
	}

	return report
}

func (r RiskDeliveryPolicyReport) Summary() string {
	parts := make([]string, 0, 3)
	for _, level := range []string{RiskLevelLow, RiskLevelMedium, RiskLevelHigh} {
		if policy, ok := r.Policies[level]; ok {
			parts = append(parts, fmt.Sprintf("%s=%s/%s", level, policy.DeliveryMode, policy.SyncStrategy))
		}
	}
	if len(parts) == 0 {
		return "风险交付矩阵为空"
	}
	return "风险交付矩阵: " + strings.Join(parts, ", ")
}

func (r RiskDeliveryPolicyReport) OK() bool {
	return len(r.Warnings) == 0
}
