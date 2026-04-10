package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// OrphanSweepConfig orphan worktree 对账配置。
type OrphanSweepConfig struct {
	AutoCleanupDiskOrphans    bool // 自动清理“磁盘存在但 DB 无记录”的 worktree
	AutoRepairMissingOnDisk   bool // 自动修复“DB 有记录但磁盘不存在”的 workspace 状态
	AutoRepairRunningMismatch bool // 自动修复 workspace.running 与 domain_task 非 running 的状态不一致
}

// DefaultOrphanSweepConfig 返回默认 orphan sweep 配置。
func DefaultOrphanSweepConfig() OrphanSweepConfig {
	return OrphanSweepConfig{
		AutoCleanupDiskOrphans:    false,
		AutoRepairMissingOnDisk:   true,
		AutoRepairRunningMismatch: true,
	}
}

// OrphanSweepReport orphan 对账报告。
type OrphanSweepReport struct {
	ScannedRoots            int
	DBWorkspaces            int
	DiskWorktrees           int
	DiskOrphans             int
	MissingOnDisk           int
	RunningMismatch         int
	CleanedDiskOrphans      int
	RepairedMissingOnDisk   int
	RepairedRunningMismatch int
	Errors                  int
}

type workspaceSweepRecord struct {
	ID            int64
	TaskID        int64
	WorkspacePath string
	Status        string
	CleanupStatus string
}

// RunOrphanSweep 扫描并对账磁盘 worktree 与 DB 记录。
func RunOrphanSweep(ctx context.Context, mgr *GitWorktreeManager, cfg OrphanSweepConfig) (*OrphanSweepReport, error) {
	if mgr == nil {
		return nil, fmt.Errorf("workspace manager 为空")
	}

	sweepCtx, cancel := context.WithTimeout(context.Background(), workspaceCleanupTimeout)
	defer cancel()

	dbRecords, err := listWorkspaceSweepRecords(sweepCtx)
	if err != nil {
		return nil, err
	}

	report := &OrphanSweepReport{
		DBWorkspaces: len(dbRecords),
	}

	dbByPath := make(map[string]workspaceSweepRecord, len(dbRecords))
	roots := make(map[string]struct{})
	runningTaskIDs := make([]int64, 0, len(dbRecords))
	for _, ws := range dbRecords {
		normalizedPath := filepath.Clean(ws.WorkspacePath)
		if normalizedPath == "" {
			continue
		}
		dbByPath[normalizedPath] = ws
		roots[resolveMainWorkDir(normalizedPath)] = struct{}{}
		if ws.Status == StatusRunning && ws.TaskID > 0 {
			runningTaskIDs = append(runningTaskIDs, ws.TaskID)
		}
	}

	diskSet := make(map[string]struct{})
	for root := range roots {
		report.ScannedRoots++
		diskPaths, scanErr := scanDiskWorktrees(root)
		if scanErr != nil {
			report.Errors++
			g.Log().Warningf(sweepCtx, "[Workspace.OrphanSweep] 扫描磁盘 worktree 失败: root=%s err=%v", root, scanErr)
			continue
		}
		report.DiskWorktrees += len(diskPaths)
		for _, diskPath := range diskPaths {
			normalizedPath := filepath.Clean(diskPath)
			diskSet[normalizedPath] = struct{}{}
			if _, ok := dbByPath[normalizedPath]; ok {
				continue
			}
			report.DiskOrphans++
			if !cfg.AutoCleanupDiskOrphans {
				continue
			}
			if cleanErr := cleanupDiskOrphanWorktree(sweepCtx, root, normalizedPath); cleanErr != nil {
				report.Errors++
				g.Log().Warningf(sweepCtx, "[Workspace.OrphanSweep] 清理磁盘 orphan 失败: root=%s path=%s err=%v", root, normalizedPath, cleanErr)
				continue
			}
			report.CleanedDiskOrphans++
		}
	}

	taskStatusMap, statusErr := queryDomainTaskStatusMap(sweepCtx, runningTaskIDs)
	if statusErr != nil {
		report.Errors++
		g.Log().Warningf(sweepCtx, "[Workspace.OrphanSweep] 查询 domain_task 状态失败: err=%v", statusErr)
	}

	for _, ws := range dbRecords {
		normalizedPath := filepath.Clean(ws.WorkspacePath)
		if normalizedPath == "" {
			continue
		}

		if _, ok := diskSet[normalizedPath]; !ok {
			report.MissingOnDisk++
			if cfg.AutoRepairMissingOnDisk {
				if repairErr := repairWorkspaceMissingOnDisk(sweepCtx, ws); repairErr != nil {
					report.Errors++
					g.Log().Warningf(sweepCtx, "[Workspace.OrphanSweep] 修复 DB orphan 失败: workspace=%d task=%d err=%v", ws.ID, ws.TaskID, repairErr)
				} else {
					report.RepairedMissingOnDisk++
				}
			}
		}

		if ws.Status != StatusRunning {
			continue
		}
		taskStatus, taskExists := taskStatusMap[ws.TaskID]
		if taskExists && taskStatus == "running" {
			continue
		}

		report.RunningMismatch++
		if !cfg.AutoRepairRunningMismatch {
			continue
		}
		if repairErr := repairWorkspaceRunningMismatch(sweepCtx, ws, taskStatus, taskExists); repairErr != nil {
			report.Errors++
			g.Log().Warningf(sweepCtx, "[Workspace.OrphanSweep] 修复假 running 失败: workspace=%d task=%d err=%v", ws.ID, ws.TaskID, repairErr)
			continue
		}
		report.RepairedRunningMismatch++
	}

	g.Log().Infof(sweepCtx, "[Workspace.OrphanSweep] 完成: roots=%d db=%d disk=%d disk_orphan=%d db_orphan=%d running_mismatch=%d repaired_missing=%d repaired_running=%d cleaned_disk=%d errors=%d",
		report.ScannedRoots, report.DBWorkspaces, report.DiskWorktrees, report.DiskOrphans, report.MissingOnDisk, report.RunningMismatch,
		report.RepairedMissingOnDisk, report.RepairedRunningMismatch, report.CleanedDiskOrphans, report.Errors)
	return report, nil
}

