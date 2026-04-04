// Package watchdog 提供 domain_task（V2 链路）的心跳检测与自动重试能力。
package watchdog

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
	domainTask "easymvp/app/mvp/internal/workflow/domain/task"
)

// SchedulerCallback 调度器回调（避免循环依赖）。
type SchedulerCallback interface {
	OnTaskFailed(ctx context.Context, taskID int64, errMsg string)
}

// RetryCallback 重试回调。
type RetryCallback func(ctx context.Context, taskID int64) error

// DomainTaskWatchdog V2 链路任务看门狗。
type DomainTaskWatchdog struct {
	checkInterval time.Duration
	maxStaleCount int
	maxRetries    int

	mu         sync.Mutex
	staleCount map[int64]int // taskID → 连续无心跳次数
	retryCount map[int64]int // taskID → 已重试次数
	cancel     context.CancelFunc

	scheduler SchedulerCallback
	retryFn   RetryCallback
}

// New 创建 V2 看门狗。
func New() *DomainTaskWatchdog {
	ctx := context.Background()
	checkInterval := engine.GetConfigInt(ctx, "watchdog.check_interval", "engine.watchdog.checkInterval", 120)
	maxStaleCount := engine.GetConfigInt(ctx, "watchdog.max_stale_count", "engine.watchdog.maxStaleCount", 3)
	maxRetries := engine.GetConfigInt(ctx, "watchdog.max_retries", "engine.watchdog.maxRetries", 3)

	return &DomainTaskWatchdog{
		checkInterval: time.Duration(checkInterval) * time.Second,
		maxStaleCount: maxStaleCount,
		maxRetries:    maxRetries,
		staleCount:    make(map[int64]int),
		retryCount:    make(map[int64]int),
	}
}

// SetScheduler 注册调度器回调。
func (w *DomainTaskWatchdog) SetScheduler(s SchedulerCallback) { w.scheduler = s }

// SetRetryFn 注册重试回调。
func (w *DomainTaskWatchdog) SetRetryFn(fn RetryCallback) { w.retryFn = fn }

// Start 启动看门狗。
func (w *DomainTaskWatchdog) Start(ctx context.Context) {
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
	go w.autoRetryLoop(childCtx)

	g.Log().Infof(ctx, "[WatchdogV2] 启动: interval=%v maxStale=%d maxRetries=%d",
		w.checkInterval, w.maxStaleCount, w.maxRetries)
}

// Stop 停止看门狗。
func (w *DomainTaskWatchdog) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
}

// ResetRetryCount 手动重试时重置计数。
func (w *DomainTaskWatchdog) ResetRetryCount(taskID int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.retryCount, taskID)
}

