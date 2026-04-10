package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/utility/snowflake"
	"easymvp/utility/worktreeguard"
)

const worktreeDir = ".mvp-worktrees"
const worktreeArtifactDir = "artifacts"

const (
	workspaceFinalizeTimeout = 45 * time.Second
	workspaceCleanupTimeout  = 90 * time.Second
)

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
	// 1. 校验主工作区是否是合法 Git 仓库；不存在则自动创建
	workDir := filepath.Clean(req.WorkDir)
	if !isGitRepo(workDir) {
		// 目录不存在或不是 git 仓库，自动初始化
		if err := os.MkdirAll(workDir, 0755); err != nil {
			return nil, fmt.Errorf("创建工作目录失败: %s: %w", workDir, err)
		}
		initCmd := exec.Command("git", "init", workDir)
		if output, err := initCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("git init 失败: %s: %s", workDir, string(output))
		}
		g.Log().Infof(ctx, "[Workspace] 自动初始化 git 仓库: %s", workDir)
	}
	req.WorkDir = workDir

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

	// 3. 获取当前 HEAD 作为基线；空仓库会自动补初始提交
	baseRef, err := ensureRepositoryBaseline(ctx, req.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("获取 HEAD 引用失败: %w", err)
	}

	// 4. 创建数据库记录（status=creating）
	ws := &TaskWorkspace{
		TaskID:         req.TaskID,
		WorkflowRunID:  req.WorkflowRunID,
		ProjectID:      req.ProjectID,
		WorkspaceType:  TypeGitWorktree,
		WorkspacePath:  worktreePath,
		BaseRef:        baseRef,
		Status:         StatusCreating,
		CleanupStatus:  CleanupPending,
		DeliveryMode:   DeliveryModePatch,
		DeliveryStatus: DeliveryStatusPending,
		SyncStrategy:   SyncStrategyAutoApply,
		SyncStatus:     SyncStatusPending,
		RiskLevel:      RiskLevelMedium,
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
	finalizeCtx, cancel := context.WithTimeout(context.Background(), workspaceFinalizeTimeout)
	defer cancel()

	ws, err := m.repo.getByTaskID(finalizeCtx, taskID)
	if err != nil {
		return fmt.Errorf("查询工作空间失败: %w", err)
	}
	if ws == nil {
		return fmt.Errorf("任务 %d 没有关联的工作空间", taskID)
	}

	extra := g.Map{}
	policy := resolveDeliveryPolicy(finalizeCtx, taskID, req)
	extra["delivery_mode"] = policy.DeliveryMode
	extra["delivery_status"] = DeliveryStatusPending
	extra["sync_strategy"] = policy.SyncStrategy
	extra["sync_status"] = SyncStatusPending
	extra["risk_level"] = policy.RiskLevel
	extra["patch_ref"] = nil
	extra["delivery_ref"] = nil
	extra["delivery_title"] = nil

	// 收集 diff 摘要
	if req.Success {
		diffSummary, diffErr := gitDiffStat(ws.WorkspacePath)
		if diffErr != nil {
			g.Log().Warningf(finalizeCtx, "[Workspace] 收集 diff 失败: taskID=%d err=%v", taskID, diffErr)
		} else {
			extra["diff_summary"] = diffSummary
		}

		patchContent, hasPatch, patchErr := gitDiffPatch(ws.WorkspacePath)
		if patchErr != nil {
			g.Log().Warningf(finalizeCtx, "[Workspace] 生成 patch 失败: taskID=%d err=%v", taskID, patchErr)
			extra["delivery_status"] = DeliveryStatusFailed
		} else if hasPatch {
			patchRef, writeErr := writePatchArtifact(resolveMainWorkDir(ws.WorkspacePath), taskID, patchContent)
			if writeErr != nil {
				g.Log().Warningf(finalizeCtx, "[Workspace] 写入 patch 失败: taskID=%d err=%v", taskID, writeErr)
				extra["delivery_status"] = DeliveryStatusFailed
			} else {
				extra["patch_ref"] = patchRef
				extra["delivery_status"] = DeliveryStatusReady
			}

			if policy.DeliveryMode == DeliveryModePR && g.NewVar(extra["patch_ref"]).String() != "" {
				deliveryTitle, deliveryRef, artifactErr := writePRArtifact(
					finalizeCtx,
					ws,
					taskID,
					policy.RiskLevel,
					g.NewVar(extra["patch_ref"]).String(),
					g.NewVar(extra["diff_summary"]).String(),
				)
				if artifactErr != nil {
					g.Log().Warningf(finalizeCtx, "[Workspace] 写入 PR 草稿失败: taskID=%d err=%v", taskID, artifactErr)
					extra["delivery_status"] = DeliveryStatusFailed
				} else {
					extra["delivery_title"] = deliveryTitle
					extra["delivery_ref"] = deliveryRef
				}
			}

			if policy.SyncStrategy == SyncStrategyAutoApply {
				if syncErr := syncWorktreeCommit(finalizeCtx, resolveMainWorkDir(ws.WorkspacePath), ws.WorkspacePath, taskID); syncErr != nil {
					extra["sync_status"] = SyncStatusFailed
					extra["error_message"] = syncErr.Error()
					_ = m.repo.updateStatus(finalizeCtx, ws.ID, StatusFailed, extra)
					repairWorkspaceTerminalState(ws.ID, taskID, StatusFailed)
					return fmt.Errorf("同步 worktree 变更失败: %w", syncErr)
				}
				extra["sync_status"] = SyncStatusApplied
			} else {
				extra["sync_status"] = SyncStatusPending
			}

			if g.NewVar(extra["delivery_status"]).String() == DeliveryStatusReady {
				payload := buildDeliveryEventPayload(ws, taskID, policy, extra)
				recordWorkspaceEvent(finalizeCtx, ws, taskID, event.EventTaskDeliveryPrepared, payload)
				if g.NewVar(extra["sync_status"]).String() == SyncStatusApplied {
					recordWorkspaceEvent(finalizeCtx, ws, taskID, event.EventTaskSyncApplied, payload)
				}
				if shouldOpenDeliveryReviewGate(policy, g.NewVar(extra["delivery_status"]).String(), g.NewVar(extra["sync_status"]).String()) {
					reviewPayload := buildDeliveryEventPayload(ws, taskID, policy, extra)
					reviewPayload["reason"] = buildDeliveryReviewReason(policy, reviewPayload)
					recordWorkspaceEvent(finalizeCtx, ws, taskID, event.EventTaskReviewRequired, reviewPayload)
				}
			}
		} else {
			extra["delivery_status"] = DeliveryStatusSkipped
			extra["sync_status"] = SyncStatusSkipped
		}
	} else {
		extra["delivery_status"] = DeliveryStatusSkipped
		extra["sync_status"] = SyncStatusSkipped
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

	if err := m.repo.updateStatus(finalizeCtx, ws.ID, newStatus, extra); err != nil {
		repairWorkspaceTerminalState(ws.ID, taskID, newStatus)
		return fmt.Errorf("更新工作空间最终状态失败: %w", err)
	}

	g.Log().Infof(finalizeCtx, "[Workspace] Finalize: taskID=%d success=%v status=%s", taskID, req.Success, newStatus)
	return nil
}

func buildDeliveryEventPayload(ws *TaskWorkspace, taskID int64, policy deliveryPolicy, extra g.Map) map[string]interface{} {
	payload := map[string]interface{}{
		"project_id":      ws.ProjectID,
		"task_id":         taskID,
		"delivery_mode":   policy.DeliveryMode,
		"delivery_status": g.NewVar(extra["delivery_status"]).String(),
		"sync_strategy":   policy.SyncStrategy,
		"sync_status":     g.NewVar(extra["sync_status"]).String(),
		"risk_level":      policy.RiskLevel,
		"workspace_path":  ws.WorkspacePath,
	}
	if patchRef := g.NewVar(extra["patch_ref"]).String(); patchRef != "" {
		payload["patch_ref"] = patchRef
	}
	if deliveryRef := g.NewVar(extra["delivery_ref"]).String(); deliveryRef != "" {
		payload["delivery_ref"] = deliveryRef
	}
	if deliveryTitle := g.NewVar(extra["delivery_title"]).String(); deliveryTitle != "" {
		payload["delivery_title"] = deliveryTitle
	}
	if diffSummary := g.NewVar(extra["diff_summary"]).String(); diffSummary != "" {
		payload["diff_summary"] = diffSummary
	}
	return payload
}

func shouldOpenDeliveryReviewGate(policy deliveryPolicy, deliveryStatus, syncStatus string) bool {
	if deliveryStatus != DeliveryStatusReady {
		return false
	}
	return policy.DeliveryMode == DeliveryModePR ||
		policy.DeliveryMode == DeliveryModeManual ||
		syncStatus == SyncStatusPending ||
		policy.RiskLevel == RiskLevelHigh
}

func buildDeliveryReviewReason(policy deliveryPolicy, payload map[string]interface{}) string {
	switch {
	case policy.DeliveryMode == DeliveryModePR:
		return "已生成 PR 草稿交付物，等待人工审核或正式提交流程"
	case policy.DeliveryMode == DeliveryModeManual:
		return "当前任务要求人工处理交付结果"
	case policy.RiskLevel == RiskLevelHigh:
		return "高风险任务默认进入人工审核闸门"
	case g.NewVar(payload["sync_status"]).String() == SyncStatusPending:
		return "交付物已准备完成，等待人工确认回写主工作区"
	default:
		return "交付物等待人工确认"
	}
}

func recordWorkspaceEvent(ctx context.Context, ws *TaskWorkspace, taskID int64, eventType string, payload map[string]interface{}) {
	var payloadJSON string
	if len(payload) > 0 {
		if content, err := json.Marshal(payload); err == nil {
			payloadJSON = string(content)
		}
	}

	data := g.Map{
		"id":              int64(snowflake.Generate()),
		"workflow_run_id": ws.WorkflowRunID,
		"entity_type":     event.EntityDomainTask,
		"entity_id":       taskID,
		"event_type":      eventType,
		"created_at":      gtime.Now(),
	}
	if payloadJSON != "" {
		data["payload"] = payloadJSON
	}

	if _, err := g.DB().Model("mvp_workflow_event").Ctx(ctx).Insert(data); err != nil {
		g.Log().Warningf(ctx, "[Workspace] 写入交付事件失败: taskID=%d event=%s err=%v", taskID, eventType, err)
	}
}

type prArtifact struct {
	TaskID         int64    `json:"taskID"`
	ProjectID      int64    `json:"projectID"`
	WorkflowRunID  int64    `json:"workflowRunID"`
	Title          string   `json:"title"`
	Body           string   `json:"body"`
	BaseRef        string   `json:"baseRef"`
	WorkspacePath  string   `json:"workspacePath"`
	PatchRef       string   `json:"patchRef"`
	DiffSummary    string   `json:"diffSummary,omitempty"`
	RiskLevel      string   `json:"riskLevel"`
	ReviewRequired bool     `json:"reviewRequired"`
	Resources      []string `json:"resources,omitempty"`
}

func writePRArtifact(ctx context.Context, ws *TaskWorkspace, taskID int64, riskLevel, patchRef, diffSummary string) (string, string, error) {
	artifact := prArtifact{
		TaskID:         taskID,
		ProjectID:      ws.ProjectID,
		WorkflowRunID:  ws.WorkflowRunID,
		Title:          fmt.Sprintf("[EasyMVP] Task #%d 交付草稿", taskID),
		BaseRef:        ws.BaseRef,
		WorkspacePath:  ws.WorkspacePath,
		PatchRef:       patchRef,
		DiffSummary:    diffSummary,
		RiskLevel:      riskLevel,
		ReviewRequired: true,
	}

	taskRecord, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).
		WhereNull("deleted_at").
		Fields("name, description, affected_resources").
		One()
	if err == nil && !taskRecord.IsEmpty() {
		taskName := strings.TrimSpace(taskRecord["name"].String())
		if taskName != "" {
			artifact.Title = fmt.Sprintf("[EasyMVP] %s", taskName)
		}

		bodyParts := []string{
			fmt.Sprintf("任务ID: %d", taskID),
			fmt.Sprintf("风险等级: %s", riskLevel),
		}
		if desc := strings.TrimSpace(taskRecord["description"].String()); desc != "" {
			bodyParts = append(bodyParts, "任务描述:\n"+desc)
		}
		if diffSummary != "" {
			bodyParts = append(bodyParts, "变更摘要:\n"+diffSummary)
		}
		bodyParts = append(bodyParts, "Patch:\n"+patchRef)
		artifact.Body = strings.Join(bodyParts, "\n\n")

		var resources []string
		if raw := strings.TrimSpace(taskRecord["affected_resources"].String()); raw != "" && raw != "[]" && raw != "null" {
			_ = json.Unmarshal([]byte(raw), &resources)
		}
		artifact.Resources = resources
	}

	if artifact.Body == "" {
		artifact.Body = fmt.Sprintf("任务ID: %d\n风险等级: %s\nPatch: %s", taskID, riskLevel, patchRef)
	}

	artifactDir := filepath.Join(resolveMainWorkDir(ws.WorkspacePath), worktreeDir, worktreeArtifactDir)
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return "", "", err
	}
	filePath := filepath.Join(artifactDir, fmt.Sprintf("task-%d-pr.json", taskID))

	content, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return "", "", err
	}
	if err := os.WriteFile(filePath, append(content, '\n'), 0644); err != nil {
		return "", "", err
	}
	return artifact.Title, filePath, nil
}

