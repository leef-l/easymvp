package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// projectRuntime 项目运行时句柄，保存可取消的 context 用于传播暂停信号
type projectRuntime struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// Scheduler 任务调度器
// 核心职责：扫描待执行任务、批次调度、依赖检查、资源冲突检测、动态推进
//
// 批次门控策略：
//   - 每个项目的活跃批次号持久化到 mvp_project.active_batch_no（进程重启可恢复）
//   - 只有当前活跃批次的任务 + batch_no=0 的紧急任务可被调度
//   - 当前批次的所有任务完成后，自动推进到下一个有任务的批次
//   - failed/bug_found 的任务会阻塞批次推进（看门狗自动重试，或人工 SkipTask 跳过）
//
// 资源锁持久化：
//   - 任务持有的资源锁写入 mvp_task.locked_resources（JSON数组）
//   - 进程重启后从 DB 恢复 running 任务的资源锁
//   - 任务完成/失败时清理 DB 和内存中的资源锁
type Scheduler struct {
	mu             sync.Mutex
	running        map[int64]bool               // 正在执行的任务 ID（内存缓存，启动时从 DB 恢复）
	lockedRes      map[string]int64             // 已锁定的资源 -> 占用任务 ID（内存缓存，与 DB 同步）
	maxConcurrency int                          // 最大并发 goroutine 数
	executor       *Executor                    // 任务执行器
	projectCtx     map[int64]*projectRuntime    // 项目级运行句柄（ctx + cancel）
}

type resourceParseResult struct {
	Resources []string
	Rejected  []string
}

// NewScheduler 创建调度器
func NewScheduler(maxConcurrency int) *Scheduler {
	s := &Scheduler{
		running:        make(map[int64]bool),
		lockedRes:      make(map[string]int64),
		maxConcurrency: maxConcurrency,
		projectCtx:     make(map[int64]*projectRuntime),
	}
	s.executor = NewExecutor(s)

	// 从 DB 恢复 running 状态任务的资源锁
	s.recoverResourceLocks()

	return s
}

// recoverResourceLocks 进程启动时从 DB 恢复 running 任务的资源锁
func (s *Scheduler) recoverResourceLocks() {
	tasks, err := g.DB().Model("mvp_task").
		Where("status", "running").
		Where("deleted_at IS NULL").
		Fields("id, locked_resources").
		All()
	if err != nil {
		g.Log().Errorf(context.Background(), "[Scheduler] 恢复资源锁失败: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	recovered := 0
	for _, t := range tasks {
		taskID := t["id"].Int64()
		s.running[taskID] = true

		lockedStr := t["locked_resources"].String()
		if lockedStr == "" || lockedStr == "null" {
			continue
		}
		var resources []string
		if err := json.Unmarshal([]byte(lockedStr), &resources); err != nil {
			continue
		}
		for _, res := range resources {
			s.lockedRes[res] = taskID
			recovered++
		}
	}

	if recovered > 0 {
		g.Log().Infof(context.Background(), "[Scheduler] 从 DB 恢复了 %d 个资源锁，%d 个 running 任务", recovered, len(tasks))
	}
}

// 全局调度器（延迟初始化，等数据库就绪后读取配置）
var defaultScheduler *Scheduler

// GetScheduler 获取全局调度器（首次调用时从配置加载参数）
func GetScheduler() *Scheduler {
	if defaultScheduler == nil {
		ctx := context.Background()
		maxConcurrent := GetConfigInt(ctx, "scheduler.max_concurrent", "engine.scheduler.maxConcurrent", 20)
		defaultScheduler = NewScheduler(maxConcurrent)
		g.Log().Infof(ctx, "[Scheduler] 配置加载: maxConcurrent=%d", maxConcurrent)
	}
	return defaultScheduler
}

// GetExecutor 返回内部执行器（供外部注入 V2 回调使用）。
func (s *Scheduler) GetExecutor() *Executor { return s.executor }

// StartProject 启动项目任务调度
func (s *Scheduler) StartProject(projectID int64) {
	ctx, cancel := context.WithCancel(context.Background())

	// 从 DB 计算初始活跃批次号并持久化
	batchNo := s.calcActiveBatch(projectID)
	s.persistActiveBatch(projectID, batchNo)

	s.mu.Lock()
	if old, ok := s.projectCtx[projectID]; ok {
		old.cancel()
	}
	s.projectCtx[projectID] = &projectRuntime{ctx: ctx, cancel: cancel}
	s.mu.Unlock()

	g.Log().Infof(ctx, "[Scheduler] 项目 %d 启动调度，初始活跃批次: %d", projectID, batchNo)

	go s.scheduleLoop(ctx, projectID)
}

// PauseProject 暂停项目调度
func (s *Scheduler) PauseProject(projectID int64) {
	// 1. 持锁取出并删除项目 runtime
	s.mu.Lock()
	rt, hasRuntime := s.projectCtx[projectID]
	if hasRuntime {
		delete(s.projectCtx, projectID)
	}
	s.mu.Unlock()

	// 2. 锁外 cancel（停止调度循环 + 传播取消信号到所有异步链路）
	if hasRuntime {
		rt.cancel()
	}

	// 3. 锁外查询该项目的 running 任务 ID
	runningTasks, _ := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "running").
		Where("deleted_at IS NULL").
		Fields("id").
		All()

	// 4. 持锁只清理内存中的 running 和 lockedRes
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

	// 5. 锁外批量更新 DB：running → pending（失败重试一次）
	pauseUpdate := func() error {
		_, err := g.DB().Model("mvp_task").
			Where("project_id", projectID).
			Where("status", "running").
			Update(g.Map{
				"status":           "pending",
				"locked_resources": nil,
				"heartbeat_at":     nil,
				"updated_at":       gtime.Now(),
			})
		return err
	}
	if err := pauseUpdate(); err != nil {
		g.Log().Errorf(context.Background(), "[Scheduler] PauseProject DB 批量回退失败，重试一次: project=%d, err=%v", projectID, err)
		time.Sleep(500 * time.Millisecond)
		if retryErr := pauseUpdate(); retryErr != nil {
			g.Log().Errorf(context.Background(), "[Scheduler] PauseProject DB 批量回退重试仍失败: project=%d, err=%v（watchdog 将兜底修正）", projectID, retryErr)
		}
	}

	// 6. 重置项目活跃批次
	s.persistActiveBatch(projectID, 0)
}

// getProjectContext 获取项目级 context（支持 cancel 传播）
// 如果项目已暂停或不存在，返回 Background context
func (s *Scheduler) getProjectContext(projectID int64) context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rt, ok := s.projectCtx[projectID]; ok {
		return rt.ctx
	}
	return context.Background()
}

