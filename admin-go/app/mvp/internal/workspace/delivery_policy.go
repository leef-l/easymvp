package workspace

import (
	"context"
	"encoding/json"
	"path"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/engine"
)

var workspaceDeliveryConfigString = func(ctx context.Context, key, yamlPath, defaultVal string) string {
	return engine.GetConfigString(ctx, key, yamlPath, defaultVal)
}

var loadTaskDeliveryProfileFn = loadTaskDeliveryProfile

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
	Name              string
	Description       string
	AffectedResources []string
}

func resolveDeliveryPolicy(ctx context.Context, taskID int64, req FinalizeRequest) deliveryPolicy {
	profile := loadTaskDeliveryProfileFn(ctx, taskID)

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
	if strings.TrimSpace(req.SyncStrategy) == "" && shouldPreferAutoApply(profile, policy) {
		policy.DeliveryMode = DeliveryModePatch
		policy.SyncStrategy = SyncStrategyAutoApply
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
		Fields("task_kind, execution_mode, name, description, affected_resources").
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
		Name:              strings.TrimSpace(record["name"].String()),
		Description:       strings.TrimSpace(record["description"].String()),
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

	if len(profile.AffectedResources) <= 3 {
		switch profile.ExecutionMode {
		case "auto", "aider", "claude_code", "codex_cli", "gemini_cli":
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

func shouldPreferAutoApply(profile *taskDeliveryProfile, policy deliveryPolicy) bool {
	if profile == nil {
		return false
	}
	if policy.DeliveryMode != DeliveryModePatch {
		return false
	}
	if policy.SyncStrategy == SyncStrategyAutoApply {
		return false
	}
	if !supportsExplicitAutoApply(profile) {
		return false
	}

	resources := normalizeExplicitDeliveryResources(profile.AffectedResources)
	if len(resources) == 0 {
		return false
	}

	if looksLikeBootstrapTask(profile) && len(resources) <= 24 {
		return true
	}

	if isNarrowBaselineScopedChange(resources) && len(resources) <= 12 {
		return true
	}

	return false
}

func supportsExplicitAutoApply(profile *taskDeliveryProfile) bool {
	switch strings.ToLower(strings.TrimSpace(profile.TaskKind)) {
	case "failure_analysis", "bug_analysis", "audit":
		return false
	}

	switch strings.ToLower(strings.TrimSpace(profile.ExecutionMode)) {
	case "openhands":
		return false
	}

	return true
}

func looksLikeBootstrapTask(profile *taskDeliveryProfile) bool {
	text := strings.ToLower(strings.TrimSpace(profile.Name + " " + profile.Description))
	if text == "" {
		return false
	}

	keywords := []string{
		"脚手架",
		"初始化",
		"bootstrap",
		"scaffold",
		"init project",
		"init repo",
		"项目骨架",
		"基础框架",
		"搭建骨架",
		"创建前端",
		"创建后端",
		"create-vite",
		"gf init",
	}
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func normalizeExplicitDeliveryResources(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, raw := range values {
		item := path.Clean(strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/"))
		switch {
		case item == "", item == ".", item == "/":
			continue
		case strings.HasPrefix(item, "/"), strings.HasPrefix(item, "../"), item == "..", strings.Contains(item, "/../"):
			return nil
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}

	return result
}

func isNarrowBaselineScopedChange(resources []string) bool {
	if len(resources) == 0 {
		return false
	}

	hasRoot := false
	scopes := make(map[string]struct{}, len(resources))
	for _, resource := range resources {
		scope := deliveryResourceScope(resource)
		if scope == "" {
			return false
		}
		if scope == "_root" {
			hasRoot = true
			continue
		}
		scopes[scope] = struct{}{}
	}

	if len(scopes) == 0 {
		return hasRoot
	}
	if len(scopes) == 1 {
		return true
	}
	return hasRoot && len(scopes) == 1
}

func deliveryResourceScope(resource string) string {
	resource = path.Clean(strings.ReplaceAll(strings.TrimSpace(resource), "\\", "/"))
	switch {
	case resource == "", resource == ".", resource == "/":
		return ""
	case strings.HasPrefix(resource, "/"), strings.HasPrefix(resource, "../"), resource == "..", strings.Contains(resource, "/../"):
		return ""
	case !strings.Contains(resource, "/"):
		return "_root"
	default:
		return strings.SplitN(resource, "/", 2)[0]
	}
}
