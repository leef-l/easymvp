package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/activity"
	providerutil "easymvp/utility/provider"
	"easymvp/utility/worktreeguard"
)

// AiderRunner 封装 Aider CLI 调用
// 根据任务角色配置，使用不同的模型和提示词调用 Aider 进行代码编辑
type AiderRunner struct{}

var defaultAiderRunner = &AiderRunner{}

// GetAiderRunner 获取全局 AiderRunner
func GetAiderRunner() *AiderRunner {
	return defaultAiderRunner
}

// AiderConfig Aider 运行配置
type AiderConfig struct {
	ModelCode            string        // 模型代码（如 tc-code-latest, glm-5）
	APIKey               string        // API Key
	BaseURL              string        // API Base URL（不带 /v1）
	ProviderType         string        // provider 类型（anthropic / openai_compatible）
	SupportedProtocols   []string      // 供应商支持的协议类型
	SystemPrompt         string        // 系统提示词
	WorkDir              string        // 工作目录（项目代码所在目录）
	Files                []string      // 需要编辑的文件列表
	AllowPaths           []string      // 允许创建/修改的路径（含目录）
	ReadFiles            []string      // 只读参考文件
	Message              string        // 任务指令
	MaxTokens            int           // 最大输出 token
	AutoCommit           bool          // 是否自动提交 git
	Timeout              time.Duration // 超时时间
	MaxSteps             int           // 最大内部执行步数
	MapTokens            int           // Repo map token 数
	MaxChatHistoryTokens int           // chat history token 上限
	CompactMode          bool          // 是否使用精简上下文模式
	OnActivity           func()        // 输出增量时更新活跃时间
}

// AiderResult Aider 执行结果
type AiderResult struct {
	Output      string        // 完整输出
	ExitCode    int           // 退出码
	Error       error         // 错误
	Duration    time.Duration // 耗时
	FailureHint string        // 归纳后的失败说明
	Category    taskFailureCategory
}

type aiderActivityWriter struct {
	buf        bytes.Buffer
	lastTouch  time.Time
	onActivity func()
}

func (w *aiderActivityWriter) Write(p []byte) (int, error) {
	n, err := w.buf.Write(p)
	if n > 0 && w.onActivity != nil {
		now := time.Now()
		if w.lastTouch.IsZero() || now.Sub(w.lastTouch) >= 2*time.Second {
			w.lastTouch = now
			w.onActivity()
		}
	}
	return n, err
}

func (w *aiderActivityWriter) String() string {
	return w.buf.String()
}

// Run 执行 Aider 任务
func (r *AiderRunner) Run(ctx context.Context, cfg *AiderConfig) *AiderResult {
	start := time.Now()
	maxSteps := cfg.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 2
	}

	result := r.runOnce(ctx, cfg)

	if result.Error != nil && r.isTokenLimitFailure(result.Output) && !cfg.CompactMode {
		if maxSteps <= 1 {
			g.Log().Warningf(ctx, "[AiderRunner] 检测到 token limit，但 maxSteps=%d，跳过精简重试", maxSteps)
		} else {
			g.Log().Warningf(ctx, "[AiderRunner] 检测到 token limit，准备使用精简上下文重试")
			compactCfg := r.buildCompactRetryConfig(cfg)
			retryResult := r.runOnce(ctx, compactCfg)
			retryResult.Output = strings.TrimSpace(result.Output) + "\n\n[AiderRunner] 检测到 token limit，已自动切换为精简上下文模式重试。\n\n" + strings.TrimSpace(retryResult.Output)
			retryResult.Duration = time.Since(start)
			if retryResult.Error == nil {
				return retryResult
			}
			result = retryResult
		}
	}

	// 对非 token-limit 错误进行分类和智能重试
	if result.Error != nil {
		category, retryable := r.classifyAiderError(result.Output, result.Error)
		result.Category = category

		if retryable && maxSteps > 1 {
			g.Log().Warningf(ctx, "[AiderRunner] 检测到可重试错误（类别=%s），等待 5 秒后重试", category)
			time.Sleep(5 * time.Second)
			retryResult := r.runOnce(ctx, cfg)
			retryResult.Output = strings.TrimSpace(result.Output) + "\n\n[AiderRunner] 检测到网络/超时错误，已自动重试。\n\n" + strings.TrimSpace(retryResult.Output)
			retryResult.Duration = time.Since(start)
			if retryResult.Error == nil {
				return retryResult
			}
			retryCategory, _ := r.classifyAiderError(retryResult.Output, retryResult.Error)
			retryResult.Category = retryCategory
			result = retryResult
		}
	}

	result.Duration = time.Since(start)
	result.FailureHint = r.buildFailureHint(result.Output, result.Error, cfg)

	g.Log().Infof(ctx, "[AiderRunner] 完成: 耗时=%v output_len=%d category=%s",
		result.Duration, len(result.Output), result.Category)

	return result
}

