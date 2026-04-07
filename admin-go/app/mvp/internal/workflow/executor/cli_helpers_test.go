package executor

import (
	"strings"
	"testing"

	"easymvp/app/mvp/internal/engine"
)

func TestShellQuoteArgEscapesSingleQuotes(t *testing.T) {
	got := shellQuoteArg("a'b")
	want := "'a'\\''b'"
	if got != want {
		t.Fatalf("shellQuoteArg() = %q, want %q", got, want)
	}
}

func TestBuildClaudeDefaultCommandUsesTaskModel(t *testing.T) {
	modelInfo := &engine.ModelInfo{ModelCode: "tc-code-latest"}
	cmd := buildClaudeDefaultCommand("/tmp/worktree", "fix bug", modelInfo)
	for _, part := range []string{
		"'claude'",
		"'--dangerously-skip-permissions'",
		"'--add-dir'",
		"'/tmp/worktree'",
		"'--model'",
		"'tc-code-latest'",
		"'fix bug'",
	} {
		if !strings.Contains(cmd, part) {
			t.Fatalf("command %q missing %q", cmd, part)
		}
	}
}

func TestBuildCodexDefaultCommandUsesExecAndTaskModel(t *testing.T) {
	modelInfo := &engine.ModelInfo{ModelCode: "gpt-5.1-codex"}
	cmd := buildCodexDefaultCommand("/tmp/worktree", "apply patch", modelInfo)
	for _, part := range []string{
		"'codex'",
		"'exec'",
		"'--full-auto'",
		"'--model'",
		"'gpt-5.1-codex'",
		"'--cd'",
		"'/tmp/worktree'",
		"'apply patch'",
	} {
		if !strings.Contains(cmd, part) {
			t.Fatalf("command %q missing %q", cmd, part)
		}
	}
}

func TestBuildGeminiDefaultCommandUsesTaskModel(t *testing.T) {
	modelInfo := &engine.ModelInfo{ModelCode: "gemini-2.5-pro"}
	cmd := buildGeminiDefaultCommand("review ui", modelInfo)
	for _, part := range []string{
		"'gemini'",
		"'-p'",
		"'--yolo'",
		"'--model'",
		"'gemini-2.5-pro'",
		"'review ui'",
	} {
		if !strings.Contains(cmd, part) {
			t.Fatalf("command %q missing %q", cmd, part)
		}
	}
}
