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

// OnTaskCompleted 任务完成回调，触发动态推进
// 合并为单个 goroutine 按顺序执行，避免回调风暴
func (s *Scheduler) OnTaskCompleted(projectID int64, taskID int64) {
	s.releaseTaskResources(taskID)

	logTaskAction(taskID, "completed", "running", "completed", "任务执行完成", "system")

	go func() {
		ctx := s.getProjectContext(projectID)
		s.advanceBatchIfDone(projectID, taskID)
		s.scheduleOnce(ctx, projectID)
		s.checkProjectDone(projectID)
	}()
}

// OnTaskFailed 任务失败回调
func (s *Scheduler) OnTaskFailed(projectID int64, taskID int64, errMsg string) {
	s.releaseTaskResources(taskID)
	logTaskAction(taskID, "failed", "running", "failed", errMsg, "system")

	// 飞书推送：任务失败通知
	go func() {
		ctx := s.getProjectContext(projectID)
		task, _ := g.DB().Ctx(ctx).Model("mvp_task").Where("id", taskID).Fields("name").One()
		taskName := ""
		if !task.IsEmpty() {
			taskName = task["name"].String()
		}
		feishuNotifyTaskFailed(ctx, projectID, taskID, taskName, errMsg)
	}()

	go s.scheduleOnce(s.getProjectContext(projectID), projectID)
}

// OnTaskEscalated 任务升级给架构师后的回调
func (s *Scheduler) OnTaskEscalated(projectID int64, taskID int64, message string) {
	s.releaseTaskResources(taskID)
	logTaskAction(taskID, "escalate_to_architect", "running", "escalated", message, "system")
	go s.scheduleOnce(s.getProjectContext(projectID), projectID)
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
				updateTaskStatus(ctx, cycleNode, "pending", "failed", g.Map{
					"error_message": "检测到循环依赖死锁，自动标记失败。请检查任务依赖关系。",
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
// 同时将资源锁持久化到 DB（mvp_task.locked_resources）
func (s *Scheduler) tryLockResources(taskID int64, resources []string) bool {
	for _, res := range resources {
		if occupier, exists := s.lockedRes[res]; exists && occupier != taskID {
			return false
		}
	}
	for _, res := range resources {
		s.lockedRes[res] = taskID
	}

	// 持久化到 DB
	if len(resources) > 0 {
		resJSON, _ := json.Marshal(resources)
		g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
			"locked_resources": string(resJSON),
		})
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

	// 计算并推进到下一个活跃批次（持久化到 DB）
	nextBatch := s.calcActiveBatch(projectID)
	s.persistActiveBatch(projectID, nextBatch)

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

// RetryTask 重新执行失败的任务
func (s *Scheduler) RetryTask(projectID int64, taskID int64) error {
	// 查询当前状态，通过状态机校验
	task, err := g.DB().Model("mvp_task").Where("id", taskID).Fields("status").One()
	if err != nil || task.IsEmpty() {
		return fmt.Errorf("任务不存在")
	}
	currentStatus := task["status"].String()

	rows, err := updateTaskStatus(context.Background(), taskID, currentStatus, "pending", g.Map{
		"error_message": nil,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("任务状态已变更，无法重试")
	}

	logTaskAction(taskID, "retry", currentStatus, "pending", "用户重新开始任务", "user")

	go s.scheduleOnce(s.getProjectContext(projectID), projectID)
	return nil
}

// SkipTask 手动跳过无法完成的任务，防止批次永久阻塞
// 将任务标记为 completed（跳过），并尝试推进批次
func (s *Scheduler) SkipTask(ctx context.Context, projectID int64, taskID int64, reason string) error {
	// 查询当前状态，只允许跳过 failed/bug_found
	task, err := g.DB().Model("mvp_task").
		Where("id", taskID).
		Where("project_id", projectID).
		Fields("status").
		One()
	if err != nil || task.IsEmpty() {
		return fmt.Errorf("任务不存在")
	}
	currentStatus := task["status"].String()
	if currentStatus != "failed" && currentStatus != "bug_found" {
		return fmt.Errorf("任务状态(%s)不允许跳过（仅 failed/bug_found 可跳过）", currentStatus)
	}

	rows, err := updateTaskStatus(ctx, taskID, currentStatus, "completed", g.Map{
		"result":       fmt.Sprintf("[已跳过] %s", reason),
		"completed_at": gtime.Now(),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("任务状态已变更，无法跳过")
	}

	logTaskAction(taskID, "skipped", currentStatus, "completed",
		fmt.Sprintf("用户手动跳过: %s", reason), "user")

	// 尝试推进批次（单 goroutine 顺序执行，避免回调风暴）
	go func() {
		s.advanceBatchIfDone(projectID, taskID)
		s.scheduleOnce(s.getProjectContext(projectID), projectID)
		s.checkProjectDone(projectID)
	}()

	return nil
}

// GetActiveBatch 获取项目当前活跃批次号（供外部查询，从 DB 读取）
func (s *Scheduler) GetActiveBatch(projectID int64) int {
	return s.getActiveBatchFromDB(projectID)
}

// getActiveBatchFromDB 从 DB 读取项目活跃批次号
func (s *Scheduler) getActiveBatchFromDB(projectID int64) int {
	project, err := g.DB().Model("mvp_project").
		Where("id", projectID).
		Fields("active_batch_no").
		One()
	if err != nil || project.IsEmpty() {
		return 0
	}
	return project["active_batch_no"].Int()
}

// persistActiveBatch 将活跃批次号持久化到 DB
func (s *Scheduler) persistActiveBatch(projectID int64, batchNo int) {
	if _, err := g.DB().Model("mvp_project").Where("id", projectID).Update(g.Map{
		"active_batch_no": batchNo,
		"updated_at":      gtime.Now(),
	}); err != nil {
		g.Log().Errorf(context.Background(), "[Scheduler] 持久化活跃批次失败: project=%d, batch=%d, err=%v", projectID, batchNo, err)
	}
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