func (r *AiderRunner) runOnce(ctx context.Context, cfg *AiderConfig) *AiderResult {
	metadataFile, err := r.writeModelMetadata(cfg)
	if err != nil {
		g.Log().Warningf(ctx, "[AiderRunner] 生成 model metadata 失败: %v", err)
	}
	if metadataFile != "" {
		defer os.Remove(metadataFile)
	}

	messageFile, err := r.writeMessageFile(cfg.Message)
	if err != nil {
		g.Log().Warningf(ctx, "[AiderRunner] 写入 message file 失败，将回退为命令行参数: %v", err)
	}
	if messageFile != "" {
		defer os.Remove(messageFile)
	}

	args := r.buildArgs(cfg, metadataFile, messageFile)
	env := r.buildEnv(cfg)

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}

	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	if _, err := exec.LookPath("aider"); err == nil {
		cmd = exec.CommandContext(cmdCtx, "aider", args...)
	} else if _, err := exec.LookPath("uv"); err == nil {
		uvArgs := []string{"tool", "run", "--python", "3.12", "--from", "aider-chat", "aider"}
		uvArgs = append(uvArgs, args...)
		cmd = exec.CommandContext(cmdCtx, "uv", uvArgs...)
	} else {
		return &AiderResult{
			Error: fmt.Errorf("未找到 aider 可执行文件，且 uv 不可用"),
		}
	}
	cmd.Dir = cfg.WorkDir
	cmd.Env = append(cmd.Environ(), env...)
	GetCommandResourcePolicy(cmdCtx).Apply(cmd)

	stdout := &aiderActivityWriter{onActivity: cfg.OnActivity}
	stderr := &aiderActivityWriter{onActivity: cfg.OnActivity}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	g.Log().Infof(ctx, "[AiderRunner] 启动: model=%s workdir=%s files=%v",
		cfg.ModelCode, cfg.WorkDir, cfg.Files)

	err = cmd.Run()

	result := &AiderResult{
		Output: stdout.String() + stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Error = err
		g.Log().Warningf(ctx, "[AiderRunner] 退出 code=%d err=%v", result.ExitCode, err)
	}

	return result
}

// buildArgs 构建 Aider 命令行参数
func (r *AiderRunner) buildArgs(cfg *AiderConfig, metadataFile string, messageFile string) []string {
	mapTokens := cfg.MapTokens
	if mapTokens == 0 && !cfg.CompactMode {
		mapTokens = 512
	}

	maxChatHistoryTokens := cfg.MaxChatHistoryTokens
	if maxChatHistoryTokens == 0 {
		if cfg.CompactMode {
			maxChatHistoryTokens = 512
		} else {
			maxChatHistoryTokens = 2048
		}
	}

	args := []string{
		"--model", r.formatModel(cfg),
		"--encoding", "utf-8",
		"--no-auto-commits",
		"--no-gitignore",
		"--no-show-model-warnings",
		"--no-pretty",
		"--no-stream",
		"--no-browser",
		"--yes-always",
		"--chat-language", "Chinese",
		"--map-tokens", strconv.Itoa(mapTokens),
		"--max-chat-history-tokens", strconv.Itoa(maxChatHistoryTokens),
	}

	// 指定 model metadata 文件
	if metadataFile != "" {
		args = append(args, "--model-metadata-file", metadataFile)
	}

	if cfg.CompactMode {
		args = append(args, "--subtree-only")
	}

	// 自动提交
	if cfg.AutoCommit {
		// 替换 --no-auto-commits（索引2）为 --auto-commits
		for i, a := range args {
			if a == "--no-auto-commits" {
				args[i] = "--auto-commits"
				break
			}
		}
	}

	if messageFile != "" {
		args = append(args, "--message-file", messageFile)
	} else if cfg.Message != "" {
		args = append(args, "--message", cfg.Message)
	}

	// 需要编辑的文件
	for _, f := range cfg.Files {
		args = append(args, "--file", f)
	}

	// 只读参考文件
	for _, f := range cfg.ReadFiles {
		args = append(args, "--read", f)
	}

	return args
}

