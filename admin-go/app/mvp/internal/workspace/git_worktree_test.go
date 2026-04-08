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

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
	return strings.TrimSpace(string(output))
}
