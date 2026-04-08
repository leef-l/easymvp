package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

// WorkflowFailedCallback 工作流失败后的清理回调（停调度器、取消 runtime 等）。
type WorkflowFailedCallback func(ctx context.Context, workflowRunID int64)

// StageReportFn 阶段完成后生成报告的回调。
type StageReportFn func(ctx context.Context, workflowRunID int64, stageType string)

// StageService 阶段编排服务。
type StageService struct {
	workflowSvc      *WorkflowService
	onWorkflowFailed WorkflowFailedCallback
	onAcceptTrigger  AcceptTriggerFn
	onStageReport    StageReportFn
}

// SetStageReportFn 注册阶段报告回调。
func (s *StageService) SetStageReportFn(fn StageReportFn) { s.onStageReport = fn }

// NewStageService 创建阶段服务。
func NewStageService(wfSvc *WorkflowService) *StageService {
	return &StageService{workflowSvc: wfSvc}
}

// SetWorkflowFailedCallback 注册工作流失败后的清理回调。
func (s *StageService) SetWorkflowFailedCallback(fn WorkflowFailedCallback) {
	s.onWorkflowFailed = fn
}

// StartStage 创建并启动指定类型的阶段。
// 整个操作在同一事务中完成，保证 stage_run 创建 + workflow_run 更新的原子性。
// 返回新创建的 stage_run ID。
func (s *StageService) StartStage(ctx context.Context, workflowRunID int64, stageType string) (int64, error) {
	now := time.Now()
	stageRunID := int64(snowflake.Generate())

	targetWfStatus := StageTypeToWorkflowStatus(stageType)

	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 查当前 workflow_run 状态，校验状态迁移合法性
		wfRun, err := tx.Model("mvp_workflow_run").Ctx(ctx).
			Where("id", workflowRunID).
			WhereNull("deleted_at").
			Fields("id, status").
			One()
		if err != nil {
			return fmt.Errorf("查询 workflow_run 失败: %w", err)
		}
		if wfRun.IsEmpty() {
			return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
		}

		currentStatus := wfRun["status"].String()
		if targetWfStatus != "" && !IsValidWorkflowTransition(currentStatus, targetWfStatus) {
			return fmt.Errorf("工作流状态迁移不合法: %s → %s (stage=%s)", currentStatus, targetWfStatus, stageType)
		}

		// 2. 获取下一个 stage_no
		maxNo, err := tx.Model("mvp_stage_run").Ctx(ctx).
			Where("workflow_run_id", workflowRunID).
			WhereNull("deleted_at").
			Max("stage_no")
		if err != nil {
			return fmt.Errorf("查询 stage_no 失败: %w", err)
		}
		stageNo := int(maxNo) + 1

		// 3. 创建 stage_run（继承项目归属字段）
		scope := repo.GetProjectScopeByWorkflowRun(ctx, workflowRunID)
		_, err = tx.Model("mvp_stage_run").Ctx(ctx).Insert(g.Map{
			"id":              stageRunID,
			"workflow_run_id": workflowRunID,
			"stage_type":      stageType,
			"stage_no":        stageNo,
			"status":          consts.StageStatusRunning,
			"created_by":      scope.CreatedBy,
			"dept_id":         scope.DeptID,
			"started_at":      now,
			"created_at":      now,
			"updated_at":      now,
		})
		if err != nil {
			return fmt.Errorf("创建 stage_run 失败: %w", err)
		}

		// 4. CAS 更新 workflow_run（WHERE status = currentStatus 防并发）
		wfUpdate := g.Map{
			"current_stage":        stageType,
			"current_stage_run_id": stageRunID,
			"updated_at":           now,
		}
		if targetWfStatus != "" {
			wfUpdate["status"] = targetWfStatus
		}
		wfResult, err := tx.Model("mvp_workflow_run").Ctx(ctx).
			Where("id", workflowRunID).
			Where("status", currentStatus).
			Update(wfUpdate)
		if err != nil {
			return fmt.Errorf("更新 workflow_run current_stage 失败: %w", err)
		}
		if rows, _ := wfResult.RowsAffected(); rows == 0 {
			return fmt.Errorf("workflow_run(%d) 状态并发冲突（期望 %s），无法创建 %s 阶段", workflowRunID, currentStatus, stageType)
		}

		// 5. 同步 mvp_project.status（项目列表依赖此字段展示状态）
		if targetWfStatus != "" {
			projectID, pidErr := tx.Model("mvp_workflow_run").Ctx(ctx).
				Where("id", workflowRunID).Value("project_id")
			if pidErr != nil {
				g.Log().Warningf(ctx, "[StageService] 查询 project_id 失败: wfRunID=%d err=%v", workflowRunID, pidErr)
			}
			if projectID.Int64() > 0 {
				if _, syncErr := tx.Model("mvp_project").Ctx(ctx).
					Where("id", projectID.Int64()).
					Update(g.Map{"status": targetWfStatus, "updated_at": now}); syncErr != nil {
					g.Log().Warningf(ctx, "[StageService] 同步 project status 失败: projectID=%d err=%v", projectID.Int64(), syncErr)
				}
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	// 事务外发布事件（只发一次，带 payload）
	if s.workflowSvc.publisher != nil {
		s.workflowSvc.publisher.Emit(ctx, event.Event{
			WorkflowRunID: workflowRunID,
			StageRunID:    &stageRunID,
			EntityType:    event.EntityStageRun,
			EntityID:      &stageRunID,
			EventType:     event.EventStageStarted,
			Payload:       map[string]string{"stage_type": stageType},
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
	wfRunID, wfErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("id", stageRunID).Value("workflow_run_id")
	if wfErr != nil {
		g.Log().Warningf(ctx, "[StageService] 查询 workflow_run_id 失败: stageRunID=%d err=%v", stageRunID, wfErr)
	}
	if s.workflowSvc.publisher != nil && wfRunID.Int64() > 0 {
		s.workflowSvc.publisher.Emit(ctx, event.Event{
			WorkflowRunID: wfRunID.Int64(),
			EntityType:    event.EntityStageRun,
			EntityID:      &stageRunID,
			EventType:     event.EventStageCompleted,
		})
	}

	g.Log().Infof(ctx, "[StageService] CompleteStage stageRunID=%d", stageRunID)

	// 异步生成阶段报告
	if s.onStageReport != nil && wfRunID.Int64() > 0 {
		stageType, stErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
			Where("id", stageRunID).Value("stage_type")
		if stErr != nil {
			g.Log().Warningf(ctx, "[StageService] 查询 stage_type 失败: stageRunID=%d err=%v", stageRunID, stErr)
		}
		if st := stageType.String(); st != "" {
			wfID := wfRunID.Int64()
			go func() {
				defer func() {
					if r := recover(); r != nil {
						g.Log().Errorf(context.Background(), "[StageService] onStageReport panic: wfRun=%d stage=%s err=%v", wfID, st, r)
					}
				}()
				s.onStageReport(context.Background(), wfID, st)
			}()
		}
	}

	return nil
}

// FailStageOnly 仅标记 stage_run 自身为 failed，不级联到 workflow_run。
// 用于调用方需要自行控制 workflow_run 状态的场景（如 review→execute 启动失败后由 review 侧统一回滚）。
func (s *StageService) FailStageOnly(ctx context.Context, stageRunID int64, reason string) {
	now := gtime.Now()
	result, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("id", stageRunID).
		Where("status", consts.StageStatusRunning).
		Data(g.Map{
			"status":        consts.StageStatusFailed,
			"error_message": reason,
			"finished_at":   now,
			"updated_at":    now,
		}).Update()
	if err != nil {
		g.Log().Errorf(ctx, "[StageService] FailStageOnly 更新失败: stageRun=%d err=%v", stageRunID, err)
		return
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		g.Log().Warningf(ctx, "[StageService] FailStageOnly 未命中（stage 可能不在 running 状态）: stageRun=%d", stageRunID)
	}
	g.Log().Infof(ctx, "[StageService] FailStageOnly stageRunID=%d reason=%s", stageRunID, reason)
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
			"finished_at":   now,
			"updated_at":    now,
		}).Update()
	if err != nil {
		return fmt.Errorf("标记 stage_run 失败: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("stage_run(%d) 不在 running 状态", stageRunID)
	}

	g.Log().Infof(ctx, "[StageService] FailStage stageRunID=%d reason=%s", stageRunID, reason)

	// 查 stage_run 关联的 workflow_run（只查一次）
	stageRun, srErr := g.DB().Model("mvp_stage_run").Ctx(ctx).Where("id", stageRunID).WhereNull("deleted_at").One()
	if srErr != nil {
		g.Log().Errorf(ctx, "[StageService] FailStage 查询 stage_run 失败: stageRun=%d err=%v", stageRunID, srErr)
		return nil
	}
	if stageRun.IsEmpty() {
		return nil
	}
	wfRunID := stageRun["workflow_run_id"].Int64()

	// 发射 stage.failed 事件
	if s.workflowSvc.publisher != nil {
		s.workflowSvc.publisher.Emit(ctx, event.Event{
			WorkflowRunID: wfRunID,
			StageRunID:    &stageRunID,
			EntityType:    event.EntityStageRun,
			EntityID:      &stageRunID,
			EventType:     event.EventStageFailed,
			Payload:       map[string]string{"reason": reason},
		})
	}

	// 同步 workflow_run 状态为 failed，并终止执行链
	// CAS 更新：从任何活跃状态 → failed
	updated := false
	for _, fromStatus := range []string{
		consts.WorkflowRunStatusExecuting,
		consts.WorkflowRunStatusAccepting,
		consts.WorkflowRunStatusReviewing,
		consts.WorkflowRunStatusReworking,
		consts.WorkflowRunStatusDesigning,
	} {
		wfRows, wfErr := s.workflowSvc.wfRepo.UpdateStatus(ctx, wfRunID, fromStatus, consts.WorkflowRunStatusFailed, g.Map{})
		if wfErr != nil {
			g.Log().Warningf(ctx, "[StageService] FailStage 更新 workflow 状态失败: wfRun=%d from=%s err=%v", wfRunID, fromStatus, wfErr)
			continue
		}
		if wfRows > 0 {
			updated = true
			break
		}
	}

	// 同步 mvp_project.status
	if updated {
		projectID, pidErr := g.DB().Model("mvp_workflow_run").Ctx(ctx).
			Where("id", wfRunID).Value("project_id")
		if pidErr != nil {
			g.Log().Warningf(ctx, "[StageService] FailStage 查询 project_id 失败: wfRun=%d err=%v", wfRunID, pidErr)
		}
		if projectID.Int64() > 0 {
			if _, upErr := g.DB().Model("mvp_project").Ctx(ctx).
				Where("id", projectID.Int64()).
				Update(g.Map{"status": consts.WorkflowRunStatusFailed, "pause_reason": reason, "updated_at": gtime.Now()}); upErr != nil {
				g.Log().Errorf(ctx, "[StageService] FailStage 同步项目状态失败: project=%d err=%v", projectID.Int64(), upErr)
			}
		}
	}

	// 终止执行链：停调度器 + 取消 runtime
	if updated && s.onWorkflowFailed != nil {
		s.onWorkflowFailed(ctx, wfRunID)
	}

	return nil
}

// AcceptTriggerFn accept 阶段触发回调。
type AcceptTriggerFn func(ctx context.Context, workflowRunID, stageRunID int64) error

// SetAcceptTriggerFn 注册 accept 阶段触发回调。
func (s *StageService) SetAcceptTriggerFn(fn AcceptTriggerFn) {
	s.onAcceptTrigger = fn
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

	// 灰度开关：当 workflow.accept.enabled=false 时，跳过 accept 直接进入 complete
	if nextStage == StageAccept && !isAcceptEnabled(ctx) {
		g.Log().Infof(ctx, "[StageService] accept 灰度未开启，跳过 accept 直接进入 complete: workflowRunID=%d", workflowRunID)
		nextStage = NextStage(StageAccept) // complete
	}

	if nextStage == "" || nextStage == StageComplete {
		// 没有下一阶段或到达 complete，结束 workflow
		return s.completeWorkflow(ctx, workflowRunID)
	}

	// 启动下一阶段
	stageRunID, err := s.StartStage(ctx, workflowRunID, nextStage)
	if err != nil {
		return err
	}

	// accept 阶段需要异步触发验收流程
	if nextStage == StageAccept && s.onAcceptTrigger != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					g.Log().Errorf(context.Background(), "[StageService] onAcceptTrigger panic: wfRun=%d stage=%d err=%v", workflowRunID, stageRunID, r)
					_ = s.FailStage(context.Background(), stageRunID, fmt.Sprintf("accept trigger panic: %v", r))
				}
			}()
			if triggerErr := s.onAcceptTrigger(ctx, workflowRunID, stageRunID); triggerErr != nil {
				g.Log().Errorf(ctx, "[StageService] accept 触发失败: workflowRunID=%d err=%v", workflowRunID, triggerErr)
				_ = s.FailStage(context.Background(), stageRunID, "accept 触发失败: "+triggerErr.Error())
			}
		}()
	}

	return nil
}

