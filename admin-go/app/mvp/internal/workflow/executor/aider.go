package executor

import (
	"context"
	"fmt"
	"strings"
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

func (e *AiderExecutor) Name() string         { return "aider" }
func (e *AiderExecutor) NeedsWorkspace() bool { return true }

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

	targets := parseResourceTargets(req.TaskRecord["affected_resources"].String())
	if len(targets.Rejected) > 0 {
		g.Log().Warningf(ctx, "[AiderExecutor] 丢弃可疑 affected_resources: task=%d rejected=%v", req.TaskID, targets.Rejected)
	}
	if len(targets.DirectoryPaths) > 0 {
		if err := ensureDirectoryTargets(workDir, targets.DirectoryPaths); err != nil {
			if req.Workspace != nil && e.wsMgr != nil {
				_ = e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{Success: false, Error: err.Error()})
			}
			return &Result{Success: false, Error: err}
		}
	}
	if len(targets.FilePaths) == 0 && len(targets.DirectoryPaths) > 0 {
		output := fmt.Sprintf("已准备目录资源: %s", strings.Join(targets.DirectoryPaths, ", "))
		if req.Workspace != nil && e.wsMgr != nil {
			if err := finalizeWorkspaceSuccess(ctx, e.wsMgr, req.TaskID, "AiderExecutor"); err != nil {
				_ = e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{Success: false, Error: err.Error(), Retain: true})
				return &Result{Success: false, Error: err}
			}
		}
		return &Result{Success: true, Output: output}
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
		buildStrictAiderTaskPrompt(req.TaskRecord["description"].String(), targets.AllowedPaths), workDir, targets.FilePaths, targets.AllowedPaths, nil)

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
		if err := finalizeWorkspaceSuccess(ctx, e.wsMgr, req.TaskID, "AiderExecutor"); err != nil {
			_ = e.wsMgr.Finalize(ctx, req.TaskID, workspace.FinalizeRequest{Success: false, Error: err.Error(), Retain: true})
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

// NewChatExecutor 创建 Chat 执行器。
func NewChatExecutor() *ChatExecutor { return &ChatExecutor{} }

func (e *ChatExecutor) Name() string         { return "chat" }
func (e *ChatExecutor) NeedsWorkspace() bool { return false }

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
