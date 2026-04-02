package engine

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Watchdog 任务看门狗
// 1. 心跳检测：每隔 checkInterval 检测 running 状态的任务是否有进展
//    连续 maxStaleCount 次无进展 → 认为卡死 → 设为 failed
// 2. 自动重启：检测 failed 状态的任务，自动重启（最多重试 maxRetries 次）
type Watchdog struct {
	checkInterval time.Duration // 心跳检测间隔
	maxStaleCount int           // 最大无进展次数
	maxRetries    int           // 最大自动重试次数

	mu          sync.Mutex
	staleCount  map[int64]int  // taskID → 连续无进展次数
	retryCount  map[int64]int  // taskID → 已重试次数
	lastChunkID map[int64]int64 // taskID → 上次检测时最新 chunk ID
	cancel      context.CancelFunc
}

// NewWatchdog 创建看门狗
func NewWatchdog(checkInterval time.Duration, maxStaleCount int, maxRetries int) *Watchdog {
	return &Watchdog{
		checkInterval: checkInterval,
		maxStaleCount: maxStaleCount,
		maxRetries:    maxRetries,
		staleCount:    make(map[int64]int),
		retryCount:    make(map[int64]int),
		lastChunkID:   make(map[int64]int64),
	}
}

// 全局看门狗（每2分钟检测，连续3次无进展认为卡死，最多自动重试3次）
var defaultWatchdog = NewWatchdog(2*time.Minute, 3, 3)

// GetWatchdog 获取全局看门狗
func GetWatchdog() *Watchdog {
	return defaultWatchdog
}

// Start 启动看门狗
func (w *Watchdog) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	go w.heartbeatLoop(ctx)
	go w.autoRestartLoop(ctx)

	g.Log().Info(ctx, "[Watchdog] 启动，检测间隔:", w.checkInterval, "最大无进展次数:", w.maxStaleCount, "最大重试:", w.maxRetries)
}

// Stop 停止看门狗
func (w *Watchdog) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

// heartbeatLoop 心跳检测循环
func (w *Watchdog) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.checkRunningTasks(ctx)
		}
	}
}

// checkRunningTasks 检测所有 running 状态的任务
func (w *Watchdog) checkRunningTasks(ctx context.Context) {
	// 查询所有 running 状态的任务
	tasks, err := g.DB().Model("mvp_task").
		Where("status", "running").
		Where("deleted_at IS NULL").
		All()
	if err != nil {
		g.Log().Errorf(ctx, "[Watchdog] 查询running任务失败: %v", err)
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, task := range tasks {
		taskID := task["id"].Int64()
		projectID := task["project_id"].Int64()

		// 查找该任务关联的最新对话消息的最新 chunk
		latestChunkID := w.getLatestChunkID(ctx, taskID)

		lastID, exists := w.lastChunkID[taskID]
		if !exists {
			// 首次检测，记录当前状态
			w.lastChunkID[taskID] = latestChunkID
			w.staleCount[taskID] = 0
			continue
		}

		if latestChunkID == lastID {
			// 无进展
			w.staleCount[taskID]++
			g.Log().Warningf(ctx, "[Watchdog] 任务 %d 无进展 (%d/%d)", taskID, w.staleCount[taskID], w.maxStaleCount)

			if w.staleCount[taskID] >= w.maxStaleCount {
				// 认为卡死，设为失败
				g.Log().Errorf(ctx, "[Watchdog] 任务 %d 检测为卡死，标记为失败", taskID)
				g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
					"status":        "failed",
					"error_message": "看门狗检测到任务无响应，自动标记为失败",
					"updated_at":    gtime.Now(),
				})
				logTaskAction(taskID, "watchdog_timeout", "running", "failed", "连续无进展，看门狗判定卡死", "system")

				// 释放调度器资源
				GetScheduler().OnTaskFailed(projectID, taskID, "watchdog timeout")

				// 清理跟踪状态
				delete(w.staleCount, taskID)
				delete(w.lastChunkID, taskID)
			}
		} else {
			// 有进展，重置计数
			w.lastChunkID[taskID] = latestChunkID
			w.staleCount[taskID] = 0
		}
	}

	// 清理已不在 running 状态的任务的跟踪记录
	activeIDs := make(map[int64]bool)
	for _, task := range tasks {
		activeIDs[task["id"].Int64()] = true
	}
	for taskID := range w.lastChunkID {
		if !activeIDs[taskID] {
			delete(w.lastChunkID, taskID)
			delete(w.staleCount, taskID)
		}
	}
}

// getLatestChunkID 获取任务关联的最新 chunk ID
func (w *Watchdog) getLatestChunkID(ctx context.Context, taskID int64) int64 {
	// 直接从 task 的 conversation_id 查 chunk，避免多表 join
	task, err := g.DB().Model("mvp_task").Where("id", taskID).Fields("conversation_id").One()
	if err != nil || task.IsEmpty() || task["conversation_id"].Int64() == 0 {
		return 0
	}
	convID := task["conversation_id"].Int64()

	result, err := g.DB().Model("mvp_message_chunk mc").
		LeftJoin("mvp_message m", "m.id = mc.message_id").
		Where("m.conversation_id", convID).
		Fields("mc.id").
		Order("mc.id DESC").
		Limit(1).
		One()
	if err != nil || result.IsEmpty() {
		return 0
	}
	return result["id"].Int64()
}

// autoRestartLoop 自动重启循环
func (w *Watchdog) autoRestartLoop(ctx context.Context) {
	// 自动重启检测间隔比心跳稍长
	ticker := time.NewTicker(w.checkInterval + 30*time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.checkFailedTasks(ctx)
		}
	}
}

// checkFailedTasks 检测 failed 任务并自动重启
func (w *Watchdog) checkFailedTasks(ctx context.Context) {
	tasks, err := g.DB().Model("mvp_task").
		Where("status", "failed").
		Where("deleted_at IS NULL").
		All()
	if err != nil {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	scheduler := GetScheduler()

	for _, task := range tasks {
		taskID := task["id"].Int64()
		projectID := task["project_id"].Int64()

		// 检查项目是否还在 running
		project, _ := g.DB().Model("mvp_project").Where("id", projectID).Fields("status").One()
		if project.IsEmpty() || project["status"].String() != "running" {
			continue
		}

		retries := w.retryCount[taskID]
		if retries >= w.maxRetries {
			// 超过最大重试次数，汇报给架构师进入 bug 流程
			g.Log().Errorf(ctx, "[Watchdog] 任务 %d 重试 %d 次仍失败，提交给架构师", taskID, retries)
			logTaskAction(taskID, "escalate_to_architect", "failed", "bug_found",
				"看门狗自动重启超过最大次数，升级给架构师处理", "system")

			errMsg := task["error_message"].String()
			go scheduler.ReportBug(context.Background(), projectID, taskID,
				"任务多次自动重启仍然失败，错误信息：\n"+errMsg)
			continue
		}

		// 自动重启
		w.retryCount[taskID] = retries + 1
		g.Log().Infof(ctx, "[Watchdog] 自动重启任务 %d (第 %d/%d 次)", taskID, retries+1, w.maxRetries)

		logTaskAction(taskID, "auto_restart", "failed", "pending",
			"看门狗自动重启 ("+time.Now().Format("15:04:05")+")", "system")

		scheduler.RetryTask(projectID, taskID)
	}
}

// ResetRetryCount 手动重启时重置重试计数
func (w *Watchdog) ResetRetryCount(taskID int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.retryCount, taskID)
}