// isAcceptEnabled ��查 accept 灰度开关。
func isAcceptEnabled(ctx context.Context) bool {
	val, err := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", "workflow.accept.enabled").
		WhereNull("deleted_at").
		Value("config_value")
	if err != nil || val.IsEmpty() {
		// 默认开启
		return true
	}
	return val.String() == "true" || val.String() == "1"
}

// TransitionTo 强制跳转到指定阶段（用于审核驳回回退 design 等场景）。
func (s *StageService) TransitionTo(ctx context.Context, workflowRunID int64, targetStage string) (int64, error) {
	return s.StartStage(ctx, workflowRunID, targetStage)
}

// ForceStartStage 强制启动指定阶段。
// 用于人工接管场景：绕过常规迁移约束，终止当前活跃阶段并创建新的 running stage_run。
func (s *StageService) ForceStartStage(ctx context.Context, workflowRunID int64, stageType, reason string) (int64, error) {
	now := time.Now()
	stageRunID := int64(snowflake.Generate())
	targetWfStatus := StageTypeToWorkflowStatus(stageType)
	if targetWfStatus == "" {
		return 0, fmt.Errorf("未知阶段类型: %s", stageType)
	}

	var projectID int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		wfRun, err := tx.Model("mvp_workflow_run").Ctx(ctx).
			Where("id", workflowRunID).
			WhereNull("deleted_at").
			Fields("id, project_id, status, current_stage_run_id").
			One()
		if err != nil {
			return fmt.Errorf("查询 workflow_run 失败: %w", err)
		}
		if wfRun.IsEmpty() {
			return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
		}
		projectID = wfRun["project_id"].Int64()

		// 终止当前 workflow 下所有 pending/running 的阶段实例，避免出现多个活跃阶段并存。
		if _, err := tx.Model("mvp_stage_run").Ctx(ctx).
			Where("workflow_run_id", workflowRunID).
			WhereIn("status", g.Slice{consts.StageStatusPending, consts.StageStatusRunning}).
			WhereNull("deleted_at").
			Data(g.Map{
				"status":        consts.StageStatusFailed,
				"error_message": buildForceStageReason(stageType, reason),
				"finished_at":   now,
				"updated_at":    now,
			}).
			Update(); err != nil {
			return fmt.Errorf("终止当前活跃阶段失败: %w", err)
		}

		maxNo, err := tx.Model("mvp_stage_run").Ctx(ctx).
			Where("workflow_run_id", workflowRunID).
			WhereNull("deleted_at").
			Max("stage_no")
		if err != nil {
			return fmt.Errorf("查询 stage_no 失败: %w", err)
		}
		stageNo := int(maxNo) + 1
		scope := repo.GetProjectScopeByWorkflowRun(ctx, workflowRunID)

		if _, err := tx.Model("mvp_stage_run").Ctx(ctx).Insert(g.Map{
			"id":              stageRunID,
			"workflow_run_id": workflowRunID,
			"stage_type":      stageType,
			"stage_no":        stageNo,
			"status":          consts.StageStatusRunning,
			"created_by":      scope.CreatedBy,
			"dept_id":         scope.DeptID,
			"started_at":      now,
			"created_at":      now,
			"updated_at":      now,
		}); err != nil {
			return fmt.Errorf("创建 stage_run 失败: %w", err)
		}

		if _, err := tx.Model("mvp_workflow_run").Ctx(ctx).
			Where("id", workflowRunID).
			Update(g.Map{
				"status":               targetWfStatus,
				"current_stage":        stageType,
				"current_stage_run_id": stageRunID,
				"pause_reason":         nil,
				"status_before_pause":  nil,
				"finished_at":          nil,
				"updated_at":           now,
			}); err != nil {
			return fmt.Errorf("更新 workflow_run 失败: %w", err)
		}

		if projectID > 0 {
			if _, err := tx.Model("mvp_project").Ctx(ctx).
				Where("id", projectID).
				Update(g.Map{
					"status":       targetWfStatus,
					"pause_reason": nil,
					"updated_at":   now,
				}); err != nil {
				return fmt.Errorf("更新项目状态失败: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	if projectID > 0 {
		s.workflowSvc.runtimeMgr.Create(workflowRunID, projectID)
	}

	if s.workflowSvc.publisher != nil {
		s.workflowSvc.publisher.Emit(ctx, event.Event{
			WorkflowRunID: workflowRunID,
			StageRunID:    &stageRunID,
			EntityType:    event.EntityStageRun,
			EntityID:      &stageRunID,
			EventType:     event.EventStageStarted,
			Payload: map[string]string{
				"stage_type": stageType,
				"forced":     "true",
				"reason":     reason,
			},
		})
	}

	g.Log().Infof(ctx, "[StageService] ForceStartStage workflowRunID=%d stageType=%s stageRunID=%d", workflowRunID, stageType, stageRunID)
	return stageRunID, nil
}

func buildForceStageReason(stageType, reason string) string {
	message := "force restarted to " + stageType
	if reason == "" {
		return message
	}
	return message + ": " + reason
}

// completeWorkflow 完成整个工作流。
func (s *StageService) completeWorkflow(ctx context.Context, workflowRunID int64) error {
	now := gtime.Now()

	// 创建 complete stage
	completeStageID, err := s.StartStage(ctx, workflowRunID, consts.StageTypeComplete)
	if err != nil {
		return err
	}

	// 完成 complete stage（先持久化 finished_at，让后续统计能包含 complete 阶段本身）
	if err := s.CompleteStage(ctx, completeStageID); err != nil {
		return err
	}

	// 写入 finished_at（StartStage("complete") 已将 status 改为 completed，此处只补 finished_at）
	// 幂等条件：status=completed AND finished_at IS NULL，防止并发重复写入
	wfResult, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).
		Where("status", consts.WorkflowRunStatusCompleted).
		WhereNull("finished_at").
		Update(g.Map{
			"finished_at": now,
			"updated_at":  now,
		})
	if err != nil {
		return fmt.Errorf("写入 workflow_run finished_at 失败: %w", err)
	}
	wfRows, _ := wfResult.RowsAffected()
	if wfRows == 0 {
		// finished_at 已被写入或 status 非 completed（被并发取消等），短路退出
		g.Log().Warningf(ctx, "[StageService] completeWorkflow CAS 失败，workflow_run(%d) 已完成或被取消", workflowRunID)
		s.workflowSvc.runtimeMgr.Cancel(workflowRunID)
		return nil
	}

	// 更新项目状态
	projectID, pidErr := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).Value("project_id")
	if pidErr != nil {
		g.Log().Warningf(ctx, "[StageService] completeWorkflow 查询 project_id 失败: wfRun=%d err=%v", workflowRunID, pidErr)
	}
	if projectID.Int64() > 0 {
		if _, upErr := g.DB().Model("mvp_project").Ctx(ctx).
			Where("id", projectID.Int64()).
			Update(g.Map{"status": "completed", "updated_at": now}); upErr != nil {
			g.Log().Errorf(ctx, "[StageService] completeWorkflow 更新项目状态失败: projectID=%d err=%v", projectID.Int64(), upErr)
		}
	}

	// 执行收尾逻辑：指标统计 + 总结生成
	// 放在所有持久化之后，基于 DB 中真实的 finished_at 统计，避免口径偏差
	if completeStageSvc != nil {
		if fErr := completeStageSvc.Finalize(ctx, completeStageID, workflowRunID); fErr != nil {
			g.Log().Warningf(ctx, "[StageService] Complete Finalize 失败（不阻塞）: %v", fErr)
		}
	}

	// 发布 workflow.completed 事件
	if s.workflowSvc.publisher != nil {
		s.workflowSvc.publisher.Emit(ctx, event.Event{
			WorkflowRunID: workflowRunID,
			EntityType:    event.EntityWorkflowRun,
			EntityID:      &workflowRunID,
			EventType:     event.EventWorkflowCompleted,
		})
	}

	s.workflowSvc.runtimeMgr.Cancel(workflowRunID)

	g.Log().Infof(ctx, "[StageService] completeWorkflow workflowRunID=%d", workflowRunID)
	return nil
}
