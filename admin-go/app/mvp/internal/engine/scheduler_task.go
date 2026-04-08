package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/worktreeguard"
)

// RetryTask 重新执行失败的任务
func (s *Scheduler) RetryTask(projectID int64, taskID int64) error {
	// 查询当前状态，通过状态机校验
	ctx := context.Background()
	task, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", taskID).WhereNull("deleted_at").Fields("status").One()
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

	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Scheduler] scheduleOnce panic: project=%d err=%v", projectID, r)
			}
		}()
		s.scheduleOnce(s.getProjectContext(projectID), projectID)
	}()
	return nil
}

// SkipTask 手动跳过无法完成的任务，防止批次永久阻塞
// 将任务标记为 completed（跳过），并尝试推进批次
func (s *Scheduler) SkipTask(ctx context.Context, projectID int64, taskID int64, reason string) error {
	// 查询当前状态，只允许跳过 failed/bug_found
	task, err := g.DB().Ctx(ctx).Model("mvp_task").
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
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Scheduler] SkipTask goroutine panic: %v", r)
			}
		}()
		s.advanceBatchIfDone(projectID, taskID)
		s.scheduleOnce(s.getProjectContext(projectID), projectID)
		s.checkProjectDone(projectID)
	}()

	return nil
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
	deps, err := g.DB().Ctx(ctx).Model("mvp_task_dependency d").
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
