package workspace

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func TestClassifyTaskRisk(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		profile *taskDeliveryProfile
		want    string
	}{
		{
			name: "nil profile defaults to medium",
			want: RiskLevelMedium,
		},
		{
			name: "openhands is high risk",
			profile: &taskDeliveryProfile{
				ExecutionMode: "openhands",
			},
			want: RiskLevelHigh,
		},
		{
			name: "path escape is high risk",
			profile: &taskDeliveryProfile{
				ExecutionMode:     "aider",
				AffectedResources: []string{"../outside.txt"},
			},
			want: RiskLevelHigh,
		},
		{
			name: "small aider task is low risk",
			profile: &taskDeliveryProfile{
				TaskKind:          "implement",
				ExecutionMode:     "aider",
				AffectedResources: []string{"README.md", "docs/guide.md"},
			},
			want: RiskLevelLow,
		},
		{
			name: "many files stay high risk",
			profile: &taskDeliveryProfile{
				ExecutionMode:     "codex_cli",
				AffectedResources: []string{"a", "b", "c", "d", "e", "f"},
			},
			want: RiskLevelHigh,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := classifyTaskRisk(tc.profile); got != tc.want {
				t.Fatalf("classifyTaskRisk() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDefaultDeliveryPolicyByRiskDefaults(t *testing.T) {
	prev := workspaceDeliveryConfigString
	t.Cleanup(func() {
		workspaceDeliveryConfigString = prev
	})
	workspaceDeliveryConfigString = func(_ context.Context, _, _, defaultVal string) string {
		return defaultVal
	}

	low := defaultDeliveryPolicyByRisk(context.Background(), RiskLevelLow)
	if low.DeliveryMode != DeliveryModePatch || low.SyncStrategy != SyncStrategyAutoApply {
		t.Fatalf("unexpected low risk policy: %+v", low)
	}

	medium := defaultDeliveryPolicyByRisk(context.Background(), RiskLevelMedium)
	if medium.DeliveryMode != DeliveryModePatch || medium.SyncStrategy != SyncStrategyManual {
		t.Fatalf("unexpected medium risk policy: %+v", medium)
	}

	high := defaultDeliveryPolicyByRisk(context.Background(), RiskLevelHigh)
	if high.DeliveryMode != DeliveryModeManual || high.SyncStrategy != SyncStrategyManual {
		t.Fatalf("unexpected high risk policy: %+v", high)
	}
}

func TestResolveDeliveryPolicyUsesRiskMatrixForImplicitDecision(t *testing.T) {
	prev := workspaceDeliveryConfigString
	t.Cleanup(func() {
		workspaceDeliveryConfigString = prev
	})
	workspaceDeliveryConfigString = func(_ context.Context, _, _, defaultVal string) string {
		return defaultVal
	}

	policy := resolveDeliveryPolicy(context.Background(), 0, FinalizeRequest{})
	if policy.RiskLevel != RiskLevelMedium {
		t.Fatalf("risk level = %s, want %s", policy.RiskLevel, RiskLevelMedium)
	}
	if policy.DeliveryMode != DeliveryModePatch || policy.SyncStrategy != SyncStrategyManual {
		t.Fatalf("unexpected implicit policy: %+v", policy)
	}
}

func TestGetRiskDeliveryPoliciesHonorsConfigOverrides(t *testing.T) {
	prev := workspaceDeliveryConfigString
	t.Cleanup(func() {
		workspaceDeliveryConfigString = prev
	})
	workspaceDeliveryConfigString = func(_ context.Context, key, _, defaultVal string) string {
		switch key {
		case "workspace.delivery.medium_risk_mode":
			return DeliveryModePR
		case "workspace.delivery.medium_risk_sync_strategy":
			return SyncStrategyManual
		case "workspace.delivery.high_risk_mode":
			return DeliveryModePR
		case "workspace.delivery.high_risk_sync_strategy":
			return SyncStrategyManual
		default:
			return defaultVal
		}
	}

	policies := GetRiskDeliveryPolicies(context.Background())
	if policies[RiskLevelMedium].DeliveryMode != DeliveryModePR {
		t.Fatalf("medium risk policy not overridden: %+v", policies[RiskLevelMedium])
	}
	if policies[RiskLevelHigh].DeliveryMode != DeliveryModePR {
		t.Fatalf("high risk policy not overridden: %+v", policies[RiskLevelHigh])
	}
}

func TestNormalizeDeliveryPolicyFields(t *testing.T) {
	t.Parallel()

	if got := normalizeDeliveryMode(" PR "); got != DeliveryModePR {
		t.Fatalf("normalizeDeliveryMode() = %q", got)
	}
	if got := normalizeSyncStrategy(" AUTO_APPLY "); got != SyncStrategyAutoApply {
		t.Fatalf("normalizeSyncStrategy() = %q", got)
	}
	if got := normalizeRiskLevel(" HIGH "); got != RiskLevelHigh {
		t.Fatalf("normalizeRiskLevel() = %q", got)
	}
	if got := normalizeDeliveryMode("unknown"); got != "" {
		t.Fatalf("expected empty delivery mode, got %q", got)
	}
}

func TestFilterLegacyWorkspaceDataRemovesDeliveryColumns(t *testing.T) {
	t.Parallel()

	input := g.Map{
		"status":          StatusCompleted,
		"delivery_mode":   DeliveryModePatch,
		"delivery_status": DeliveryStatusReady,
		"sync_strategy":   SyncStrategyAutoApply,
		"sync_status":     SyncStatusApplied,
		"risk_level":      RiskLevelLow,
		"patch_ref":       "/tmp/task.patch",
	}

	got := filterLegacyWorkspaceData(input)
	if got["status"] != StatusCompleted {
		t.Fatalf("status not preserved: %#v", got)
	}
	for _, key := range []string{"delivery_mode", "delivery_status", "sync_strategy", "sync_status", "risk_level", "patch_ref"} {
		if _, ok := got[key]; ok {
			t.Fatalf("legacy data should not contain %q: %#v", key, got)
		}
	}
}

func TestIsUnknownColumnErr(t *testing.T) {
	t.Parallel()

	if !isUnknownColumnErr(assertError("Error 1054 (42S22): Unknown column 'delivery_mode' in 'field list'")) {
		t.Fatal("expected unknown column error to be detected")
	}
	if isUnknownColumnErr(assertError("duplicate entry")) {
		t.Fatal("did not expect non-column error to match")
	}
}

func assertError(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
