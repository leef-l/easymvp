package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/repo"
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

func (e *AiderExecutor) Name() string         { return "aider" }
func (e *AiderExecutor) NeedsWorkspace() bool { return true }

// Execute 执行 Aider 任务。
func (e *AiderExecutor) Execute(ctx context.Context, req *Request) *Result {
	project, err := repo.NewProjectRepo().GetByID(ctx, req.ProjectID, "work_dir")
	if err != nil || len(project) == 0 {
		return &Result{Success: false, Error: fmt.Errorf("项目 %d 不存在或查询失败: %v", req.ProjectID, err)}
	}
	workDir := g.NewVar(project["work_dir"]).String()

	// 如果有 workspace 隔离，使用 worktree 路径
	if req.Workspace != nil {
		workDir = req.Workspace.WorkspacePath
		if mrErr := e.wsMgr.MarkRunning(ctx, req.TaskID); mrErr != nil {
			g.Log().Warningf(ctx, "[AiderExecutor] MarkRunning 失败: task=%d err=%v", req.TaskID, mrErr)
		}
		g.Log().Infof(ctx, "[AiderExecutor] 使用 worktree 隔离: task=%d path=%s", req.TaskID, workDir)
	}

	targets := parseResourceTargets(req.TaskRecord["affected_resources"].String())
	if len(targets.Rejected) > 0 {
		g.Log().Warningf(ctx, "[AiderExecutor] 丢弃可疑 affected_resources: task=%d rejected=%v", req.TaskID, targets.Rejected)
	}
	promptAllowPaths := promptAllowedPathsForExecution(workDir, targets)
	workDir, targets = applyExecutionSubdir(workDir, targets)
	if len(targets.DirectoryPaths) > 0 {
		if err := ensureDirectoryTargets(workDir, targets.DirectoryPaths); err != nil {
			if req.Workspace != nil && e.wsMgr != nil {
				finalizeWorkspaceFailure(ctx, e.wsMgr, req.TaskID, "AiderExecutor", err.Error(), false)
			}
			return &Result{Success: false, Error: err}
		}
	}
	if len(targets.FilePaths) == 0 && len(targets.DirectoryPaths) > 0 {
		output := fmt.Sprintf("已准备目录资源: %s", strings.Join(targets.DirectoryPaths, ", "))
		if req.Workspace != nil && e.wsMgr != nil {
			if err := finalizeWorkspaceSuccess(ctx, e.wsMgr, req.TaskID, "AiderExecutor"); err != nil {
				return &Result{Success: false, Error: err}
			}
		}
		return &Result{Success: true, Output: output}
	}

	// 从引擎配置读取超时，与其他执行器保持一致
	timeoutSeconds := 1800
	engineCfg, cfgErr := repo.NewAIEngineConfigRepo().GetEnabledByCode(ctx, "aider", "timeout_seconds")
	if cfgErr == nil && len(engineCfg) > 0 && g.NewVar(engineCfg["timeout_seconds"]).Int() > 0 {
		timeoutSeconds = g.NewVar(engineCfg["timeout_seconds"]).Int()
	}
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	runner := engine.GetAiderRunner()
	aiderResult := runner.RunTask(execCtx, req.ProjectID, req.TaskID, req.ModelInfo,
		buildStrictAiderTaskPrompt(req.TaskRecord["description"].String(), promptAllowPaths), workDir, targets.FilePaths, targets.AllowedPaths, nil)

	if aiderResult.Error != nil {
		// workspace finalize: 标记失败
		if req.Workspace != nil && e.wsMgr != nil {
			finalizeWorkspaceFailure(ctx, e.wsMgr, req.TaskID, "AiderExecutor", aiderResult.Error.Error(), false)
		}
		return &Result{Success: false, Error: aiderResult.Error}
	}

	// workspace finalize: 标记成功
	if req.Workspace != nil && e.wsMgr != nil {
		if err := finalizeWorkspaceSuccess(ctx, e.wsMgr, req.TaskID, "AiderExecutor"); err != nil {
			return &Result{Success: false, Error: err}
		}
	}

	return &Result{Success: true, Output: aiderResult.Output}
}

