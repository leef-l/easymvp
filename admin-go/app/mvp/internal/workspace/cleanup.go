package workspace

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// CleanupConfig 清理配置。
type CleanupConfig struct {
	SuccessRetainHours  int // 成功任务保留时长（小时），默认 24
	FailedRetainHours   int // 失败任务保留时长（小时），默认 72
	CanceledRetainHours int // 取消任务保留时长（小时），默认 24
}

// DefaultCleanupConfig 返回默认清理配置。
func DefaultCleanupConfig() CleanupConfig {
	return CleanupConfig{
		SuccessRetainHours:  24,
		FailedRetainHours:   72,
		CanceledRetainHours: 24,
	}
}

// RunCleanup 执行一轮清理：扫描已过保留期的工作空间并清理。
func RunCleanup(ctx context.Context, mgr *GitWorktreeManager, cfg CleanupConfig) (cleaned int, errs int) {
	if mgr == nil {
		return 0, 0
	}
	cleanupCtx, cancel := context.WithTimeout(context.Background(), workspaceCleanupTimeout+30*time.Second)
	defer cancel()

	// 按最短保留期查询候选（24小时）
	minHours := cfg.SuccessRetainHours
	if cfg.CanceledRetainHours < minHours {
		minHours = cfg.CanceledRetainHours
	}

	candidates, err := mgr.repo.listPendingCleanup(cleanupCtx, minHours)
	if err != nil {
		g.Log().Errorf(cleanupCtx, "[Workspace.Cleanup] 查询待清理工作空间失败: %v", err)
		return 0, 1
	}

	for _, ws := range candidates {
		// 按状态判断是否超过保留期
		var retainHours int
		switch ws.Status {
		case StatusCompleted:
			retainHours = cfg.SuccessRetainHours
		case StatusFailed:
			retainHours = cfg.FailedRetainHours
		case StatusCanceled:
			retainHours = cfg.CanceledRetainHours
		default:
			retainHours = cfg.SuccessRetainHours
		}

		// 重新检查是否超过对应状态的保留期
		// （listPendingCleanup 已按最小值过滤，这里做精确判断）
		recheckList, _ := mgr.repo.listPendingCleanup(cleanupCtx, retainHours)
		found := false
		for _, rc := range recheckList {
			if rc.ID == ws.ID {
				found = true
				break
			}
		}
		if !found {
			continue // 尚未超过此状态的保留期
		}

		if cleanErr := mgr.Cleanup(cleanupCtx, ws.TaskID); cleanErr != nil {
			g.Log().Warningf(cleanupCtx, "[Workspace.Cleanup] 清理失败: taskID=%d err=%v", ws.TaskID, cleanErr)
			errs++
		} else {
			cleaned++
		}
	}

	report, sweepErr := RunOrphanSweep(cleanupCtx, mgr, DefaultOrphanSweepConfig())
	if sweepErr != nil {
		errs++
		g.Log().Warningf(cleanupCtx, "[Workspace.Cleanup] orphan sweep 失败: %v", sweepErr)
	} else if report != nil {
		if report.Errors > 0 {
			errs += report.Errors
		}
		g.Log().Infof(cleanupCtx, "[Workspace.Cleanup] orphan sweep: roots=%d db=%d disk=%d disk_orphan=%d db_orphan=%d running_mismatch=%d repaired_missing=%d repaired_running=%d cleaned_disk=%d errors=%d",
			report.ScannedRoots, report.DBWorkspaces, report.DiskWorktrees, report.DiskOrphans, report.MissingOnDisk,
			report.RunningMismatch, report.RepairedMissingOnDisk, report.RepairedRunningMismatch, report.CleanedDiskOrphans, report.Errors)
	}

	g.Log().Infof(cleanupCtx, "[Workspace.Cleanup] 完成: cleaned=%d errors=%d", cleaned, errs)
	return
}
