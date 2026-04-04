package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/worktreeguard"
)

// Scheduler 任务调度器
// 核心职责：扫描待执行任务、批次调度、依赖检查、资源冲突检测、动态推进
//
// 批次门控策略：
//   - 每个项目维护一个"当前活跃批次号"（内存缓存，事件驱动更新）
//   - 只有当前活跃批次的任务 + batch_no=0 的紧急任务可被调度
//   - 当前批次的所有任务完成后，自动推进到下一个有任务的批次
//   - failed/bug_found 的任务会阻塞批次推进（看门狗自动重试，或人工 SkipTask 跳过）
type Scheduler struct {
	mu             sync.Mutex
	running        map[int64]bool               // 正在执行的任务 ID
	lockedRes      map[string]int64             // 已锁定的资源 -> 占用任务 ID
	activeBatch    map[int64]int                // 项目 ID -> 当前活跃批次号（0 表示无活跃批次）
	maxConcurrency int                          // 最大并发 goroutine 数
	executor       *Executor                    // 任务执行器
	projectCtx     map[int64]context.CancelFunc // 项目级 cancel 函数（暂停用）
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
		activeBatch:    make(map[int64]int),
		maxConcurrency: maxConcurrency,
		projectCtx:     make(map[int64]context.CancelFunc),
	}
	s.executor = NewExecutor(s)
	return s
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

// StartProject 启动项目任务调度
func (s *Scheduler) StartProject(projectID int64) {
	ctx, cancel := context.WithCancel(context.Background())

	// 从 DB 计算初始活跃批次号（进程恢复时也走这里）
	batchNo := s.calcActiveBatch(projectID)

	s.mu.Lock()
	if oldCancel, ok := s.projectCtx[projectID]; ok {
		oldCancel()
	}
	s.projectCtx[projectID] = cancel
	s.activeBatch[projectID] = batchNo
	s.mu.Unlock()

	g.Log().Infof(ctx, "[Scheduler] 项目 %d 启动调度，初始活跃批次: %d", projectID, batchNo)

	go s.scheduleLoop(ctx, projectID)
}

// PauseProject 暂停项目调度
func (s *Scheduler) PauseProject(projectID int64) {
	s.mu.Lock()
	if cancel, ok := s.projectCtx[projectID]; ok {
		cancel()
		delete(s.projectCtx, projectID)
	}
	delete(s.activeBatch, projectID)
	s.mu.Unlock()

	// DB 操作移到锁外，避免持锁期间阻塞
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
	for res, tid := range s.lockedRes {
		if tid == taskID {
			delete(s.lockedRes, res)
		}
	}
	s.mu.Unlock()

	logTaskAction(taskID, "completed", "running", "completed", "任务执行完成", "system")

	// 检查批次完成 → 推进活跃批次（事件驱动，非轮询）
	go s.advanceBatchIfDone(projectID, taskID)

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

	// 释放资源后触发调度，让同批次其他 pending 任务有机会执行
	go s.scheduleOnce(context.Background(), projectID)
}

// OnTaskEscalated 任务升级给架构师后的回调
func (s *Scheduler) OnTaskEscalated(projectID int64, taskID int64, message string) {
	s.mu.Lock()
	delete(s.running, taskID)
	for res, tid := range s.lockedRes {
		if tid == taskID {
			delete(s.lockedRes, res)
		}
	}
	s.mu.Unlock()

	logTaskAction(taskID, "escalate_to_architect", "running", "escalated", message, "system")
	go s.scheduleOnce(context.Background(), projectID)
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
	// --- 阶段 1：锁外读取内存中的活跃批次号 + 并发度快检 ---
	s.mu.Lock()
	currentRunning := len(s.running)
	if currentRunning >= s.maxConcurrency {
		s.mu.Unlock()
		return
	}
	batchNo := s.activeBatch[projectID]
	s.mu.Unlock()

	// --- 阶段 2：锁外查 DB，获取候选任务 ---
	query := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "pending").
		Where("deleted_at IS NULL")

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

		s.running[taskID] = true
		currentRunning++

		// 锁内更新 DB（轻量写操作，可接受）
		g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
			"status":     "running",
			"started_at": gtime.Now(),
			"updated_at": gtime.Now(),
		})

		logTaskAction(taskID, "started", "pending", "running",
			fmt.Sprintf("任务开始执行 (batch=%d)", task["batch_no"].Int()), "system")

		go s.executor.Execute(ctx, projectID, taskID)
	}
}

