// Package orchestrator 驱动工作流阶段切换，是新内核的总协调器。
package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/app/mvp/internal/workflow/runtime"
	"easymvp/utility/snowflake"
)

// WorkflowService 工作流编排服务。
type WorkflowService struct {
	runtimeMgr *runtime.Manager
	publisher  *event.Publisher
	wfRepo     *repo.WorkflowRunRepo
	stageRepo  *repo.StageRunRepo
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
// 返回 workflowRunID。
func (s *WorkflowService) CreateRun(ctx context.Context, projectID int64) (int64, error) {
	now := time.Now()

	// 1. 获取下一个 run_no
	runNo, err := s.wfRepo.NextRunNo(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("获取 run_no 失败: %w", err)
	}

	// 2. 创建 workflow_run
	wfRunID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_workflow_run").Ctx(ctx).Insert(g.Map{
		"id":            wfRunID,
		"project_id":    projectID,
		"run_no":        runNo,
		"status":        consts.WorkflowRunStatusDesigning,
		"current_stage": consts.StageTypeDesign,
		"created_at":    now,
		"updated_at":    now,
	})
	if err != nil {
		return 0, fmt.Errorf("创建 workflow_run 失败: %w", err)
	}

	// 3. 创建 design stage_run
	stageRunID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_stage_run").Ctx(ctx).Insert(g.Map{
		"id":              stageRunID,
		"workflow_run_id": wfRunID,
		"stage_type":      consts.StageTypeDesign,
		"stage_no":        1,
		"status":          consts.StageStatusPending,
		"created_at":      now,
		"updated_at":      now,
	})
	if err != nil {
		return 0, fmt.Errorf("创建 design stage_run 失败: %w", err)
	}

	// 4. 回写 current_stage_run_id
	_, err = g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", wfRunID).
		Update(g.Map{"current_stage_run_id": stageRunID, "updated_at": now})
	if err != nil {
		return 0, fmt.Errorf("更新 current_stage_run_id 失败: %w", err)
	}

	// 5. 发布事件
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

	// workflow_run 已在 CreateRun 中设为 designing，补 started_at
	_, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).
		Where("status", consts.WorkflowRunStatusDesigning).
		Update(g.Map{
			"started_at": now,
			"updated_at": now,
		})
	if err != nil {
		return fmt.Errorf("启动 workflow_run 失败: %w", err)
	}

	// design stage_run: pending → running
	_, err = g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("stage_type", consts.StageTypeDesign).
		Where("status", consts.StageStatusPending).
		Update(g.Map{
			"status":     consts.StageStatusRunning,
			"started_at": now,
			"updated_at": now,
		})
	if err != nil {
		return fmt.Errorf("启动 design stage_run 失败: %w", err)
	}

	// 创建运行时
	s.runtimeMgr.Create(workflowRunID, 0)

	g.Log().Infof(ctx, "[WorkflowService] StartDesign workflowRunID=%d", workflowRunID)
	return nil
}

// Pause 暂停工作流（从任何活跃阶段状态暂停）。
func (s *WorkflowService) Pause(ctx context.Context, workflowRunID int64, reason string) error {
	now := gtime.Now()

	// 查当前状态
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).WhereNull("deleted_at").One()
	if err != nil || wfRun.IsEmpty() {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}

	currentStatus := wfRun["status"].String()
	activeStatuses := map[string]bool{
		consts.WorkflowRunStatusDesigning: true,
		consts.WorkflowRunStatusReviewing: true,
		consts.WorkflowRunStatusExecuting: true,
		consts.WorkflowRunStatusReworking: true,
	}
	if !activeStatuses[currentStatus] {
		return fmt.Errorf("工作流状态(%s)不允许暂停", currentStatus)
	}

	rows, err := s.wfRepo.UpdateStatus(ctx, workflowRunID, currentStatus, consts.WorkflowRunStatusPaused, g.Map{
		"pause_reason":      reason,
		"status_before_pause": currentStatus,
		"updated_at":        now,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("工作流状态并发冲突，暂停失败")
	}
	s.runtimeMgr.Cancel(workflowRunID)
	return nil
}

// Resume 恢复工作流（恢复到暂停前的阶段状态）。
func (s *WorkflowService) Resume(ctx context.Context, workflowRunID int64) error {
	now := gtime.Now()

	// 查暂停前状态
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).Where("status", consts.WorkflowRunStatusPaused).
		WhereNull("deleted_at").One()
	if err != nil || wfRun.IsEmpty() {
		return fmt.Errorf("工作流不在暂停状态，无法恢复")
	}

	// 恢复到暂停前的状态，若无记录则根据 current_stage 推断
	resumeStatus := wfRun["status_before_pause"].String()
	if resumeStatus == "" {
		currentStage := wfRun["current_stage"].String()
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
	return nil
}

// Cancel 取消工作流（从任何非终态都可取消）。
func (s *WorkflowService) Cancel(ctx context.Context, workflowRunID int64, reason string) error {
	now := gtime.Now()

	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).WhereNull("deleted_at").One()
	if err != nil || wfRun.IsEmpty() {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}

	currentStatus := wfRun["status"].String()
	terminalStatuses := map[string]bool{
		consts.WorkflowRunStatusCompleted: true,
		consts.WorkflowRunStatusCanceled:  true,
	}
	if terminalStatuses[currentStatus] {
		return fmt.Errorf("工作流已处于终态(%s)，不可取消", currentStatus)
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
	s.runtimeMgr.Cancel(workflowRunID)
	return nil
}
