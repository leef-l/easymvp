package workspace

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncWorktreeCommitCherryPicksBackToMain(t *testing.T) {
	mainDir := t.TempDir()
	runGit(t, mainDir, "init")
	runGit(t, mainDir, "config", "user.name", "Test User")
	runGit(t, mainDir, "config", "user.email", "test@example.com")

	mainFile := filepath.Join(mainDir, "README.md")
	if err := os.WriteFile(mainFile, []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write main file: %v", err)
	}
	runGit(t, mainDir, "add", "README.md")
	runGit(t, mainDir, "commit", "-m", "init")

	worktreePath := filepath.Join(mainDir, ".mvp-worktrees", "task-1")
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		t.Fatalf("mkdir worktree parent: %v", err)
	}
	runGit(t, mainDir, "worktree", "add", "-b", "mvp-task-1", worktreePath, "HEAD")

	if err := os.WriteFile(filepath.Join(worktreePath, "README.md"), []byte("hello world\n"), 0644); err != nil {
		t.Fatalf("update tracked file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worktreePath, "new.txt"), []byte("new file\n"), 0644); err != nil {
		t.Fatalf("create untracked file: %v", err)
	}

	if err := syncWorktreeCommit(context.Background(), mainDir, worktreePath, 1); err != nil {
		t.Fatalf("syncWorktreeCommit() error = %v", err)
	}

	readmeContent, err := os.ReadFile(filepath.Join(mainDir, "README.md"))
	if err != nil {
		t.Fatalf("read synced README: %v", err)
	}
	if got := string(readmeContent); got != "hello world\n" {
		t.Fatalf("README content = %q", got)
	}

	newContent, err := os.ReadFile(filepath.Join(mainDir, "new.txt"))
	if err != nil {
		t.Fatalf("read synced new file: %v", err)
	}
	if got := string(newContent); got != "new file\n" {
		t.Fatalf("new.txt content = %q", got)
	}

	logOutput := runGit(t, mainDir, "log", "--oneline", "-1")
	if !strings.Contains(logOutput, "mvp task 1: apply workspace changes") {
		t.Fatalf("unexpected latest commit: %q", logOutput)
	}
}

func TestSyncWorktreeCommitFallsBackWhenMainHasUnrelatedDirtyChanges(t *testing.T) {
	mainDir := t.TempDir()
	runGit(t, mainDir, "init")
	runGit(t, mainDir, "config", "user.name", "Test User")
	runGit(t, mainDir, "config", "user.email", "test@example.com")

	if err := os.WriteFile(filepath.Join(mainDir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write main file: %v", err)
	}
	runGit(t, mainDir, "add", "README.md")
	runGit(t, mainDir, "commit", "-m", "init")

	if err := os.WriteFile(filepath.Join(mainDir, "notes.txt"), []byte("dirty\n"), 0644); err != nil {
		t.Fatalf("write dirty file: %v", err)
	}

	worktreePath := filepath.Join(mainDir, ".mvp-worktrees", "task-2")
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		t.Fatalf("mkdir worktree parent: %v", err)
	}
	runGit(t, mainDir, "worktree", "add", "-b", "mvp-task-2", worktreePath, "HEAD")

	if err := os.WriteFile(filepath.Join(worktreePath, "README.md"), []byte("hello from task\n"), 0644); err != nil {
		t.Fatalf("update tracked file: %v", err)
	}

	if err := syncWorktreeCommit(context.Background(), mainDir, worktreePath, 2); err != nil {
		t.Fatalf("syncWorktreeCommit() error = %v", err)
	}

	readmeContent, err := os.ReadFile(filepath.Join(mainDir, "README.md"))
	if err != nil {
		t.Fatalf("read synced README: %v", err)
	}
	if got := string(readmeContent); got != "hello from task\n" {
		t.Fatalf("README content = %q", got)
	}

	dirtyContent, err := os.ReadFile(filepath.Join(mainDir, "notes.txt"))
	if err != nil {
		t.Fatalf("read dirty file: %v", err)
	}
	if got := string(dirtyContent); got != "dirty\n" {
		t.Fatalf("notes.txt content = %q", got)
	}
}