// scheduleLoop 调度主循环
func (s *Scheduler) scheduleLoop(ctx context.Context, projectID int64) {
	// 首次立即调度
	s.scheduleOnce(ctx, projectID)

	pollInterval := GetConfigInt(ctx, "scheduler.poll_interval", "engine.scheduler.pollInterval", 2)
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
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
// 两阶段设计：锁外查 DB，锁内做决策
func (s *Scheduler) scheduleOnce(ctx context.Context, projectID int64) {
	// --- 阶段 1：从 DB 读取活跃批次号 + 并发度快检 ---
	s.mu.Lock()
	currentRunning := len(s.running)
	if currentRunning >= s.maxConcurrency {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	batchNo := s.getActiveBatchFromDB(projectID)

	// --- 阶段 1.5：阶段锁 —— 按项目状态过滤允许执行的角色 ---
	projectStatus := ""
	proj, _ := g.DB().Model("mvp_project").Where("id", projectID).Fields("status").One()
	if !proj.IsEmpty() {
		projectStatus = proj["status"].String()
	}
	allowedRoles := stageAllowedRoles(projectStatus)

	// --- 阶段 2：锁外查 DB，获取候选任务 ---
	query := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "pending").
		Where("deleted_at IS NULL")

	// 阶段锁：只调度允许的角色
	if len(allowedRoles) > 0 {
		query = query.WhereIn("role_type", allowedRoles)
	} else if projectStatus == "paused" || projectStatus == "designing" || projectStatus == "reviewing" {
		return // 这些状态不应执行任何任务
	}

	if batchNo > 0 {
		// 有活跃批次：只调度 batch_no=0（紧急）和当前活跃批次
		query = query.Where("(batch_no = 0 OR batch_no = ?)", batchNo)
	} else {
		// 无活跃批次（所有常规批次已完成），只调度 batch_no=0 的紧急任务
		query = query.Where("batch_no = 0")
	}

	tasks, err := query.Order("batch_no ASC, sort ASC").All()
	if err != nil {
		g.Log().Errorf(ctx, "[Scheduler] 调度扫描失败: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	// 锁外批量查询所有候选任务的依赖状态（消除 N+1）
	candidateIDs := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		candidateIDs = append(candidateIDs, t["id"].Int64())
	}
	depSatisfied := s.batchCheckDependencies(ctx, candidateIDs)

	// --- 阶段 3：锁内做调度决策 ---
	s.mu.Lock()
	defer s.mu.Unlock()

	currentRunning = len(s.running) // 重新读取，可能已变化

	for _, task := range tasks {
		if currentRunning >= s.maxConcurrency {
			break
		}

		taskID := task["id"].Int64()

		if s.running[taskID] {
			continue
		}

		// 依赖检查（使用锁外预计算结果）
		if !depSatisfied[taskID] {
			continue
		}

		// 资源冲突检测
		resources := parseResources(task["affected_resources"].String())
		if !s.tryLockResources(taskID, resources) {
			continue
		}

		// 用状态机 CAS 防止重复分发：只有 pending 才能变 running
		rows, _ := updateTaskStatus(ctx, taskID, "pending", "running", g.Map{
			"started_at": gtime.Now(),
		})
		if rows == 0 {
			// 已被其他 goroutine 拿走，释放刚加的资源锁
			for res, tid := range s.lockedRes {
				if tid == taskID {
					delete(s.lockedRes, res)
				}
			}
			continue
		}

		s.running[taskID] = true
		currentRunning++

		logTaskAction(taskID, "started", "pending", "running",
			fmt.Sprintf("任务开始执行 (batch=%d)", task["batch_no"].Int()), "system")

		go s.executor.Execute(ctx, projectID, taskID)
	}
}

// releaseTaskResources 释放任务持有的资源锁
// 短锁优先：持锁只清内存，锁外做 DB 清理
func (s *Scheduler) releaseTaskResources(taskID int64) {
	s.mu.Lock()
	_, wasRunning := s.running[taskID]
	delete(s.running, taskID)
	for res, tid := range s.lockedRes {
		if tid == taskID {
			delete(s.lockedRes, res)
		}
	}
	s.mu.Unlock()

	// 锁外清理 DB，失败只记日志不回滚内存
	if wasRunning {
		if _, err := g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
			"locked_resources": nil,
			"heartbeat_at":     nil,
		}); err != nil {
			g.Log().Errorf(context.Background(), "[Scheduler] releaseTaskResources DB 清理失败: task=%d, err=%v", taskID, err)
		}
	}
}

