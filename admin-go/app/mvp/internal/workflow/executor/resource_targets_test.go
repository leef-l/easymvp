package executor

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseResourceTargetsSplitsDirectoriesAndFiles(t *testing.T) {
	targets := parseResourceTargets(`["cmd/server/","internal/handler/","go.mod","README.md","  ../bad  "]`)

	if !reflect.DeepEqual(targets.AllowedPaths, []string{"cmd/server", "internal/handler", "go.mod", "README.md"}) {
		t.Fatalf("unexpected allowed paths: %#v", targets.AllowedPaths)
	}
	if !reflect.DeepEqual(targets.DirectoryPaths, []string{"cmd/server", "internal/handler"}) {
		t.Fatalf("unexpected directory paths: %#v", targets.DirectoryPaths)
	}
	if !reflect.DeepEqual(targets.FilePaths, []string{"go.mod", "README.md"}) {
		t.Fatalf("unexpected file paths: %#v", targets.FilePaths)
	}
	if !reflect.DeepEqual(targets.Rejected, []string{"../bad"}) {
		t.Fatalf("unexpected rejected paths: %#v", targets.Rejected)
	}
}

func TestEnsureDirectoryTargetsCreatesNestedDirectories(t *testing.T) {
	baseDir := t.TempDir()
	if err := ensureDirectoryTargets(baseDir, []string{"cmd/server", "internal/handler"}); err != nil {
		t.Fatalf("ensureDirectoryTargets() error = %v", err)
	}

	for _, relative := range []string{"cmd/server", "internal/handler"} {
		info, err := os.Stat(filepath.Join(baseDir, relative))
		if err != nil {
			t.Fatalf("stat %s: %v", relative, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", relative)
		}
	}
}

func TestApplyExecutionSubdirRebasesBackendTask(t *testing.T) {
	baseDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(baseDir, "backend", "manifest"), 0755); err != nil {
		t.Fatalf("mkdir backend manifest: %v", err)
	}
	targets := resourceTargets{
		AllowedPaths:   []string{"backend/go.mod", "backend/main.go", "backend/manifest/config.yaml"},
		FilePaths:      []string{"backend/go.mod", "backend/main.go", "backend/manifest/config.yaml"},
		DirectoryPaths: []string{"backend/manifest"},
	}

	gotDir, gotTargets := applyExecutionSubdir(baseDir, targets)
	wantDir := filepath.Join(baseDir, "backend")
	if gotDir != wantDir {
		t.Fatalf("applyExecutionSubdir() workdir = %q, want %q", gotDir, wantDir)
	}
	if !reflect.DeepEqual(gotTargets.AllowedPaths, []string{"backend/go.mod", "backend/main.go", "backend/manifest/config.yaml"}) {
		t.Fatalf("applyExecutionSubdir() allowed paths = %#v", gotTargets.AllowedPaths)
	}
	if !reflect.DeepEqual(gotTargets.FilePaths, []string{"go.mod", "main.go", "manifest/config.yaml"}) {
		t.Fatalf("applyExecutionSubdir() file paths = %#v", gotTargets.FilePaths)
	}
	if !reflect.DeepEqual(gotTargets.DirectoryPaths, []string{"manifest"}) {
		t.Fatalf("applyExecutionSubdir() directory paths = %#v", gotTargets.DirectoryPaths)
	}
}

func TestApplyExecutionSubdirKeepsRootWhenTargetDirMissing(t *testing.T) {
	baseDir := t.TempDir()
	targets := resourceTargets{
		AllowedPaths: []string{
			"backend/go.mod",
			"backend/main.go",
			"backend/internal/cmd/cmd.go",
		},
		FilePaths: []string{
			"backend/go.mod",
			"backend/main.go",
			"backend/internal/cmd/cmd.go",
		},
	}

	gotDir, gotTargets := applyExecutionSubdir(baseDir, targets)
	if gotDir != baseDir {
		t.Fatalf("applyExecutionSubdir() workdir = %q, want %q", gotDir, baseDir)
	}
	if !reflect.DeepEqual(gotTargets.AllowedPaths, targets.AllowedPaths) {
		t.Fatalf("applyExecutionSubdir() allowed paths = %#v, want %#v", gotTargets.AllowedPaths, targets.AllowedPaths)
	}
	if !reflect.DeepEqual(gotTargets.FilePaths, targets.FilePaths) {
		t.Fatalf("applyExecutionSubdir() file paths = %#v, want %#v", gotTargets.FilePaths, targets.FilePaths)
	}
}

func TestApplyExecutionSubdirRebasesFrontendCommonPrefix(t *testing.T) {
	baseDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(baseDir, "frontend", "src", "components"), 0755); err != nil {
		t.Fatalf("mkdir frontend src components: %v", err)
	}
	targets := resourceTargets{
		AllowedPaths: []string{
			"frontend/src/components/GameBoard.tsx",
			"frontend/src/components/GameControls.tsx",
			"frontend/src/hooks/useSnakeGame.ts",
		},
		FilePaths: []string{
			"frontend/src/components/GameBoard.tsx",
			"frontend/src/components/GameControls.tsx",
			"frontend/src/hooks/useSnakeGame.ts",
		},
	}

	gotDir, gotTargets := applyExecutionSubdir(baseDir, targets)
	wantDir := filepath.Join(baseDir, "frontend", "src")
	if gotDir != wantDir {
		t.Fatalf("applyExecutionSubdir() workdir = %q, want %q", gotDir, wantDir)
	}
	if !reflect.DeepEqual(gotTargets.AllowedPaths, []string{"frontend/src/components/GameBoard.tsx", "frontend/src/components/GameControls.tsx", "frontend/src/hooks/useSnakeGame.ts"}) {
		t.Fatalf("applyExecutionSubdir() allowed paths = %#v", gotTargets.AllowedPaths)
	}
	if !reflect.DeepEqual(gotTargets.FilePaths, []string{"components/GameBoard.tsx", "components/GameControls.tsx", "hooks/useSnakeGame.ts"}) {
		t.Fatalf("applyExecutionSubdir() file paths = %#v", gotTargets.FilePaths)
	}
}

func TestApplyExecutionSubdirDoesNotRebaseMultiRootTargets(t *testing.T) {
	baseDir := t.TempDir()
	targets := resourceTargets{
		AllowedPaths: []string{
			"frontend/src/App.tsx",
			"backend/main.go",
			"docs/README.md",
		},
		FilePaths: []string{
			"frontend/src/App.tsx",
			"backend/main.go",
			"docs/README.md",
		},
	}

	gotDir, gotTargets := applyExecutionSubdir(baseDir, targets)
	if gotDir != baseDir {
		t.Fatalf("applyExecutionSubdir() workdir = %q, want %q", gotDir, baseDir)
	}
	if !reflect.DeepEqual(gotTargets.AllowedPaths, targets.AllowedPaths) {
		t.Fatalf("applyExecutionSubdir() allowed paths = %#v, want %#v", gotTargets.AllowedPaths, targets.AllowedPaths)
	}
	if !reflect.DeepEqual(gotTargets.FilePaths, targets.FilePaths) {
		t.Fatalf("applyExecutionSubdir() file paths = %#v, want %#v", gotTargets.FilePaths, targets.FilePaths)
	}
}