// heartbeatLoop 心跳检测循环。
func (w *DomainTaskWatchdog) heartbeatLoop(ctx context.Context) {
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

// checkRunningTasks 检测 running 状态的 domain_task。
func (w *DomainTaskWatchdog) checkRunningTasks(ctx context.Context) {
	tasks, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("status", domainTask.StatusRunning).
		WhereNull("deleted_at").
		Fields("id, workflow_run_id, heartbeat_at, started_at").
		All()
	if err != nil {
		g.Log().Errorf(ctx, "[WatchdogV2] 查询 running 任务失败: %v", err)
		return
	}

	heartbeatTimeout := w.checkInterval * time.Duration(w.maxStaleCount)

	type staleTask struct {
		taskID int64
		reason string
	}
	var staleTasks []staleTask

	w.mu.Lock()
	for _, task := range tasks {
		taskID := task["id"].Int64()
		hb := task["heartbeat_at"].GTime()
		startedAt := task["started_at"].GTime()

		// 确定检测基准时间
		var refTime *gtime.Time
		if hb != nil && !hb.IsZero() {
			refTime = hb
		} else if startedAt != nil && !startedAt.IsZero() {
			refTime = startedAt
		}

		if refTime == nil {
			// 无参考时间，跳过本轮
			continue
		}

		elapsed := time.Since(refTime.Time)
		if elapsed > heartbeatTimeout {
			w.staleCount[taskID]++
			if w.staleCount[taskID] >= 1 {
				// 已超过心跳超时阈值
				staleTasks = append(staleTasks, staleTask{
					taskID: taskID,
					reason: "心跳超时：最后活跃 " + refTime.String(),
				})
				delete(w.staleCount, taskID)
			}
		} else {
			w.staleCount[taskID] = 0
		}
	}

	// 清理不再 running 的跟踪记录
	activeIDs := make(map[int64]bool)
	for _, task := range tasks {
		activeIDs[task["id"].Int64()] = true
	}
	for id := range w.staleCount {
		if !activeIDs[id] {
			delete(w.staleCount, id)
		}
	}
	w.mu.Unlock()

	// 锁外执行状态更新
	for _, st := range staleTasks {
		now := gtime.Now()
		result, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("id", st.taskID).
			Where("status", domainTask.StatusRunning).
			Update(g.Map{
				"status":     domainTask.StatusFailed,
				"result":     "看门狗检测: " + st.reason,
				"updated_at": now,
			})
		if err != nil {
			g.Log().Errorf(ctx, "[WatchdogV2] 标记任务 %d 失败出错: %v", st.taskID, err)
			continue
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			continue // 任务已不是 running
		}

		g.Log().Warningf(ctx, "[WatchdogV2] 任务 %d 判定卡死: %s", st.taskID, st.reason)

		if w.scheduler != nil {
			w.scheduler.OnTaskFailed(ctx, st.taskID, "watchdog: "+st.reason)
		}
	}
}

// autoRetryLoop 自动重试循环。
func (w *DomainTaskWatchdog) autoRetryLoop(ctx context.Context) {
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

// checkFailedTasks 检测 failed 任务并自动重试。
func (w *DomainTaskWatchdog) checkFailedTasks(ctx context.Context) {
	tasks, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("status", domainTask.StatusFailed).
		WhereNull("deleted_at").
		Fields("id, workflow_run_id, retry_count").
		All()
	if err != nil {
		return
	}

	// 过滤：只处理所属 workflow_run 仍在活跃状态的任务
	type candidate struct {
		taskID     int64
		retryCount int
	}
	var candidates []candidate
	for _, task := range tasks {
		wfRunID := task["workflow_run_id"].Int64()
		wfStatus, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
			Where("id", wfRunID).Value("status")
		status := wfStatus.String()
		// 只在 executing/reworking 状态下自动重试
		if status != "executing" && status != "reworking" {
			continue
		}
		candidates = append(candidates, candidate{
			taskID:     task["id"].Int64(),
			retryCount: task["retry_count"].Int(),
		})
	}

	// 决策
	type retryItem struct {
		taskID  int64
		retryNo int
	}
	type escalateItem struct {
		taskID int64
	}
	var retryList []retryItem
	var escalateList []escalateItem

	w.mu.Lock()
	for _, c := range candidates {
		totalRetries := w.retryCount[c.taskID] + c.retryCount
		if totalRetries >= w.maxRetries {
			escalateList = append(escalateList, escalateItem{taskID: c.taskID})
		} else {
			w.retryCount[c.taskID]++
			retryList = append(retryList, retryItem{c.taskID, totalRetries + 1})
		}
	}
	w.mu.Unlock()

	// 执行重试
	for _, item := range retryList {
		g.Log().Infof(ctx, "[WatchdogV2] 自动重试任务 %d (第 %d/%d 次)", item.taskID, item.retryNo, w.maxRetries)

		now := gtime.Now()
		result, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("id", item.taskID).
			Where("status", domainTask.StatusFailed).
			Update(g.Map{
				"status":      domainTask.StatusPending,
				"retry_count": item.retryNo,
				"result":      nil,
				"updated_at":  now,
			})
		rows, _ := result.RowsAffected()
		if rows == 0 {
			continue
		}

		if w.retryFn != nil {
			_ = w.retryFn(ctx, item.taskID)
		}
	}

	// 执行升级
	for _, item := range escalateList {
		g.Log().Errorf(ctx, "[WatchdogV2] 任务 %d 重试超限，升级为 escalated", item.taskID)

		now := gtime.Now()
		_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("id", item.taskID).
			Where("status", domainTask.StatusFailed).
			Update(g.Map{
				"status":     domainTask.StatusEscalated,
				"updated_at": now,
			})
	}
}
