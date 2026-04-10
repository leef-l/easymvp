package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workspace"
)

// ClaudeCodeExecutor Claude Code CLI 执行器。
// 通过 claude 命令行工具执行编码任务。
type ClaudeCodeExecutor struct {
	wsMgr workspace.Manager
}

type claudeCLIResponse struct {
	Subtype           string            `json:"subtype"`
	IsError           bool              `json:"is_error"`
	Result            string            `json:"result"`
	StopReason        string            `json:"stop_reason"`
	TerminalReason    string            `json:"terminal_reason"`
	PermissionDenials []json.RawMessage `json:"permission_denials"`
}

// NewClaudeCodeExecutor 创建 Claude Code 执行器。
func NewClaudeCodeExecutor(wsMgr workspace.Manager) *ClaudeCodeExecutor {
	return &ClaudeCodeExecutor{wsMgr: wsMgr}
}

func (e *ClaudeCodeExecutor) Name() string         { return "claude_code" }
func (e *ClaudeCodeExecutor) NeedsWorkspace() bool { return true }

// Execute 执行 Claude Code 任务。
func (e *ClaudeCodeExecutor) Execute(ctx context.Context, req *Request) *Result {
	// 加载引擎配置
	engineCfg, err := g.DB().Model("ai_engine_config").Ctx(ctx).
		Where("engine_code", "claude_code").
		Where("status", 1).
		WhereNull("deleted_at").
		One()
	if err != nil || engineCfg.IsEmpty() {
		return &Result{Success: false, Error: fmt.Errorf("Claude Code 引擎未配置或已禁用")}
	}

	timeoutSeconds := engineCfg["timeout_seconds"].Int()
	if timeoutSeconds <= 0 {
		timeoutSeconds = 1800
	}

	// 确定工作目录
	project, projErr := g.DB().Model("mvp_project").Ctx(ctx).Where("id", req.ProjectID).WhereNull("deleted_at").Fields("work_dir").One()
	if projErr != nil || project.IsEmpty() {
		return &Result{Success: false, Error: fmt.Errorf("项目 %d 不存在或查询失败: %v", req.ProjectID, projErr)}
	}
	workDir := project["work_dir"].String()
	if req.Workspace != nil {
		workDir = req.Workspace.WorkspacePath
		if mrErr := e.wsMgr.MarkRunning(ctx, req.TaskID); mrErr != nil {
			g.Log().Warningf(ctx, "[ClaudeCodeExecutor] MarkRunning 失败: task=%d err=%v", req.TaskID, mrErr)
		}
	}

	// 构建任务指令
	taskInstruction := req.TaskRecord["description"].String()
	if req.ModelInfo != nil && req.ModelInfo.SystemPrompt != "" {
		taskInstruction = strings.TrimSpace(req.ModelInfo.SystemPrompt) + "\n\n" + taskInstruction
	}

	targets := parseResourceTargets(req.TaskRecord["affected_resources"].String())
	if len(targets.Rejected) > 0 {
		g.Log().Warningf(ctx, "[ClaudeCodeExecutor] 丢弃可疑 affected_resources: task=%d rejected=%v", req.TaskID, targets.Rejected)
	}
	workDir, targets = applyExecutionSubdir(workDir, targets)
	if len(targets.DirectoryPaths) > 0 {
		if err := ensureDirectoryTargets(workDir, targets.DirectoryPaths); err != nil {
			if req.Workspace != nil && e.wsMgr != nil {
				finalizeWorkspaceFailure(ctx, e.wsMgr, req.TaskID, "ClaudeCodeExecutor", err.Error(), false)
			}
			return &Result{Success: false, Error: err}
		}
	}
	if len(targets.FilePaths) == 0 && len(targets.DirectoryPaths) > 0 {
		output := fmt.Sprintf("已准备目录资源: %s", strings.Join(targets.DirectoryPaths, ", "))
		if req.Workspace != nil && e.wsMgr != nil {
			if err := finalizeWorkspaceSuccess(ctx, e.wsMgr, req.TaskID, "ClaudeCodeExecutor"); err != nil {
				return &Result{Success: false, Error: err}
			}
		}
		return &Result{Success: true, Output: output}
	}

	// 支持两种模式：command_template 或默认 claude CLI
	cmdTemplate := engineCfg["command_template"].String()
	var cmdStr string
	if cmdTemplate != "" {
		envVars := map[string]string{
			"AI_TASK_WORKTREE_PATH": workDir,
			"AI_TASK_INSTRUCTION":   strings.ReplaceAll(taskInstruction, "'", "'\\''"),
			"AI_TASK_FILES":         strings.Join(targets.FilePaths, ","),
		}
		if req.ModelInfo != nil {
			envVars["AI_MODEL_API_KEY"] = req.ModelInfo.APIKey
			envVars["AI_MODEL_CODE"] = req.ModelInfo.ModelCode
			envVars["AI_MODEL_BASE_URL"] = resolveProtocolBaseURL(req.ModelInfo, engineCfg["base_url"].String(), "anthropic")
		}
		cmdStr = renderCommandTemplate(cmdTemplate, envVars)
	} else {
		// 默认：Claude CLI 非交互执行，并显式使用任务模型。
		cmdStr = buildClaudeDefaultCommand(workDir, taskInstruction, req.ModelInfo)
	}

	g.Log().Infof(ctx, "[ClaudeCodeExecutor] 启动: task=%d workDir=%s", req.TaskID, workDir)

	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "bash", "-c", cmdStr)
	cmd.Dir = workDir
	cmd.Env = os.Environ()
	// Claude CLI 使用项目级凭据时，Anthropic 基础地址不能带 /v1。
	if shouldUseClaudeProjectCredentials(req.ModelInfo) {
		if req.ModelInfo != nil && req.ModelInfo.APIKey != "" {
			cmd.Env = append(cmd.Env, "ANTHROPIC_API_KEY="+req.ModelInfo.APIKey)
		}
		if baseURL := resolveProtocolBaseURL(req.ModelInfo, engineCfg["base_url"].String(), "anthropic"); baseURL != "" {
			cmd.Env = append(cmd.Env, "ANTHROPIC_BASE_URL="+baseURL)
		}
	}
	engine.GetCommandResourcePolicy(ctx).Apply(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := strings.TrimSpace(stdout.String() + "\n" + stderr.String())

	if err != nil {
		errMsg := fmt.Sprintf("Claude Code 执行失败: %v", err)
		if output != "" {
			errMsg = errMsg + "\n" + truncateOutput(output, 2000)
		}
		if req.Workspace != nil && e.wsMgr != nil {
			finalizeWorkspaceFailure(ctx, e.wsMgr, req.TaskID, "ClaudeCodeExecutor", errMsg, false)
		}
		return &Result{Success: false, Error: fmt.Errorf("%s", errMsg)}
	}
	if blockedErr := detectClaudeBlockedByPermissions(output); blockedErr != nil {
		errMsg := blockedErr.Error()
		if output != "" {
			errMsg = errMsg + "\n" + truncateOutput(output, 2000)
		}
		if req.Workspace != nil && e.wsMgr != nil {
			finalizeWorkspaceFailure(ctx, e.wsMgr, req.TaskID, "ClaudeCodeExecutor", errMsg, false)
		}
		return &Result{Success: false, Error: fmt.Errorf("%s", errMsg)}
	}

	if req.Workspace != nil && e.wsMgr != nil {
		if err := finalizeWorkspaceSuccess(ctx, e.wsMgr, req.TaskID, "ClaudeCodeExecutor"); err != nil {
			return &Result{Success: false, Error: err}
		}
	}

	return &Result{Success: true, Output: truncateOutput(output, 10000)}
}

