package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/utility/snowflake"
)

// WorkflowFailedCallback 工作流失败后的清理回调（停调度器、取消 runtime 等）。
type WorkflowFailedCallback func(ctx context.Context, workflowRunID int64)

// StageService 阶段编排服务。
type StageService struct {
	workflowSvc      *WorkflowService
	onWorkflowFailed WorkflowFailedCallback
}

// NewStageService 创建阶段服务。
func NewStageService(wfSvc *WorkflowService) *StageService {
	return &StageService{workflowSvc: wfSvc}
}

// SetWorkflowFailedCallback 注册工作流失败后的清理回调。
func (s *StageService) SetWorkflowFailedCallback(fn WorkflowFailedCallback) {
	s.onWorkflowFailed = fn
}

// StartStage 创建并启动指定类型的阶段。
// 返回新创建的 stage_run ID。
func (s *StageService) StartStage(ctx context.Context, workflowRunID int64, stageType string) (int64, error) {
	now := time.Now()

	// 获取下一个 stage_no
	maxNo, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		WhereNull("deleted_at").
		Max("stage_no")
	if err != nil {
		return 0, fmt.Errorf("查询 stage_no 失败: %w", err)
	}
	stageNo := int(maxNo) + 1

	// 创建 stage_run
	stageRunID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_stage_run").Ctx(ctx).Insert(g.Map{
		"id":              stageRunID,
		"workflow_run_id": workflowRunID,
		"stage_type":      stageType,
		"stage_no":        stageNo,
		"status":          consts.StageStatusRunning,
		"started_at":      now,
		"created_at":      now,
		"updated_at":      now,
	})
	if err != nil {
		return 0, fmt.Errorf("创建 stage_run 失败: %w", err)
	}

	// 更新 workflow_run 的 current_stage、current_stage_run_id 和 status
	wfStatus := StageTypeToWorkflowStatus(stageType)
	wfUpdate := g.Map{
		"current_stage":        stageType,
		"current_stage_run_id": stageRunID,
		"updated_at":           now,
	}
	if wfStatus != "" {
		wfUpdate["status"] = wfStatus
	}
	_, err = g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).
		Update(wfUpdate)
	if err != nil {
		return 0, fmt.Errorf("更新 workflow_run current_stage 失败: %w", err)
	}

	// 发布事件
	if s.workflowSvc.publisher != nil {
		s.workflowSvc.publisher.Emit(ctx, event.Event{
			WorkflowRunID: workflowRunID,
			EntityType:    event.EntityStageRun,
			EntityID:      &stageRunID,
			EventType:     event.EventStageStarted,
		})
	}

	g.Log().Infof(ctx, "[StageService] StartStage workflowRunID=%d stageType=%s stageRunID=%d", workflowRunID, stageType, stageRunID)
	return stageRunID, nil
}

