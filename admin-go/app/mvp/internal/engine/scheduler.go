package engine

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Scheduler 任务调度器
// 核心职责：扫描待执行任务、批次调度、依赖检查、资源冲突检测、动态推进
type Scheduler struct {
	mu             sync.Mutex
	running        map[int64]bool     // 正在执行的任务 ID
	lockedRes      map[string]int64   // 已锁定的资源 -> 占用任务 ID
	maxConcurrency int                // 最大并发 goroutine 数
	executor       *Executor          // 任务执行器
	projectCtx     map[int64]context.CancelFunc // 项目级 cancel 函数（暂停用）
}

// NewScheduler 创建调度器
func NewScheduler(maxConcurrency int) *Scheduler {
	s := &Scheduler{
		running:        make(map[int64]bool),
		lockedRes:      make(map[string]int64),
		maxConcurrency: maxConcurrency,
		projectCtx:     make(map[int64]context.CancelFunc),
	}
	s.executor = NewExecutor(s)
	return s
}

// 全局调度器
var defaultScheduler = NewScheduler(20)

// GetScheduler 获取全局调度器
func GetScheduler() *Scheduler {
	return defaultScheduler
}

// StartProject 启动项目任务调度
func (s *Scheduler) StartProject(projectID int64) {
	ctx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	// 如果有旧的 cancel，先取消
	if oldCancel, ok := s.projectCtx[projectID]; ok {
		oldCancel()
	}
	s.projectCtx[projectID] = cancel
	s.mu.Unlock()

	go s.scheduleLoop(ctx, projectID)
}

// PauseProject 暂停项目调度
func (s *Scheduler) PauseProject(projectID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, ok := s.projectCtx[projectID]; ok {
		cancel()
		delete(s.projectCtx, projectID)
	}

	// 更新所有 running 状态的任务为 pending
	g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "running").
		Update(g.Map{
			"status":     "pending",
			"updated_at": gtime.Now(),
		})
}

// OnTaskCompleted 任务完成回调，触发动态推进
func (s *Scheduler) OnTaskCompleted(projectID int64, taskID int64) {
	s.mu.Lock()
	delete(s.running, taskID)
	// 释放该任务锁定的资源
	for res, tid := range s.lockedRes {
		if tid == taskID {
			delete(s.lockedRes, res)
		}
	}
	s.mu.Unlock()

	// 记录日志
	logTaskAction(taskID, "completed", "running", "completed", "任务执行完成", "system")

	// 检查该任务所在批次是否全部完成，是则触发批次压缩
	go s.checkBatchDone(projectID, taskID)

	// 触发立即调度（不等待下一个 tick）
	go s.scheduleOnce(context.Background(), projectID)

	// 检查项目是否全部完成
	go s.checkProjectDone(projectID)
}

// OnTaskFailed 任务失败回调
func (s *Scheduler) OnTaskFailed(projectID int64, taskID int64, errMsg string) {
	s.mu.Lock()
	delete(s.running, taskID)
	for res, tid := range s.lockedRes {
		if tid == taskID {
			delete(s.lockedRes, res)
		}
	}
	s.mu.Unlock()

	logTaskAction(taskID, "failed", "running", "failed", errMsg, "system")
}

// scheduleLoop 调度主循环
func (s *Scheduler) scheduleLoop(ctx context.Context, projectID int64) {
	// 首次立即调度
	s.scheduleOnce(ctx, projectID)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.scheduleOnce(ctx, projectID)
		}
	}
}

// scheduleOnce 执行一次调度扫描
func (s *Scheduler) scheduleOnce(ctx context.Context, projectID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查并发度
	currentRunning := len(s.running)
	if currentRunning >= s.maxConcurrency {
		return
	}

	// 查询所有待执行的任务，按 batch_no 和 sort 排序
	tasks, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "pending").
		Where("deleted_at IS NULL").
		Order("batch_no ASC, sort ASC").
		All()
	if err != nil {
		g.Log().Errorf(ctx, "调度扫描失败: %v", err)
		return
	}

	for _, task := range tasks {
		if currentRunning >= s.maxConcurrency {
			break
		}

		taskID := task["id"].Int64()

		// 跳过已在运行的
		if s.running[taskID] {
			continue
		}

		// 依赖检查
		if !s.checkDependencies(ctx, taskID) {
			continue
		}

		// 资源冲突检测
		resources := parseResources(task["affected_resources"].String())
		if !s.tryLockResources(taskID, resources) {
			continue
		}

		// 可以启动
		s.running[taskID] = true
		currentRunning++

		// 更新状态
		g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
			"status":     "running",
			"started_at": gtime.Now(),
			"updated_at": gtime.Now(),
		})

		logTaskAction(taskID, "started", "pending", "running", "任务开始执行", "system")

		// 启动执行器
		go s.executor.Execute(ctx, projectID, taskID)
	}
}

