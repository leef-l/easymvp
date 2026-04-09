package workspace

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/engine"
)

var workspaceDeliveryConfigString = func(ctx context.Context, key, yamlPath, defaultVal string) string {
	return engine.GetConfigString(ctx, key, yamlPath, defaultVal)
}

type deliveryPolicy struct {
	DeliveryMode string
	SyncStrategy string
	RiskLevel    string
}

type RiskDeliveryPolicy struct {
	RiskLevel    string
	DeliveryMode string
	SyncStrategy string
}

type taskDeliveryProfile struct {
	TaskKind          string
	ExecutionMode     string
	AffectedResources []string
}

func resolveDeliveryPolicy(ctx context.Context, taskID int64, req FinalizeRequest) deliveryPolicy {
	profile := loadTaskDeliveryProfile(ctx, taskID)

	policy := deliveryPolicy{
		DeliveryMode: normalizeDeliveryMode(req.DeliveryMode),
		SyncStrategy: normalizeSyncStrategy(req.SyncStrategy),
		RiskLevel:    normalizeRiskLevel(req.RiskLevel),
	}

	if policy.DeliveryMode == "" {
		policy.DeliveryMode = normalizeDeliveryMode(workspaceDeliveryConfigString(
			ctx,
			"workspace.delivery.default_mode",
			"engine.workspace.delivery.defaultMode",
			DeliveryModePatch,
		))
	}
	if policy.DeliveryMode == "" {
		policy.DeliveryMode = DeliveryModePatch
	}

	if policy.RiskLevel == "" {
		policy.RiskLevel = classifyTaskRisk(profile)
	}
	if policy.RiskLevel == "" {
		policy.RiskLevel = RiskLevelMedium
	}

	implicit := defaultDeliveryPolicyByRisk(ctx, policy.RiskLevel)
	if policy.DeliveryMode == "" {
		policy.DeliveryMode = implicit.DeliveryMode
	}
	if policy.SyncStrategy == "" {
		policy.SyncStrategy = implicit.SyncStrategy
	}

	if policy.DeliveryMode == "" {
		policy.DeliveryMode = normalizeDeliveryMode(workspaceDeliveryConfigString(
			ctx,
			"workspace.delivery.default_mode",
			"engine.workspace.delivery.defaultMode",
			DeliveryModePatch,
		))
	}
	if policy.DeliveryMode == "" {
		policy.DeliveryMode = DeliveryModePatch
	}

	if policy.SyncStrategy == "" {
		policy.SyncStrategy = normalizeSyncStrategy(workspaceDeliveryConfigString(
			ctx,
			"workspace.delivery.default_sync_strategy",
			"engine.workspace.delivery.defaultSyncStrategy",
			SyncStrategyAutoApply,
		))
	}
	if policy.SyncStrategy == "" {
		if policy.DeliveryMode == DeliveryModePatch {
			policy.SyncStrategy = SyncStrategyAutoApply
		} else {
			policy.SyncStrategy = SyncStrategyManual
		}
	}

	if policy.DeliveryMode == DeliveryModePR || policy.DeliveryMode == DeliveryModeManual {
		policy.SyncStrategy = SyncStrategyManual
	}

	return policy
}

func GetRiskDeliveryPolicies(ctx context.Context) map[string]RiskDeliveryPolicy {
	policies := make(map[string]RiskDeliveryPolicy, 3)
	for _, riskLevel := range []string{RiskLevelLow, RiskLevelMedium, RiskLevelHigh} {
		policy := defaultDeliveryPolicyByRisk(ctx, riskLevel)
		policies[riskLevel] = RiskDeliveryPolicy{
			RiskLevel:    riskLevel,
			DeliveryMode: policy.DeliveryMode,
			SyncStrategy: policy.SyncStrategy,
		}
	}
	return policies
}

