package chat

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"easymvp/app/mvp/internal/workflow/eventstream"
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

func TestInspectExperienceReviewerReadinessOK(t *testing.T) {
	t.Parallel()

	status, message := inspectExperienceReviewerReadiness(
		[]experienceReviewerPresetCheck{
			{CategoryCode: "software_dev", ModelID: 315100000000000007},
			{CategoryCode: "game_dev", ModelID: 315100000000000008},
		},
		map[int64]experienceReviewerModelCheck{
			315100000000000007: {Exists: true, Enabled: true, Name: "hunyuan-turbos"},
			315100000000000008: {Exists: true, Enabled: true, Name: "Tencent HY 2.0 Instruct"},
		},
	)
	if status != "ok" {
		t.Fatalf("status = %s, want ok; message=%s", status, message)
	}
	for _, fragment := range []string{
		"software_dev=hunyuan-turbos(315100000000000007)",
		"game_dev=Tencent HY 2.0 Instruct(315100000000000008)",
	} {
		if !strings.Contains(message, fragment) {
			t.Fatalf("message missing %q: %s", fragment, message)
		}
	}
}

func TestInspectExperienceReviewerReadinessMissingCategory(t *testing.T) {
	t.Parallel()

	status, message := inspectExperienceReviewerReadiness(
		[]experienceReviewerPresetCheck{
			{CategoryCode: "software_dev", ModelID: 315100000000000007},
		},
		map[int64]experienceReviewerModelCheck{
			315100000000000007: {Exists: true, Enabled: true, Name: "hunyuan-turbos"},
		},
	)
	if status != "error" {
		t.Fatalf("status = %s, want error; message=%s", status, message)
	}
	if !strings.Contains(message, "game_dev 缺少 experience_reviewer/max 默认预设") {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestInspectExperienceReviewerReadinessInvalidModel(t *testing.T) {
	t.Parallel()

	status, message := inspectExperienceReviewerReadiness(
		[]experienceReviewerPresetCheck{
			{CategoryCode: "software_dev", ModelID: 0},
			{CategoryCode: "game_dev", ModelID: 315100000000000007},
		},
		map[int64]experienceReviewerModelCheck{
			315100000000000007: {Exists: true, Enabled: false, Name: "hunyuan-turbos"},
		},
	)
	if status != "error" {
		t.Fatalf("status = %s, want error; message=%s", status, message)
	}
	for _, fragment := range []string{
		"software_dev 预设不可用：model_id=0",
		"game_dev 预设不可用：model_id=315100000000000007 已禁用",
	} {
		if !strings.Contains(message, fragment) {
			t.Fatalf("message missing %q: %s", fragment, message)
		}
	}
}

func TestSummarizeWorkflowEventConsumerSnapshotCreatedButNotStarted(t *testing.T) {
	status, message := summarizeWorkflowEventConsumerSnapshot(eventstream.RuntimeSnapshot{
		ConsumerCreated: true,
		StreamName:      "test-stream",
		ConsumerGroup:   "group-a",
		ConsumerName:    "consumer-a",
	})
	if status != "warning" {
		t.Fatalf("status = %s, want warning; message=%s", status, message)
	}
	if !strings.Contains(message, "consumer 已创建但未启动") {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestSummarizeWorkflowEventConsumerSnapshotHealthy(t *testing.T) {
	status, message := summarizeWorkflowEventConsumerSnapshot(eventstream.RuntimeSnapshot{
		ConsumerCreated:   true,
		ConsumerStarted:   true,
		StreamName:        "test-stream",
		ConsumerGroup:     "group-a",
		ConsumerName:      "consumer-a",
		PendingKnown:      true,
		Pending:           2,
		LagKnown:          true,
		Lag:               0,
		WorkerHeartbeatAt: time.Now(),
		StartedAt:         time.Now(),
		ReclaimedMessages: 1,
	})
	if status != "ok" {
		t.Fatalf("status = %s, want ok; message=%s", status, message)
	}
	if !strings.Contains(message, "pending=2") || !strings.Contains(message, "heartbeat=") {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestSummarizeWorkflowEventConsumerSnapshotWarnsOnStaleHeartbeat(t *testing.T) {
	status, message := summarizeWorkflowEventConsumerSnapshot(eventstream.RuntimeSnapshot{
		ConsumerCreated:   true,
		ConsumerStarted:   true,
		StreamName:        "test-stream",
		ConsumerGroup:     "group-a",
		ConsumerName:      "consumer-a",
		WorkerHeartbeatAt: time.Now().Add(-31 * time.Second),
		StartedAt:         time.Now().Add(-2 * time.Minute),
	})
	if status != "warning" {
		t.Fatalf("status = %s, want warning; message=%s", status, message)
	}
	if !strings.Contains(message, "heartbeat=") {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestInspectWorkflowEventDurableSchemaReady(t *testing.T) {
	originalEventProbe := inspectWorkflowEventMetadataColumnsFn
	originalLedgerProbe := inspectWorkflowEventLedgerTableFn
	defer func() {
		inspectWorkflowEventMetadataColumnsFn = originalEventProbe
		inspectWorkflowEventLedgerTableFn = originalLedgerProbe
	}()

	inspectWorkflowEventMetadataColumnsFn = func(ctx context.Context) error { return nil }
	inspectWorkflowEventLedgerTableFn = func(ctx context.Context) error { return nil }

	status, message := inspectWorkflowEventDurableSchema(context.Background())
	if status != "ok" {
		t.Fatalf("status = %s, want ok; message=%s", status, message)
	}
	if !strings.Contains(message, "已就绪") {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestInspectWorkflowEventDurableSchemaWarnsOnMissingLedger(t *testing.T) {
	originalEventProbe := inspectWorkflowEventMetadataColumnsFn
	originalLedgerProbe := inspectWorkflowEventLedgerTableFn
	defer func() {
		inspectWorkflowEventMetadataColumnsFn = originalEventProbe
		inspectWorkflowEventLedgerTableFn = originalLedgerProbe
	}()

	inspectWorkflowEventMetadataColumnsFn = func(ctx context.Context) error { return nil }
	inspectWorkflowEventLedgerTableFn = func(ctx context.Context) error {
		return errors.New("Error 1146 (42S02): Table 'easymvp.mvp_workflow_event_ledger' doesn't exist")
	}

	status, message := inspectWorkflowEventDurableSchema(context.Background())
	if status != "warning" {
		t.Fatalf("status = %s, want warning; message=%s", status, message)
	}
	if !strings.Contains(message, "durable ledger 表未就绪") {
		t.Fatalf("unexpected message: %s", message)
	}
}