// checkProjectDone 检查项目所有任务是否全部完成
func (s *Scheduler) checkProjectDone(projectID int64) {
	count, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("deleted_at IS NULL").
		WhereNotIn("status", []string{"completed", "draft"}).
		Count()
	if err != nil {
		return
	}

	if count == 0 {
		g.DB().Model("mvp_project").Where("id", projectID).Update(g.Map{
			"status":          "completed",
			"active_batch_no": 0,
			"updated_at":      gtime.Now(),
		})

		// 飞书推送：项目完成通知
		go feishuNotifyProjectCompleted(context.Background(), projectID)

		s.mu.Lock()
		if rt, ok := s.projectCtx[projectID]; ok {
			rt.cancel()
			delete(s.projectCtx, projectID)
		}
		s.mu.Unlock()
	}
}

// --- 辅助函数 ---

// stageAllowedRoles 阶段锁：根据项目状态返回允许执行的角色类型
// designing/reviewing/paused 状态下不允许调度任何任务（返回 nil 表示不限制，空切片由调用方判断）
func stageAllowedRoles(projectStatus string) g.Slice {
	switch projectStatus {
	case "running":
		// running 状态：实施员、审计员可执行；架构师仅限系统创建的分析任务（batch_no=0）
		return g.Slice{"implementer", "auditor", "architect"}
	case "reviewing":
		// reviewing 状态：只允许审计员和协调员（审核流程内部使用）
		return g.Slice{"auditor", "coordinator"}
	default:
		return nil
	}
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

// feishuNotifyTaskFailed / feishuNotifyProjectCompleted
// 通过函数变量注入，避免 engine 包直接 import collab/notifier（循环引用）。
// 在 cmd_custom.go 启动时由 collab/notifier 包注册实现。
var feishuNotifyTaskFailed = func(ctx context.Context, projectID, taskID int64, taskName, errMsg string) {}
var feishuNotifyProjectCompleted = func(ctx context.Context, projectID int64) {}
