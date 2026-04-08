package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workspace"
)

// AiderExecutor Aider CLI 执行器。
type AiderExecutor struct {
	wsMgr workspace.Manager
}

// NewAiderExecutor 创建 Aider 执行器。
func NewAiderExecutor(wsMgr workspace.Manager) *AiderExecutor {
	return &AiderExecutor{wsMgr: wsMgr}
}

func (e *AiderExecutor) Name() string          { return "aider" }
func (e *AiderExecutor) NeedsWorkspace() bool   { return true }

// Execute 执行 Aider 任务。
func (e *AiderExecutor) Execute(ctx context.Context, req *Request) *Result {
	project, err := g.DB().Model("mvp_project").Ctx(ctx).Where("id", req.ProjectID).WhereNull("deleted_at").One()
	if err != nil || project.IsEmpty() {
		return &Result{Success: false, Error: fmt.Errorf("项目 %d 不存在或查询失败: %v", req.ProjectID, err)}
	}
	workDir := project["work_dir"].String()

	// 如果有 workspace 隔离，使用 worktree 路径
	if req.Workspace != nil {
		workDir = req.Workspace.WorkspacePath
		if mrErr := e.wsMgr.MarkRunning(ctx, req.TaskID); mrErr != nil {
			g.Log().Warningf(ctx, "[AiderExecutor] MarkRunning 失败: task=%d err=%v", req.TaskID, mrErr)
		}
		g.Log().Infof(ctx, "[AiderExecutor] 使用 worktree 隔离: task=%d path=%s", req.TaskID, workDir)
	}

	// 解析 affected_resources 作为文件列表
	var files []string
	resJSON := req.TaskRecord["affected_resources"].String()
	if resJSON != "" && resJSON != "[]" && resJSON != "null" {
		if umErr := json.Unmarshal([]byte(resJSON), &files); umErr != nil {
			g.Log().Warningf(ctx, "[AiderExecutor] 解析 affected_resources 失败: task=%d err=%v", req.TaskID, umErr)
		}
	}

	// 从引擎配置读取超时，与其他执行器保持一致
	timeoutSeconds := 1800
	engineCfg, cfgErr := g.DB().Model("ai_engine_config").Ctx(ctx).
		Where("engine_code", "aider").Where("status", 1).WhereNull("deleted_at").One()
	if cfgErr == nil && !engineCfg.IsEmpty() && engineCfg["timeout_seconds"].Int() > 0 {
		timeoutSeconds = engineCfg["timeout_seconds"].Int()
	}
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	runner := engine.GetAiderRunner()
	aiderResult := runner.RunTask(execCtx, req.ProjectID, req.TaskID, req.ModelInfo,
		req.TaskRecord["description"].String(), workDir, files, nil)

	if aiderResult.Error != nil {
		// workspace finalize: 标记失败
		if req.Workspace != nil && e.wsMgr != nil {
			if fErr := e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{
				Success: false,
				Error:   aiderResult.Error.Error(),
			}); fErr != nil {
				g.Log().Warningf(ctx, "[AiderExecutor] workspace finalize(失败) 失败: task=%d err=%v", req.TaskID, fErr)
			}
		}
		return &Result{Success: false, Error: aiderResult.Error}
	}

	// workspace finalize: 标记成功
	if req.Workspace != nil && e.wsMgr != nil {
		if err := e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{Success: true}); err != nil {
			g.Log().Warningf(ctx, "[AiderExecutor] workspace finalize 失败: task=%d err=%v", req.TaskID, err)
		} else {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						g.Log().Errorf(context.Background(), "[AiderExecutor] workspace cleanup panic: task=%d err=%v", req.TaskID, r)
					}
				}()
				if cleanErr := e.wsMgr.Cleanup(context.Background(), req.TaskID); cleanErr != nil {
					g.Log().Warningf(context.Background(), "[AiderExecutor] workspace cleanup 失败: task=%d err=%v", req.TaskID, cleanErr)
				}
			}()
		}
	}

	return &Result{Success: true, Output: aiderResult.Output}
}

// ChatExecutor ChatStream 执行器。
type ChatExecutor struct{}

// NewChatExecutor 创建 Chat 执行器。
func NewChatExecutor() *ChatExecutor { return &ChatExecutor{} }

func (e *ChatExecutor) Name() string          { return "chat" }
func (e *ChatExecutor) NeedsWorkspace() bool   { return false }

// Execute 执行 Chat 任务。
func (e *ChatExecutor) Execute(ctx context.Context, req *Request) *Result {
	// 创建或获取任务对话
	convID, err := engine.EnsureDomainTaskConversation(ctx, req.ProjectID, req.TaskID,
		req.TaskRecord["role_type"].String(), req.TaskRecord["name"].String())
	if err != nil {
		return &Result{Success: false, Error: err}
	}

	// 发送任务描述到对话
	_, _, err = engine.GetEngine().SendMessage(ctx, convID, req.TaskRecord["description"].String(), 0, 0)
	if err != nil {
		return &Result{Success: false, Error: fmt.Errorf("chat 执行失败: %w", err)}
	}

	return &Result{Success: true, Output: "chat execution completed"}
}