func detectClaudeBlockedByPermissions(output string) error {
	candidate := strings.TrimSpace(output)
	if start := strings.Index(candidate, "{"); start >= 0 {
		if end := strings.LastIndex(candidate, "}"); end > start {
			candidate = candidate[start : end+1]
		}
	}

	var response claudeCLIResponse
	if err := json.Unmarshal([]byte(candidate), &response); err == nil {
		if response.IsError {
			msg := strings.TrimSpace(response.Result)
			if msg == "" {
				msg = "Claude Code 返回错误结果"
			}
			return fmt.Errorf("%s", msg)
		}
		if len(response.PermissionDenials) > 0 {
			return fmt.Errorf("Claude Code 运行过程中遭遇权限拒绝，任务未实际落地")
		}
		if strings.Contains(response.Result, "请授权后我将执行") || strings.Contains(response.Result, "请授权后我会执行") {
			return fmt.Errorf("Claude Code 需要人工授权后才能继续写入，任务未实际落地")
		}
		return nil
	}

	lower := strings.ToLower(output)
	if strings.Contains(lower, "\"permission_denials\"") || strings.Contains(output, "请授权后我将执行") || strings.Contains(output, "请授权后我会执行") {
		return fmt.Errorf("Claude Code 需要人工授权后才能继续写入，任务未实际落地")
	}
	return nil
}