func listWorkspaceSweepRecords(ctx context.Context) ([]workspaceSweepRecord, error) {
	records, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		WhereNull("deleted_at").
		Fields("id, task_id, workspace_path, status, cleanup_status").
		All()
	if err != nil {
		return nil, err
	}

	result := make([]workspaceSweepRecord, 0, len(records))
	for _, record := range records {
		result = append(result, workspaceSweepRecord{
			ID:            record["id"].Int64(),
			TaskID:        record["task_id"].Int64(),
			WorkspacePath: record["workspace_path"].String(),
			Status:        record["status"].String(),
			CleanupStatus: record["cleanup_status"].String(),
		})
	}
	return result, nil
}

func queryDomainTaskStatusMap(ctx context.Context, taskIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	if len(taskIDs) == 0 {
		return result, nil
	}
	rows, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		WhereIn("id", taskIDs).
		WhereNull("deleted_at").
		Fields("id, status").
		All()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row["id"].Int64()] = strings.TrimSpace(row["status"].String())
	}
	return result, nil
}

func scanDiskWorktrees(mainWorkDir string) ([]string, error) {
	worktreesRoot := filepath.Join(mainWorkDir, worktreeDir)
	entries, err := os.ReadDir(worktreesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !isTaskWorktreeName(entry.Name()) {
			continue
		}
		result = append(result, filepath.Join(worktreesRoot, entry.Name()))
	}
	return result, nil
}

func cleanupDiskOrphanWorktree(ctx context.Context, mainWorkDir, worktreePath string) error {
	if err := gitWorktreeRemove(mainWorkDir, worktreePath); err != nil && !isBenignWorktreeRemoveErr(err) {
		return err
	}

	if taskID, ok := parseTaskIDFromWorktreeName(filepath.Base(worktreePath)); ok {
		_ = gitDeleteBranch(mainWorkDir, fmt.Sprintf("mvp-task-%d", taskID))
	}
	return os.RemoveAll(worktreePath)
}

func repairWorkspaceMissingOnDisk(ctx context.Context, ws workspaceSweepRecord) error {
	now := gtime.Now()

	workspaceStatus := ws.Status
	cleanupStatus := CleanupDone
	if ws.Status == StatusRunning || ws.Status == StatusReady || ws.Status == StatusCreating {
		workspaceStatus = StatusFailed
		cleanupStatus = CleanupPending
	}
	if ws.CleanupStatus == CleanupRetained {
		cleanupStatus = CleanupRetained
	}

	data := g.Map{
		"status":         workspaceStatus,
		"cleanup_status": cleanupStatus,
		"updated_at":     now,
	}
	if _, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		Where("id", ws.ID).
		Data(data).
		Update(); err != nil {
		return err
	}

	if workspaceStatus == StatusFailed {
		_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("id", ws.TaskID).
			Where("status", "running").
			Data(g.Map{
				"status":     "failed",
				"result":     "workspace orphan 修复：磁盘 worktree 丢失，任务从 running 修正为 failed",
				"updated_at": now,
			}).
			Update()
	}

	return nil
}

func repairWorkspaceRunningMismatch(ctx context.Context, ws workspaceSweepRecord, taskStatus string, taskExists bool) error {
	now := gtime.Now()
	reason := "workspace running 修复：domain_task 状态不匹配"
	if !taskExists {
		reason = "workspace running 修复：domain_task 记录不存在"
	} else if strings.TrimSpace(taskStatus) != "" {
		reason = fmt.Sprintf("workspace running 修复：domain_task 当前状态=%s", taskStatus)
	}

	if _, err := g.DB().Model("mvp_task_workspace").Ctx(ctx).
		Where("id", ws.ID).
		Where("status", StatusRunning).
		Data(g.Map{
			"status":         StatusFailed,
			"cleanup_status": CleanupPending,
			"updated_at":     now,
			"error_message":  reason,
		}).
		Update(); err != nil {
		// 兼容旧表结构，error_message 列不存在时降级写入。
		if _, fallbackErr := g.DB().Model("mvp_task_workspace").Ctx(ctx).
			Where("id", ws.ID).
			Where("status", StatusRunning).
			Data(g.Map{
				"status":         StatusFailed,
				"cleanup_status": CleanupPending,
				"updated_at":     now,
			}).
			Update(); fallbackErr != nil {
			return fallbackErr
		}
	}

	_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", ws.TaskID).
		Where("status", "running").
		Data(g.Map{
			"status":     "failed",
			"result":     reason,
			"updated_at": now,
		}).
		Update()
	return nil
}

func isTaskWorktreeName(name string) bool {
	name = strings.TrimSpace(name)
	if !strings.HasPrefix(name, "task-") {
		return false
	}
	_, ok := parseTaskIDFromWorktreeName(name)
	return ok
}

func parseTaskIDFromWorktreeName(name string) (int64, bool) {
	name = strings.TrimSpace(name)
	if !strings.HasPrefix(name, "task-") {
		return 0, false
	}
	raw := strings.TrimPrefix(name, "task-")
	if raw == "" {
		return 0, false
	}
	taskID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || taskID <= 0 {
		return 0, false
	}
	return taskID, true
}
