package worktreeguard

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestValidateRejectsDeltaWhenAllowPathsEmpty(t *testing.T) {
	t.Parallel()

	repoDir := initGitRepo(t)
	writeFile(t, repoDir, "README.md", "# test\n")

	snapshot, err := Capture(context.Background(), repoDir)
	if err != nil {
		t.Fatalf("capture snapshot: %v", err)
	}

	writeFile(t, repoDir, "extra.txt", "unexpected\n")

	result, err := snapshot.Validate(context.Background(), repoDir, nil)
	if err != nil {
		t.Fatalf("validate snapshot: %v", err)
	}

	if len(result.Invalid) != 1 || result.Invalid[0] != "extra.txt" {
		t.Fatalf("expected extra.txt to be invalid, got %#v", result.Invalid)
	}
}

func TestValidateMarksSuspiciousTitleLikePath(t *testing.T) {
	t.Parallel()

	repoDir := initGitRepo(t)
	snapshot, err := Capture(context.Background(), repoDir)
	if err != nil {
		t.Fatalf("capture snapshot: %v", err)
	}

	writeFile(t, repoDir, "运行方式：", "bad\n")

	result, err := snapshot.Validate(context.Background(), repoDir, []string{"README.md"})
	if err != nil {
		t.Fatalf("validate snapshot: %v", err)
	}

	if len(result.Suspicious) != 1 || result.Suspicious[0] != "运行方式：" {
		t.Fatalf("expected suspicious path to be flagged, got %#v", result.Suspicious)
	}
	if len(result.Invalid) != 0 {
		t.Fatalf("expected suspicious path to bypass invalid list, got %#v", result.Invalid)
	}
}

func TestIsSuspiciousPathRejectsColonTitles(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{
		"main.go":   false,
		"README.md": false,
		"运行方式：":     true,
		"验证方式：":     true,
		"note:tmp":  true,
	}
	for input, want := range cases {
		if got := IsSuspiciousPath(input); got != want {
			t.Fatalf("IsSuspiciousPath(%q)=%v want %v", input, got, want)
		}
	}
}

func TestReadGitStatusTrimsQuotedUTF8Paths(t *testing.T) {
	t.Parallel()

	repoDir := initGitRepo(t)
	writeFile(t, repoDir, "含 空格.txt", "hello\n")

	status, err := readGitStatus(context.Background(), repoDir)
	if err != nil {
		t.Fatalf("read git status: %v", err)
	}

	if _, ok := status["含 空格.txt"]; !ok {
		t.Fatalf("expected UTF-8 path without quotes, got %#v", status)
	}
}

func TestPruneEmbeddedAllowedDuplicatesRemovesRepeatedRootPath(t *testing.T) {
	t.Parallel()

	repoDir := initGitRepo(t)
	workDir := filepath.Join(repoDir, "frontend")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir frontend: %v", err)
	}
	writeFile(t, repoDir, "backend/internal/cmd/cmd.go", "package cmd\n")
	writeFile(t, repoDir, "backend/backend/internal/cmd/cmd.go", "package cmd\n")

	pruned, err := PruneEmbeddedAllowedDuplicates(context.Background(), workDir, []string{"backend/internal/cmd/cmd.go"})
	if err != nil {
		t.Fatalf("PruneEmbeddedAllowedDuplicates() error = %v", err)
	}
	if len(pruned) != 1 || pruned[0] != "backend/backend/internal/cmd/cmd.go" {
		t.Fatalf("unexpected pruned paths: %#v", pruned)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "backend/backend/internal/cmd/cmd.go")); !os.IsNotExist(err) {
		t.Fatalf("expected duplicate path to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "backend/internal/cmd/cmd.go")); err != nil {
		t.Fatalf("expected allowed path to remain, stat err=%v", err)
	}
}

func TestPruneEmbeddedAllowedDuplicatesKeepsPathWhenAllowedTargetMissing(t *testing.T) {
	t.Parallel()

	repoDir := initGitRepo(t)
	workDir := filepath.Join(repoDir, "frontend")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir frontend: %v", err)
	}
	writeFile(t, repoDir, "backend/backend/internal/cmd/cmd.go", "package cmd\n")

	pruned, err := PruneEmbeddedAllowedDuplicates(context.Background(), workDir, []string{"backend/internal/cmd/cmd.go"})
	if err != nil {
		t.Fatalf("PruneEmbeddedAllowedDuplicates() error = %v", err)
	}
	if len(pruned) != 0 {
		t.Fatalf("expected nothing pruned when allowed target missing, got %#v", pruned)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "backend/backend/internal/cmd/cmd.go")); err != nil {
		t.Fatalf("expected duplicate path to remain, stat err=%v", err)
	}
}

func TestPruneEmbeddedAllowedDuplicatesRemovesRepeatedAncestorPath(t *testing.T) {
	t.Parallel()

	repoDir := initGitRepo(t)
	workDir := filepath.Join(repoDir, "frontend", "src", "components")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir frontend src components: %v", err)
	}
	writeFile(t, repoDir, "frontend/src/components/ControlPanel/ControlPanel.tsx", "export const ControlPanel = () => null\n")
	writeFile(t, repoDir, "frontend/src/components/frontend/src/components/ControlPanel/ControlPanel.tsx", "export const ControlPanel = () => null\n")

	pruned, err := PruneEmbeddedAllowedDuplicates(context.Background(), workDir, []string{"frontend/src/components/ControlPanel/ControlPanel.tsx"})
	if err != nil {
		t.Fatalf("PruneEmbeddedAllowedDuplicates() error = %v", err)
	}
	if len(pruned) != 1 || pruned[0] != "frontend/src/components/frontend/src/components/ControlPanel/ControlPanel.tsx" {
		t.Fatalf("unexpected pruned paths: %#v", pruned)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "frontend/src/components/frontend/src/components/ControlPanel/ControlPanel.tsx")); !os.IsNotExist(err) {
		t.Fatalf("expected duplicate nested path to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "frontend/src/components/ControlPanel/ControlPanel.tsx")); err != nil {
		t.Fatalf("expected allowed path to remain, stat err=%v", err)
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.name", "Codex")
	runGit(t, repoDir, "config", "user.email", "codex@example.com")
	return repoDir
}

func runGit(t *testing.T, repoDir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repoDir}, args...)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func writeFile(t *testing.T, repoDir, relPath, content string) {
	t.Helper()

	absPath := filepath.Join(repoDir, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", relPath, err)
	}
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}
