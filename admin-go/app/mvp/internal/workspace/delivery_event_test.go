package workspace

import (
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func TestBuildDeliveryEventPayloadOmitsEmptyFields(t *testing.T) {
	t.Parallel()

	ws := &TaskWorkspace{
		ProjectID:     11,
		WorkflowRunID: 22,
		WorkspacePath: "/tmp/ws",
	}
	policy := deliveryPolicy{
		DeliveryMode: DeliveryModePatch,
		SyncStrategy: SyncStrategyAutoApply,
		RiskLevel:    RiskLevelLow,
	}
	payload := buildDeliveryEventPayload(ws, 33, policy, g.Map{
		"delivery_status": DeliveryStatusReady,
		"sync_status":     SyncStatusApplied,
		"patch_ref":       "/tmp/task.patch",
		"diff_summary":    " README.md | 2 +-",
	})

	if payload["project_id"] != int64(11) || payload["task_id"] != int64(33) {
		t.Fatalf("unexpected base payload: %+v", payload)
	}
	if payload["patch_ref"] != "/tmp/task.patch" || payload["diff_summary"] != " README.md | 2 +-" {
		t.Fatalf("expected patch fields in payload: %+v", payload)
	}
	if _, ok := payload["delivery_ref"]; ok {
		t.Fatalf("did not expect empty delivery_ref in payload: %+v", payload)
	}
	if _, ok := payload["delivery_title"]; ok {
		t.Fatalf("did not expect empty delivery_title in payload: %+v", payload)
	}
}

func TestShouldOpenDeliveryReviewGate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		policy         deliveryPolicy
		deliveryStatus string
		syncStatus     string
		want           bool
	}{
		{
			name: "patch auto applied stays closed",
			policy: deliveryPolicy{
				DeliveryMode: DeliveryModePatch,
				SyncStrategy: SyncStrategyAutoApply,
				RiskLevel:    RiskLevelLow,
			},
			deliveryStatus: DeliveryStatusReady,
			syncStatus:     SyncStatusApplied,
			want:           false,
		},
		{
			name: "pending sync opens gate",
			policy: deliveryPolicy{
				DeliveryMode: DeliveryModePatch,
				SyncStrategy: SyncStrategyManual,
				RiskLevel:    RiskLevelMedium,
			},
			deliveryStatus: DeliveryStatusReady,
			syncStatus:     SyncStatusPending,
			want:           true,
		},
		{
			name: "pr opens gate",
			policy: deliveryPolicy{
				DeliveryMode: DeliveryModePR,
				SyncStrategy: SyncStrategyManual,
				RiskLevel:    RiskLevelMedium,
			},
			deliveryStatus: DeliveryStatusReady,
			syncStatus:     SyncStatusPending,
			want:           true,
		},
		{
			name: "non ready stays closed",
			policy: deliveryPolicy{
				DeliveryMode: DeliveryModeManual,
				SyncStrategy: SyncStrategyManual,
				RiskLevel:    RiskLevelHigh,
			},
			deliveryStatus: DeliveryStatusPending,
			syncStatus:     SyncStatusPending,
			want:           false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldOpenDeliveryReviewGate(tc.policy, tc.deliveryStatus, tc.syncStatus); got != tc.want {
				t.Fatalf("shouldOpenDeliveryReviewGate() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBuildDeliveryReviewReason(t *testing.T) {
	t.Parallel()

	if got := buildDeliveryReviewReason(deliveryPolicy{DeliveryMode: DeliveryModePR}, map[string]interface{}{}); got != "已生成 PR 草稿交付物，等待人工审核或正式提交流程" {
		t.Fatalf("unexpected PR reason: %s", got)
	}
	if got := buildDeliveryReviewReason(deliveryPolicy{DeliveryMode: DeliveryModeManual}, map[string]interface{}{}); got != "当前任务要求人工处理交付结果" {
		t.Fatalf("unexpected manual reason: %s", got)
	}
	if got := buildDeliveryReviewReason(deliveryPolicy{RiskLevel: RiskLevelHigh}, map[string]interface{}{}); got != "高风险任务默认进入人工审核闸门" {
		t.Fatalf("unexpected high risk reason: %s", got)
	}
	if got := buildDeliveryReviewReason(deliveryPolicy{}, map[string]interface{}{"sync_status": SyncStatusPending}); got != "交付物已准备完成，等待人工确认回写主工作区" {
		t.Fatalf("unexpected pending sync reason: %s", got)
	}
}
