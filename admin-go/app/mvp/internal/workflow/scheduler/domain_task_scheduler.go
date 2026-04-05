package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
	domainTask "easymvp/app/mvp/internal/workflow/domain/task"
)

// TaskExecutor 任务执行回调接口。
type TaskExecutor interface {
	ExecuteDomainTask(ctx context.Context, workflowRunID, taskID int64)
}

// CompletionCallback 所有任务完成时的回调。
type CompletionCallback func(ctx context.Context, workflowRunID int64)

// DomainTaskScheduler 领域任务调度器。
// 核心策略与旧 engine.Scheduler 一致：批次门控 + 依赖检查 + 资源冲突 + 并发控制。
type DomainTaskScheduler struct {
	mu             sync.Mutex
	running        map[int64]bool          // 正在执行的任务 ID
	lockedRes      map[string]int64        // 资源 → 占用任务 ID
	maxConcurrency int
	lockMgr        *ResourceLockManager
	executor       TaskExecutor
	onAllDone      CompletionCallback
	cancelFns      map[int64]context.CancelFunc // workflowRunID → cancel
}

// NewDomainTaskScheduler 创建领域任务调度器。
func NewDomainTaskScheduler() *DomainTaskScheduler {
	ctx := context.Background()
	maxConcurrent := engine.GetConfigInt(ctx, "scheduler.max_concurrent", "engine.scheduler.maxConcurrent", 20)
	return &DomainTaskScheduler{
		running:        make(map[int64]bool),
		lockedRes:      make(map[string]int64),
		maxConcurrency: maxConcurrent,
		lockMgr:        NewResourceLockManager(),
		cancelFns:      make(map[int64]context.CancelFunc),
	}
}

// SetExecutor 注册任务执行器。
func (s *DomainTaskScheduler) SetExecutor(e TaskExecutor) { s.executor = e }

// SetCompletionCallback 注册所有任务完成回调。
func (s *DomainTaskScheduler) SetCompletionCallback(fn CompletionCallback) { s.onAllDone = fn }

// Start 启动调度循环。
func (s *DomainTaskScheduler) Start(ctx context.Context, workflowRunID int64) error {
	schedCtx, cancel := context.WithCancel(ctx)

	s.mu.Lock()
	if oldCancel, ok := s.cancelFns[workflowRunID]; ok {
		oldCancel()
	}
	s.cancelFns[workflowRunID] = cancel
	s.mu.Unlock()

	g.Log().Infof(ctx, "[DomainTaskScheduler] Start workflowRunID=%d", workflowRunID)

	go s.scheduleLoop(schedCtx, workflowRunID)
	return nil
}

// Stop 停止调度。
func (s *DomainTaskScheduler) Stop(workflowRunID int64) {
	s.mu.Lock()
	if cancel, ok := s.cancelFns[workflowRunID]; ok {
		cancel()
		delete(s.cancelFns, workflowRunID)
	}
	s.mu.Unlock()
}

// OnTaskCompleted 任务完成回调。
func (s *DomainTaskScheduler) OnTaskCompleted(ctx context.Context, taskID int64) error {
	s.releaseTaskResources(taskID)

	// 查 workflow_run_id
	wfRunID, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).Value("workflow_run_id")

	go func() {
		s.scheduleOnce(context.Background(), wfRunID.Int64())
		s.checkAllDone(context.Background(), wfRunID.Int64())
	}()
	return nil
}

// OnTaskFailed 任务失败回调。
func (s *DomainTaskScheduler) OnTaskFailed(ctx context.Context, taskID int64, errMsg string) {
	s.releaseTaskResources(taskID)
	wfRunID, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).Value("workflow_run_id")
	go s.scheduleOnce(context.Background(), wfRunID.Int64())
}

// scheduleLoop 调度主循环。
func (s *DomainTaskScheduler) scheduleLoop(ctx context.Context, workflowRunID int64) {
	s.scheduleOnce(ctx, workflowRunID)

	pollInterval := engine.GetConfigInt(ctx, "scheduler.poll_interval", "engine.scheduler.pollInterval", 2)
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.scheduleOnce(ctx, workflowRunID)
		}
	}
}

