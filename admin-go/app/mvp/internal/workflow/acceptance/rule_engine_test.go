package acceptance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"easymvp/app/mvp/internal/workflow/qualitygate"
)

func TestRiskLevelAtLeast(t *testing.T) {
	t.Parallel()

	cases := []struct {
		actual    string
		threshold string
		want      bool
	}{
		{actual: "high", threshold: "medium", want: true},
		{actual: "medium", threshold: "medium", want: true},
		{actual: "low", threshold: "medium", want: false},
		{actual: "low", threshold: "", want: false},
		{actual: "", threshold: "low", want: false},
	}

	for _, tc := range cases {
		if got := riskLevelAtLeast(tc.actual, tc.threshold); got != tc.want {
			t.Fatalf("riskLevelAtLeast(%q, %q) = %v, want %v", tc.actual, tc.threshold, got, tc.want)
		}
	}
}

func TestNewRuleEngine(t *testing.T) {
	t.Parallel()

	engine := NewRuleEngine(nil)
	if engine == nil {
		t.Fatal("NewRuleEngine() returned nil")
	}
	if engine.ruleRepo != nil {
		t.Fatalf("expected nil repo, got %+v", engine.ruleRepo)
	}
}

func TestCheckRequiredFiles(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "README.md"), []byte("# demo\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	engine := &RuleEngine{}
	hits := engine.checkRequiredFiles(context.Background(), &AcceptContext{WorkDir: workDir},
		"software.required_file_exists", "required file", "artifact", "project",
		&ruleConfig{RequiredFiles: []string{"README.md", "docs/spec.md", "../escape.md"}},
	)

	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d: %+v", len(hits), hits)
	}
	if hits[0].Severity != SeverityError {
		t.Fatalf("severity = %s, want %s", hits[0].Severity, SeverityError)
	}
	if hits[0].Title != "必需文件 docs/spec.md 不存在" {
		t.Fatalf("unexpected title: %s", hits[0].Title)
	}
}

func TestCheckRequiredExtensions(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "summary.md"), []byte("done\n"), 0o644); err != nil {
		t.Fatalf("write summary.md: %v", err)
	}

	engine := &RuleEngine{}
	hits := engine.checkRequiredExtensions(context.Background(), &AcceptContext{WorkDir: workDir},
		"document.required_output_exists", "required output", "artifact", "project",
		&ruleConfig{RequiredExtensions: []string{".md", ".pdf"}},
	)

	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d: %+v", len(hits), hits)
	}
	if hits[0].Severity != SeverityError {
		t.Fatalf("severity = %s, want %s", hits[0].Severity, SeverityError)
	}
	if hits[0].Title != "未找到 .pdf 格式的文档产物" {
		t.Fatalf("unexpected title: %s", hits[0].Title)
	}
}

func TestCheckRequiredExtensionsWarnsOnReadDirError(t *testing.T) {
	t.Parallel()

	engine := &RuleEngine{}
	hits := engine.checkRequiredExtensions(context.Background(), &AcceptContext{WorkDir: filepath.Join(t.TempDir(), "missing")},
		"document.required_output_exists", "required output", "artifact", "project",
		&ruleConfig{RequiredExtensions: []string{".md"}},
	)

	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d: %+v", len(hits), hits)
	}
	if hits[0].Severity != SeverityWarn {
		t.Fatalf("severity = %s, want %s", hits[0].Severity, SeverityWarn)
	}
	if hits[0].SuggestedAction != "人工确认文档产物是否存在" {
		t.Fatalf("unexpected suggested action: %s", hits[0].SuggestedAction)
	}
}

func TestBuildDeliveryReviewReasons(t *testing.T) {
	t.Parallel()

	allowedModes := map[string]struct{}{
		"manual": {},
		"pr":     {},
	}
	allowedSyncStatuses := map[string]struct{}{
		"pending": {},
		"failed":  {},
	}

	if got := buildDeliveryReviewReasons("patch", "skipped", "skipped", "high", allowedModes, allowedSyncStatuses, "high"); got != nil {
		t.Fatalf("non-ready delivery should not require review, got %+v", got)
	}

	got := buildDeliveryReviewReasons("patch", "ready", "pending", "medium", allowedModes, allowedSyncStatuses, "high")
	if len(got) != 1 || got[0] != "回写状态=pending" {
		t.Fatalf("unexpected reasons for pending ready delivery: %+v", got)
	}

	got = buildDeliveryReviewReasons("manual", "ready", "applied", "high", allowedModes, allowedSyncStatuses, "high")
	if len(got) != 2 {
		t.Fatalf("expected 2 reasons for manual high-risk ready delivery, got %+v", got)
	}
}

func TestEvaluateVerificationSnapshot(t *testing.T) {
	t.Parallel()

	standard := qualitygate.VerificationStandard{
		Code:                      "coding.interactive_delivery",
		RequirePassedVerification: true,
		RequireBrowserEvidence:    true,
	}
	if hit := evaluateVerificationSnapshot(standard, verificationSnapshot{}); hit == nil || hit.RuleCode != "software.verification_required" {
		t.Fatalf("expected missing verification hit, got %+v", hit)
	}
	if hit := evaluateVerificationSnapshot(standard, verificationSnapshot{
		Present:            true,
		Status:             "completed",
		Decision:           DecisionPassed,
		HasBrowserEvidence: false,
	}); hit == nil || hit.RuleCode != "software.browser_verification_required" {
		t.Fatalf("expected browser verification hit, got %+v", hit)
	}
	if hit := evaluateVerificationSnapshot(standard, verificationSnapshot{
		Present:            true,
		Status:             "completed",
		Decision:           DecisionPassed,
		HasBrowserEvidence: true,
	}); hit != nil {
		t.Fatalf("expected verification snapshot to pass, got %+v", hit)
	}
}

func TestEvaluateRequiredProjectRoles(t *testing.T) {
	t.Parallel()

	prev := resolveRequiredProjectRole
	resolveRequiredProjectRole = func(ctx context.Context, projectID int64, requirement qualitygate.ProjectRoleRequirement) error {
		if requirement.RoleType == qualitygate.RoleTypeExperienceReviewer {
			return fmt.Errorf("分类 software_dev 没有 %s 的 V2 默认预设", requirement.RoleType)
		}
		return nil
	}
	t.Cleanup(func() {
		resolveRequiredProjectRole = prev
	})

	engine := &RuleEngine{}
	hits := engine.evaluateRequiredProjectRoles(context.Background(), &AcceptContext{ProjectID: 1}, qualitygate.VerificationStandard{
		DisplayName: "Coding Interactive Delivery Standard",
		RequiredProjectRoles: []qualitygate.ProjectRoleRequirement{{
			RoleType:    qualitygate.RoleTypeExperienceReviewer,
			DisplayName: "产品体验评审师",
			Purpose:     "验收关键用户路径",
			Blocking:    true,
		}},
	})
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d: %+v", len(hits), hits)
	}
	if hits[0].RuleCode != "software.required_project_role_missing" {
		t.Fatalf("unexpected rule code: %+v", hits[0])
	}
	if hits[0].Severity != SeverityError {
		t.Fatalf("unexpected severity: %+v", hits[0])
	}
}

func TestVerificationEvidenceCheckKind(t *testing.T) {
	t.Parallel()

	got := verificationEvidenceCheckKind(`{"name":"frontend e2e","command":"'pnpm' 'run' 'test:e2e'","runner":"local"}`, "")
	if got != qualitygate.CheckKindBrowser {
		t.Fatalf("verificationEvidenceCheckKind() = %q, want %q", got, qualitygate.CheckKindBrowser)
	}
}