// batchCheckDependencies 批量检查多个任务的依赖是否已满足
// 返回 taskID -> bool，true 表示所有依赖已完成
// 单次查询替代 N+1，大幅减少 DB 开销
// 同时检测循环依赖：如果发现环，标记环中一个任务为 failed
func (s *Scheduler) batchCheckDependencies(ctx context.Context, taskIDs []int64) map[int64]bool {
	result := make(map[int64]bool, len(taskIDs))
	for _, id := range taskIDs {
		result[id] = true // 默认满足（无依赖的任务）
	}

	if len(taskIDs) == 0 {
		return result
	}

	// 一次性查出所有候选任务的依赖关系及其状态
	deps, err := g.DB().Model("mvp_task_dependency d").
		LeftJoin("mvp_task t", "t.id = d.depends_on_id").
		Fields("d.task_id, d.depends_on_id, t.status").
		WhereIn("d.task_id", taskIDs).
		All()
	if err != nil {
		g.Log().Errorf(ctx, "[Scheduler] 批量依赖检查失败: %v", err)
		for _, id := range taskIDs {
			result[id] = false
		}
		return result
	}

	// 构建依赖图并检测循环依赖
	graph := make(map[int64][]int64)        // taskID -> 依赖的任务列表
	depStatusMap := make(map[int64]string)   // depends_on_id -> status
	for _, dep := range deps {
		taskID := dep["task_id"].Int64()
		depOnID := dep["depends_on_id"].Int64()
		depStatus := dep["status"].String()
		graph[taskID] = append(graph[taskID], depOnID)
		depStatusMap[depOnID] = depStatus

		if depStatus != "completed" {
			result[taskID] = false
		}
	}

	// DFS 循环依赖检测：只检查当前候选任务集中的互相依赖
	candidateSet := make(map[int64]bool, len(taskIDs))
	for _, id := range taskIDs {
		candidateSet[id] = true
	}

	// 检测候选集内的环：如果 A→B→A 且 A、B 都在 pending 候选中，则形成死锁
	visited := make(map[int64]int) // 0=未访问, 1=访问中, 2=已完成
	var cycleNode int64

	var dfs func(node int64) bool
	dfs = func(node int64) bool {
		visited[node] = 1 // 标记为访问中
		for _, dep := range graph[node] {
			if !candidateSet[dep] {
				continue // 依赖不在候选集中，不构成死锁
			}
			status := depStatusMap[dep]
			if status == "completed" {
				continue // 已完成的依赖不构成环
			}
			if visited[dep] == 1 {
				// 发现环！
				cycleNode = dep
				return true
			}
			if visited[dep] == 0 {
				if dfs(dep) {
					return true
				}
			}
		}
		visited[node] = 2 // 标记为已完成
		return false
	}

	for _, id := range taskIDs {
		if visited[id] == 0 {
			if dfs(id) {
				// 发现循环依赖，标记环中的一个任务为 failed 以打破死锁
				g.Log().Errorf(ctx, "[Scheduler] 检测到循环依赖，强制标记任务 %d 为失败以打破死锁", cycleNode)
				g.DB().Model("mvp_task").Where("id", cycleNode).Update(g.Map{
					"status":        "failed",
					"error_message": "检测到循环依赖死锁，自动标记失败。请检查任务依赖关系。",
					"updated_at":    gtime.Now(),
				})
				logTaskAction(cycleNode, "cycle_detected", "pending", "failed",
					"循环依赖检测：打破死锁", "system")
				result[cycleNode] = false
			}
		}
	}

	return result
}

// tryLockResources 尝试锁定资源，如果有冲突返回 false
func (s *Scheduler) tryLockResources(taskID int64, resources []string) bool {
	for _, res := range resources {
		if occupier, exists := s.lockedRes[res]; exists && occupier != taskID {
			return false
		}
	}
	for _, res := range resources {
		s.lockedRes[res] = taskID
	}
	return true
}