func defaultDeliveryPolicyByRisk(ctx context.Context, riskLevel string) deliveryPolicy {
	configKeys := map[string][2]string{
		RiskLevelLow: {
			"workspace.delivery.low_risk_mode",
			"engine.workspace.delivery.lowRiskMode",
		},
		RiskLevelMedium: {
			"workspace.delivery.medium_risk_mode",
			"engine.workspace.delivery.mediumRiskMode",
		},
		RiskLevelHigh: {
			"workspace.delivery.high_risk_mode",
			"engine.workspace.delivery.highRiskMode",
		},
	}
	syncKeys := map[string][2]string{
		RiskLevelLow: {
			"workspace.delivery.low_risk_sync_strategy",
			"engine.workspace.delivery.lowRiskSyncStrategy",
		},
		RiskLevelMedium: {
			"workspace.delivery.medium_risk_sync_strategy",
			"engine.workspace.delivery.mediumRiskSyncStrategy",
		},
		RiskLevelHigh: {
			"workspace.delivery.high_risk_sync_strategy",
			"engine.workspace.delivery.highRiskSyncStrategy",
		},
	}
	defaults := map[string]deliveryPolicy{
		RiskLevelLow: {
			DeliveryMode: DeliveryModePatch,
			SyncStrategy: SyncStrategyAutoApply,
		},
		RiskLevelMedium: {
			DeliveryMode: DeliveryModePatch,
			SyncStrategy: SyncStrategyManual,
		},
		RiskLevelHigh: {
			DeliveryMode: DeliveryModeManual,
			SyncStrategy: SyncStrategyManual,
		},
	}

	fallback, ok := defaults[riskLevel]
	if !ok {
		fallback = defaults[RiskLevelMedium]
	}

	modeKeys, ok := configKeys[riskLevel]
	if !ok {
		modeKeys = configKeys[RiskLevelMedium]
	}
	strategyKeys, ok := syncKeys[riskLevel]
	if !ok {
		strategyKeys = syncKeys[RiskLevelMedium]
	}

	return deliveryPolicy{
		DeliveryMode: normalizeDeliveryMode(workspaceDeliveryConfigString(ctx, modeKeys[0], modeKeys[1], fallback.DeliveryMode)),
		SyncStrategy: normalizeSyncStrategy(workspaceDeliveryConfigString(ctx, strategyKeys[0], strategyKeys[1], fallback.SyncStrategy)),
		RiskLevel:    riskLevel,
	}
}

func loadTaskDeliveryProfile(ctx context.Context, taskID int64) *taskDeliveryProfile {
	if taskID <= 0 {
		return nil
	}

	record, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).
		WhereNull("deleted_at").
		Fields("task_kind, execution_mode, affected_resources").
		One()
	if err != nil || record.IsEmpty() {
		return nil
	}

	var affectedResources []string
	rawResources := strings.TrimSpace(record["affected_resources"].String())
	if rawResources != "" && rawResources != "[]" && rawResources != "null" {
		_ = json.Unmarshal([]byte(rawResources), &affectedResources)
	}

	return &taskDeliveryProfile{
		TaskKind:          strings.ToLower(strings.TrimSpace(record["task_kind"].String())),
		ExecutionMode:     strings.ToLower(strings.TrimSpace(record["execution_mode"].String())),
		AffectedResources: affectedResources,
	}
}

func classifyTaskRisk(profile *taskDeliveryProfile) string {
	if profile == nil {
		return RiskLevelMedium
	}

	if profile.ExecutionMode == "openhands" {
		return RiskLevelHigh
	}

	switch profile.TaskKind {
	case "failure_analysis", "bug_analysis", "audit":
		return RiskLevelHigh
	}

	if len(profile.AffectedResources) == 0 {
		return RiskLevelMedium
	}

	if len(profile.AffectedResources) > 5 {
		return RiskLevelHigh
	}

	for _, resource := range profile.AffectedResources {
		resource = strings.TrimSpace(resource)
		if resource == "" {
			continue
		}
		if strings.HasPrefix(resource, "/") || strings.HasPrefix(resource, "\\") || strings.Contains(resource, "..") {
			return RiskLevelHigh
		}
	}

	if len(profile.AffectedResources) <= 2 {
		switch profile.ExecutionMode {
		case "aider", "claude_code", "codex_cli", "gemini_cli":
			return RiskLevelLow
		}
	}

	return RiskLevelMedium
}

func normalizeDeliveryMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case DeliveryModePatch:
		return DeliveryModePatch
	case DeliveryModePR:
		return DeliveryModePR
	case DeliveryModeManual:
		return DeliveryModeManual
	default:
		return ""
	}
}

func normalizeSyncStrategy(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case SyncStrategyAutoApply:
		return SyncStrategyAutoApply
	case SyncStrategyManual:
		return SyncStrategyManual
	default:
		return ""
	}
}

func normalizeRiskLevel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case RiskLevelLow:
		return RiskLevelLow
	case RiskLevelMedium:
		return RiskLevelMedium
	case RiskLevelHigh:
		return RiskLevelHigh
	default:
		return ""
	}
}
