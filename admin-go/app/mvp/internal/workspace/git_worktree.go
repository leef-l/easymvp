package workspace

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

const worktreeDir = ".mvp-worktrees"

// GitWorktreeManager 基于 git worktree 的工作空间管理器。
type GitWorktreeManager struct {
	repo *repo
}

// NewGitWorktreeManager 创建 Git Worktree 管理器。
func NewGitWorktreeManager() *GitWorktreeManager {
	return &GitWorktreeManager{repo: defaultRepo}
}

// Prepare 为任务创建独立的 git worktree。
func (m *GitWorktreeManager) Prepare(ctx context.Context, req PrepareRequest) (*TaskWorkspace, error) {
	// 1. 校验主工作区是否是合法 Git 仓库
	if !isGitRepo(req.WorkDir) {
		return nil, fmt.Errorf("工作目录不是 Git 仓库: %s", req.WorkDir)
	}

	worktreePath := filepath.Join(req.WorkDir, worktreeDir, fmt.Sprintf("task-%d", req.TaskID))
	branchName := fmt.Sprintf("mvp-task-%d", req.TaskID)

	// 2. 幂等检查：如果已有记录，按状态恢复
	existing, err := m.repo.getByTaskID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("查询已有工作空间失败: %w", err)
	}
	if existing != nil {
		switch existing.Status {
		case StatusReady, StatusRunning:
			// 校验磁盘目录和 git 可用性
			if isGitRepo(existing.WorkspacePath) {
				g.Log().Infof(ctx, "[Workspace] 复用已有 worktree: taskID=%d status=%s", req.TaskID, existing.Status)
				return existing, nil
			}
			// 磁盘已损坏/被清理，按残留处理
			g.Log().Warningf(ctx, "[Workspace] worktree 目录不可用，清理重建: taskID=%d path=%s", req.TaskID, existing.WorkspacePath)
			_ = gitWorktreeRemove(req.WorkDir, worktreePath)
			_ = gitDeleteBranch(req.WorkDir, branchName)
			_ = os.RemoveAll(worktreePath)
			_ = m.repo.softDelete(ctx, existing.ID)
		case StatusCreating, StatusFailed:
			// 清理残留后重建
			g.Log().Infof(ctx, "[Workspace] 清理残留 worktree 后重建: taskID=%d status=%s", req.TaskID, existing.Status)
			_ = gitWorktreeRemove(req.WorkDir, worktreePath)
			_ = gitDeleteBranch(req.WorkDir, branchName)
			_ = os.RemoveAll(worktreePath)
			_ = m.repo.softDelete(ctx, existing.ID)
		default:
			// completed/canceled 等终态，清理后重建
			_ = gitWorktreeRemove(req.WorkDir, worktreePath)
			_ = gitDeleteBranch(req.WorkDir, branchName)
			_ = os.RemoveAll(worktreePath)
			_ = m.repo.softDelete(ctx, existing.ID)
		}
	}

	// 3. 获取当前 HEAD 作为基线
	baseRef, err := gitHeadRef(req.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("获取 HEAD 引用失败: %w", err)
	}

	// 4. 创建数据库记录（status=creating）
	ws := &TaskWorkspace{
		TaskID:        req.TaskID,
		WorkflowRunID: req.WorkflowRunID,
		ProjectID:     req.ProjectID,
		WorkspaceType: TypeGitWorktree,
		WorkspacePath: worktreePath,
		BaseRef:       baseRef,
		Status:        StatusCreating,
		CleanupStatus: CleanupPending,
	}
	if err := m.repo.create(ctx, ws); err != nil {
		return nil, fmt.Errorf("创建工作空间记录失败: %w", err)
	}

	// 5. 确保父目录存在
	parentDir := filepath.Dir(worktreePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		_ = m.repo.updateStatus(ctx, ws.ID, StatusFailed, g.Map{"error_message": err.Error()})
		return nil, fmt.Errorf("创建 worktree 父目录失败: %w", err)
	}

	// 6. 创建 git worktree（先清理可能残留的同名分支）
	_ = gitDeleteBranch(req.WorkDir, branchName)
	if err := gitWorktreeAdd(req.WorkDir, worktreePath, baseRef, branchName); err != nil {
		_ = m.repo.updateStatus(ctx, ws.ID, StatusFailed, g.Map{"error_message": err.Error()})
		return nil, fmt.Errorf("创建 git worktree 失败: %w", err)
	}

	// 7. 更新状态为 ready
	if err := m.repo.updateStatus(ctx, ws.ID, StatusReady, nil); err != nil {
		return nil, fmt.Errorf("更新工作空间状态失败: %w", err)
	}
	ws.Status = StatusReady

	g.Log().Infof(ctx, "[Workspace] 创建 worktree: taskID=%d path=%s baseRef=%s", req.TaskID, worktreePath, baseRef)
	return ws, nil
}

// MarkRunning 标记工作空间为执行中。
func (m *GitWorktreeManager) MarkRunning(ctx context.Context, taskID int64) error {
	ws, err := m.repo.getByTaskID(ctx, taskID)
	if err != nil {
		return err
	}
	if ws == nil {
		return fmt.Errorf("任务 %d 没有关联的工作空间", taskID)
	}
	if ws.Status != StatusReady {
		return nil // 非 ready 状态不更新
	}
	return m.repo.updateStatus(ctx, ws.ID, StatusRunning, nil)
}

