package workspace

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

func TestSyncWorktreeCommitFallsBackToCopyOnCherryPickConflict(t *testing.T) {
	mainDir := t.TempDir()
	runGit(t, mainDir, "init")
	runGit(t, mainDir, "config", "user.name", "Test User")
	runGit(t, mainDir, "config", "user.email", "test@example.com")

	targetPath := filepath.Join(mainDir, "shared.txt")
	if err := os.WriteFile(targetPath, []byte("base\n"), 0644); err != nil {
		t.Fatalf("write shared file: %v", err)
	}
	runGit(t, mainDir, "add", "shared.txt")
	runGit(t, mainDir, "commit", "-m", "init")

	worktreeA := filepath.Join(mainDir, ".mvp-worktrees", "task-11")
	worktreeB := filepath.Join(mainDir, ".mvp-worktrees", "task-12")
	if err := os.MkdirAll(filepath.Dir(worktreeA), 0755); err != nil {
		t.Fatalf("mkdir worktree parent: %v", err)
	}
	runGit(t, mainDir, "worktree", "add", "-b", "mvp-task-11", worktreeA, "HEAD")
	runGit(t, mainDir, "worktree", "add", "-b", "mvp-task-12", worktreeB, "HEAD")

	if err := os.WriteFile(filepath.Join(worktreeA, "shared.txt"), []byte("first\n"), 0644); err != nil {
		t.Fatalf("write worktreeA file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worktreeB, "shared.txt"), []byte("second\n"), 0644); err != nil {
		t.Fatalf("write worktreeB file: %v", err)
	}

	if err := syncWorktreeCommit(context.Background(), mainDir, worktreeA, 11); err != nil {
		t.Fatalf("syncWorktreeCommit(worktreeA) error = %v", err)
	}
	if err := syncWorktreeCommit(context.Background(), mainDir, worktreeB, 12); err != nil {
		t.Fatalf("syncWorktreeCommit(worktreeB) error = %v", err)
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read synced file: %v", err)
	}
	if got := string(content); got != "second\n" {
		t.Fatalf("shared.txt content = %q", got)
	}

	status := runGit(t, mainDir, "status", "--short")
	var kept []string
	for _, line := range strings.Split(status, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "?? .mvp-worktrees/" {
			continue
		}
		kept = append(kept, line)
	}
	status = strings.Join(kept, "\n")
	if strings.TrimSpace(status) != "" {
		t.Fatalf("expected clean main worktree after fallback commit, got %q", status)
	}

	logOutput := runGit(t, mainDir, "log", "--oneline", "-1")
	if !strings.Contains(logOutput, "mvp task 12: apply workspace changes") {
		t.Fatalf("unexpected latest commit after fallback: %q", logOutput)
	}
}

func TestPatchChangedFiles(t *testing.T) {
	patchFile := filepath.Join(t.TempDir(), "task.patch")
	content := strings.Join([]string{
		"diff --git a/README.md b/README.md",
		"index 1111111..2222222 100644",
		"--- a/README.md",
		"+++ b/README.md",
		"@@ -1 +1 @@",
		"-old",
		"+new",
		"diff --git a/docs/old.md b/docs/new.md",
		"similarity index 100%",
		"rename from docs/old.md",
		"rename to docs/new.md",
		"diff --git a/tmp.txt b/tmp.txt",
		"deleted file mode 100644",
		"--- a/tmp.txt",
		"+++ /dev/null",
		"",
	}, "\n")
	if err := os.WriteFile(patchFile, []byte(content), 0644); err != nil {
		t.Fatalf("write patch file: %v", err)
	}

	files, err := patchChangedFiles(patchFile)
	if err != nil {
		t.Fatalf("patchChangedFiles() error = %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("patchChangedFiles() count = %d", len(files))
	}
	if files[0].Status != "M" || files[0].NewPath != "README.md" {
		t.Fatalf("unexpected modified file: %+v", files[0])
	}
	if files[1].Status != "R" || files[1].OldPath != "docs/old.md" || files[1].NewPath != "docs/new.md" {
		t.Fatalf("unexpected renamed file: %+v", files[1])
	}
	if files[2].Status != "D" || files[2].OldPath != "tmp.txt" || files[2].NewPath != "" {
		t.Fatalf("unexpected deleted file: %+v", files[2])
	}
}

func TestExtractAddedFileSnapshots(t *testing.T) {
	patchFile := filepath.Join(t.TempDir(), "task.patch")
	content := strings.Join([]string{
		"diff --git a/Makefile b/Makefile",
		"new file mode 100644",
		"--- /dev/null",
		"+++ b/Makefile",
		"@@ -0,0 +1,3 @@",
		"+build:",
		"+\tnode scripts/build.js all",
		"+",
		"diff --git a/scripts/build.js b/scripts/build.js",
		"new file mode 100644",
		"--- /dev/null",
		"+++ b/scripts/build.js",
		"@@ -0,0 +1,2 @@",
		"+console.log('ok')",
		"+",
		"",
	}, "\n")
	if err := os.WriteFile(patchFile, []byte(content), 0644); err != nil {
		t.Fatalf("write patch file: %v", err)
	}

	snapshots, err := extractAddedFileSnapshots(patchFile)
	if err != nil {
		t.Fatalf("extractAddedFileSnapshots() error = %v", err)
	}
	if got := string(snapshots["Makefile"]); got != "build:\n\tnode scripts/build.js all\n\n" {
		t.Fatalf("unexpected Makefile snapshot: %q", got)
	}
	if got := string(snapshots["scripts/build.js"]); got != "console.log('ok')\n\n" {
		t.Fatalf("unexpected scripts/build.js snapshot: %q", got)
	}
}

func TestValidateSyncBackPathsRejectsSuspiciousFile(t *testing.T) {
	changedFiles := []gitChangedFile{
		{Status: "A", NewPath: "运行方式："},
		{Status: "A", NewPath: "frontend/e2e/snake.spec.ts"},
	}

	err := validateSyncBackPaths(changedFiles, []string{"frontend/e2e/snake.spec.ts", "frontend/playwright.config.ts"})
	if err == nil {
		t.Fatal("expected suspicious syncBack path to be rejected")
	}
	if !strings.Contains(err.Error(), "检测到可疑文件: 运行方式：") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSyncBackPathsRejectsOutOfScopeFile(t *testing.T) {
	changedFiles := []gitChangedFile{
		{Status: "A", NewPath: "frontend/e2e/snake.spec.ts"},
		{Status: "A", NewPath: "frontend/extra.md"},
	}

	err := validateSyncBackPaths(changedFiles, []string{"frontend/e2e/snake.spec.ts", "frontend/playwright.config.ts"})
	if err == nil {
		t.Fatal("expected out-of-scope syncBack path to be rejected")
	}
	if !strings.Contains(err.Error(), "检测到越界修改: frontend/extra.md") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSyncBackPathsRejectsInternalWorkspaceMetadata(t *testing.T) {
	changedFiles := []gitChangedFile{
		{Status: "A", NewPath: ".mvp-worktrees/artifacts/task-1.patch"},
	}

	err := validateSyncBackPaths(changedFiles, nil)
	if err == nil {
		t.Fatal("expected internal workspace metadata to be rejected")
	}
	if !strings.Contains(err.Error(), "检测到内部工作区元数据: .mvp-worktrees/artifacts/task-1.patch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSyncBackPathsAllowsDecoratedAllowedPath(t *testing.T) {
	changedFiles := []gitChangedFile{
		{Status: "A", NewPath: "└── .gitignore         # 版本控制忽略规则"},
		{Status: "A", NewPath: "├── scripts/"},
	}

	err := validateSyncBackPaths(changedFiles, []string{".gitignore", "scripts/"})
	if err != nil {
		t.Fatalf("expected decorated allowed paths to pass, got %v", err)
	}
}

func TestValidateSyncBackPathsNormalizesDecoratedOutOfScopeFile(t *testing.T) {
	changedFiles := []gitChangedFile{
		{Status: "A", NewPath: "└── frontend/extra.md         # 越界文件"},
	}

	err := validateSyncBackPaths(changedFiles, []string{"frontend/e2e/snake.spec.ts", "frontend/playwright.config.ts"})
	if err == nil {
		t.Fatal("expected decorated out-of-scope syncBack path to be rejected")
	}
	if !strings.Contains(err.Error(), "检测到越界修改: frontend/extra.md") {
		t.Fatalf("unexpected error: %v", err)
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

func TestCommitSyncedFilesToMainStagesOnlyExplicitPaths(t *testing.T) {
	mainDir := t.TempDir()
	runGit(t, mainDir, "init")
	runGit(t, mainDir, "config", "user.name", "Test User")
	runGit(t, mainDir, "config", "user.email", "test@example.com")

	if err := os.WriteFile(filepath.Join(mainDir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	runGit(t, mainDir, "add", "README.md")
	runGit(t, mainDir, "commit", "-m", "init")

	if err := os.WriteFile(filepath.Join(mainDir, "README.md"), []byte("hello world\n"), 0644); err != nil {
		t.Fatalf("update README: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(mainDir, ".mvp-worktrees", "artifacts"), 0755); err != nil {
		t.Fatalf("mkdir internal metadata: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mainDir, ".mvp-worktrees", "artifacts", "task-9.patch"), []byte("diff --git a/x b/x\n"), 0644); err != nil {
		t.Fatalf("write internal metadata: %v", err)
	}

	if err := commitSyncedFilesToMain(mainDir, 9, []gitChangedFile{
		{Status: "M", NewPath: "README.md"},
	}); err != nil {
		t.Fatalf("commitSyncedFilesToMain() error = %v", err)
	}

	if got := strings.TrimSpace(runGit(t, mainDir, "show", "--name-only", "--pretty=format:", "HEAD")); got != "README.md" {
		t.Fatalf("unexpected committed files: %q", got)
	}
	if got := strings.TrimSpace(runGit(t, mainDir, "ls-files", ".mvp-worktrees")); got != "" {
		t.Fatalf("internal metadata should not be tracked, got %q", got)
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

func TestEnsureRepositoryBaselineConcurrentCallsShareOneInitialCommit(t *testing.T) {
	mainDir := t.TempDir()
	runGit(t, mainDir, "init")

	if err := os.WriteFile(filepath.Join(mainDir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write initial file: %v", err)
	}

	const workers = 4

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
		refs []string
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ref, err := ensureRepositoryBaseline(context.Background(), mainDir)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			refs = append(refs, ref)
		}()
	}
	wg.Wait()

	if len(errs) > 0 {
		t.Fatalf("ensureRepositoryBaseline() concurrent errors = %v", errs)
	}
	if len(refs) != workers {
		t.Fatalf("ensureRepositoryBaseline() refs = %d, want %d", len(refs), workers)
	}

	refSet := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		if strings.TrimSpace(ref) == "" {
			t.Fatal("expected non-empty head ref")
		}
		refSet[ref] = struct{}{}
	}
	if len(refSet) != 1 {
		t.Fatalf("expected one shared head ref, got %v", refs)
	}

	if got := runGit(t, mainDir, "rev-list", "--count", "HEAD"); got != "1" {
		t.Fatalf("unexpected commit count after concurrent baseline ensure: %q", got)
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

func TestIsBenignWorktreeRemoveErr(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err  error
		want bool
	}{
		{err: fmt.Errorf("fatal: '/tmp/demo' is not a working tree"), want: true},
		{err: fmt.Errorf("remove /tmp/demo: no such file or directory"), want: true},
		{err: fmt.Errorf("path does not exist"), want: true},
		{err: fmt.Errorf("permission denied"), want: false},
		{err: nil, want: false},
	}
	for _, tc := range cases {
		if got := isBenignWorktreeRemoveErr(tc.err); got != tc.want {
			t.Fatalf("isBenignWorktreeRemoveErr(%v) = %v, want %v", tc.err, got, tc.want)
		}
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
