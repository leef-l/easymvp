package acceptance

import (
	"context"
	"os"
	"path/filepath"
	"testing"
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
