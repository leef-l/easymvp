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
func (w *Watchdog) Start(ctx context.Context) {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
	}
	w.mu.Unlock()

	childCtx, cancel := context.WithCancel(ctx)

	w.mu.Lock()
	w.cancel = cancel
	w.mu.Unlock()

	go w.heartbeatLoop(childCtx)
	go w.autoRestartLoop(childCtx)

	g.Log().Info(childCtx, "[Watchdog] 启动，检测间隔:", w.checkInterval, "最大无进展次数:", w.maxStaleCount, "最大重试:", w.maxRetries)
}

// Stop 停止看门狗
func (w *Watchdog) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
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
	// 查询所有 running 状态的任务（锁外执行 DB 操作）
	tasks, err := g.DB().Model("mvp_task").
		Where("status", "running").
		Where("deleted_at IS NULL").
		All()
	if err != nil {
		g.Log().Errorf(ctx, "[Watchdog] 查询running任务失败: %v", err)
		return
	}

	// 锁外获取所有任务的最新 chunkID（DB 操作不持锁）
	type taskInfo struct {
		taskID      int64
		projectID   int64
		latestChunk int64
	}
	infos := make([]taskInfo, 0, len(tasks))
	for _, task := range tasks {
		tid := task["id"].Int64()
		pid := task["project_id"].Int64()
		chunkID := w.getLatestChunkID(ctx, tid)
		infos = append(infos, taskInfo{tid, pid, chunkID})
	}

	// 决策阶段：持锁，纯内存操作，确定哪些任务需要被标记为失败
	type staleTask struct {
		taskID    int64
		projectID int64
	}
	var staleTasks []staleTask

	w.mu.Lock()
	for _, info := range infos {
		taskID := info.taskID
		latestChunkID := info.latestChunk

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
				// 认为卡死，加入待处理列表
				g.Log().Errorf(ctx, "[Watchdog] 任务 %d 检测为卡死，标记为失败", taskID)
				staleTasks = append(staleTasks, staleTask{taskID, info.projectID})
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
	for _, info := range infos {
		activeIDs[info.taskID] = true
	}
	for taskID := range w.lastChunkID {
		if !activeIDs[taskID] {
			delete(w.lastChunkID, taskID)
			delete(w.staleCount, taskID)
		}
	}
	w.mu.Unlock()

	// 锁外执行 DB 更新操作
	for _, st := range staleTasks {
		g.DB().Model("mvp_task").Where("id", st.taskID).Update(g.Map{
			"status":        "failed",
			"error_message": "看门狗检测到任务无响应，自动标记为失败",
			"updated_at":    gtime.Now(),
		})
		logTaskAction(st.taskID, "watchdog_timeout", "running", "failed", "连续无进展，看门狗判定卡死", "system")

		// 释放调度器资源
		GetScheduler().OnTaskFailed(st.projectID, st.taskID, "watchdog timeout")
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
	// 锁外执行所有 DB 查询
	tasks, err := g.DB().Model("mvp_task").
		Where("status", "failed").
		Where("deleted_at IS NULL").
		All()
	if err != nil {
		return
	}

	// 过滤出项目仍在 running 状态的任务（锁外执行 DB 操作）
	type failedTaskInfo struct {
		taskID    int64
		projectID int64
		taskType  string
		errMsg    string
	}
	var candidates []failedTaskInfo
	for _, task := range tasks {
		projectID := task["project_id"].Int64()
		project, _ := g.DB().Model("mvp_project").Where("id", projectID).Fields("status").One()
		if project.IsEmpty() || project["status"].String() != "running" {
			continue
		}
		candidates = append(candidates, failedTaskInfo{
			taskID:    task["id"].Int64(),
			projectID: projectID,
			taskType:  task["task_type"].String(),
			errMsg:    task["error_message"].String(),
		})
	}

	// 持锁阶段：纯内存决策，确定哪些任务需要重试或上报
	type retryItem struct {
		taskID    int64
		projectID int64
		retryNo   int
	}
	type escalateItem struct {
		taskID    int64
		projectID int64
		errMsg    string
		isAuditor bool
	}
	var retryList []retryItem
	var escalateList []escalateItem

	w.mu.Lock()
	for _, info := range candidates {
		retries := w.retryCount[info.taskID]
		if retries >= w.maxRetries {
			escalateList = append(escalateList, escalateItem{
				taskID:    info.taskID,
				projectID: info.projectID,
				errMsg:    info.errMsg,
				isAuditor: info.taskType == "auditor",
			})
		} else {
			w.retryCount[info.taskID] = retries + 1
			retryList = append(retryList, retryItem{info.taskID, info.projectID, retries + 1})
		}
	}
	w.mu.Unlock()

	scheduler := GetScheduler()

	// 锁外执行实际操作
	for _, item := range escalateList {
		g.Log().Errorf(ctx, "[Watchdog] 任务 %d 重试超过最大次数，提交给架构师", item.taskID)
		logTaskAction(item.taskID, "escalate_to_architect", "failed", "bug_found",
			"看门狗自动重启超过最大次数，升级给架构师处理", "system")

		// 只有 auditor 类型任务才调用 ReportBug
		if item.isAuditor {
			go scheduler.ReportBug(context.Background(), item.projectID, item.taskID,
				"任务多次自动重启仍然失败，错误信息：\n"+item.errMsg)
		} else {
			g.Log().Warningf(ctx, "[Watchdog] 任务 %d 非 auditor 类型（跳过 ReportBug），项目 %d 需人工介入",
				item.taskID, item.projectID)
		}
	}

	for _, item := range retryList {
		g.Log().Infof(ctx, "[Watchdog] 自动重启任务 %d (第 %d/%d 次)", item.taskID, item.retryNo, w.maxRetries)
		logTaskAction(item.taskID, "auto_restart", "failed", "pending",
			"看门狗自动重启 ("+time.Now().Format("15:04:05")+")", "system")
		scheduler.RetryTask(item.projectID, item.taskID)
	}
}

// ResetRetryCount 手动重启时重置重试计数
func (w *Watchdog) ResetRetryCount(taskID int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.retryCount, taskID)
}
