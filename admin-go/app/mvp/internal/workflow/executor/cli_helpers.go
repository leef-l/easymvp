package executor

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workspace"
	providerutil "easymvp/utility/provider"
)

func shellQuoteArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}

func shellQuoteArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, shellQuoteArg(arg))
	}
	return strings.Join(quoted, " ")
}

func resolveModelCode(modelInfo *engine.ModelInfo) string {
	if modelInfo == nil {
		return ""
	}
	return strings.TrimSpace(modelInfo.ModelCode)
}

func resolveModelBaseURL(modelInfo *engine.ModelInfo, fallback string) string {
	if modelInfo != nil && strings.TrimSpace(modelInfo.BaseURL) != "" {
		return strings.TrimSpace(modelInfo.BaseURL)
	}
	return strings.TrimSpace(fallback)
}

func resolveProtocolBaseURL(modelInfo *engine.ModelInfo, fallback string, protocol string) string {
	raw := resolveModelBaseURL(modelInfo, fallback)
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	if modelInfo == nil {
		return providerutil.ResolveCLIBaseURLForProtocol(providerutil.Config{BaseURL: raw}, protocol)
	}
	return providerutil.ResolveCLIBaseURLForProtocol(providerutil.Config{
		ProviderType:       modelInfo.ProviderType,
		SupportedProtocols: modelInfo.SupportedProtocols,
		BaseURL:            raw,
	}, protocol)
}

func shouldUseClaudeProjectCredentials(modelInfo *engine.ModelInfo) bool {
	if modelInfo == nil {
		return false
	}
	providerType := strings.ToLower(strings.TrimSpace(modelInfo.ProviderType))
	if providerutil.SupportsProtocol(modelInfo.ProviderType, modelInfo.SupportedProtocols, providerutil.TypeAnthropic) {
		return true
	}
	return strings.Contains(providerType, "anthropic") ||
		providerType == "tencent_coding" ||
		providerType == "baidu_coding"
}

func resolveClaudeModelCode(modelInfo *engine.ModelInfo) string {
	if !shouldUseClaudeProjectCredentials(modelInfo) {
		return ""
	}
	return resolveModelCode(modelInfo)
}

func normalizeClaudeBaseURL(raw string) string {
	return providerutil.ResolveCLIBaseURLForProtocol(providerutil.Config{BaseURL: raw}, providerutil.TypeAnthropic)
}

func shouldUseClaudePermissionBypass() bool {
	return os.Geteuid() != 0
}

func shouldAvoidClaudeCodeInCurrentEnv() bool {
	return os.Geteuid() == 0
}

func buildClaudeDefaultCommand(workDir, taskInstruction string, modelInfo *engine.ModelInfo) string {
	args := []string{
		"claude",
		"-p",
		"--output-format",
		"json",
	}
	if shouldUseClaudePermissionBypass() {
		args = append(args, "--dangerously-skip-permissions")
	}
	if strings.TrimSpace(workDir) != "" {
		args = append(args, "--add-dir", workDir)
	}
	if modelCode := resolveClaudeModelCode(modelInfo); modelCode != "" {
		args = append(args, "--model", modelCode)
	}
	args = append(args, strings.TrimSpace(taskInstruction))
	return shellQuoteArgs(args)
}

func buildCodexDefaultCommand(workDir, taskInstruction string, modelInfo *engine.ModelInfo) string {
	args := []string{
		"codex",
		"exec",
		"--full-auto",
	}
	if modelCode := resolveModelCode(modelInfo); modelCode != "" {
		args = append(args, "--model", modelCode)
	}
	if strings.TrimSpace(workDir) != "" {
		args = append(args, "--cd", workDir)
	}
	args = append(args, strings.TrimSpace(taskInstruction))
	return shellQuoteArgs(args)
}

func buildGeminiDefaultCommand(taskInstruction string, modelInfo *engine.ModelInfo) string {
	args := []string{
		"gemini",
		"-p",
		"--yolo",
	}
	if modelCode := resolveModelCode(modelInfo); modelCode != "" {
		args = append(args, "--model", modelCode)
	}
	args = append(args, strings.TrimSpace(taskInstruction))
	return shellQuoteArgs(args)
}

func finalizeWorkspaceSuccess(ctx context.Context, wsMgr workspace.Manager, taskID int64, executorName string) error {
	if wsMgr == nil {
		return nil
	}
	if err := wsMgr.Finalize(ctx, taskID, workspace.FinalizeRequest{Success: true}); err != nil {
		return fmt.Errorf("%s workspace finalize failed: %w", executorName, err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[%s] workspace cleanup panic: task=%d err=%v", executorName, taskID, r)
			}
		}()
		if cleanErr := wsMgr.Cleanup(context.Background(), taskID); cleanErr != nil {
			g.Log().Warningf(context.Background(), "[%s] workspace cleanup 失败: task=%d err=%v", executorName, taskID, cleanErr)
		}
	}()
	return nil
}
