// Package workspace 提供任务级工作空间隔离能力。
// 当前实现基于 git worktree，为写仓执行器（aider/claude_code/codex_cli/gemini_cli）提供独立工作目录。
package workspace

import (
	"context"
)

// 工作空间类型常量。
const (
	TypeGitWorktree = "git_worktree"
)

// 工作空间状态常量。
const (
	StatusCreating  = "creating"
	StatusReady     = "ready"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"
)

// 清理状态常量。
const (
	CleanupPending  = "pending"
	CleanupDone     = "done"
	CleanupRetained = "retained"
	CleanupFailed   = "failed"
)

// TaskWorkspace 任务工作空间信息。
type TaskWorkspace struct {
	ID            int64  // 雪花 ID
	TaskID        int64  // 任务 ID
	WorkflowRunID int64  // 工作流运行 ID（可选）
	ProjectID     int64  // 项目 ID
	WorkspaceType string // 工作空间类型
	WorkspacePath string // 工作空间绝对路径
	BaseRef       string // 基线引用
	Status        string // 状态
	CleanupStatus string // 清理状态
}

// PrepareRequest 创建工作空间请求。
type PrepareRequest struct {
	TaskID        int64
	WorkflowRunID int64  // 0 表示 legacy 模式
	ProjectID     int64
	WorkDir       string // 项目主工作目录
}

// FinalizeRequest 完成工作空间请求。
type FinalizeRequest struct {
	Success bool
	Retain  bool   // 是否保留（用于调试）
	Error   string // 错误信息
}

// Manager 工作空间管理器接口。
type Manager interface {
	// Prepare 为任务准备隔离工作空间，返回工作空间信息。
	Prepare(ctx context.Context, req PrepareRequest) (*TaskWorkspace, error)

	// MarkRunning 标记工作空间为执行中（任务真正开始执行时调用）。
	MarkRunning(ctx context.Context, taskID int64) error

	// Get 获取任务工作空间信息。
	Get(ctx context.Context, taskID int64) (*TaskWorkspace, error)

	// Finalize 任务结束后收尾：收集 diff、更新状态。
	Finalize(ctx context.Context, taskID int64, req FinalizeRequest) error

	// Cleanup 清理工作空间目录和数据库记录。
	Cleanup(ctx context.Context, taskID int64) error
}

// NeedsIsolation 判断执行模式是否需要工作空间隔离。
// 当前仅 aider 已接入隔离链路；后续接入: openhands, claude_code, codex_cli, gemini_cli。
func NeedsIsolation(executionMode string) bool {
	switch executionMode {
	case "aider":
		return true
	default:
		return false
	}
}