// scheduleOnce 执行一次调度扫描。
func (s *DomainTaskScheduler) scheduleOnce(ctx context.Context, workflowRunID int64) {
	// 并发度快检
	s.mu.Lock()
	if len(s.running) >= s.maxConcurrency {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	// 获取活跃批次
	batchNo := s.getActiveBatch(ctx, workflowRunID)

	// 查候选任务
	query := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("status", domainTask.StatusPending).
		WhereNull("deleted_at")

	if batchNo > 0 {
		query = query.Where("(batch_no = 0 OR batch_no = ?)", batchNo)
	} else {
		query = query.Where("batch_no = 0")
	}

	tasks, err := query.OrderAsc("batch_no").OrderAsc("sort").All()
	if err != nil || len(tasks) == 0 {
		return
	}

	// 检查依赖
	s.mu.Lock()
	defer s.mu.Unlock()

	currentRunning := len(s.running)

	for _, task := range tasks {
		if currentRunning >= s.maxConcurrency {
			break
		}

		taskID := task["id"].Int64()
		if s.running[taskID] {
			continue
		}

		// 依赖检查：所有依赖任务必须已完成
		if !s.allDepsCompleted(ctx, task) {
			continue
		}

		// 资源冲突检测
		var resources []string
		resJSON := task["affected_resources"].String()
		if resJSON != "" && resJSON != "[]" && resJSON != "null" {
			json.Unmarshal([]byte(resJSON), &resources)
		}

		hasConflict := false
		for _, res := range resources {
			if _, conflict := s.lockedRes[res]; conflict {
				hasConflict = true
				break
			}
		}
		if hasConflict {
			continue
		}

		// CAS 更新状态 pending → running（必须先确认数据库更新成功再写内存锁）
		now := gtime.Now()
		casResult, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("id", taskID).
			Where("status", domainTask.StatusPending).
			Update(g.Map{
				"status":     domainTask.StatusRunning,
				"started_at": now,
				"updated_at": now,
			})
		casRows, _ := casResult.RowsAffected()
		if casRows == 0 {
			// CAS 失败：任务已不在 pending 状态（被并发 Pause/Resume/手动重试改走），跳过
			continue
		}

		// CAS 成功，锁定资源
		for _, res := range resources {
			s.lockedRes[res] = taskID
		}
		s.running[taskID] = true
		currentRunning++

		// 持久化资源锁
		if len(resources) > 0 {
			lockedJSON, _ := json.Marshal(resources)
			_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
				Where("id", taskID).
				Update(g.Map{"locked_resources": string(lockedJSON)})
			_ = s.lockMgr.AcquireLocks(ctx, workflowRunID, taskID, resources)
		}

		// 执行
		if s.executor != nil {
			go s.executor.ExecuteDomainTask(context.Background(), workflowRunID, taskID)
		}
	}
}

// getActiveBatch 获取活跃批次号。
func (s *DomainTaskScheduler) getActiveBatch(ctx context.Context, workflowRunID int64) int {
	// 找最小的有 pending 任务的 batch_no（排除 batch_no=0）
	val, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("status", domainTask.StatusPending).
		Where("batch_no > 0").
		WhereNull("deleted_at").
		Min("batch_no")
	if err != nil || val == 0 {
		return 0
	}

	batchNo := int(val)

	// 检查该批次之前是否有未完成的任务（非 completed/skipped）
	prevUnfinished, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("batch_no > 0").
		Where("batch_no < ?", batchNo).
		WhereNotIn("status", g.Slice{domainTask.StatusCompleted, "skipped"}).
		WhereNull("deleted_at").
		Count()
	if prevUnfinished > 0 {
		// 之前批次还有未完成的，不推进
		prevBatch, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("workflow_run_id", workflowRunID).
			Where("batch_no > 0").
			Where("batch_no < ?", batchNo).
			WhereNotIn("status", g.Slice{domainTask.StatusCompleted, "skipped"}).
			WhereNull("deleted_at").
			Min("batch_no")
		return int(prevBatch)
	}

	return batchNo
}

// releaseTaskResources 释放任务资源锁。
func (s *DomainTaskScheduler) releaseTaskResources(taskID int64) {
	s.mu.Lock()
	delete(s.running, taskID)
	for res, tid := range s.lockedRes {
		if tid == taskID {
			delete(s.lockedRes, res)
		}
	}
	s.mu.Unlock()

	// DB 清理
	_, _ = g.DB().Model("mvp_domain_task").Where("id", taskID).
		Update(g.Map{"locked_resources": nil, "heartbeat_at": nil})
	_ = s.lockMgr.ReleaseLocks(context.Background(), taskID)
}

// checkAllDone 检查 workflow 的所有任务是否完成。
func (s *DomainTaskScheduler) checkAllDone(ctx context.Context, workflowRunID int64) {
	unfinished, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNotIn("status", g.Slice{domainTask.StatusCompleted, "skipped", domainTask.StatusEscalated}).
		WhereNull("deleted_at").
		Count()

	if unfinished == 0 {
		g.Log().Infof(ctx, "[DomainTaskScheduler] 所有任务完成 workflowRunID=%d", workflowRunID)
		s.Stop(workflowRunID)
		if s.onAllDone != nil {
			s.onAllDone(ctx, workflowRunID)
		}
	}
}

// GetRunningCount 获取正在运行的任务数。
func (s *DomainTaskScheduler) GetRunningCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.running)
}