// buildEnv 构建环境变量
func (r *AiderRunner) buildEnv(cfg *AiderConfig) []string {
	var env []string

	// 强制 UTF-8 编码，避免 Windows GBK 环境下 Rich 库输出 Unicode 字符崩溃
	env = append(env,
		"PYTHONIOENCODING=utf-8",
		"PYTHONLEGACYWINDOWSSTDIO=0",
	)

	switch r.resolveProtocol(cfg) {
	case providerutil.TypeAnthropic:
		env = append(env,
			"ANTHROPIC_API_KEY="+cfg.APIKey,
			"ANTHROPIC_BASE_URL="+cfg.BaseURL,
		)
	case providerutil.TypeOpenAICompatible:
		env = append(env,
			"OPENAI_API_KEY="+cfg.APIKey,
			"OPENAI_API_BASE="+cfg.BaseURL,
		)
	default:
		// 兜底用 anthropic
		env = append(env,
			"ANTHROPIC_API_KEY="+cfg.APIKey,
			"ANTHROPIC_BASE_URL="+cfg.BaseURL,
		)
	}

	return GetCommandResourcePolicy(context.Background()).EnvSlice(env)
}

// writeModelMetadata 生成临时 model metadata JSON 文件
// 告诉 Aider 模型的真实 context window 大小，避免使用默认值误判 token limits
func (r *AiderRunner) writeModelMetadata(cfg *AiderConfig) (string, error) {
	modelName := r.formatModel(cfg)
	maxOutput := cfg.MaxTokens
	if maxOutput == 0 {
		maxOutput = 4096
	}

	// 从数据库查询模型的 context_window
	contextWindow := 0
	model, err := g.DB().Model("ai_model").Ctx(context.Background()).
		Fields("context_window").
		Where("model_code", cfg.ModelCode).
		WhereNull("deleted_at").
		One()
	if err == nil && !model.IsEmpty() {
		contextWindow = model["context_window"].Int()
	}
	if contextWindow == 0 {
		contextWindow = 128000 // 兜底默认值
	}

	metadata := map[string]map[string]int{
		modelName: {
			"max_tokens":        maxOutput,
			"max_input_tokens":  contextWindow,
			"max_output_tokens": maxOutput,
		},
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", err
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("aider-metadata-%d.json", time.Now().UnixNano()))
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return "", err
	}

	return tmpFile, nil
}

func (r *AiderRunner) writeMessageFile(message string) (string, error) {
	if strings.TrimSpace(message) == "" {
		return "", nil
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("aider-message-%d.txt", time.Now().UnixNano()))
	if err := os.WriteFile(tmpFile, []byte(message), 0600); err != nil {
		return "", err
	}
	return tmpFile, nil
}

// formatModel 格式化模型名称（Aider 需要 provider/model 格式）
func (r *AiderRunner) formatModel(cfg *AiderConfig) string {
	switch r.resolveProtocol(cfg) {
	case providerutil.TypeAnthropic:
		if !strings.HasPrefix(cfg.ModelCode, "anthropic/") {
			return "anthropic/" + cfg.ModelCode
		}
	case providerutil.TypeOpenAICompatible:
		if !strings.HasPrefix(cfg.ModelCode, "openai/") {
			return "openai/" + cfg.ModelCode
		}
	}
	return cfg.ModelCode
}

func (r *AiderRunner) resolveProtocol(cfg *AiderConfig) string {
	if cfg == nil {
		return providerutil.TypeAnthropic
	}
	return providerutil.ResolveProtocol(providerutil.Config{
		ProviderType:       cfg.ProviderType,
		SupportedProtocols: cfg.SupportedProtocols,
		BaseURL:            cfg.BaseURL,
	})
}

// BuildConfigFromModel 从数据库模型信息构建 AiderConfig
func (r *AiderRunner) BuildConfigFromModel(ctx context.Context, modelInfo *ModelInfo, workDir string) *AiderConfig {
	if modelInfo == nil {
		g.Log().Errorf(ctx, "[AiderRunner] BuildConfigFromModel: modelInfo 为 nil")
		return &AiderConfig{WorkDir: workDir}
	}
	protocol := providerutil.ResolveProtocol(providerutil.Config{
		ProviderType:       modelInfo.ProviderType,
		SupportedProtocols: modelInfo.SupportedProtocols,
		BaseURL:            modelInfo.BaseURL,
	})
	baseURL := providerutil.ResolveCLIBaseURLForProtocol(providerutil.Config{
		ProviderType:       modelInfo.ProviderType,
		SupportedProtocols: modelInfo.SupportedProtocols,
		BaseURL:            modelInfo.BaseURL,
	}, protocol)

	timeoutSeconds := GetConfigInt(ctx, "runtime.task_timeout_seconds", "engine.runtime.taskTimeoutSeconds", 600)
	maxSteps := GetConfigInt(ctx, "runtime.max_steps", "engine.runtime.maxSteps", 2)

	return &AiderConfig{
		ModelCode:            modelInfo.ModelCode,
		APIKey:               modelInfo.APIKey,
		BaseURL:              baseURL,
		ProviderType:         modelInfo.ProviderType,
		SupportedProtocols:   modelInfo.SupportedProtocols,
		SystemPrompt:         modelInfo.SystemPrompt,
		WorkDir:              workDir,
		MaxTokens:            modelInfo.MaxTokens,
		AutoCommit:           false,
		Timeout:              time.Duration(timeoutSeconds) * time.Second,
		MaxSteps:             maxSteps,
		MapTokens:            512,
		MaxChatHistoryTokens: 2048,
	}
}

