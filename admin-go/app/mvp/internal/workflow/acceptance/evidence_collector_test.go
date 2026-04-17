package acceptance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectCIArtifactFilesDetectsKnownFiles(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	ciDir := filepath.Join(workDir, ".easymvp", "ci")
	if err := os.MkdirAll(ciDir, 0o755); err != nil {
		t.Fatalf("mkdir ci dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ciDir, "latest.json"), []byte(`{"status":"passed","tool":"github_actions","pipeline":"backend","summary":"all green"}`), 0o644); err != nil {
		t.Fatalf("write latest.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, ".gitlab-ci.yml"), []byte("stages:\n  - test\n"), 0o644); err != nil {
		t.Fatalf("write gitlab ci: %v", err)
	}
	workflowDir := filepath.Join(workDir, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		t.Fatalf("mkdir workflow dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workflowDir, "backend.yml"), []byte("name: backend\n"), 0o644); err != nil {
		t.Fatalf("write github workflow: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workflowDir, "deploy.yaml"), []byte("name: deploy\n"), 0o644); err != nil {
		t.Fatalf("write github workflow yaml: %v", err)
	}

	items := collectCIArtifactFiles(workDir)
	if len(items) != 3 {
		t.Fatalf("expected 3 CI items, got %d: %+v", len(items), items)
	}
	if items[0].EvidenceType != "ci" || items[0].SourceType != "project_repo" {
		t.Fatalf("unexpected latest.json item: %+v", items[0])
	}
	if !strings.Contains(items[0].Summary, "CI 结果：status=passed tool=github_actions pipeline=backend summary=all green") {
		t.Fatalf("unexpected latest.json summary: %s", items[0].Summary)
	}
	if items[1].Summary != "检测到 CI 文件: .gitlab-ci.yml" {
		t.Fatalf("unexpected .gitlab-ci.yml summary: %s", items[1].Summary)
	}
	if items[2].Summary != "检测到 GitHub Actions 工作流 2 个" {
		t.Fatalf("unexpected workflow summary: %s", items[2].Summary)
	}
}

func TestCollectCIArtifactFilesFallsBackToRepoRootLatestJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	ciDir := filepath.Join(root, ".easymvp", "ci")
	if err := os.MkdirAll(ciDir, 0o755); err != nil {
		t.Fatalf("mkdir ci dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ciDir, "latest.json"), []byte(`{"status":"passed","tool":"github_actions","pipeline":"backend"}`), 0o644); err != nil {
		t.Fatalf("write latest.json: %v", err)
	}

	workDir := filepath.Join(root, "frontend", "apps", "web")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir work dir: %v", err)
	}

	items := collectCIArtifactFiles(workDir)
	if len(items) == 0 {
		t.Fatal("expected CI evidence items")
	}
	if items[0].ContentRef != filepath.Join(root, ".easymvp", "ci", "latest.json") {
		t.Fatalf("ContentRef = %q, want repo root latest.json", items[0].ContentRef)
	}
}

func TestCollectCIArtifactFilesFallsBackFromWorktreeToMainRepoLatestJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	ciDir := filepath.Join(root, ".easymvp", "ci")
	if err := os.MkdirAll(ciDir, 0o755); err != nil {
		t.Fatalf("mkdir ci dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ciDir, "latest.json"), []byte(`{"status":"passed","tool":"github_actions","pipeline":"web-antd-guard"}`), 0o644); err != nil {
		t.Fatalf("write latest.json: %v", err)
	}

	worktreePath := filepath.Join(root, ".mvp-worktrees", "task-42")
	if err := os.MkdirAll(worktreePath, 0o755); err != nil {
		t.Fatalf("mkdir worktree path: %v", err)
	}

	items := collectCIArtifactFiles(worktreePath)
	if len(items) == 0 {
		t.Fatal("expected CI evidence items")
	}
	if items[0].ContentRef != filepath.Join(root, ".easymvp", "ci", "latest.json") {
		t.Fatalf("ContentRef = %q, want main repo latest.json", items[0].ContentRef)
	}
}

func TestSummarizeCIJSONFallsBackOnInvalidPayload(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	path := filepath.Join(workDir, "latest.json")
	if err := os.WriteFile(path, []byte(`{broken`), 0o644); err != nil {
		t.Fatalf("write invalid json: %v", err)
	}

	if got := summarizeCIJSON(path); got != "检测到 CI 结果文件 latest.json" {
		t.Fatalf("unexpected summary: %s", got)
	}
}

func TestIsCIRelatedLog(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		action  string
		message string
		want    bool
	}{
		{name: "english build log", action: "build", message: "go test ./...", want: true},
		{name: "chinese static check log", action: "执行", message: "开始静态检查", want: true},
		{name: "non ci log", action: "chat_reply", message: "生成总结", want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isCIRelatedLog(tc.action, tc.message); got != tc.want {
				t.Fatalf("isCIRelatedLog(%q, %q) = %v, want %v", tc.action, tc.message, got, tc.want)
			}
		})
	}
}

func TestTrimSummary(t *testing.T) {
	t.Parallel()

	if got := trimSummary("  short text  ", 20); got != "short text" {
		t.Fatalf("trimSummary() = %q, want %q", got, "short text")
	}
	if got := trimSummary(" 0123456789ABC ", 10); got != "0123456789..." {
		t.Fatalf("trimSummary() truncated = %q, want %q", got, "0123456789...")
	}
}