// GetLockedResources 获取锁定的资源列表（调试用）。
func (s *DomainTaskScheduler) GetLockedResources() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]int64, len(s.lockedRes))
	for k, v := range s.lockedRes {
		result[k] = v
	}
	return result
}

// Wakeup 唤醒一次调度扫描（不重建调度循环，仅触发单次 scheduleOnce）。
// 用于单任务重试后让调度器拾取，比 Start() 更轻量。
func (s *DomainTaskScheduler) Wakeup(ctx context.Context, workflowRunID int64) {
	go s.scheduleOnce(ctx, workflowRunID)
}

// HasUnfinished 检查是否还有未完成任务（供外部查询）。
func (s *DomainTaskScheduler) HasUnfinished(ctx context.Context, workflowRunID int64) bool {
	count, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNotIn("status", g.Slice{domainTask.StatusCompleted, "skipped", domainTask.StatusEscalated}).
		WhereNull("deleted_at").
		Count()
	return count > 0
}

// Pause 暂停调度并回退 running 任务。
func (s *DomainTaskScheduler) Pause(ctx context.Context, workflowRunID int64) {
	s.Stop(workflowRunID)

	// 查 running 任务
	runningTasks, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("status", domainTask.StatusRunning).
		WhereNull("deleted_at").
		Fields("id").All()

	s.mu.Lock()
	for _, t := range runningTasks {
		tid := t["id"].Int64()
		delete(s.running, tid)
		for res, occupier := range s.lockedRes {
			if occupier == tid {
				delete(s.lockedRes, res)
			}
		}
	}
	s.mu.Unlock()

	// 释放 DB 资源锁 + 回退状态
	for _, t := range runningTasks {
		tid := t["id"].Int64()
		_ = s.lockMgr.ReleaseLocks(ctx, tid)
	}

	// running → pending
	_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("status", domainTask.StatusRunning).
		Update(g.Map{
			"status":           domainTask.StatusPending,
			"locked_resources": nil,
			"heartbeat_at":     nil,
			"updated_at":       gtime.Now(),
		})
}

// allDepsCompleted 检查任务的所有依赖是否已完成。
// 优先使用 depends_on_task_ids（完整依赖列表），回退到 parent_task_id（单依赖兼容）。
func (s *DomainTaskScheduler) allDepsCompleted(ctx context.Context, task gdb.Record) bool {
	// 优先检查多依赖字段
	depsJSON := task["depends_on_task_ids"].String()
	if depsJSON != "" && depsJSON != "null" && depsJSON != "[]" {
		var depIDs []int64
		if err := json.Unmarshal([]byte(depsJSON), &depIDs); err == nil && len(depIDs) > 0 {
			for _, depID := range depIDs {
				status, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
					Where("id", depID).Value("status")
				st := status.String()
				if st != domainTask.StatusCompleted && st != "skipped" {
					return false
				}
			}
			return true
		}
	}

	// 回退：单依赖 parent_task_id
	parentID := task["parent_task_id"].Int64()
	if parentID > 0 {
		status, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("id", parentID).Value("status")
		st := status.String()
		return st == domainTask.StatusCompleted || st == "skipped"
	}

	return true // 无依赖
}

// ErrorMsg 用于错误追踪的调度失败描述。
func ErrorMsg(taskID int64, msg string) string {
	return fmt.Sprintf("task=%d: %s", taskID, msg)
}
