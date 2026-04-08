package executor

import (
	"os"
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
	modelInfo := &engine.ModelInfo{
		ModelCode:    "claude-sonnet-4-5",
		ProviderType: "anthropic",
	}
	cmd := buildClaudeDefaultCommand("/tmp/worktree", "fix bug", modelInfo)
	for _, part := range []string{
		"'claude'",
		"'--add-dir'",
		"'/tmp/worktree'",
		"'--model'",
		"'claude-sonnet-4-5'",
		"'fix bug'",
	} {
		if !strings.Contains(cmd, part) {
			t.Fatalf("command %q missing %q", cmd, part)
		}
	}
	if os.Geteuid() == 0 {
		if strings.Contains(cmd, "'--dangerously-skip-permissions'") {
			t.Fatalf("root command %q should not contain permissions bypass flag", cmd)
		}
	} else if !strings.Contains(cmd, "'--dangerously-skip-permissions'") {
		t.Fatalf("non-root command %q missing permissions bypass flag", cmd)
	}
}

func TestBuildClaudeDefaultCommandFallsBackToLocalProfileForNonAnthropicProvider(t *testing.T) {
	modelInfo := &engine.ModelInfo{
		ModelCode:    "deepseek-chat",
		ProviderType: "openai_compatible",
	}
	cmd := buildClaudeDefaultCommand("/tmp/worktree", "fix bug", modelInfo)
	if strings.Contains(cmd, "'--model'") {
		t.Fatalf("command %q should not pin model for non-Claude provider", cmd)
	}
	if !strings.Contains(cmd, "'fix bug'") {
		t.Fatalf("command %q missing prompt", cmd)
	}
}

func TestBuildClaudeDefaultCommandUsesTencentCodingModel(t *testing.T) {
	modelInfo := &engine.ModelInfo{
		ModelCode:          "kimi-k2.5",
		ProviderType:       "tencent_coding",
		SupportedProtocols: []string{"anthropic", "openai_compatible"},
	}
	cmd := buildClaudeDefaultCommand("/tmp/worktree", "fix bug", modelInfo)
	for _, part := range []string{
		"'--model'",
		"'kimi-k2.5'",
	} {
		if !strings.Contains(cmd, part) {
			t.Fatalf("command %q missing %q", cmd, part)
		}
	}
}

func TestBuildClaudeDefaultCommandUsesSupportedAnthropicProtocol(t *testing.T) {
	modelInfo := &engine.ModelInfo{
		ModelCode:          "claude-sonnet-4-5",
		ProviderType:       "vendor_router",
		SupportedProtocols: []string{"anthropic", "openai_compatible"},
	}
	cmd := buildClaudeDefaultCommand("/tmp/worktree", "fix bug", modelInfo)
	if !strings.Contains(cmd, "'--model'") {
		t.Fatalf("command %q should pin model when anthropic is supported", cmd)
	}
}

func TestNormalizeClaudeBaseURLStripsAnthropicV1(t *testing.T) {
	got := normalizeClaudeBaseURL("https://api.lkeap.cloud.tencent.com/coding/anthropic/v1")
	want := "https://api.lkeap.cloud.tencent.com/coding/anthropic"
	if got != want {
		t.Fatalf("normalizeClaudeBaseURL() = %q, want %q", got, want)
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