func buildStrictAiderTaskPrompt(description string, allowPaths []string) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(description))
	b.WriteString("\n\n执行约束：\n")
	if len(allowPaths) > 0 {
		b.WriteString("- 只允许创建或修改以下路径：\n")
		for _, allowPath := range allowPaths {
			allowPath = strings.TrimSpace(allowPath)
			if allowPath == "" {
				continue
			}
			b.WriteString("  - ")
			b.WriteString(allowPath)
			b.WriteString("\n")
		}
	} else {
		b.WriteString("- 当前任务未声明允许写入的文件路径；若无法在现有文件内完成，请直接失败，不要自行扩展范围。\n")
	}
	b.WriteString("- 严禁创建任何未在允许列表中的文件、目录、临时文件、总结文件或说明文件。\n")
	b.WriteString("- 严禁把回答中的标题、项目符号、代码块说明、'运行方式：'、'验证：'、'说明：' 等自然语言文本当成文件名。\n")
	b.WriteString("- 直接编辑目标文件，不要额外生成 Markdown 报告、操作说明、验证清单或同义标题文件。\n")
	b.WriteString("- 如果需要说明如何运行或验证，只能写入已允许的目标文件内容，或在标准输出中说明，不要新增文件。\n")
	b.WriteString("- 如果发现任务描述与允许路径冲突，优先遵守允许路径并直接失败，不要越界修改。\n")
	return b.String()
}

// ChatExecutor ChatStream 执行器。
type ChatExecutor struct{}

type chatReplySnapshot struct {
	Status  string
	Content string
}

var loadChatReplySnapshot = func(ctx context.Context, replyID int64) (chatReplySnapshot, error) {
	reply, err := repo.NewMessageRepo().GetByID(ctx, replyID, "status", "content")
	if err != nil {
		return chatReplySnapshot{}, err
	}
	if len(reply) == 0 {
		return chatReplySnapshot{}, fmt.Errorf("chat reply %d 不存在", replyID)
	}

	return chatReplySnapshot{
		Status:  strings.TrimSpace(g.NewVar(reply["status"]).String()),
		Content: g.NewVar(reply["content"]).String(),
	}, nil
}

var chatReplyPollInterval = 500 * time.Millisecond

func waitForChatReply(ctx context.Context, replyID int64) (string, error) {
	for {
		snapshot, err := loadChatReplySnapshot(ctx, replyID)
		if err != nil {
			return "", err
		}

		switch snapshot.Status {
		case "completed":
			return snapshot.Content, nil
		case "failed":
			errMsg := strings.TrimSpace(snapshot.Content)
			if errMsg == "" {
				errMsg = fmt.Sprintf("chat reply %d 执行失败", replyID)
			}
			return "", fmt.Errorf("%s", errMsg)
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(chatReplyPollInterval):
		}
	}
}

// NewChatExecutor 创建 Chat 执行器。
func NewChatExecutor() *ChatExecutor { return &ChatExecutor{} }

func (e *ChatExecutor) Name() string         { return "chat" }
func (e *ChatExecutor) NeedsWorkspace() bool { return false }

// Execute 执行 Chat 任务。
func (e *ChatExecutor) Execute(ctx context.Context, req *Request) *Result {
	timeoutSeconds := engine.GetConfigInt(ctx, "runtime.task_timeout_seconds", "engine.runtime.taskTimeoutSeconds", 600)
	if timeoutSeconds <= 0 {
		timeoutSeconds = 600
	}
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// 创建或获取任务对话
	convID, err := engine.EnsureDomainTaskConversation(execCtx, req.ProjectID, req.TaskID,
		req.TaskRecord["role_type"].String(), req.TaskRecord["name"].String())
	if err != nil {
		return &Result{Success: false, Error: err}
	}

	// 发送任务描述到对话
	_, replyID, err := engine.GetEngine().SendMessage(execCtx, convID, req.TaskRecord["description"].String(), 0, 0)
	if err != nil {
		return &Result{Success: false, Error: fmt.Errorf("chat 执行失败: %w", err)}
	}

	output, err := waitForChatReply(execCtx, replyID)
	if err != nil {
		return &Result{Success: false, Error: fmt.Errorf("chat 执行失败: %w", err)}
	}

	return &Result{Success: true, Output: output}
}