func TestEnsureRepositoryBaselineCreatesInitialCommitForEmptyRepo(t *testing.T) {
	mainDir := t.TempDir()
	runGit(t, mainDir, "init")

	if err := os.WriteFile(filepath.Join(mainDir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write initial file: %v", err)
	}

	headRef, err := ensureRepositoryBaseline(context.Background(), mainDir)
	if err != nil {
		t.Fatalf("ensureRepositoryBaseline() error = %v", err)
	}
	if strings.TrimSpace(headRef) == "" {
		t.Fatalf("expected non-empty head ref")
	}

	if got := runGit(t, mainDir, "rev-list", "--count", "HEAD"); got != "1" {
		t.Fatalf("unexpected commit count: %q", got)
	}
	if got := runGit(t, mainDir, "show", "HEAD:README.md"); got != "hello" {
		t.Fatalf("unexpected committed README content: %q", got)
	}
}

func TestEnsureRepositoryBaselineKeepsExistingHead(t *testing.T) {
	mainDir := t.TempDir()
	runGit(t, mainDir, "init")
	runGit(t, mainDir, "config", "user.name", "Test User")
	runGit(t, mainDir, "config", "user.email", "test@example.com")

	if err := os.WriteFile(filepath.Join(mainDir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write initial file: %v", err)
	}
	runGit(t, mainDir, "add", "README.md")
	runGit(t, mainDir, "commit", "-m", "init")

	originalHead := runGit(t, mainDir, "rev-parse", "HEAD")
	headRef, err := ensureRepositoryBaseline(context.Background(), mainDir)
	if err != nil {
		t.Fatalf("ensureRepositoryBaseline() error = %v", err)
	}
	if headRef != originalHead {
		t.Fatalf("head changed unexpectedly: got %q want %q", headRef, originalHead)
	}
	if got := runGit(t, mainDir, "rev-list", "--count", "HEAD"); got != "1" {
		t.Fatalf("unexpected commit count after baseline ensure: %q", got)
	}
}

func TestGitDiffStatIncludesUntrackedFiles(t *testing.T) {
	mainDir := t.TempDir()
	runGit(t, mainDir, "init")
	runGit(t, mainDir, "config", "user.name", "Test User")
	runGit(t, mainDir, "config", "user.email", "test@example.com")

	if err := os.WriteFile(filepath.Join(mainDir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write initial file: %v", err)
	}
	runGit(t, mainDir, "add", "README.md")
	runGit(t, mainDir, "commit", "-m", "init")

	worktreePath := filepath.Join(mainDir, ".mvp-worktrees", "task-3")
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		t.Fatalf("mkdir worktree parent: %v", err)
	}
	runGit(t, mainDir, "worktree", "add", "-b", "mvp-task-3", worktreePath, "HEAD")

	if err := os.WriteFile(filepath.Join(worktreePath, "README.md"), []byte("hello stat\n"), 0644); err != nil {
		t.Fatalf("update tracked file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worktreePath, "new.txt"), []byte("new\n"), 0644); err != nil {
		t.Fatalf("create untracked file: %v", err)
	}

	diffStat, err := gitDiffStat(worktreePath)
	if err != nil {
		t.Fatalf("gitDiffStat() error = %v", err)
	}
	if !strings.Contains(diffStat, "README.md") {
		t.Fatalf("diffStat missing tracked file: %q", diffStat)
	}
	if !strings.Contains(diffStat, "new.txt") {
		t.Fatalf("diffStat missing untracked file: %q", diffStat)
	}
}

func TestGitDiffPatchIncludesUntrackedFiles(t *testing.T) {
	mainDir := t.TempDir()
	runGit(t, mainDir, "init")
	runGit(t, mainDir, "config", "user.name", "Test User")
	runGit(t, mainDir, "config", "user.email", "test@example.com")

	if err := os.WriteFile(filepath.Join(mainDir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write initial file: %v", err)
	}
	runGit(t, mainDir, "add", "README.md")
	runGit(t, mainDir, "commit", "-m", "init")

	worktreePath := filepath.Join(mainDir, ".mvp-worktrees", "task-4")
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		t.Fatalf("mkdir worktree parent: %v", err)
	}
	runGit(t, mainDir, "worktree", "add", "-b", "mvp-task-4", worktreePath, "HEAD")

	if err := os.WriteFile(filepath.Join(worktreePath, "README.md"), []byte("hello patch\n"), 0644); err != nil {
		t.Fatalf("update tracked file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worktreePath, "new.txt"), []byte("new\n"), 0644); err != nil {
		t.Fatalf("create untracked file: %v", err)
	}

	patchContent, hasPatch, err := gitDiffPatch(worktreePath)
	if err != nil {
		t.Fatalf("gitDiffPatch() error = %v", err)
	}
	if !hasPatch {
		t.Fatal("expected patch content")
	}
	if !strings.Contains(patchContent, "diff --git a/README.md b/README.md") {
		t.Fatalf("patch missing tracked file diff: %q", patchContent)
	}
	if !strings.Contains(patchContent, "diff --git a/new.txt b/new.txt") {
		t.Fatalf("patch missing untracked file diff: %q", patchContent)
	}
}

func TestResolveMainWorkDir(t *testing.T) {
	t.Parallel()

	worktreePath := filepath.Join("/tmp/demo", ".mvp-worktrees", "task-42")
	if got := resolveMainWorkDir(worktreePath); got != filepath.Join("/tmp/demo") {
		t.Fatalf("resolveMainWorkDir() = %q", got)
	}
}

func TestIsGitRepoRejectsParentRepoSubdirectory(t *testing.T) {
	mainDir := t.TempDir()
	runGit(t, mainDir, "init")

	projectDir := filepath.Join(mainDir, "nested", "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	if isGitRepo(projectDir) {
		t.Fatalf("expected nested subdirectory inside parent repo to be treated as non-repo")
	}
}

func TestIsGitRepoAcceptsRepositoryRoot(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")

	if !isGitRepo(repoDir) {
		t.Fatalf("expected repo root to be treated as git repo")
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
	return strings.TrimSpace(string(output))
}
