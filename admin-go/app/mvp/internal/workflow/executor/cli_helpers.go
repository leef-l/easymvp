package executor

import (
	"strings"

	"easymvp/app/mvp/internal/engine"
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

func buildClaudeDefaultCommand(workDir, taskInstruction string, modelInfo *engine.ModelInfo) string {
	args := []string{
		"claude",
		"-p",
		"--output-format",
		"json",
		"--dangerously-skip-permissions",
	}
	if strings.TrimSpace(workDir) != "" {
		args = append(args, "--add-dir", workDir)
	}
	if modelCode := resolveModelCode(modelInfo); modelCode != "" {
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
