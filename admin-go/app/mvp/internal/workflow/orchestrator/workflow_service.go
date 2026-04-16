// Package orchestrator 驱动工作流阶段切换，是新内核的总协调器。
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
	"easymvp/app/mvp/internal/workflow/runtime"
	"easymvp/utility/snowflake"
)

// WorkflowPausedCallback 工作流暂停后的清理回调（停调度器等）。
type WorkflowPausedCallback func(ctx context.Context, workflowRunID int64)

// WorkflowResumedCallback 工作流恢复后的回调（重启调度器等）。
type WorkflowResumedCallback func(ctx context.Context, workflowRunID int64, resumeStatus string)

// WorkflowCanceledCallback 工作流取消后的回调（停调度器等）。
type WorkflowCanceledCallback func(ctx context.Context, workflowRunID int64)

// WorkflowService 工作流编排服务。
type WorkflowService struct {
	runtimeMgr         *runtime.Manager
	publisher          *event.Publisher
	wfRepo             *repo.WorkflowRunRepo
	stageRepo          *repo.StageRunRepo
	onWorkflowPaused   WorkflowPausedCallback
	onWorkflowResumed  WorkflowResumedCallback
	onWorkflowCanceled WorkflowCanceledCallback
}

// SetWorkflowPausedCallback 注册工作流暂停后的清理回调。
func (s *WorkflowService) SetWorkflowPausedCallback(fn WorkflowPausedCallback) {
	s.onWorkflowPaused = fn
}

// SetWorkflowResumedCallback 注册工作流恢复后的回调（重启调度器等）。
func (s *WorkflowService) SetWorkflowResumedCallback(fn WorkflowResumedCallback) {
	s.onWorkflowResumed = fn
}

// SetWorkflowCanceledCallback 注册工作流取消后的清理回调（停调度器等）。
func (s *WorkflowService) SetWorkflowCanceledCallback(fn WorkflowCanceledCallback) {
	s.onWorkflowCanceled = fn
}

// NewWorkflowService 创建工作流服务。
func NewWorkflowService(rtMgr *runtime.Manager, pub *event.Publisher, wfRepo *repo.WorkflowRunRepo, stageRepo *repo.StageRunRepo) *WorkflowService {
	return &WorkflowService{
		runtimeMgr: rtMgr,
		publisher:  pub,
		wfRepo:     wfRepo,
		stageRepo:  stageRepo,
	}
}

