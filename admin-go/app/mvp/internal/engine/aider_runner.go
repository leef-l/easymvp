package engine

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
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
	ModelCode    string // 模型代码（如 tc-code-latest, glm-5）
	APIKey       string // API Key
	BaseURL      string // API Base URL（不带 /v1）
	ProviderType string // provider 类型（anthropic / openai_compatible）
	SystemPrompt string // 系统提示词
	WorkDir      string // 工作目录（项目代码所在目录）
	Files        []string // 需要编辑的文件列表
	ReadFiles    []string // 只读参考文件
	Message      string // 任务指令
	MaxTokens    int    // 最大输出 token
	AutoCommit   bool   // 是否自动提交 git
	Timeout      time.Duration // 超时时间
}

// AiderResult Aider 执行结果
type AiderResult struct {
	Output   string // 完整输出
	ExitCode int    // 退出码
	Error    error  // 错误
	Duration time.Duration // 耗时
}

// Run 执行 Aider 任务
func (r *AiderRunner) Run(ctx context.Context, cfg *AiderConfig) *AiderResult {
	start := time.Now()

	args := r.buildArgs(cfg)
	env := r.buildEnv(cfg)

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}

	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "aider", args...)
	cmd.Dir = cfg.WorkDir
	cmd.Env = append(cmd.Environ(), env...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	g.Log().Infof(ctx, "[AiderRunner] 启动: model=%s workdir=%s files=%v",
		cfg.ModelCode, cfg.WorkDir, cfg.Files)

	err := cmd.Run()

	result := &AiderResult{
		Output:   stdout.String() + stderr.String(),
		Duration: time.Since(start),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Error = err
		g.Log().Warningf(ctx, "[AiderRunner] 退出 code=%d err=%v", result.ExitCode, err)
	}

	g.Log().Infof(ctx, "[AiderRunner] 完成: 耗时=%v output_len=%d",
		result.Duration, len(result.Output))

	return result
}

// buildArgs 构建 Aider 命令行参数
func (r *AiderRunner) buildArgs(cfg *AiderConfig) []string {
	args := []string{
		"--model", r.formatModel(cfg),
		"--no-auto-commits",
		"--no-show-model-warnings",
		"--no-pretty",
		"--no-stream",
		"--no-browser",
		"--yes-always",
		"--chat-language", "Chinese",
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

	// 系统提示词（通过 --message-file 或 inline）
	if cfg.Message != "" {
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

	switch cfg.ProviderType {
	case "anthropic":
		env = append(env,
			"ANTHROPIC_API_KEY="+cfg.APIKey,
			"ANTHROPIC_BASE_URL="+cfg.BaseURL,
		)
	case "openai_compatible":
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

	return env
}

// formatModel 格式化模型名称（Aider 需要 provider/model 格式）
func (r *AiderRunner) formatModel(cfg *AiderConfig) string {
	switch cfg.ProviderType {
	case "anthropic":
		if !strings.HasPrefix(cfg.ModelCode, "anthropic/") {
			return "anthropic/" + cfg.ModelCode
		}
	case "openai_compatible":
		if !strings.HasPrefix(cfg.ModelCode, "openai/") {
			return "openai/" + cfg.ModelCode
		}
	}
	return cfg.ModelCode
}

// BuildConfigFromModel 从数据库模型信息构建 AiderConfig
func (r *AiderRunner) BuildConfigFromModel(ctx context.Context, modelInfo *ModelInfo, workDir string) *AiderConfig {
	// 处理 base URL：腾讯云的 anthropic URL 末尾是 /v1，Aider 需要去掉
	baseURL := modelInfo.BaseURL
	baseURL = strings.TrimSuffix(baseURL, "/v1")
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &AiderConfig{
		ModelCode:    modelInfo.ModelCode,
		APIKey:       modelInfo.APIKey,
		BaseURL:      baseURL,
		ProviderType: modelInfo.ProviderType,
		SystemPrompt: modelInfo.SystemPrompt,
		WorkDir:      workDir,
		MaxTokens:    modelInfo.MaxTokens,
		AutoCommit:   false,
		Timeout:      10 * time.Minute,
	}
}

// RunTask 为 MVP 任务执行 Aider 代码编辑
// 整合：解析角色模型 → 构建配置 → 调用 Aider → 返回结果
func (r *AiderRunner) RunTask(ctx context.Context, projectID int64, taskID int64, modelInfo *ModelInfo, taskPrompt string, workDir string, files []string, readFiles []string) *AiderResult {
	cfg := r.BuildConfigFromModel(ctx, modelInfo, workDir)
	cfg.Message = taskPrompt
	cfg.Files = files
	cfg.ReadFiles = readFiles

	// 如果有 system prompt，拼到 message 前面
	if cfg.SystemPrompt != "" {
		cfg.Message = fmt.Sprintf("## 角色设定\n%s\n\n## 任务指令\n%s", cfg.SystemPrompt, taskPrompt)
	}

	return r.Run(ctx, cfg)
}
