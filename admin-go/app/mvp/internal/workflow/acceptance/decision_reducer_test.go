package acceptance

import (
	"context"
	"strings"
	"testing"
)

func TestRequiresManualReview(t *testing.T) {
	t.Parallel()

	if requiresManualReview(nil) {
		t.Fatal("nil hits should not require manual review")
	}
	if requiresManualReview([]RuleHit{{RuleCode: "software.no_failed_tasks"}}) {
		t.Fatal("unrelated rules should not require manual review")
	}
	if !requiresManualReview([]RuleHit{{RuleCode: "software.delivery_review_required"}}) {
		t.Fatal("delivery review rule should require manual review")
	}
}

func TestDecisionReducerReduceBySeverity(t *testing.T) {
	t.Parallel()

	reducer := NewDecisionReducer(nil, nil)
	result, err := reducer.Reduce(context.Background(), &AcceptContext{ProjectType: "software_dev"}, []RuleHit{
		{RuleCode: "software.warning", Severity: SeverityWarn},
	}, nil)
	if err != nil {
		t.Fatalf("Reduce() error = %v", err)
	}
	if result.Decision != DecisionPassed {
		t.Fatalf("decision = %s, want %s", result.Decision, DecisionPassed)
	}
	if result.Score != 95 {
		t.Fatalf("score = %.1f, want 95", result.Score)
	}

	result, err = reducer.Reduce(context.Background(), &AcceptContext{ProjectType: "software_dev"}, []RuleHit{
		{RuleCode: "software.error", Severity: SeverityError},
	}, nil)
	if err != nil {
		t.Fatalf("Reduce() error = %v", err)
	}
	if result.Decision != DecisionFailed {
		t.Fatalf("decision = %s, want %s", result.Decision, DecisionFailed)
	}
	if result.Score != 40 {
		t.Fatalf("score = %.1f, want 40", result.Score)
	}
}

func TestDecisionReducerManualReviewAndDowngrade(t *testing.T) {
	prevGlobal := manualReviewGloballyEnabled
	prevProjectType := manualReviewProjectTypeEnabled
	t.Cleanup(func() {
		manualReviewGloballyEnabled = prevGlobal
		manualReviewProjectTypeEnabled = prevProjectType
	})

	manualReviewGloballyEnabled = func(context.Context) bool { return true }
	manualReviewProjectTypeEnabled = func(context.Context, string) bool { return true }

	reducer := NewDecisionReducer(nil, nil)
	result, err := reducer.Reduce(context.Background(), &AcceptContext{ProjectType: "software_dev"}, []RuleHit{
		{RuleCode: "software.delivery_review_required", Severity: SeverityInfo},
	}, nil)
	if err != nil {
		t.Fatalf("Reduce() error = %v", err)
	}
	if result.Decision != DecisionManualReview {
		t.Fatalf("decision = %s, want %s", result.Decision, DecisionManualReview)
	}
	if result.Score != 85 {
		t.Fatalf("score = %.1f, want 85", result.Score)
	}
	if !strings.Contains(result.Summary, "交付结果命中人工审核规则") {
		t.Fatalf("summary missing manual review marker: %s", result.Summary)
	}

	manualReviewGloballyEnabled = func(context.Context) bool { return false }
	result, err = reducer.Reduce(context.Background(), &AcceptContext{ProjectType: "software_dev"}, []RuleHit{
		{RuleCode: "software.delivery_review_required", Severity: SeverityInfo},
	}, nil)
	if err != nil {
		t.Fatalf("Reduce() error = %v", err)
	}
	if result.Decision != DecisionPassed {
		t.Fatalf("decision = %s, want %s", result.Decision, DecisionPassed)
	}
	if !strings.Contains(result.Summary, "自动放行") {
		t.Fatalf("summary missing auto pass marker: %s", result.Summary)
	}
}
