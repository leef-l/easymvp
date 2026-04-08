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

var (
	defaultWatchdog     *Watchdog
	defaultWatchdogOnce sync.Once
)

// GetWatchdog 获取全局看门狗（首次调用时从配置加载参数）
func GetWatchdog() *Watchdog {
	defaultWatchdogOnce.Do(func() {
		ctx := context.Background()
		checkInterval := GetConfigInt(ctx, "watchdog.check_interval", "engine.watchdog.checkInterval", 120)
		maxStaleCount := GetConfigInt(ctx, "watchdog.max_stale_count", "engine.watchdog.maxStaleCount", 3)
		maxRetries := GetConfigInt(ctx, "watchdog.max_retries", "engine.watchdog.maxRetries", 3)
		defaultWatchdog = NewWatchdog(
			time.Duration(checkInterval)*time.Second,
			maxStaleCount,
			maxRetries,
		)
		g.Log().Infof(ctx, "[Watchdog] 配置加载: checkInterval=%ds, maxStaleCount=%d, maxRetries=%d",
			checkInterval, maxStaleCount, maxRetries)
	})
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
// 双重检测机制：优先检查心跳时间戳，其次检查 chunk 增长
func (w *Watchdog) checkRunningTasks(ctx context.Context) {
	// 查询所有 running 状态的任务（锁外执行 DB 操作）
	tasks, err := g.DB().Model("mvp_task").
		Where("status", "running").
		WhereNull("deleted_at").
		Fields("id, project_id, heartbeat_at, conversation_id").
		All()
	if err != nil {
		g.Log().Errorf(ctx, "[Watchdog] 查询running任务失败: %v", err)
		return
	}

	// 默认心跳超时阈值（后续会按项目分类动态调整）
	defaultHeartbeatTimeout := w.checkInterval * time.Duration(w.maxStaleCount)

	// 锁外收集项目分类信息，用于按分类族调整心跳超时
	projectCategories := make(map[int64]string)
	projectIDs := make(map[int64]bool)
	for _, task := range tasks {
		pid := task["project_id"].Int64()
		projectIDs[pid] = true
	}
	if len(projectIDs) > 0 {
		pidList := make([]int64, 0, len(projectIDs))
		for pid := range projectIDs {
			pidList = append(pidList, pid)
		}
		projects, pErr := g.DB().Model("mvp_project").Ctx(ctx).WhereIn("id", pidList).Fields("id, project_category").All()
		if pErr == nil {
			for _, p := range projects {
				projectCategories[p["id"].Int64()] = p["project_category"].String()
			}
		}
	}

	// 锁外收集信息
	type taskInfo struct {
		taskID       int64
		projectID    int64
		latestChunk  int64
		heartbeatAt  *gtime.Time
		hasHeartbeat bool // 任务是否支持心跳（有 heartbeat_at 字段值）
	}
	infos := make([]taskInfo, 0, len(tasks))
	for _, task := range tasks {
		tid := task["id"].Int64()
		pid := task["project_id"].Int64()
		hb := task["heartbeat_at"].GTime()
		hasHB := hb != nil && !hb.IsZero()

		chunkID := int64(0)
		if !hasHB {
			// 没有心跳的任务，退化到 chunk 检测
			chunkID = w.getLatestChunkID(ctx, tid)
		}
		infos = append(infos, taskInfo{tid, pid, chunkID, hb, hasHB})
	}

	// 决策阶段：持锁
	type staleTask struct {
		taskID    int64
		projectID int64
		reason    string
	}
	var staleTasks []staleTask

	w.mu.Lock()
	for _, info := range infos {
		taskID := info.taskID

		if info.hasHeartbeat {
			// 心跳模式：根据项目分类族动态调整超时时间
			heartbeatTimeout := defaultHeartbeatTimeout
			if cat, ok := projectCategories[info.projectID]; ok && cat != "" {
				timeoutSec := GetHeartbeatTimeout(ctx, cat)
				heartbeatTimeout = time.Duration(timeoutSec) * time.Second
			}
			elapsed := time.Since(info.heartbeatAt.Time)
			if elapsed > heartbeatTimeout {
				g.Log().Errorf(ctx, "[Watchdog] 任务 %d 心跳超时 (%.0fs > %.0fs)，标记为失败",
					taskID, elapsed.Seconds(), heartbeatTimeout.Seconds())
				staleTasks = append(staleTasks, staleTask{taskID, info.projectID,
					"心跳超时：最后心跳 " + info.heartbeatAt.String()})
				delete(w.staleCount, taskID)
				delete(w.lastChunkID, taskID)
			} else {
				// 心跳正常，重置
				w.staleCount[taskID] = 0
			}
			continue
		}

		// Chunk 模式：退化检测（兼容无心跳的旧任务）
		latestChunkID := info.latestChunk
		lastID, exists := w.lastChunkID[taskID]
		if !exists {
			w.lastChunkID[taskID] = latestChunkID
			w.staleCount[taskID] = 0
			continue
		}

		if latestChunkID == lastID {
			w.staleCount[taskID]++
			g.Log().Warningf(ctx, "[Watchdog] 任务 %d 无进展 (%d/%d)", taskID, w.staleCount[taskID], w.maxStaleCount)
			if w.staleCount[taskID] >= w.maxStaleCount {
				g.Log().Errorf(ctx, "[Watchdog] 任务 %d 检测为卡死，标记为失败", taskID)
				staleTasks = append(staleTasks, staleTask{taskID, info.projectID, "连续无进展，看门狗判定卡死"})
				delete(w.staleCount, taskID)
				delete(w.lastChunkID, taskID)
			}
		} else {
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

	// 锁外执行 DB 更新操作（通过 updateTaskStatus 使用 CAS，防止与正常完成竞争）
	for _, st := range staleTasks {
		rows, err := updateTaskStatus(ctx, st.taskID, "running", "failed", g.Map{
			"error_message": "看门狗检测: " + st.reason,
		})
		if err != nil {
			g.Log().Errorf(ctx, "[Watchdog] 更新任务 %d 状态失败: %v", st.taskID, err)
			continue
		}
		if rows == 0 {
			g.Log().Infof(ctx, "[Watchdog] 任务 %d 已不是 running 状态，跳过标记失败", st.taskID)
			continue
		}

		logTaskAction(st.taskID, "watchdog_timeout", "running", "failed", st.reason, "system")

		// 释放调度器资源
		GetScheduler().OnTaskFailed(st.projectID, st.taskID, "watchdog timeout: "+st.reason)
	}
}

// TouchHeartbeat 更新任务心跳时间戳（执行器定期调用）
func TouchHeartbeat(ctx context.Context, taskID int64) {
	if _, err := g.DB().Model("mvp_task").Ctx(ctx).Where("id", taskID).Update(g.Map{
		"heartbeat_at": gtime.Now(),
	}); err != nil {
		g.Log().Warningf(ctx, "[Watchdog] TouchHeartbeat 失败: task=%d err=%v", taskID, err)
	}
}

// getLatestChunkID 获取任务关联的最新 chunk ID
func (w *Watchdog) getLatestChunkID(ctx context.Context, taskID int64) int64 {
	// 直接从 task 的 conversation_id 查 chunk，避免多表 join
	task, err := g.DB().Model("mvp_task").Ctx(ctx).Where("id", taskID).Fields("conversation_id").One()
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
		WhereNull("deleted_at").
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

	// 批量收集项目 ID 并一次查询状态（避免 N+1）
	projectIDs := make(map[int64]struct{})
	for _, task := range tasks {
		projectIDs[task["project_id"].Int64()] = struct{}{}
	}
	pidList := make([]int64, 0, len(projectIDs))
	for pid := range projectIDs {
		pidList = append(pidList, pid)
	}
	runningProjects := make(map[int64]bool)
	if len(pidList) > 0 {
		projects, projErr := g.DB().Model("mvp_project").Ctx(ctx).
			WhereIn("id", pidList).
			Where("status", "running").
			Fields("id").All()
		if projErr != nil {
			g.Log().Warningf(ctx, "[Watchdog] 查询运行中项目失败: err=%v", projErr)
		}
		for _, p := range projects {
			runningProjects[p["id"].Int64()] = true
		}
	}

	var candidates []failedTaskInfo
	for _, task := range tasks {
		projectID := task["project_id"].Int64()
		if !runningProjects[projectID] {
			continue
		}
		candidates = append(candidates, failedTaskInfo{
			taskID:    task["id"].Int64(),
			projectID: projectID,
			taskType:  task["role_type"].String(),
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
		taskType  string
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
				taskType:  info.taskType,
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
		g.Log().Errorf(ctx, "[Watchdog] 任务 %d（%s）重试超过最大次数，升级处理", item.taskID, item.taskType)

		// 更新任务状态为 escalated，防止下轮重复扫描
		updateTaskStatus(ctx, item.taskID, "failed", "escalated", nil)

		logTaskAction(item.taskID, "escalate_to_architect", "failed", "escalated",
			"看门狗自动重启超过最大次数，升级给架构师处理", "system")

		if item.taskType == "auditor" {
			// auditor 类型：通过 ReportBug 走 bug 闭环（关联实施员）
			go scheduler.ReportBug(scheduler.getProjectContext(item.projectID), item.projectID, item.taskID,
				"审计任务多次重试仍然失败，错误信息：\n"+item.errMsg)
		} else {
			// 非 auditor 类型：直接创建架构师分析任务
			go scheduler.EscalateFailedTask(scheduler.getProjectContext(item.projectID), item.projectID, item.taskID, item.taskType, item.errMsg)
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
