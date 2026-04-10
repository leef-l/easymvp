package accept

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func TestManualCompleteBypassContext(t *testing.T) {
	t.Parallel()

	if HasManualCompleteBypass(nil) {
		t.Fatal("nil context should not bypass")
	}
	if HasManualCompleteBypass(context.Background()) {
		t.Fatal("background context should not bypass")
	}

	ctx := WithManualCompleteBypass(context.Background())
	if !HasManualCompleteBypass(ctx) {
		t.Fatal("expected manual bypass flag")
	}
}

func TestAcceptIssueRecordsToRuleHitsPreservesDomainTaskID(t *testing.T) {
	t.Parallel()

	hits := acceptIssueRecordsToRuleHits([]g.Map{
		{
			"rule_code":        "software.delivery_review_required",
			"issue_type":       "process",
			"severity":         "warn",
			"title":            "任务需要人工审核",
			"detail":           "命中高风险规则",
			"expected_value":   "自动或人工确认",
			"actual_value":     "risk=high",
			"suggested_action": "继续返工",
			"domain_task_id":   int64(12345),
			"resource_ref":     "/tmp/task.patch",
		},
	})
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].DomainTaskID != 12345 {
		t.Fatalf("DomainTaskID = %d, want 12345", hits[0].DomainTaskID)
	}
	if hits[0].RuleType != "process" {
		t.Fatalf("RuleType = %q, want process", hits[0].RuleType)
	}
	if hits[0].ResourceRef != "/tmp/task.patch" {
		t.Fatalf("ResourceRef = %q, want /tmp/task.patch", hits[0].ResourceRef)
	}
}
