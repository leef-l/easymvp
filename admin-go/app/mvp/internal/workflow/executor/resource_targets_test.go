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