// Cleanup 清理工作空间目录。
func (m *GitWorktreeManager) Cleanup(ctx context.Context, taskID int64) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), workspaceCleanupTimeout)
	defer cancel()

	ws, err := m.repo.getByTaskID(cleanupCtx, taskID)
	if err != nil {
		return fmt.Errorf("查询工作空间失败: %w", err)
	}
	if ws == nil {
		return nil // 没有工作空间，无需清理
	}

	if ws.CleanupStatus == CleanupDone {
		return nil // 已清理
	}
	if ws.CleanupStatus == CleanupRetained {
		return nil // 显式保留
	}

	// 获取主工作目录（从 worktree 路径反推）
	mainWorkDir := resolveMainWorkDir(ws.WorkspacePath)

	// 移除 git worktree
	if err := gitWorktreeRemove(mainWorkDir, ws.WorkspacePath); err != nil {
		if !isBenignWorktreeRemoveErr(err) {
			_ = m.repo.updateCleanupStatus(cleanupCtx, ws.ID, CleanupFailed)
			return fmt.Errorf("移除 git worktree 失败: %w", err)
		}
		g.Log().Warningf(cleanupCtx, "[Workspace] worktree 路径已失效，按已清理处理: taskID=%d path=%s err=%v", taskID, ws.WorkspacePath, err)
	}

	// 删除临时分支
	branchName := fmt.Sprintf("mvp-task-%d", taskID)
	_ = gitDeleteBranch(mainWorkDir, branchName) // 忽略错误，分支可能不存在

	_ = os.RemoveAll(ws.WorkspacePath)

	// 更新清理状态
	if err := m.repo.updateCleanupStatus(cleanupCtx, ws.ID, CleanupDone); err != nil {
		return fmt.Errorf("更新清理状态失败: %w", err)
	}

	g.Log().Infof(cleanupCtx, "[Workspace] Cleanup: taskID=%d path=%s", taskID, ws.WorkspacePath)
	return nil
}