// RunTask 为 MVP 任务执行 Aider 代码编辑
// 整合：解析角色模型 → 构建配置 → 调用 Aider → 返回结果
func (r *AiderRunner) RunTask(ctx context.Context, projectID int64, taskID int64, modelInfo *ModelInfo, taskPrompt string, workDir string, files []string, allowPaths []string, readFiles []string) *AiderResult {
	cfg := r.BuildConfigFromModel(ctx, modelInfo, workDir)
	cfg.Message = taskPrompt
	cfg.Files = files
	cfg.AllowPaths = allowPaths
	cfg.ReadFiles = readFiles
	cfg.OnActivity = func() {
		activity.TouchTaskActivity(context.Background(), taskID)
		TouchHeartbeat(context.Background(), taskID)
	}

	// 如果有 system prompt，拼到 message 前面
	if cfg.SystemPrompt != "" {
		cfg.Message = strings.TrimSpace(cfg.SystemPrompt) + "\n\n" + taskPrompt
	}

	snapshot, err := worktreeguard.Capture(ctx, workDir)
	if err != nil {
		g.Log().Warningf(ctx, "[AiderRunner] 捕获 git 基线失败: %v", err)
	}

	result := r.Run(ctx, cfg)
	if cleanupErr := cleanupAiderArtifacts(workDir, cfg.AllowPaths); cleanupErr != nil {
		g.Log().Warningf(ctx, "[AiderRunner] 清理 Aider 临时文件失败: %v", cleanupErr)
	}
	if pruned, pruneErr := worktreeguard.PruneEmbeddedAllowedDuplicates(ctx, workDir, cfg.AllowPaths); pruneErr != nil {
		g.Log().Warningf(ctx, "[AiderRunner] 清理重复嵌入路径失败: %v", pruneErr)
	} else if len(pruned) > 0 {
		g.Log().Infof(ctx, "[AiderRunner] 已清理重复嵌入路径: %v", pruned)
	}
	if result.Error != nil || snapshot == nil {
		return result
	}

	validation, err := snapshot.Validate(ctx, workDir, cfg.AllowPaths)
	if err != nil {
		g.Log().Warningf(ctx, "[AiderRunner] 校验 git 变更失败: %v", err)
		return result
	}
	if len(validation.Suspicious) > 0 {
		if pruned, pruneErr := worktreeguard.PruneSuspiciousDeltaPaths(workDir, validation.Suspicious); pruneErr != nil {
			g.Log().Warningf(ctx, "[AiderRunner] 清理可疑标题文件失败: %v", pruneErr)
		} else if len(pruned) > 0 {
			g.Log().Infof(ctx, "[AiderRunner] 已清理可疑标题文件: %v", pruned)
			validation, err = snapshot.Validate(ctx, workDir, cfg.AllowPaths)
			if err != nil {
				g.Log().Warningf(ctx, "[AiderRunner] 清理后重新校验 git 变更失败: %v", err)
				return result
			}
		}
	}
	if validation.HasIssues() {
		summary := validation.Summary()
		if summary == "" {
			summary = "检测到异常文件变更"
		}
		result.Output = strings.TrimSpace(result.Output + "\n\n[guard] " + summary)
		result.Error = fmt.Errorf("%s", summary)
		result.Category = taskFailurePolicyGuard
	}
	return result
}