// CreateRun 为项目创建新的工作流运行实例 + design 阶段。
// 整个操作在同一事务中完成，保证 workflow_run + stage_run + 回填的原子性。
// 返回 workflowRunID。
func (s *WorkflowService) CreateRun(ctx context.Context, projectID int64) (int64, error) {
	now := time.Now()
	wfRunID := int64(snowflake.Generate())
	stageRunID := int64(snowflake.Generate())

	err := repo.WithTx(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 获取下一个 run_no
		runNo, err := s.wfRepo.NextRunNoInTx(ctx, tx, projectID)
		if err != nil {
			return fmt.Errorf("获取 run_no 失败: %w", err)
		}

		// 2. 获取项目归属字段
		scope := repo.GetProjectScopeByProjectInTx(ctx, tx, projectID)

		// 3. 创建 workflow_run
		err = s.wfRepo.CreateInTx(ctx, tx, g.Map{
			"id":            wfRunID,
			"project_id":    projectID,
			"run_no":        runNo,
			"status":        consts.WorkflowRunStatusDesigning,
			"current_stage": consts.StageTypeDesign,
			"created_by":    scope.CreatedBy,
			"dept_id":       scope.DeptID,
			"created_at":    now,
			"updated_at":    now,
		})
		if err != nil {
			return fmt.Errorf("创建 workflow_run 失败: %w", err)
		}

		// 4. 创建 design stage_run
		err = s.stageRepo.CreateInTx(ctx, tx, g.Map{
			"id":              stageRunID,
			"workflow_run_id": wfRunID,
			"stage_type":      consts.StageTypeDesign,
			"stage_no":        1,
			"status":          consts.StageStatusPending,
			"created_by":      scope.CreatedBy,
			"dept_id":         scope.DeptID,
			"created_at":      now,
			"updated_at":      now,
		})
		if err != nil {
			return fmt.Errorf("创建 design stage_run 失败: %w", err)
		}

		// 4. 回写 current_stage_run_id
		if err := s.wfRepo.UpdateInTx(ctx, tx, wfRunID, g.Map{
			"current_stage_run_id": stageRunID,
			"updated_at":           now,
		}); err != nil {
			return fmt.Errorf("更新 current_stage_run_id 失败: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	// 事务外发布事件（避免事务内做 I/O 拉长事务）
	if s.publisher != nil {
		s.publisher.Emit(ctx, event.Event{
			WorkflowRunID: wfRunID,
			EntityType:    event.EntityWorkflowRun,
			EntityID:      &wfRunID,
			EventType:     event.EventWorkflowCreated,
		})
	}

	g.Log().Infof(ctx, "[WorkflowService] CreateRun projectID=%d wfRunID=%d stageRunID=%d", projectID, wfRunID, stageRunID)
	return wfRunID, nil
}

// StartDesign 启动设计阶段（workflow 已是 designing，启动 stage_run）。
func (s *WorkflowService) StartDesign(ctx context.Context, workflowRunID int64) error {
	now := gtime.Now()
	projectRepo := repo.NewProjectRepo()

	// 查 project_id（runtime 需要真实 projectID，不能传 0）
	wfRun, err := s.wfRepo.GetByIDMap(ctx, workflowRunID, "project_id")
	if err != nil {
		return fmt.Errorf("查询 workflow_run(%d) 的 project_id 失败: %v", workflowRunID, err)
	}
	projectID := g.NewVar(wfRun["project_id"]).Int64()
	if projectID == 0 {
		return fmt.Errorf("查询 workflow_run(%d) 的 project_id 失败: project_id 为空", workflowRunID)
	}

	// workflow_run 已在 CreateRun 中设为 designing，补 started_at（CAS 校验）
	wfRows, err := s.wfRepo.UpdateFieldsIfStatuses(ctx, workflowRunID, []string{consts.WorkflowRunStatusDesigning}, g.Map{
		"started_at": now,
		"updated_at": now,
	})
	if err != nil {
		return fmt.Errorf("启动 workflow_run 失败: %w", err)
	}
	if wfRows == 0 {
		return fmt.Errorf("workflow_run(%d) 不在 designing 状态，无法启动设计阶段", workflowRunID)
	}

	// design stage_run: pending → running（CAS 校验）
	stageRun, stageErr := s.stageRepo.GetLatestByWorkflowTypeStatuses(ctx, workflowRunID, consts.StageTypeDesign, []string{consts.StageStatusPending}, "id")
	if stageErr != nil || stageRun == nil {
		return fmt.Errorf("workflow_run(%d) 的 design stage_run 不在 pending 状态", workflowRunID)
	}
	stageRows, err := s.stageRepo.UpdateFieldsIfStatus(ctx, stageRun["id"].Int64(), consts.StageStatusPending, g.Map{
		"status":     consts.StageStatusRunning,
		"started_at": now,
		"updated_at": now,
	})
	if err != nil {
		return fmt.Errorf("启动 design stage_run 失败: %w", err)
	}
	if stageRows == 0 {
		return fmt.Errorf("workflow_run(%d) 的 design stage_run 不在 pending 状态", workflowRunID)
	}

	if syncErr := projectRepo.UpdateFields(ctx, projectID, g.Map{
		"status":     consts.WorkflowRunStatusDesigning,
		"updated_at": now,
	}); syncErr != nil {
		g.Log().Warningf(ctx, "[WorkflowService] StartDesign 同步 project status 失败: projectID=%d err=%v", projectID, syncErr)
	}

	// 创建运行时（传真实 projectID，与 Resume 行为一致）
	s.runtimeMgr.Create(workflowRunID, projectID)

	g.Log().Infof(ctx, "[WorkflowService] StartDesign workflowRunID=%d projectID=%d", workflowRunID, projectID)
	return nil
}

// Pause 暂停工作流（从任何活跃阶段状态暂停）。
func (s *WorkflowService) Pause(ctx context.Context, workflowRunID int64, reason string) error {
	now := gtime.Now()
	projectRepo := repo.NewProjectRepo()

	// 查当前状态
	wfRun, err := s.wfRepo.GetByIDMap(ctx, workflowRunID, "status", "project_id")
	if err != nil || len(wfRun) == 0 {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}

	currentStatus := g.NewVar(wfRun["status"]).String()
	activeStatuses := map[string]bool{
		consts.WorkflowRunStatusDesigning: true,
		consts.WorkflowRunStatusReviewing: true,
		consts.WorkflowRunStatusExecuting: true,
		consts.WorkflowRunStatusAccepting: true,
		consts.WorkflowRunStatusReworking: true,
	}
	if !activeStatuses[currentStatus] {
		return fmt.Errorf("工作流状态(%s)不允许暂停", currentStatus)
	}

	rows, err := s.wfRepo.UpdateStatus(ctx, workflowRunID, currentStatus, consts.WorkflowRunStatusPaused, g.Map{
		"pause_reason":        reason,
		"status_before_pause": currentStatus,
		"updated_at":          now,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("工作流状态并发冲突，暂停失败")
	}
	s.runtimeMgr.Cancel(workflowRunID)

	// 停调度器 + 释放资源锁
	if s.onWorkflowPaused != nil {
		s.onWorkflowPaused(ctx, workflowRunID)
	}

	// 同步 mvp_project.status
	projectID := g.NewVar(wfRun["project_id"]).Int64()
	if projectID > 0 {
		if syncErr := projectRepo.UpdateFields(ctx, projectID, g.Map{
			"status":       consts.WorkflowRunStatusPaused,
			"pause_reason": reason,
			"updated_at":   now,
		}); syncErr != nil {
			g.Log().Errorf(ctx, "[WorkflowService] Pause 同步 project status 失败: projectID=%d err=%v", projectID, syncErr)
		}
	}

	if s.publisher != nil {
		s.publisher.Emit(ctx, event.Event{
			WorkflowRunID: workflowRunID,
			EntityType:    event.EntityWorkflowRun,
			EntityID:      &workflowRunID,
			EventType:     event.EventWorkflowPaused,
			Payload:       map[string]string{"reason": reason, "from_status": currentStatus},
		})
	}
	return nil
}

// Resume 恢复工作流（恢复到暂停前的阶段状态）。
func (s *WorkflowService) Resume(ctx context.Context, workflowRunID int64) error {
	now := gtime.Now()
	projectRepo := repo.NewProjectRepo()

	// 查暂停前状态
	wfRun, err := s.wfRepo.GetByIDMap(ctx, workflowRunID, "status", "status_before_pause", "current_stage", "project_id")
	if err != nil || len(wfRun) == 0 {
		return fmt.Errorf("工作流不在暂停状态，无法恢复")
	}
	if g.NewVar(wfRun["status"]).String() != consts.WorkflowRunStatusPaused {
		return fmt.Errorf("工作流不在暂停状态，无法恢复")
	}

	// 恢复到暂停前的状态，若无记录则根据 current_stage 推断
	resumeStatus := g.NewVar(wfRun["status_before_pause"]).String()
	if resumeStatus == "" {
		currentStage := g.NewVar(wfRun["current_stage"]).String()
		resumeStatus = StageTypeToWorkflowStatus(currentStage)
	}
	if resumeStatus == "" {
		resumeStatus = consts.WorkflowRunStatusDesigning
	}

	rows, err := s.wfRepo.UpdateStatus(ctx, workflowRunID, consts.WorkflowRunStatusPaused, resumeStatus, g.Map{
		"pause_reason":        nil,
		"status_before_pause": nil,
		"updated_at":          now,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("工作流恢复失败（并发冲突）")
	}

	// 重建 runtime context（Pause 时已 Cancel）
	projectID := g.NewVar(wfRun["project_id"]).Int64()
	s.runtimeMgr.Create(workflowRunID, projectID)

	// 同步 mvp_project.status
	if projectID > 0 {
		if syncErr := projectRepo.UpdateFields(ctx, projectID, g.Map{
			"status":       resumeStatus,
			"pause_reason": nil,
			"updated_at":   now,
		}); syncErr != nil {
			g.Log().Errorf(ctx, "[WorkflowService] Resume 同步 project status 失败: projectID=%d err=%v", projectID, syncErr)
		}
	}

	// 恢复后重启调度器（execute/rework 阶段需要调度器推进任务）
	if s.onWorkflowResumed != nil {
		s.onWorkflowResumed(ctx, workflowRunID, resumeStatus)
	}

	if s.publisher != nil {
		s.publisher.Emit(ctx, event.Event{
			WorkflowRunID: workflowRunID,
			EntityType:    event.EntityWorkflowRun,
			EntityID:      &workflowRunID,
			EventType:     event.EventWorkflowResumed,
			Payload:       map[string]string{"resume_status": resumeStatus},
		})
	}
	return nil
}

// Cancel 取消工作流（从任何非终态都可取消）。
func (s *WorkflowService) Cancel(ctx context.Context, workflowRunID int64, reason string) error {
	now := gtime.Now()
	var (
		domainTaskRepo = repo.NewDomainTaskRepo()
		projectRepo    = repo.NewProjectRepo()
		stageTaskRepo  = repo.NewStageTaskRepo()
	)

	wfRun, err := s.wfRepo.GetByIDMap(ctx, workflowRunID, "status", "project_id")
	if err != nil || len(wfRun) == 0 {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}

	currentStatus := g.NewVar(wfRun["status"]).String()
	terminalStatuses := map[string]bool{
		consts.WorkflowRunStatusCompleted: true,
		consts.WorkflowRunStatusCanceled:  true,
	}
	if terminalStatuses[currentStatus] {
		return fmt.Errorf("工作流已处于终态(%s)，不可取消", currentStatus)
	}

	cancelReason := buildWorkflowCancelReason(reason)

	activeStageRuns, stageRunErr := s.stageRepo.ListByWorkflowStatuses(ctx, workflowRunID, []string{
		consts.StageStatusPending,
		consts.StageStatusRunning,
	}, "id")
	if stageRunErr != nil {
		g.Log().Warningf(ctx, "[WorkflowService] Cancel 查询活跃 stage_run 失败: workflowRunID=%d err=%v", workflowRunID, stageRunErr)
	}
	activeStageRunIDs := make([]int64, 0, len(activeStageRuns))
	for _, item := range activeStageRuns {
		activeStageRunIDs = append(activeStageRunIDs, g.NewVar(item["id"]).Int64())
	}

	runningTasks, runningTaskErr := domainTaskRepo.ListByWorkflowAndStatuses(ctx, workflowRunID, []string{consts.TaskStatusRunning}, "id")
	if runningTaskErr != nil {
		g.Log().Warningf(ctx, "[WorkflowService] Cancel 查询运行中任务失败: workflowRunID=%d err=%v", workflowRunID, runningTaskErr)
	}
	runningTaskIDs := make([]int64, 0, len(runningTasks))
	for _, item := range runningTasks {
		runningTaskIDs = append(runningTaskIDs, g.NewVar(item["id"]).Int64())
	}

	rows, err := s.wfRepo.UpdateStatus(ctx, workflowRunID, currentStatus, consts.WorkflowRunStatusCanceled, g.Map{
		"cancel_reason": reason,
		"finished_at":   now,
		"updated_at":    now,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("工作流取消失败（并发冲突）")
	}

	if len(activeStageRunIDs) > 0 {
		if upErr := s.stageRepo.UpdateByIDs(ctx, activeStageRunIDs, g.Map{
			"status":        consts.StageStatusFailed,
			"error_message": cancelReason,
			"finished_at":   now,
			"updated_at":    now,
		}); upErr != nil {
			g.Log().Warningf(ctx, "[WorkflowService] Cancel 更新 stage_run 终态失败: workflowRunID=%d err=%v", workflowRunID, upErr)
		}

		if upErr := stageTaskRepo.UpdateByStageRunsStatuses(ctx, activeStageRunIDs, []string{
			consts.StageStatusPending,
			consts.StageStatusRunning,
		}, g.Map{
			"status":        consts.StageStatusFailed,
			"error_message": cancelReason,
			"completed_at":  now,
			"updated_at":    now,
		}); upErr != nil {
			g.Log().Warningf(ctx, "[WorkflowService] Cancel 更新 stage_task 终态失败: workflowRunID=%d err=%v", workflowRunID, upErr)
		}
	}

	projectID := g.NewVar(wfRun["project_id"]).Int64()
	if projectID > 0 {
		if syncErr := projectRepo.UpdateFields(ctx, projectID, g.Map{
			"status":       consts.WorkflowRunStatusCanceled,
			"pause_reason": nil,
			"updated_at":   now,
		}); syncErr != nil {
			g.Log().Errorf(ctx, "[WorkflowService] Cancel 同步 project status 失败: projectID=%d err=%v", projectID, syncErr)
		}
	}

	if s.onWorkflowCanceled != nil {
		s.onWorkflowCanceled(ctx, workflowRunID)
	}

	if len(runningTaskIDs) > 0 {
		if upErr := domainTaskRepo.UpdateByIDsStatuses(ctx, runningTaskIDs, []string{
			consts.TaskStatusPending,
			consts.TaskStatusRunning,
		}, g.Map{
			"status":           consts.TaskStatusFailed,
			"result":           cancelReason,
			"error_message":    cancelReason,
			"completed_at":     now,
			"heartbeat_at":     nil,
			"locked_resources": nil,
			"updated_at":       now,
		}); upErr != nil {
			g.Log().Warningf(ctx, "[WorkflowService] Cancel 更新运行中任务终态失败: workflowRunID=%d err=%v", workflowRunID, upErr)
		}
	}

	s.runtimeMgr.Cancel(workflowRunID)

	if s.publisher != nil {
		s.publisher.Emit(ctx, event.Event{
			WorkflowRunID: workflowRunID,
			EntityType:    event.EntityWorkflowRun,
			EntityID:      &workflowRunID,
			EventType:     event.EventWorkflowCanceled,
			Payload:       map[string]string{"reason": reason},
		})
	}
	return nil
}

func buildWorkflowCancelReason(reason string) string {
	if reason == "" {
		return "workflow canceled"
	}
	return "workflow canceled: " + reason
}