// Get 获取任务工作空间信息。
func (m *GitWorktreeManager) Get(ctx context.Context, taskID int64) (*TaskWorkspace, error) {
	return m.repo.getByTaskID(ctx, taskID)
}

// Finalize 任务结束后收尾。
func (m *GitWorktreeManager) Finalize(ctx context.Context, taskID int64, req FinalizeRequest) error {
	ws, err := m.repo.getByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("查询工作空间失败: %w", err)
	}
	if ws == nil {
		return fmt.Errorf("任务 %d 没有关联的工作空间", taskID)
	}

	extra := g.Map{}

	// 收集 diff 摘要
	if req.Success {
		diffSummary, diffErr := gitDiffStat(ws.WorkspacePath)
		if diffErr != nil {
			g.Log().Warningf(ctx, "[Workspace] 收集 diff 失败: taskID=%d err=%v", taskID, diffErr)
		} else {
			extra["diff_summary"] = diffSummary
		}
	}

	// 设置最终状态
	newStatus := StatusCompleted
	if !req.Success {
		newStatus = StatusFailed
		extra["error_message"] = req.Error
	}

	// 显式保留（调试用），否则留给定时清理按保留期处理
	if req.Retain {
		extra["cleanup_status"] = CleanupRetained
	}
	// 失败任务保持 cleanup_status=pending，由 RunCleanup 按 72h 策略自动清理

	if err := m.repo.updateStatus(ctx, ws.ID, newStatus, extra); err != nil {
		return fmt.Errorf("更新工作空间最终状态失败: %w", err)
	}

	g.Log().Infof(ctx, "[Workspace] Finalize: taskID=%d success=%v status=%s", taskID, req.Success, newStatus)
	return nil
}

// Cleanup 清理工作空间目录。
func (m *GitWorktreeManager) Cleanup(ctx context.Context, taskID int64) error {
	ws, err := m.repo.getByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("查询工作空间失败: %w", err)
	}
	if ws == nil {
		return nil // 没有工作空间，无需清理
	}

	if ws.CleanupStatus == CleanupDone {
		return nil // 已清理
	}

	// 获取主工作目录（从 worktree 路径反推）
	mainWorkDir := resolveMainWorkDir(ws.WorkspacePath)

	// 移除 git worktree
	if err := gitWorktreeRemove(mainWorkDir, ws.WorkspacePath); err != nil {
		_ = m.repo.updateCleanupStatus(ctx, ws.ID, CleanupFailed)
		return fmt.Errorf("移除 git worktree 失败: %w", err)
	}

	// 删除临时分支
	branchName := fmt.Sprintf("mvp-task-%d", taskID)
	_ = gitDeleteBranch(mainWorkDir, branchName) // 忽略错误，分支可能不存在

	// 更新清理状态
	if err := m.repo.updateCleanupStatus(ctx, ws.ID, CleanupDone); err != nil {
		return fmt.Errorf("更新清理状态失败: %w", err)
	}

	g.Log().Infof(ctx, "[Workspace] Cleanup: taskID=%d path=%s", taskID, ws.WorkspacePath)
	return nil
}

// --- git 命令封装 ---

// isGitRepo 检查目录是否是 Git 仓库。
func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

// gitHeadRef 获取当前 HEAD 的 commit hash。
func gitHeadRef(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// gitWorktreeAdd 创建 worktree 并基于指定 commit 创建新分支。
func gitWorktreeAdd(mainDir, worktreePath, baseRef, branchName string) error {
	cmd := exec.Command("git", "-C", mainDir, "worktree", "add", "-b", branchName, worktreePath, baseRef)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

// gitWorktreeRemove 移除 worktree。
func gitWorktreeRemove(mainDir, worktreePath string) error {
	// 先尝试正常移除
	cmd := exec.Command("git", "-C", mainDir, "worktree", "remove", worktreePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果目录已不存在，尝试 prune
		cmd2 := exec.Command("git", "-C", mainDir, "worktree", "prune")
		_ = cmd2.Run()

		// 如果仍然失败，强制移除
		cmd3 := exec.Command("git", "-C", mainDir, "worktree", "remove", "--force", worktreePath)
		output2, err2 := cmd3.CombinedOutput()
		if err2 != nil {
			return fmt.Errorf("移除失败: %s / 强制移除: %s", string(output), string(output2))
		}
	}
	return nil
}

// gitDeleteBranch 删除本地分支。
func gitDeleteBranch(mainDir, branchName string) error {
	cmd := exec.Command("git", "-C", mainDir, "branch", "-D", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

// gitDiffStat 获取 worktree 中的变更统计。
func gitDiffStat(worktreePath string) (string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "diff", "--stat", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// resolveMainWorkDir 从 worktree 路径反推主工作目录。
// worktree 路径格式: {work_dir}/.mvp-worktrees/task-{id}
func resolveMainWorkDir(worktreePath string) string {
	// 向上两级: task-xxx -> .mvp-worktrees -> work_dir
	return filepath.Dir(filepath.Dir(worktreePath))
}
