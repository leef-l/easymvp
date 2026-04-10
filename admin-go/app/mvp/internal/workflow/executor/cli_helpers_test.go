package executor

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workspace"
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

type workspaceFinalizeCall struct {
	ctxErr error
	req    workspace.FinalizeRequest
}

type fakeWorkspaceManager struct {
	mu                    sync.Mutex
	finalizeCalls         []workspaceFinalizeCall
	cleanupCtxErrs        []error
	cleanupDone           chan struct{}
	failOnSuccessFinalize bool
}

func (m *fakeWorkspaceManager) Prepare(ctx context.Context, req workspace.PrepareRequest) (*workspace.TaskWorkspace, error) {
	return nil, nil
}

func (m *fakeWorkspaceManager) MarkRunning(ctx context.Context, taskID int64) error {
	return nil
}

func (m *fakeWorkspaceManager) Get(ctx context.Context, taskID int64) (*workspace.TaskWorkspace, error) {
	return nil, nil
}

func (m *fakeWorkspaceManager) Finalize(ctx context.Context, taskID int64, req workspace.FinalizeRequest) error {
	m.mu.Lock()
	m.finalizeCalls = append(m.finalizeCalls, workspaceFinalizeCall{
		ctxErr: ctx.Err(),
		req:    req,
	})
	fail := m.failOnSuccessFinalize && req.Success
	m.mu.Unlock()

	if fail {
		return errors.New("finalize failed")
	}
	return nil
}

func (m *fakeWorkspaceManager) Cleanup(ctx context.Context, taskID int64) error {
	m.mu.Lock()
	m.cleanupCtxErrs = append(m.cleanupCtxErrs, ctx.Err())
	m.mu.Unlock()
	if m.cleanupDone != nil {
		select {
		case m.cleanupDone <- struct{}{}:
		default:
		}
	}
	return nil
}

func TestFinalizeWorkspaceSuccessUsesDetachedContexts(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mgr := &fakeWorkspaceManager{
		cleanupDone: make(chan struct{}, 1),
	}
	if err := finalizeWorkspaceSuccess(ctx, mgr, 99, "TestExecutor"); err != nil {
		t.Fatalf("finalizeWorkspaceSuccess() error = %v", err)
	}

	select {
	case <-mgr.cleanupDone:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup was not called")
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	if len(mgr.finalizeCalls) != 1 {
		t.Fatalf("expected 1 finalize call, got %d", len(mgr.finalizeCalls))
	}
	if mgr.finalizeCalls[0].ctxErr != nil {
		t.Fatalf("expected detached finalize context, got err=%v", mgr.finalizeCalls[0].ctxErr)
	}
	if !mgr.finalizeCalls[0].req.Success {
		t.Fatalf("expected success finalize request, got %#v", mgr.finalizeCalls[0].req)
	}
	if len(mgr.cleanupCtxErrs) != 1 || mgr.cleanupCtxErrs[0] != nil {
		t.Fatalf("expected detached cleanup context, got %#v", mgr.cleanupCtxErrs)
	}
}

func TestFinalizeWorkspaceSuccessFallsBackToFailureFinalize(t *testing.T) {
	t.Parallel()

	mgr := &fakeWorkspaceManager{
		failOnSuccessFinalize: true,
	}

	if err := finalizeWorkspaceSuccess(context.Background(), mgr, 101, "TestExecutor"); err == nil {
		t.Fatal("expected finalizeWorkspaceSuccess() to return error")
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	if len(mgr.finalizeCalls) != 2 {
		t.Fatalf("expected 2 finalize calls, got %d", len(mgr.finalizeCalls))
	}
	first := mgr.finalizeCalls[0]
	second := mgr.finalizeCalls[1]
	if !first.req.Success {
		t.Fatalf("expected first finalize to be success request, got %#v", first.req)
	}
	if second.req.Success {
		t.Fatalf("expected fallback finalize to be failure request, got %#v", second.req)
	}
	if !second.req.Retain {
		t.Fatalf("expected fallback finalize to retain workspace, got %#v", second.req)
	}
	if first.ctxErr != nil || second.ctxErr != nil {
		t.Fatalf("expected detached contexts, got first=%v second=%v", first.ctxErr, second.ctxErr)
	}
}