func repairWorkspaceTerminalState(workspaceID, taskID int64, workspaceStatus string) {
	now := gtime.Now()
	_, _ = g.DB().Model("mvp_task_workspace").Ctx(context.Background()).
		Where("id", workspaceID).
		Data(g.Map{
			"status":     workspaceStatus,
			"updated_at": now,
		}).
		Update()

	if workspaceStatus == StatusFailed {
		_, _ = g.DB().Model("mvp_domain_task").Ctx(context.Background()).
			Where("id", taskID).
			Where("status", "running").
			Data(g.Map{
				"status":     "failed",
				"result":     "workspace finalize 修复链触发：任务状态从 running 修正为 failed",
				"updated_at": now,
			}).
			Update()
	}
}

func isBenignWorktreeRemoveErr(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(text, "is not a working tree") ||
		strings.Contains(text, "no such file or directory") ||
		strings.Contains(text, "does not exist")
}

// ApplyDelivery 人工确认并将 pending 交付回写到主工作区。
func (m *GitWorktreeManager) ApplyDelivery(ctx context.Context, taskID int64) error {
	ws, err := m.repo.getByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("查询工作空间失败: %w", err)
	}
	if ws == nil {
		return fmt.Errorf("任务 %d 没有关联的工作空间", taskID)
	}
	if ws.DeliveryStatus != DeliveryStatusReady {
		return fmt.Errorf("任务 %d 交付状态为 %s，不允许回写", taskID, ws.DeliveryStatus)
	}
	if ws.SyncStatus == SyncStatusApplied {
		return nil
	}
	if ws.SyncStatus != SyncStatusPending && ws.SyncStatus != SyncStatusFailed {
		return fmt.Errorf("任务 %d 回写状态为 %s，不允许回写", taskID, ws.SyncStatus)
	}
	if !isGitRepo(ws.WorkspacePath) {
		return fmt.Errorf("任务 %d 的工作空间不可用: %s", taskID, ws.WorkspacePath)
	}

	mainWorkDir := resolveMainWorkDir(ws.WorkspacePath)
	if !isGitRepo(mainWorkDir) {
		return fmt.Errorf("主工作区不是 git 仓库: %s", mainWorkDir)
	}

	if err := syncWorktreeCommit(ctx, mainWorkDir, ws.WorkspacePath, taskID); err != nil {
		_ = m.repo.updateStatus(ctx, ws.ID, ws.Status, g.Map{
			"sync_status":   SyncStatusFailed,
			"error_message": err.Error(),
		})
		return err
	}

	extra := g.Map{
		"delivery_status": ws.DeliveryStatus,
		"sync_status":     SyncStatusApplied,
		"risk_level":      ws.RiskLevel,
		"patch_ref":       ws.PatchRef,
		"delivery_ref":    ws.DeliveryRef,
		"delivery_title":  ws.DeliveryTitle,
		"diff_summary":    ws.DiffSummary,
		"error_message":   nil,
	}
	if err := m.repo.updateStatus(ctx, ws.ID, ws.Status, extra); err != nil {
		return fmt.Errorf("更新工作空间回写状态失败: %w", err)
	}

	payload := buildDeliveryEventPayload(ws, taskID, deliveryPolicy{
		DeliveryMode: ws.DeliveryMode,
		SyncStrategy: ws.SyncStrategy,
		RiskLevel:    ws.RiskLevel,
	}, extra)
	recordWorkspaceEvent(ctx, ws, taskID, event.EventTaskSyncApplied, payload)

	g.Log().Infof(ctx, "[Workspace] ManualApplyDelivery: taskID=%d workspace=%s", taskID, ws.WorkspacePath)
	return nil
}

