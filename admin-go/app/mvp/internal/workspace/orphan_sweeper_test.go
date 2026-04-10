package workspace

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestIsTaskWorktreeName(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{
		"task-1":       true,
		"task-42":      true,
		"task-0":       false,
		"task-x":       false,
		"task-":        false,
		"artifacts":    false,
		"tmp-task-123": false,
	}
	for input, want := range cases {
		if got := isTaskWorktreeName(input); got != want {
			t.Fatalf("isTaskWorktreeName(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestParseTaskIDFromWorktreeName(t *testing.T) {
	t.Parallel()

	if taskID, ok := parseTaskIDFromWorktreeName("task-123"); !ok || taskID != 123 {
		t.Fatalf("parseTaskIDFromWorktreeName(task-123) = (%d,%v), want (123,true)", taskID, ok)
	}
	for _, invalid := range []string{"task-0", "task-abc", "task-", "foo-1"} {
		if _, ok := parseTaskIDFromWorktreeName(invalid); ok {
			t.Fatalf("parseTaskIDFromWorktreeName(%q) expected invalid", invalid)
		}
	}
}

func TestScanDiskWorktreesOnlyIncludesTaskDirs(t *testing.T) {
	t.Parallel()

	mainDir := t.TempDir()
	root := filepath.Join(mainDir, worktreeDir)
	if err := os.MkdirAll(filepath.Join(root, "task-11"), 0755); err != nil {
		t.Fatalf("mkdir task-11: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "task-22"), 0755); err != nil {
		t.Fatalf("mkdir task-22: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "artifacts"), 0755); err != nil {
		t.Fatalf("mkdir artifacts: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "task-bad"), 0755); err != nil {
		t.Fatalf("mkdir task-bad: %v", err)
	}

	got, err := scanDiskWorktrees(mainDir)
	if err != nil {
		t.Fatalf("scanDiskWorktrees() error = %v", err)
	}

	want := []string{
		filepath.Join(root, "task-11"),
		filepath.Join(root, "task-22"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scanDiskWorktrees() = %#v, want %#v", got, want)
	}
}

func TestDefaultOrphanSweepConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultOrphanSweepConfig()
	if cfg.AutoCleanupDiskOrphans {
		t.Fatalf("AutoCleanupDiskOrphans should default false")
	}
	if !cfg.AutoRepairMissingOnDisk {
		t.Fatalf("AutoRepairMissingOnDisk should default true")
	}
	if !cfg.AutoRepairRunningMismatch {
		t.Fatalf("AutoRepairRunningMismatch should default true")
	}
}
