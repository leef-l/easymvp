package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanupAiderArtifactsRemovesInternalFiles(t *testing.T) {
	workDir := t.TempDir()

	for _, name := range []string{".aider.chat.history.md", ".aider.input.history"} {
		if err := os.WriteFile(filepath.Join(workDir, name), []byte("temp\n"), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(workDir, ".gitignore"), []byte(".aider*\n"), 0644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	if err := cleanupAiderArtifacts(workDir, []string{"cmd/server/"}); err != nil {
		t.Fatalf("cleanupAiderArtifacts() error = %v", err)
	}

	for _, name := range []string{".aider.chat.history.md", ".aider.input.history", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(workDir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, stat err=%v", name, err)
		}
	}
}

func TestCleanupAiderArtifactsKeepsAllowedGitignore(t *testing.T) {
	workDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(workDir, ".gitignore"), []byte(".aider*\n"), 0644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	if err := cleanupAiderArtifacts(workDir, []string{".gitignore"}); err != nil {
		t.Fatalf("cleanupAiderArtifacts() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(workDir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if string(content) != ".aider*\n" {
		t.Fatalf("unexpected .gitignore content: %q", string(content))
	}
}

func TestIsAiderArtifactGitignore(t *testing.T) {
	if !isAiderArtifactGitignore("# comment\n.aider*\n") {
		t.Fatalf("expected aider-only gitignore to be detected")
	}
	if isAiderArtifactGitignore(".aider*\nnode_modules/\n") {
		t.Fatalf("expected mixed gitignore rules to be preserved")
	}
}