// --- git 命令封装 ---

// isGitRepo 检查目录是否是 Git 仓库。
func isGitRepo(dir string) bool {
	dir = filepath.Clean(dir)

	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	root := filepath.Clean(strings.TrimSpace(string(output)))
	return root == dir
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

func ensureRepositoryBaseline(ctx context.Context, dir string) (string, error) {
	hasHead, err := gitHasHead(dir)
	if err != nil {
		return "", err
	}
	if hasHead {
		return gitHeadRef(dir)
	}

	if err := ensureGitIdentity(dir); err != nil {
		return "", err
	}
	if err := gitAddAll(dir); err != nil {
		return "", fmt.Errorf("暂存初始仓库内容失败: %w", err)
	}
	if err := gitCommitAllowEmpty(dir, "init: project workspace"); err != nil {
		return "", fmt.Errorf("创建初始提交失败: %w", err)
	}

	baseRef, err := gitHeadRef(dir)
	if err != nil {
		return "", fmt.Errorf("读取初始提交失败: %w", err)
	}
	g.Log().Infof(ctx, "[Workspace] 空仓库自动补初始提交: %s baseRef=%s", dir, baseRef)
	return baseRef, nil
}

func gitHasHead(dir string) (bool, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--verify", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 {
			text := string(output)
			if strings.Contains(text, "Needed a single revision") || strings.Contains(text, "unknown revision or path not in the working tree") {
				return false, nil
			}
		}
		return false, fmt.Errorf("%s: %s", err, string(output))
	}
	return true, nil
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
	if err := gitAddAll(worktreePath); err != nil {
		return "", fmt.Errorf("暂存 diff 统计变更失败: %w", err)
	}
	hasChanges, err := gitHasStagedChanges(worktreePath)
	if err != nil {
		return "", fmt.Errorf("检查 diff 统计变更失败: %w", err)
	}
	if !hasChanges {
		return "", nil
	}

	cmd := exec.Command("git", "-C", worktreePath, "diff", "--cached", "--stat", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func gitDiffPatch(worktreePath string) (string, bool, error) {
	if err := gitAddAll(worktreePath); err != nil {
		return "", false, fmt.Errorf("暂存 patch 变更失败: %w", err)
	}
	hasChanges, err := gitHasStagedChanges(worktreePath)
	if err != nil {
		return "", false, fmt.Errorf("检查 patch 变更失败: %w", err)
	}
	if !hasChanges {
		return "", false, nil
	}

	cmd := exec.Command("git", "-C", worktreePath, "diff", "--cached", "--binary", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", false, err
	}
	return string(output), true, nil
}

func writePatchArtifact(mainWorkDir string, taskID int64, content string) (string, error) {
	dir := filepath.Join(mainWorkDir, worktreeDir, worktreeArtifactDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, fmt.Sprintf("task-%d.patch", taskID))
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func syncWorktreeCommit(ctx context.Context, mainWorkDir, worktreePath string, taskID int64) error {
	if err := ensureGitIdentity(mainWorkDir); err != nil {
		return err
	}
	if err := ensureGitIdentity(worktreePath); err != nil {
		return err
	}

	allowPaths, err := loadTaskAllowedPaths(ctx, taskID)
	if err != nil {
		g.Log().Warningf(ctx, "[Workspace] 读取任务允许路径失败，退化为仅校验可疑文件: taskID=%d err=%v", taskID, err)
		allowPaths = nil
	}

	changedFiles, err := stageSyncBackCandidates(ctx, worktreePath, allowPaths, taskID)
	if err != nil {
		return err
	}
	if len(changedFiles) == 0 {
		return nil
	}

	commitMessage := fmt.Sprintf("mvp task %d: apply workspace changes", taskID)
	if err := gitCommit(worktreePath, commitMessage); err != nil {
		return fmt.Errorf("提交 worktree 变更失败: %w", err)
	}

	commitHash, err := gitHeadRef(worktreePath)
	if err != nil {
		return fmt.Errorf("读取 worktree 提交失败: %w", err)
	}

	committedFiles, err := gitChangedFiles(worktreePath, commitHash)
	if err != nil {
		return fmt.Errorf("读取 worktree 变更文件失败: %w", err)
	}
	if err := validateSyncBackPaths(committedFiles, allowPaths); err != nil {
		return err
	}

	dirtyFiles, err := listDirtyMainWorktreeFiles(mainWorkDir)
	if err != nil {
		return err
	}
	if len(dirtyFiles) == 0 {
		if err := gitCherryPick(mainWorkDir, commitHash); err != nil {
			g.Log().Warningf(ctx, "[Workspace] cherry-pick 失败，降级为文件同步: taskID=%d commit=%s err=%v", taskID, commitHash, err)
		} else {
			g.Log().Infof(ctx, "[Workspace] SyncBack: taskID=%d commit=%s mode=cherry-pick", taskID, commitHash)
			return nil
		}
	}

	if err := syncChangedFilesToMain(mainWorkDir, worktreePath, committedFiles, dirtyFiles, allowPaths); err != nil {
		return err
	}

	if len(dirtyFiles) == 0 {
		if err := commitSyncedFilesToMain(mainWorkDir, taskID); err != nil {
			return err
		}
	}

	g.Log().Infof(ctx, "[Workspace] SyncBack: taskID=%d commit=%s mode=copy", taskID, commitHash)
	return nil
}

func stageSyncBackCandidates(ctx context.Context, worktreePath string, allowPaths []string, taskID int64) ([]gitChangedFile, error) {
	if err := gitAddAll(worktreePath); err != nil {
		return nil, fmt.Errorf("暂存 worktree 变更失败: %w", err)
	}
	hasChanges, err := gitHasStagedChanges(worktreePath)
	if err != nil {
		return nil, fmt.Errorf("检查 worktree 变更失败: %w", err)
	}
	if !hasChanges {
		return nil, nil
	}

	if pruned, pruneErr := worktreeguard.PruneEmbeddedAllowedDuplicates(ctx, worktreePath, allowPaths); pruneErr != nil {
		g.Log().Warningf(ctx, "[Workspace] syncBack 前清理重复嵌入路径失败: taskID=%d err=%v", taskID, pruneErr)
	} else if len(pruned) > 0 {
		g.Log().Infof(ctx, "[Workspace] syncBack 前已清理重复嵌入路径: taskID=%d paths=%v", taskID, pruned)
		if err := gitAddAll(worktreePath); err != nil {
			return nil, fmt.Errorf("重复路径清理后重新暂存失败: %w", err)
		}
	}

	changedFiles, err := gitStagedChangedFiles(worktreePath)
	if err != nil {
		return nil, fmt.Errorf("读取 staged 变更文件失败: %w", err)
	}
	suspicious, _ := collectSyncBackPathIssues(changedFiles, allowPaths)
	if len(suspicious) > 0 {
		if pruned, pruneErr := worktreeguard.PruneSuspiciousDeltaPaths(worktreePath, suspicious); pruneErr != nil {
			g.Log().Warningf(ctx, "[Workspace] syncBack 前清理可疑标题文件失败: taskID=%d err=%v", taskID, pruneErr)
		} else if len(pruned) > 0 {
			g.Log().Infof(ctx, "[Workspace] syncBack 前已清理可疑标题文件: taskID=%d paths=%v", taskID, pruned)
			if err := gitAddAll(worktreePath); err != nil {
				return nil, fmt.Errorf("可疑文件清理后重新暂存失败: %w", err)
			}
			changedFiles, err = gitStagedChangedFiles(worktreePath)
			if err != nil {
				return nil, fmt.Errorf("清理后读取 staged 变更文件失败: %w", err)
			}
		}
	}

	hasChanges, err = gitHasStagedChanges(worktreePath)
	if err != nil {
		return nil, fmt.Errorf("清理后检查 worktree 变更失败: %w", err)
	}
	if !hasChanges {
		return nil, nil
	}
	if err := validateSyncBackPaths(changedFiles, allowPaths); err != nil {
		return nil, err
	}
	return changedFiles, nil
}

func ensureGitIdentity(dir string) error {
	name, _ := gitConfigGet(dir, "user.name")
	if strings.TrimSpace(name) == "" {
		if err := gitConfigSet(dir, "user.name", "EasyMVP"); err != nil {
			return fmt.Errorf("设置 git user.name 失败: %w", err)
		}
	}

	email, _ := gitConfigGet(dir, "user.email")
	if strings.TrimSpace(email) == "" {
		if err := gitConfigSet(dir, "user.email", "mvp@easymvp.local"); err != nil {
			return fmt.Errorf("设置 git user.email 失败: %w", err)
		}
	}
	return nil
}

func gitConfigGet(dir, key string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "config", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func gitConfigSet(dir, key, value string) error {
	cmd := exec.Command("git", "-C", dir, "config", key, value)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

func listDirtyMainWorktreeFiles(mainWorkDir string) (map[string]struct{}, error) {
	cmd := exec.Command("git", "-C", mainWorkDir, "status", "--porcelain=v1", "--untracked-files=all")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("读取主工作区 git 状态失败: %w", err)
	}

	result := make(map[string]struct{})
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			path = strings.TrimSpace(parts[len(parts)-1])
		}
		path = strings.ReplaceAll(path, "\\", "/")
		if path == "" || path == "." || path == ".mvp-worktrees" || strings.HasPrefix(path, ".mvp-worktrees/") {
			continue
		}
		result[path] = struct{}{}
	}
	return result, nil
}

func gitAddAll(dir string) error {
	cmd := exec.Command("git", "-C", dir, "add", "-A")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

func gitHasStagedChanges(dir string) (bool, error) {
	cmd := exec.Command("git", "-C", dir, "diff", "--cached", "--quiet", "--exit-code")
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func gitCommit(dir, message string) error {
	cmd := exec.Command("git", "-C", dir, "commit", "--no-verify", "-m", message)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

func gitCommitAllowEmpty(dir, message string) error {
	cmd := exec.Command("git", "-C", dir, "commit", "--allow-empty", "--no-verify", "-m", message)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

func gitCherryPick(dir, commitHash string) error {
	cmd := exec.Command("git", "-C", dir, "cherry-pick", commitHash)
	if output, err := cmd.CombinedOutput(); err != nil {
		abortCmd := exec.Command("git", "-C", dir, "cherry-pick", "--abort")
		_, _ = abortCmd.CombinedOutput()
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

type gitChangedFile struct {
	OldPath string
	NewPath string
	Status  string
}

func gitChangedFiles(dir, commitHash string) ([]gitChangedFile, error) {
	cmd := exec.Command("git", "-C", dir, "diff-tree", "--no-commit-id", "--name-status", "-r", "-M", commitHash)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseGitChangedFiles(string(output)), nil
}

func gitStagedChangedFiles(dir string) ([]gitChangedFile, error) {
	cmd := exec.Command("git", "-C", dir, "diff", "--cached", "--name-status", "-M")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseGitChangedFiles(string(output)), nil
}

func parseGitChangedFiles(output string) []gitChangedFile {
	var result []gitChangedFile
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		item := gitChangedFile{Status: fields[0]}
		if strings.HasPrefix(fields[0], "R") || strings.HasPrefix(fields[0], "C") {
			if len(fields) < 3 {
				continue
			}
			item.OldPath = filepath.Clean(fields[1])
			item.NewPath = filepath.Clean(fields[2])
		} else {
			item.NewPath = filepath.Clean(fields[1])
		}
		result = append(result, item)
	}
	return result
}

func syncChangedFilesToMain(mainWorkDir, worktreePath string, changedFiles []gitChangedFile, dirtyFiles map[string]struct{}, allowPaths []string) error {
	if err := validateSyncBackPaths(changedFiles, allowPaths); err != nil {
		return err
	}
	for _, file := range changedFiles {
		targets := []string{}
		if file.OldPath != "" {
			targets = append(targets, file.OldPath)
		}
		if file.NewPath != "" {
			targets = append(targets, file.NewPath)
		}
		for _, target := range targets {
			if _, exists := dirtyFiles[target]; exists {
				return fmt.Errorf("主工作区存在冲突中的未提交变更: %s", target)
			}
		}

		if strings.HasPrefix(file.Status, "D") || strings.HasPrefix(file.Status, "R") {
			if file.OldPath != "" {
				_ = os.Remove(filepath.Join(mainWorkDir, file.OldPath))
			}
		}
		if strings.HasPrefix(file.Status, "D") {
			continue
		}

		srcPath := filepath.Join(worktreePath, file.NewPath)
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			return fmt.Errorf("读取 worktree 文件失败: %s: %w", file.NewPath, err)
		}
		dstPath := filepath.Join(mainWorkDir, file.NewPath)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("创建目标目录失败: %s: %w", file.NewPath, err)
		}
		if err := copyFile(srcPath, dstPath, srcInfo.Mode()); err != nil {
			return fmt.Errorf("同步文件失败: %s: %w", file.NewPath, err)
		}
	}
	return nil
}

func commitSyncedFilesToMain(mainWorkDir string, taskID int64) error {
	if err := gitAddAll(mainWorkDir); err != nil {
		return fmt.Errorf("暂存主工作区同步结果失败: %w", err)
	}
	hasChanges, err := gitHasStagedChanges(mainWorkDir)
	if err != nil {
		return fmt.Errorf("检查主工作区同步结果失败: %w", err)
	}
	if !hasChanges {
		return nil
	}

	commitMessage := fmt.Sprintf("mvp task %d: apply workspace changes", taskID)
	if err := gitCommit(mainWorkDir, commitMessage); err != nil {
		return fmt.Errorf("提交主工作区同步结果失败: %w", err)
	}
	return nil
}

func loadTaskAllowedPaths(ctx context.Context, taskID int64) (_ []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("读取任务允许路径 panic: %v", r)
		}
	}()

	record, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).
		WhereNull("deleted_at").
		Fields("affected_resources").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, nil
	}

	raw := strings.TrimSpace(record["affected_resources"].String())
	if raw == "" || raw == "null" {
		return nil, nil
	}

	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	normalized, _ := worktreeguard.NormalizeRelativePaths(values)
	return normalized, nil
}