// checkDependencies 检查任务的所有依赖是否已完成
func (s *Scheduler) checkDependencies(ctx context.Context, taskID int64) bool {
	// 从 mvp_task_dependency 表查依赖
	deps, err := g.DB().Model("mvp_task_dependency").
		Where("task_id", taskID).
		All()
	if err != nil {
		g.Log().Errorf(ctx, "查询依赖失败: %v", err)
		return false
	}

	if len(deps) == 0 {
		return true
	}

	for _, dep := range deps {
		depID := dep["depends_on_id"].Int64()
		// 检查依赖任务的状态
		depTask, err := g.DB().Model("mvp_task").
			Where("id", depID).
			Fields("status").
			One()
		if err != nil || depTask.IsEmpty() {
			return false
		}
		if depTask["status"].String() != "completed" {
			return false
		}
	}

	return true
}

// tryLockResources 尝试锁定资源，如果有冲突返回 false
func (s *Scheduler) tryLockResources(taskID int64, resources []string) bool {
	// 先检查是否有冲突
	for _, res := range resources {
		if occupier, exists := s.lockedRes[res]; exists && occupier != taskID {
			return false
		}
	}
	// 全部可用，锁定
	for _, res := range resources {
		s.lockedRes[res] = taskID
	}
	return true
}

// checkBatchDone 检查任务所在批次是否全部完成，完成则触发批次压缩
func (s *Scheduler) checkBatchDone(projectID int64, taskID int64) {
	// 查任务的 batch_no
	task, err := g.DB().Model("mvp_task").Where("id", taskID).Fields("batch_no").One()
	if err != nil || task.IsEmpty() {
		return
	}
	batchNo := task["batch_no"].Int()

	// 统计该批次未完成的任务数
	count, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("batch_no", batchNo).
		Where("deleted_at IS NULL").
		WhereNotIn("status", []string{"completed"}).
		Count()
	if err != nil || count > 0 {
		return
	}

	// 批次全部完成，触发批次压缩 → 合并进全局摘要
	GetCompressor().CompressBatchContext(context.Background(), projectID, batchNo)
}

// checkProjectDone 检查项目所有任务是否全部完成
func (s *Scheduler) checkProjectDone(projectID int64) {
	// 统计未完成的任务数
	count, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("deleted_at IS NULL").
		WhereNotIn("status", []string{"completed"}).
		Count()
	if err != nil {
		return
	}

	if count == 0 {
		// 所有任务完成，更新项目状态
		g.DB().Model("mvp_project").Where("id", projectID).Update(g.Map{
			"status":     "completed",
			"updated_at": gtime.Now(),
		})

		// 停止调度
		s.mu.Lock()
		if cancel, ok := s.projectCtx[projectID]; ok {
			cancel()
			delete(s.projectCtx, projectID)
		}
		s.mu.Unlock()
	}
}

// RetryTask 重新执行失败的任务
func (s *Scheduler) RetryTask(projectID int64, taskID int64) error {
	_, err := g.DB().Model("mvp_task").Where("id", taskID).WhereIn("status", g.Slice{"failed", "submit_error"}).Update(g.Map{
		"status":        "pending",
		"error_message": nil,
		"updated_at":    gtime.Now(),
	})
	if err != nil {
		return err
	}

	logTaskAction(taskID, "retry", "failed", "pending", "用户重新开始任务", "user")

	// 触发调度
	go s.scheduleOnce(context.Background(), projectID)
	return nil
}

// --- 辅助函数 ---

// parseResources 解析 JSON 资源列表
func parseResources(jsonStr string) []string {
	if jsonStr == "" || jsonStr == "null" {
		return nil
	}
	var resources []string
	json.Unmarshal([]byte(jsonStr), &resources)
	return resources
}

// logTaskAction 记录任务日志
func logTaskAction(taskID int64, action, fromStatus, toStatus, message, operator string) {
	g.DB().Model("mvp_task_log").Insert(g.Map{
		"task_id":     taskID,
		"action":      action,
		"from_status": fromStatus,
		"to_status":   toStatus,
		"message":     message,
		"operator":    operator,
		"created_at":  gtime.Now(),
	})
}
