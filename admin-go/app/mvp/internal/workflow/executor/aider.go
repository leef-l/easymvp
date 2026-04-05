package executor

import (
	"context"
	"encoding/json"
	"fmt"

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
	project, _ := g.DB().Model("mvp_project").Ctx(ctx).Where("id", req.ProjectID).One()
	workDir := project["work_dir"].String()

	// 如果有 workspace 隔离，使用 worktree 路径
	if req.Workspace != nil {
		workDir = req.Workspace.WorkspacePath
		_ = e.wsMgr.MarkRunning(ctx, req.TaskID)
		g.Log().Infof(ctx, "[AiderExecutor] 使用 worktree 隔离: task=%d path=%s", req.TaskID, workDir)
	}

	// 解析 affected_resources 作为文件列表
	var files []string
	resJSON := req.TaskRecord["affected_resources"].String()
	if resJSON != "" && resJSON != "[]" && resJSON != "null" {
		json.Unmarshal([]byte(resJSON), &files)
	}

	runner := engine.GetAiderRunner()
	aiderResult := runner.RunTask(ctx, req.ProjectID, req.TaskID, req.ModelInfo,
		req.TaskRecord["description"].String(), workDir, files, nil)

	if aiderResult.Error != nil {
		// workspace finalize: 标记失败
		if req.Workspace != nil && e.wsMgr != nil {
			_ = e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{
				Success: false,
				Error:   aiderResult.Error.Error(),
			})
		}
		return &Result{Success: false, Error: aiderResult.Error}
	}

	// workspace finalize: 标记成功
	if req.Workspace != nil && e.wsMgr != nil {
		if err := e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{Success: true}); err != nil {
			g.Log().Warningf(ctx, "[AiderExecutor] workspace finalize 失败: task=%d err=%v", req.TaskID, err)
		} else {
			go func() {
				if cleanErr := e.wsMgr.Cleanup(context.Background(), req.TaskID); cleanErr != nil {
					g.Log().Warningf(ctx, "[AiderExecutor] workspace cleanup 失败: task=%d err=%v", req.TaskID, cleanErr)
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