func cleanupAiderArtifacts(workDir string, allowPaths []string) error {
	workDir = worktreeguard.ResolveRepoRoot(workDir)
	normalizedAllowPaths, _ := worktreeguard.NormalizeRelativePaths(allowPaths)
	allowSet := make(map[string]struct{}, len(normalizedAllowPaths))
	for _, allowPath := range normalizedAllowPaths {
		allowSet[allowPath] = struct{}{}
	}

	matches, err := filepath.Glob(filepath.Join(workDir, ".aider*"))
	if err != nil {
		return err
	}
	for _, match := range matches {
		if removeErr := os.RemoveAll(match); removeErr != nil && !os.IsNotExist(removeErr) {
			return removeErr
		}
	}

	if _, allowed := allowSet[".gitignore"]; allowed {
		return nil
	}

	gitignorePath := filepath.Join(workDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !isAiderArtifactGitignore(string(content)) {
		return nil
	}
	if err := os.Remove(gitignorePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func isAiderArtifactGitignore(content string) bool {
	content = strings.ReplaceAll(content, "\r", "")
	var rules []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rules = append(rules, line)
	}
	return len(rules) == 1 && rules[0] == ".aider*"
}

func (r *AiderRunner) buildCompactRetryConfig(cfg *AiderConfig) *AiderConfig {
	compact := *cfg
	compact.CompactMode = true
	if compact.MaxSteps > 0 {
		compact.MaxSteps--
	}
	compact.MapTokens = 0
	compact.MaxChatHistoryTokens = 512
	compact.ReadFiles = nil
	compact.Files = limitFiles(cfg.Files, 6)
	compact.Message = r.compactMessage(cfg.Message)
	return &compact
}

func (r *AiderRunner) compactMessage(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > 2500 {
		message = message[:2500] + "\n...(已自动截断上下文)"
	}
	return message + "\n\n请仅完成最小必要修改，优先修改最核心文件；如果变更过大，请先完成一部分可落地修改。"
}

func (r *AiderRunner) isTokenLimitFailure(output string) bool {
	lower := strings.ToLower(output)
	keywords := []string{
		"token-limits.html",
		"hit a token limit",
		"exceeded output limit",
		"input tokens:",
		"output tokens:",
		"context window",
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

// classifyAiderError 对 Aider 执行错误进行分类
// 返回错误类别和是否可重试
func (r *AiderRunner) classifyAiderError(output string, runErr error) (category taskFailureCategory, retryable bool) {
	lower := strings.ToLower(output)
	errStr := ""
	if runErr != nil {
		errStr = strings.ToLower(runErr.Error())
	}

	// Token limit → 不可重试（已在 Run() 中特殊处理过）
	if r.isTokenLimitFailure(output) {
		return taskFailureExecution, false
	}

	// 网络/超时错误 → 可重试
	networkKeywords := []string{
		"connection reset", "connection refused", "timeout", "timed out",
		"eof", "broken pipe", "network unreachable", "dns resolve",
		"rate limit", "429", "503", "502", "500",
	}
	for _, kw := range networkKeywords {
		if strings.Contains(lower, kw) || strings.Contains(errStr, kw) {
			return taskFailureExecution, true
		}
	}

	// 权限/路径错误 → 不可重试，升级给架构师
	permKeywords := []string{
		"permission denied", "no such file", "not found",
		"access denied", "read-only", "disk full", "no space",
	}
	for _, kw := range permKeywords {
		if strings.Contains(lower, kw) || strings.Contains(errStr, kw) {
			return taskFailurePlanning, false
		}
	}

	// 认证错误 → 不可重试
	authKeywords := []string{
		"invalid api key", "authentication", "unauthorized", "401", "403",
	}
	for _, kw := range authKeywords {
		if strings.Contains(lower, kw) || strings.Contains(errStr, kw) {
			return taskFailurePolicyGuard, false
		}
	}

	// 默认：执行失败，不确定是否可重试
	return taskFailureExecution, false
}

func (r *AiderRunner) buildFailureHint(output string, runErr error, cfg *AiderConfig) string {
	if strings.TrimSpace(output) == "" {
		if runErr != nil {
			return "Aider 无输出即退出，底层错误: " + runErr.Error()
		}
		return "Aider 无输出即退出，请检查 aider 可执行文件、模型配置和工作目录。"
	}

	if r.isTokenLimitFailure(output) {
		return fmt.Sprintf("Aider 命中了 token limit。已尝试自动精简上下文重试，但仍失败。建议拆小任务、减少 affected_resources 数量，或改用上下文更大的模型。当前文件数=%d。", len(cfg.Files))
	}

	snippet := shortenForHint(output, 600)
	return "Aider 执行失败，关键信息：" + snippet
}

func limitFiles(files []string, max int) []string {
	if len(files) <= max {
		return files
	}
	trimmed := make([]string, 0, max)
	seen := make(map[string]struct{}, max)
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		if _, ok := seen[file]; ok {
			continue
		}
		seen[file] = struct{}{}
		trimmed = append(trimmed, file)
		if len(trimmed) == max {
			break
		}
	}
	return trimmed
}

func shortenForHint(text string, max int) string {
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.TrimSpace(text)
	if len(text) <= max {
		return text
	}
	return text[:max] + "...(截断)"
}
