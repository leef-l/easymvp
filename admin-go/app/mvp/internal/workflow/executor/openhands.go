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

// OpenHandsExecutor OpenHands CLI 执行器。
// 通过 ai_engine_config.command_template 启动 OpenHands Docker 容器执行任务。
type OpenHandsExecutor struct {
	wsMgr workspace.Manager
}

// NewOpenHandsExecutor 创建 OpenHands 执行器。
func NewOpenHandsExecutor(wsMgr workspace.Manager) *OpenHandsExecutor {
	return &OpenHandsExecutor{wsMgr: wsMgr}
}

func (e *OpenHandsExecutor) Name() string         { return "openhands" }
func (e *OpenHandsExecutor) NeedsWorkspace() bool { return true }

// Execute 执行 OpenHands 任务。
func (e *OpenHandsExecutor) Execute(ctx context.Context, req *Request) *Result {
	// 1. 加载 command_template 配置
	engineCfg, err := g.DB().Model("ai_engine_config").Ctx(ctx).
		Where("engine_code", "openhands").
		Where("status", 1).
		WhereNull("deleted_at").
		One()
	if err != nil || engineCfg.IsEmpty() {
		return &Result{Success: false, Error: fmt.Errorf("OpenHands 引擎未配置或已禁用")}
	}

	cmdTemplate := engineCfg["command_template"].String()
	if cmdTemplate == "" {
		return &Result{Success: false, Error: fmt.Errorf("OpenHands command_template 未配置")}
	}
	if commandTemplateUsesDocker(cmdTemplate) {
		return &Result{Success: false, Error: fmt.Errorf("OpenHands command_template 依赖 Docker，当前环境已禁用 Docker，请改用非 Docker 模板或切换其他执行模式")}
	}

	timeoutSeconds := engineCfg["timeout_seconds"].Int()
	if timeoutSeconds <= 0 {
		timeoutSeconds = 1800
	}

	// 2. 确定工作目录
	project, projErr := g.DB().Model("mvp_project").Ctx(ctx).Where("id", req.ProjectID).Fields("work_dir").One()
	if projErr != nil || project.IsEmpty() {
		return &Result{Success: false, Error: fmt.Errorf("项目 %d 不存在或查询失败: %v", req.ProjectID, projErr)}
	}
	workDir := project["work_dir"].String()
	if req.Workspace != nil {
		workDir = req.Workspace.WorkspacePath
		if mrErr := e.wsMgr.MarkRunning(ctx, req.TaskID); mrErr != nil {
			g.Log().Warningf(ctx, "[OpenHandsExecutor] MarkRunning 失败: task=%d err=%v", req.TaskID, mrErr)
		}
		g.Log().Infof(ctx, "[OpenHandsExecutor] 使用 worktree 隔离: task=%d path=%s", req.TaskID, workDir)
	}

	// 3. 构建环境变量替换表
	taskInstruction := req.TaskRecord["description"].String()
	if req.ModelInfo != nil && req.ModelInfo.SystemPrompt != "" {
		taskInstruction = strings.TrimSpace(req.ModelInfo.SystemPrompt) + "\n\n" + taskInstruction
	}
	// 转义单引号，防止 shell 注入
	taskInstruction = strings.ReplaceAll(taskInstruction, "'", "'\\''")

	envVars := map[string]string{
		"AI_TASK_WORKTREE_PATH": workDir,
		"AI_TASK_INSTRUCTION":   taskInstruction,
	}
	if req.ModelInfo != nil {
		envVars["AI_MODEL_API_KEY"] = req.ModelInfo.APIKey
		envVars["AI_MODEL_CODE"] = req.ModelInfo.ModelCode
		envVars["AI_MODEL_BASE_URL"] = resolveModelBaseURL(req.ModelInfo, engineCfg["base_url"].String())
	}

	// 解析 affected_resources 附加到指令
	var files []string
	resJSON := req.TaskRecord["affected_resources"].String()
	if resJSON != "" && resJSON != "[]" && resJSON != "null" {
		if err := json.Unmarshal([]byte(resJSON), &files); err != nil {
			g.Log().Warningf(ctx, "[OpenHandsExecutor] affected_resources JSON 解析失败: %v", err)
		}
	}
	if len(files) > 0 {
		envVars["AI_TASK_FILES"] = strings.Join(files, ",")
	}

	// 4. 渲染 command_template
	cmdStr := renderCommandTemplate(cmdTemplate, envVars)

	g.Log().Infof(ctx, "[OpenHandsExecutor] 启动: task=%d workDir=%s timeout=%ds",
		req.TaskID, workDir, timeoutSeconds)

	// 5. 执行命令
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "bash", "-c", cmdStr)
	cmd.Dir = workDir

	// 传递必要的环境变量（某些 template 使用 $ENV 而非 ${env:} 占位符）
	cmd.Env = append(os.Environ(),
		"AI_TASK_WORKTREE_PATH="+workDir,
		"AI_TASK_INSTRUCTION="+req.TaskRecord["description"].String(),
	)
	if req.ModelInfo != nil {
		cmd.Env = append(cmd.Env,
			"AI_MODEL_API_KEY="+req.ModelInfo.APIKey,
			"AI_MODEL_CODE="+req.ModelInfo.ModelCode,
			"AI_MODEL_BASE_URL="+resolveModelBaseURL(req.ModelInfo, engineCfg["base_url"].String()),
		)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output = output + "\n" + stderr.String()
	}
	output = strings.TrimSpace(output)

	if err != nil {
		errMsg := fmt.Sprintf("OpenHands 执行失败: %v", err)
		if output != "" {
			errMsg = errMsg + "\n" + truncateOutput(output, 2000)
		}
		g.Log().Warningf(ctx, "[OpenHandsExecutor] 失败: task=%d err=%v", req.TaskID, err)

		// workspace finalize: 标记失败
		if req.Workspace != nil && e.wsMgr != nil {
			_ = e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{
				Success: false,
				Error:   errMsg,
			})
		}
		return &Result{Success: false, Error: fmt.Errorf("%s", errMsg)}
	}

	g.Log().Infof(ctx, "[OpenHandsExecutor] 成功: task=%d output_len=%d", req.TaskID, len(output))

	// workspace finalize: 标记成功
	if req.Workspace != nil && e.wsMgr != nil {
		if fErr := e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{Success: true}); fErr != nil {
			g.Log().Warningf(ctx, "[OpenHandsExecutor] workspace finalize 失败: task=%d err=%v", req.TaskID, fErr)
		} else {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						g.Log().Errorf(context.Background(), "[OpenHandsExecutor] workspace cleanup panic: task=%d err=%v", req.TaskID, r)
					}
				}()
				if cleanErr := e.wsMgr.Cleanup(context.Background(), req.TaskID); cleanErr != nil {
					g.Log().Warningf(context.Background(), "[OpenHandsExecutor] workspace cleanup 失败: task=%d err=%v", req.TaskID, cleanErr)
				}
			}()
		}
	}

	return &Result{Success: true, Output: truncateOutput(output, 10000)}
}

// renderCommandTemplate 渲染 command_template，将 ${env:VAR} 和 $env:VAR 替换为实际值。
func renderCommandTemplate(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		// 替换 ${env:VAR} 格式
		result = strings.ReplaceAll(result, "${env:"+k+"}", v)
		// 替换 $env:VAR 格式（非括号形式，需注意边界）
		result = strings.ReplaceAll(result, "$env:"+k, v)
	}
	return result
}

// truncateOutput 截断输出到指定长度。
func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n...(truncated)"
}
