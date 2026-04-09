package chat

import (
	"strings"
	"testing"

	workspacepkg "easymvp/app/mvp/internal/workspace"
)

func TestSummarizeRiskDeliveryPoliciesOK(t *testing.T) {
	t.Parallel()

	status, message := summarizeRiskDeliveryPolicies(map[string]workspacepkg.RiskDeliveryPolicy{
		workspacepkg.RiskLevelLow: {
			RiskLevel:    workspacepkg.RiskLevelLow,
			DeliveryMode: workspacepkg.DeliveryModePatch,
			SyncStrategy: workspacepkg.SyncStrategyAutoApply,
		},
		workspacepkg.RiskLevelMedium: {
			RiskLevel:    workspacepkg.RiskLevelMedium,
			DeliveryMode: workspacepkg.DeliveryModePatch,
			SyncStrategy: workspacepkg.SyncStrategyManual,
		},
		workspacepkg.RiskLevelHigh: {
			RiskLevel:    workspacepkg.RiskLevelHigh,
			DeliveryMode: workspacepkg.DeliveryModeManual,
			SyncStrategy: workspacepkg.SyncStrategyManual,
		},
	})
	if status != "ok" {
		t.Fatalf("status = %s, want ok; message=%s", status, message)
	}
	if !strings.Contains(message, "low=patch/auto_apply") || !strings.Contains(message, "high=manual/manual") {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestSummarizeRiskDeliveryPoliciesWarnsOnUnsafePolicy(t *testing.T) {
	t.Parallel()

	status, message := summarizeRiskDeliveryPolicies(map[string]workspacepkg.RiskDeliveryPolicy{
		workspacepkg.RiskLevelLow: {
			RiskLevel:    workspacepkg.RiskLevelLow,
			DeliveryMode: workspacepkg.DeliveryModePatch,
			SyncStrategy: workspacepkg.SyncStrategyAutoApply,
		},
		workspacepkg.RiskLevelMedium: {
			RiskLevel:    workspacepkg.RiskLevelMedium,
			DeliveryMode: workspacepkg.DeliveryModePatch,
			SyncStrategy: workspacepkg.SyncStrategyAutoApply,
		},
		workspacepkg.RiskLevelHigh: {
			RiskLevel:    workspacepkg.RiskLevelHigh,
			DeliveryMode: workspacepkg.DeliveryModePatch,
			SyncStrategy: workspacepkg.SyncStrategyAutoApply,
		},
	})
	if status != "warning" {
		t.Fatalf("status = %s, want warning; message=%s", status, message)
	}
	for _, fragment := range []string{"medium 风险仍允许 auto_apply", "high 风险未进入 PR/人工交付路径"} {
		if !strings.Contains(message, fragment) {
			t.Fatalf("message missing %q: %s", fragment, message)
		}
	}
}
