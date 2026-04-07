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

	"easymvp/app/mvp/internal/workspace"
)

// CodexCLIExecutor OpenAI Codex CLI 执行器。
// 通过 codex 命令行工具执行编码任务。
type CodexCLIExecutor struct {
	wsMgr workspace.Manager
}

// NewCodexCLIExecutor 创建 Codex CLI 执行器。
func NewCodexCLIExecutor(wsMgr workspace.Manager) *CodexCLIExecutor {
	return &CodexCLIExecutor{wsMgr: wsMgr}
}

func (e *CodexCLIExecutor) Name() string         { return "codex_cli" }
func (e *CodexCLIExecutor) NeedsWorkspace() bool { return true }

// Execute 执行 Codex CLI 任务。
func (e *CodexCLIExecutor) Execute(ctx context.Context, req *Request) *Result {
	// 加载引擎配置
	engineCfg, err := g.DB().Model("ai_engine_config").Ctx(ctx).
		Where("engine_code", "codex_cli").
		Where("status", 1).
		WhereNull("deleted_at").
		One()
	if err != nil || engineCfg.IsEmpty() {
		return &Result{Success: false, Error: fmt.Errorf("Codex CLI 引擎未配置或已禁用")}
	}

	timeoutSeconds := engineCfg["timeout_seconds"].Int()
	if timeoutSeconds <= 0 {
		timeoutSeconds = 1800
	}

	// 确定工作目录
	project, projErr := g.DB().Model("mvp_project").Ctx(ctx).Where("id", req.ProjectID).One()
	if projErr != nil || project.IsEmpty() {
		return &Result{Success: false, Error: fmt.Errorf("项目 %d 不存在或查询失败: %v", req.ProjectID, projErr)}
	}
	workDir := project["work_dir"].String()
	if req.Workspace != nil {
		workDir = req.Workspace.WorkspacePath
		if mrErr := e.wsMgr.MarkRunning(ctx, req.TaskID); mrErr != nil {
			g.Log().Warningf(ctx, "[CodexCLIExecutor] MarkRunning 失败: task=%d err=%v", req.TaskID, mrErr)
		}
	}

	// 构建任务指令
	taskInstruction := req.TaskRecord["description"].String()
	if req.ModelInfo != nil && req.ModelInfo.SystemPrompt != "" {
		taskInstruction = strings.TrimSpace(req.ModelInfo.SystemPrompt) + "\n\n" + taskInstruction
	}

	// 解析 affected_resources
	var files []string
	resJSON := req.TaskRecord["affected_resources"].String()
	if resJSON != "" && resJSON != "[]" && resJSON != "null" {
		_ = json.Unmarshal([]byte(resJSON), &files)
	}

	// 支持 command_template 或默认 codex CLI
	cmdTemplate := engineCfg["command_template"].String()
	var cmdStr string
	if cmdTemplate != "" {
		envVars := map[string]string{
			"AI_TASK_WORKTREE_PATH": workDir,
			"AI_TASK_INSTRUCTION":   strings.ReplaceAll(taskInstruction, "'", "'\\''"),
			"AI_TASK_FILES":         strings.Join(files, ","),
		}
		if req.ModelInfo != nil {
			envVars["AI_MODEL_API_KEY"] = req.ModelInfo.APIKey
			envVars["AI_MODEL_CODE"] = req.ModelInfo.ModelCode
			envVars["AI_MODEL_BASE_URL"] = resolveModelBaseURL(req.ModelInfo, engineCfg["base_url"].String())
		}
		cmdStr = renderCommandTemplate(cmdTemplate, envVars)
	} else {
		// 默认：Codex CLI 非交互执行，并显式使用任务模型。
		cmdStr = buildCodexDefaultCommand(workDir, taskInstruction, req.ModelInfo)
	}

	g.Log().Infof(ctx, "[CodexCLIExecutor] 启动: task=%d workDir=%s", req.TaskID, workDir)

	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "bash", "-c", cmdStr)
	cmd.Dir = workDir
	cmd.Env = os.Environ()
	if req.ModelInfo != nil && req.ModelInfo.APIKey != "" {
		cmd.Env = append(cmd.Env, "OPENAI_API_KEY="+req.ModelInfo.APIKey)
	}
	if baseURL := resolveModelBaseURL(req.ModelInfo, engineCfg["base_url"].String()); baseURL != "" {
		cmd.Env = append(cmd.Env, "OPENAI_BASE_URL="+baseURL)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := strings.TrimSpace(stdout.String() + "\n" + stderr.String())

	if err != nil {
		errMsg := fmt.Sprintf("Codex CLI 执行失败: %v", err)
		if output != "" {
			errMsg = errMsg + "\n" + truncateOutput(output, 2000)
		}
		if req.Workspace != nil && e.wsMgr != nil {
			_ = e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{Success: false, Error: errMsg})
		}
		return &Result{Success: false, Error: fmt.Errorf("%s", errMsg)}
	}

	if req.Workspace != nil && e.wsMgr != nil {
		if fErr := e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{Success: true}); fErr != nil {
			g.Log().Warningf(ctx, "[CodexCLIExecutor] workspace finalize 失败: task=%d err=%v", req.TaskID, fErr)
		} else {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						g.Log().Errorf(context.Background(), "[CodexCLIExecutor] workspace cleanup panic: task=%d err=%v", req.TaskID, r)
					}
				}()
				if cleanErr := e.wsMgr.Cleanup(context.Background(), req.TaskID); cleanErr != nil {
					g.Log().Warningf(context.Background(), "[CodexCLIExecutor] workspace cleanup 失败: task=%d err=%v", req.TaskID, cleanErr)
				}
			}()
		}
	}

	return &Result{Success: true, Output: truncateOutput(output, 10000)}
}