// CompleteStage 完成阶段（CAS: running→completed）。
func (s *StageService) CompleteStage(ctx context.Context, stageRunID int64) error {
	now := gtime.Now()

	result, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("id", stageRunID).
		Where("status", consts.StageStatusRunning).
		Data(g.Map{
			"status":      consts.StageStatusCompleted,
			"finished_at": now,
			"updated_at":  now,
		}).Update()
	if err != nil {
		return fmt.Errorf("完成 stage_run 失败: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("stage_run(%d) 不在 running 状态", stageRunID)
	}

	// 查 workflow_run_id 发事件
	wfRunID, _ := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("id", stageRunID).Value("workflow_run_id")
	if s.workflowSvc.publisher != nil && wfRunID.Int64() > 0 {
		s.workflowSvc.publisher.Emit(ctx, event.Event{
			WorkflowRunID: wfRunID.Int64(),
			EntityType:    event.EntityStageRun,
			EntityID:      &stageRunID,
			EventType:     event.EventStageCompleted,
		})
	}

	g.Log().Infof(ctx, "[StageService] CompleteStage stageRunID=%d", stageRunID)
	return nil
}

// FailStage 标记阶段失败（CAS: running→failed）。
func (s *StageService) FailStage(ctx context.Context, stageRunID int64, reason string) error {
	now := gtime.Now()

	result, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("id", stageRunID).
		Where("status", consts.StageStatusRunning).
		Data(g.Map{
			"status":        consts.StageStatusFailed,
			"error_message": reason,
			"finished_at":  now,
			"updated_at":   now,
		}).Update()
	if err != nil {
		return fmt.Errorf("标记 stage_run 失败: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("stage_run(%d) 不在 running 状态", stageRunID)
	}

	g.Log().Infof(ctx, "[StageService] FailStage stageRunID=%d reason=%s", stageRunID, reason)

	// 同步 workflow_run 状态为 failed，并终止执行链
	stageRun, _ := g.DB().Model("mvp_stage_run").Ctx(ctx).Where("id", stageRunID).One()
	if !stageRun.IsEmpty() {
		wfRunID := stageRun["workflow_run_id"].Int64()
		// CAS 更新：从任何活跃状态 → failed
		updated := false
		for _, fromStatus := range []string{
			consts.WorkflowRunStatusExecuting,
			consts.WorkflowRunStatusReviewing,
			consts.WorkflowRunStatusReworking,
			consts.WorkflowRunStatusDesigning,
		} {
			rows, _ := s.workflowSvc.wfRepo.UpdateStatus(ctx, wfRunID, fromStatus, consts.WorkflowRunStatusFailed, g.Map{})
			if rows > 0 {
				updated = true
				break
			}
		}

		// 终止执行链：停调度器 + 取消 runtime
		if updated && s.onWorkflowFailed != nil {
			s.onWorkflowFailed(ctx, wfRunID)
		}
	}

	return nil
}

// TransitionNext 完成当前阶段并推进到下一阶段。
// 如果没有下一阶段（即 complete 之后），则完成整个 workflow_run。
func (s *StageService) TransitionNext(ctx context.Context, workflowRunID int64) error {
	// 查当前 stage
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).
		WhereNull("deleted_at").
		One()
	if err != nil || wfRun.IsEmpty() {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}

	currentStage := wfRun["current_stage"].String()
	nextStage := NextStage(currentStage)

	if nextStage == "" || nextStage == StageComplete {
		// 没有下一阶段或到达 complete，结束 workflow
		return s.completeWorkflow(ctx, workflowRunID)
	}

	// 启动下一阶段
	_, err = s.StartStage(ctx, workflowRunID, nextStage)
	return err
}

// TransitionTo 强制跳转到指定阶段（用于审核驳回回退 design 等场景）。
func (s *StageService) TransitionTo(ctx context.Context, workflowRunID int64, targetStage string) (int64, error) {
	return s.StartStage(ctx, workflowRunID, targetStage)
}

// completeWorkflow 完成整个工作流。
func (s *StageService) completeWorkflow(ctx context.Context, workflowRunID int64) error {
	now := gtime.Now()

	// 创建 complete stage
	completeStageID, err := s.StartStage(ctx, workflowRunID, consts.StageTypeComplete)
	if err != nil {
		return err
	}
	// 立即完成 complete stage
	if err := s.CompleteStage(ctx, completeStageID); err != nil {
		return err
	}

	// workflow_run → completed（从任何活跃阶段状态完成）
	_, err = g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).
		WhereNotIn("status", g.Slice{consts.WorkflowRunStatusCompleted, consts.WorkflowRunStatusCanceled}).
		Update(g.Map{
			"status":      consts.WorkflowRunStatusCompleted,
			"finished_at": now,
			"updated_at":  now,
		})
	if err != nil {
		return fmt.Errorf("完成 workflow_run 失败: %w", err)
	}

	// 更新项目状态
	projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).Value("project_id")
	if projectID.Int64() > 0 {
		_, _ = g.DB().Model("mvp_project").Ctx(ctx).
			Where("id", projectID.Int64()).
			Update(g.Map{"status": "completed", "updated_at": now})
	}

	s.workflowSvc.runtimeMgr.Cancel(workflowRunID)

	g.Log().Infof(ctx, "[StageService] completeWorkflow workflowRunID=%d", workflowRunID)
	return nil
}