func validateSyncBackPaths(changedFiles []gitChangedFile, allowPaths []string) error {
	suspicious, invalid := collectSyncBackPathIssues(changedFiles, allowPaths)
	if len(suspicious) == 0 && len(invalid) == 0 {
		return nil
	}

	sort.Strings(suspicious)
	sort.Strings(invalid)

	var issues []string
	if len(suspicious) > 0 {
		issues = append(issues, "检测到可疑文件: "+strings.Join(suspicious, ", "))
	}
	if len(invalid) > 0 {
		issues = append(issues, "检测到越界修改: "+strings.Join(invalid, ", "))
	}
	return fmt.Errorf("syncBack 校验失败: %s", strings.Join(issues, "；"))
}

func collectSyncBackPathIssues(changedFiles []gitChangedFile, allowPaths []string) ([]string, []string) {
	var (
		invalid        []string
		suspicious     []string
		seenInvalid    = map[string]struct{}{}
		seenSuspicious = map[string]struct{}{}
		allowSet       = make(map[string]struct{}, len(allowPaths))
	)
	for _, allowPath := range allowPaths {
		allowSet[allowPath] = struct{}{}
	}

	for _, item := range changedFiles {
		for _, filePath := range changedFileTargets(item) {
			filePath = path.Clean(strings.ReplaceAll(strings.TrimSpace(filePath), "\\", "/"))
			if filePath == "" || filePath == "." {
				continue
			}
			if worktreeguard.IsSuspiciousPath(filePath) {
				if _, exists := seenSuspicious[filePath]; !exists {
					seenSuspicious[filePath] = struct{}{}
					suspicious = append(suspicious, filePath)
				}
				continue
			}
			if len(allowSet) > 0 && !isAllowedSyncBackPath(filePath, allowSet) {
				if _, exists := seenInvalid[filePath]; !exists {
					seenInvalid[filePath] = struct{}{}
					invalid = append(invalid, filePath)
				}
			}
		}
	}
	return suspicious, invalid
}

func changedFileTargets(item gitChangedFile) []string {
	targets := make([]string, 0, 2)
	if strings.TrimSpace(item.OldPath) != "" {
		targets = append(targets, item.OldPath)
	}
	if strings.TrimSpace(item.NewPath) != "" {
		targets = append(targets, item.NewPath)
	}
	return targets
}

func isAllowedSyncBackPath(filePath string, allowSet map[string]struct{}) bool {
	for allowPath := range allowSet {
		if filePath == allowPath || strings.HasPrefix(filePath, allowPath+"/") {
			return true
		}
	}
	return false
}

func copyFile(srcPath, dstPath string, mode os.FileMode) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

// resolveMainWorkDir 从 worktree 路径反推主工作目录。
// worktree 路径格式: {work_dir}/.mvp-worktrees/task-{id}
func resolveMainWorkDir(worktreePath string) string {
	// 向上两级: task-xxx -> .mvp-worktrees -> work_dir
	return filepath.Dir(filepath.Dir(worktreePath))
}
