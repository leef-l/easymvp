package workspace

import (
	"strings"
	"testing"
)

func TestInspectRiskDeliveryPoliciesFromMapOK(t *testing.T) {
	t.Parallel()

	report := InspectRiskDeliveryPoliciesFromMap(map[string]RiskDeliveryPolicy{
		RiskLevelLow: {
			RiskLevel:    RiskLevelLow,
			DeliveryMode: DeliveryModePatch,
			SyncStrategy: SyncStrategyAutoApply,
		},
		RiskLevelMedium: {
			RiskLevel:    RiskLevelMedium,
			DeliveryMode: DeliveryModePatch,
			SyncStrategy: SyncStrategyManual,
		},
		RiskLevelHigh: {
			RiskLevel:    RiskLevelHigh,
			DeliveryMode: DeliveryModeManual,
			SyncStrategy: SyncStrategyManual,
		},
	})
	if !report.OK() {
		t.Fatalf("expected OK report, got warnings=%v", report.Warnings)
	}
	if !strings.Contains(report.Summary(), "low=patch/auto_apply") {
		t.Fatalf("unexpected summary: %s", report.Summary())
	}
}

func TestInspectRiskDeliveryPoliciesFromMapWarns(t *testing.T) {
	t.Parallel()

	report := InspectRiskDeliveryPoliciesFromMap(map[string]RiskDeliveryPolicy{
		RiskLevelLow: {
			RiskLevel:    RiskLevelLow,
			DeliveryMode: DeliveryModePatch,
			SyncStrategy: SyncStrategyAutoApply,
		},
		RiskLevelMedium: {
			RiskLevel:    RiskLevelMedium,
			DeliveryMode: DeliveryModePatch,
			SyncStrategy: SyncStrategyAutoApply,
		},
		RiskLevelHigh: {
			RiskLevel:    RiskLevelHigh,
			DeliveryMode: DeliveryModePatch,
			SyncStrategy: SyncStrategyAutoApply,
		},
	})
	if report.OK() {
		t.Fatal("expected warnings for unsafe report")
	}
	for _, fragment := range []string{"medium 风险仍允许 auto_apply", "high 风险未进入 PR/人工交付路径"} {
		found := false
		for _, warning := range report.Warnings {
			if strings.Contains(warning, fragment) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing warning %q in %v", fragment, report.Warnings)
		}
	}
}

func TestRiskDeliveryPolicyReportSummaryEmpty(t *testing.T) {
	t.Parallel()

	report := RiskDeliveryPolicyReport{}
	if got := report.Summary(); got != "风险交付矩阵为空" {
		t.Fatalf("Summary() = %q, want %q", got, "风险交付矩阵为空")
	}
}
