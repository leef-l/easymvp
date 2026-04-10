package review

import (
	"testing"

	"easymvp/app/mvp/internal/workflow/qualitygate"
)

func TestFinalizeReviewConclusionBlocksWarnings(t *testing.T) {
	t.Parallel()

	got := finalizeReviewConclusion(true, 2, "审核通过")
	if got.passed {
		t.Fatalf("expected warnings to block execution, got %+v", got)
	}
	if !got.blockedByWarnings {
		t.Fatalf("expected blockedByWarnings to be true, got %+v", got)
	}
	wantSummary := "审核发现 2 条警告，需全部修复后才能进入执行阶段"
	if got.summary != wantSummary {
		t.Fatalf("summary = %q, want %q", got.summary, wantSummary)
	}
}

func TestFinalizeReviewConclusionAllowsCleanPass(t *testing.T) {
	t.Parallel()

	got := finalizeReviewConclusion(true, 0, "审核通过")
	if !got.passed {
		t.Fatalf("expected clean review to pass, got %+v", got)
	}
	if got.blockedByWarnings {
		t.Fatalf("expected blockedByWarnings to be false, got %+v", got)
	}
	if got.summary != "审核通过" {
		t.Fatalf("summary = %q, want %q", got.summary, "审核通过")
	}
}

func TestFinalizeReviewConclusionPreservesOriginalFailure(t *testing.T) {
	t.Parallel()

	got := finalizeReviewConclusion(false, 3, "系统预检发现错误")
	if got.passed {
		t.Fatalf("expected failed review to stay failed, got %+v", got)
	}
	if got.blockedByWarnings {
		t.Fatalf("expected blockedByWarnings to remain false when review already failed, got %+v", got)
	}
	if got.summary != "系统预检发现错误" {
		t.Fatalf("summary = %q, want %q", got.summary, "系统预检发现错误")
	}
}

func TestShouldWarnMissingBrowserVerificationPlan(t *testing.T) {
	t.Parallel()

	if !shouldWarnMissingBrowserVerificationPlan(qualitygate.VerificationStandard{RequireBrowserPlan: true}, false) {
		t.Fatal("expected coding interactive plan without browser verification to warn")
	}
	if shouldWarnMissingBrowserVerificationPlan(qualitygate.VerificationStandard{RequireBrowserPlan: true}, true) {
		t.Fatal("did not expect warning when browser verification plan exists")
	}
	if shouldWarnMissingBrowserVerificationPlan(qualitygate.VerificationStandard{RequireBrowserPlan: false}, false) {
		t.Fatal("did not expect analysis family to require browser verification plan")
	}
}