// advanceBatchIfDone 检查任务所在批次是否全部完成，完成则推进活跃批次
// 事件驱动：只在任务完成时调用，不在轮询中反复检查
func (s *Scheduler) advanceBatchIfDone(projectID int64, taskID int64) {
	task, err := g.DB().Model("mvp_task").Where("id", taskID).Fields("batch_no").One()
	if err != nil || task.IsEmpty() {
		return
	}
	batchNo := task["batch_no"].Int()

	// batch_no=0 是紧急任务，不参与批次门控
	if batchNo == 0 {
		return
	}

	// 统计该批次未完成的任务数
	// pending/running/failed/bug_found/bug_dispatched 都算未完成
	// draft 尚未规划、completed 已完成，不阻塞
	count, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("batch_no", batchNo).
		Where("deleted_at IS NULL").
		WhereNotIn("status", []string{"completed", "draft"}).
		Count()
	if err != nil || count > 0 {
		return
	}

	g.Log().Infof(context.Background(),
		"[Scheduler] 项目 %d 批次 %d 全部完成，触发批次压缩", projectID, batchNo)

	// 触发批次压缩
	GetCompressor().CompressBatchContext(context.Background(), projectID, batchNo)

	// 计算并推进到下一个活跃批次
	nextBatch := s.calcActiveBatch(projectID)

	s.mu.Lock()
	s.activeBatch[projectID] = nextBatch
	s.mu.Unlock()

	if nextBatch > 0 {
		g.Log().Infof(context.Background(),
			"[Scheduler] 项目 %d 推进到批次 %d", projectID, nextBatch)
	} else {
		g.Log().Infof(context.Background(),
			"[Scheduler] 项目 %d 所有常规批次已完成", projectID)
	}
}

// calcActiveBatch 从 DB 计算项目当前活跃批次号
// 返回最小的、还有未完成任务的 batch_no（排除 draft 和 batch_no=0）
// 返回 0 表示所有常规批次已完成
func (s *Scheduler) calcActiveBatch(projectID int64) int {
	result, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("batch_no > 0").
		Where("deleted_at IS NULL").
		WhereNotIn("status", []string{"completed", "draft"}).
		Fields("MIN(batch_no) as min_batch").
		One()
	if err != nil || result.IsEmpty() || result["min_batch"].IsNil() {
		return 0
	}
	return result["min_batch"].Int()
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
			"status":     "completed",
			"updated_at": gtime.Now(),
		})

		s.mu.Lock()
		if cancel, ok := s.projectCtx[projectID]; ok {
			cancel()
			delete(s.projectCtx, projectID)
		}
		delete(s.activeBatch, projectID)
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

	go s.scheduleOnce(context.Background(), projectID)
	return nil
}

// SkipTask 手动跳过无法完成的任务，防止批次永久阻塞
// 将任务标记为 completed（跳过），并尝试推进批次
func (s *Scheduler) SkipTask(ctx context.Context, projectID int64, taskID int64, reason string) error {
	// 只允许跳过 failed/bug_found 状态的任务
	result, err := g.DB().Model("mvp_task").
		Where("id", taskID).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"failed", "bug_found"}).
		Update(g.Map{
			"status":       "completed",
			"result":       fmt.Sprintf("[已跳过] %s", reason),
			"completed_at": gtime.Now(),
			"updated_at":   gtime.Now(),
		})
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("任务不存在或状态不允许跳过（仅 failed/bug_found 可跳过）")
	}

	logTaskAction(taskID, "skipped", "failed", "completed",
		fmt.Sprintf("用户手动跳过: %s", reason), "user")

	// 尝试推进批次
	go s.advanceBatchIfDone(projectID, taskID)
	go s.scheduleOnce(context.Background(), projectID)
	go s.checkProjectDone(projectID)

	return nil
}

// GetActiveBatch 获取项目当前活跃批次号（供外部查询）
func (s *Scheduler) GetActiveBatch(projectID int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeBatch[projectID]
}

// --- 辅助函数 ---

// parseResources 解析 JSON 资源列表
func parseResources(jsonStr string) []string {
	return parseResourcesDetail(jsonStr).Resources
}

func parseResourcesDetail(jsonStr string) resourceParseResult {
	result := resourceParseResult{}
	if jsonStr == "" || jsonStr == "null" {
		return result
	}
	var resources []string
	if err := json.Unmarshal([]byte(jsonStr), &resources); err != nil {
		g.Log().Warningf(context.Background(), "[Scheduler] parseResources 解析失败，跳过资源锁定，原始值: %s, 错误: %v", jsonStr, err)
		return result
	}
	normalized, dropped := worktreeguard.NormalizeRelativePaths(resources)
	if len(dropped) > 0 {
		g.Log().Warningf(context.Background(), "[Scheduler] parseResources 丢弃可疑资源: %v", dropped)
	}
	result.Resources = normalized
	result.Rejected = dropped
	return result
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
