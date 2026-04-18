package chat

import "testing"

func TestBuildDeliveryReviewReasons(t *testing.T) {
	t.Parallel()

	reasons := buildDeliveryReviewReasons(taskWorkspaceMeta{
		DeliveryMode:   "pr",
		DeliveryStatus: "ready",
		SyncStatus:     "pending",
		RiskLevel:      "high",
	})
	if len(reasons) != 3 {
		t.Fatalf("expected 3 reasons, got %v", reasons)
	}
	want := []string{"PR 草稿待人工确认", "变更待人工回写", "高风险任务需人工复核"}
	for i, item := range want {
		if reasons[i] != item {
			t.Fatalf("reason[%d]=%q want %q", i, reasons[i], item)
		}
	}

	if got := buildDeliveryReviewReasons(taskWorkspaceMeta{
		DeliveryMode:   "manual",
		DeliveryStatus: "pending",
		SyncStatus:     "pending",
		RiskLevel:      "high",
	}); got != nil {
		t.Fatalf("non-ready delivery should not require review: %v", got)
	}
}

func TestDeliveryReviewRiskRank(t *testing.T) {
	t.Parallel()

	if deliveryReviewRiskRank("high") <= deliveryReviewRiskRank("medium") {
		t.Fatal("high risk should outrank medium")
	}
	if deliveryReviewRiskRank("medium") <= deliveryReviewRiskRank("low") {
		t.Fatal("medium risk should outrank low")
	}
	if deliveryReviewRiskRank("unknown") != 0 {
		t.Fatalf("unexpected rank for unknown risk: %d", deliveryReviewRiskRank("unknown"))
	}
}

func TestDecodeTaskReplayHandoffPayloadParsesSplitChildren(t *testing.T) {
	t.Parallel()

	mode, taskIDs := decodeTaskReplayHandoffPayload(`{"mode":"split_tasks","created_task_ids":[101,102,0]}`)
	if mode != "split_tasks" {
		t.Fatalf("unexpected mode: %q", mode)
	}
	if len(taskIDs) != 2 || int64(taskIDs[0]) != 101 || int64(taskIDs[1]) != 102 {
		t.Fatalf("unexpected task ids: %+v", taskIDs)
	}
}

func TestDecodeTaskReplayHandoffPayloadIgnoresInvalidJSON(t *testing.T) {
	t.Parallel()

	mode, taskIDs := decodeTaskReplayHandoffPayload("not-json")
	if mode != "" || len(taskIDs) != 0 {
		t.Fatalf("expected empty payload decode, got mode=%q ids=%+v", mode, taskIDs)
	}
}
