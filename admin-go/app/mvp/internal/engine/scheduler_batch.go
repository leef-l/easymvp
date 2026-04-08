package engine

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

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
	ctx := context.Background()
	if _, err := g.DB().Model("mvp_project").Ctx(ctx).Where("id", projectID).Update(g.Map{
		"active_batch_no": batchNo,
		"updated_at":      gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[Scheduler] 持久化活跃批次失败: project=%d, batch=%d, err=%v", projectID, batchNo, err)
	}
}

// advanceBatchIfDone 检查任务所在批次是否全部完成，完成则推进活跃批次
// 事件驱动：只在任务完成时调用，不在轮询中反复检查
func (s *Scheduler) advanceBatchIfDone(projectID int64, taskID int64) {
	task, err := g.DB().Model("mvp_task").Ctx(context.Background()).Where("id", taskID).Fields("batch_no").One()
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

// OnTaskCompleted 任务完成回调，触发动态推进
// 合并为单个 goroutine 按顺序执行，避免回调风暴
func (s *Scheduler) OnTaskCompleted(projectID int64, taskID int64) {
	s.releaseTaskResources(taskID)

	logTaskAction(taskID, "completed", "running", "completed", "任务执行完成", "system")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Scheduler] OnTaskCompleted panic: projectID=%d taskID=%d err=%v", projectID, taskID, r)
			}
		}()
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
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Scheduler] feishuNotify panic: %v", r)
			}
		}()
		ctx := s.getProjectContext(projectID)
		task, _ := g.DB().Ctx(ctx).Model("mvp_task").Where("id", taskID).Fields("name").One()
		taskName := ""
		if !task.IsEmpty() {
			taskName = task["name"].String()
		}
		feishuNotifyTaskFailed(ctx, projectID, taskID, taskName, errMsg)
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[Scheduler] scheduleOnce panic: %v", r)
			}
		}()
		s.scheduleOnce(s.getProjectContext(projectID), projectID)
	}()
}

// OnTaskEscalated 任务升级给架构师后的回调
func (s *Scheduler) OnTaskEscalated(projectID int64, taskID int64, message string) {
	s.releaseTaskResources(taskID)
	logTaskAction(taskID, "escalate_to_architect", "running", "escalated", message, "system")
	go s.scheduleOnce(s.getProjectContext(projectID), projectID)
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
		resJSON, jsonErr := json.Marshal(resources)
		if jsonErr != nil {
			g.Log().Warningf(context.Background(), "[Scheduler] 序列化资源锁信息失败: task=%d err=%v", taskID, jsonErr)
		} else if _, upErr := g.DB().Model("mvp_task").Ctx(context.Background()).Where("id", taskID).Update(g.Map{
			"locked_resources": string(resJSON),
		}); upErr != nil {
			g.Log().Warningf(context.Background(), "[Scheduler] 更新任务资源锁失败: task=%d err=%v", taskID, upErr)
		}
	}

	return true
}
