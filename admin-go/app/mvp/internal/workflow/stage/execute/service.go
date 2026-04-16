// Package execute 管理执行阶段：蓝图实例化 → domain_task → 调度执行。
package execute

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
	domainTask "easymvp/app/mvp/internal/workflow/domain/task"
	"easymvp/app/mvp/internal/workflow/executor"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/app/mvp/internal/workflow/scheduler"
	"easymvp/app/mvp/internal/workspace"
)

// StageCompleter 阶段完成回调（避免循环依赖）。
type StageCompleter interface {
	CompleteStage(ctx context.Context, stageRunID int64) error
	FailStage(ctx context.Context, stageRunID int64, reason string) error
	TransitionNext(ctx context.Context, workflowRunID int64) error
}

// Service 执行阶段服务。
type Service struct {
	taskSvc             *domainTask.TaskService
	scheduler           *scheduler.DomainTaskScheduler
	stageCompleter      StageCompleter
	executorRegistry    *executor.Registry
	onAnalysisCompleted AnalysisCompletedFn
	stageRunRepo        *repo.StageRunRepo
	workflowRunRepo     *repo.WorkflowRunRepo
	projectRepo         *repo.ProjectRepo
	taskRepo            *repo.DomainTaskRepo
}

// NewService 创建执行阶段服务。
func NewService(ts *domainTask.TaskService, sched *scheduler.DomainTaskScheduler, sc StageCompleter, reg *executor.Registry) *Service {
	return &Service{
		taskSvc:          ts,
		scheduler:        sched,
		stageCompleter:   sc,
		executorRegistry: reg,
		stageRunRepo:     repo.NewStageRunRepo(),
		workflowRunRepo:  repo.NewWorkflowRunRepo(),
		projectRepo:      repo.NewProjectRepo(),
		taskRepo:         repo.NewDomainTaskRepo(),
	}
}

// SetAnalysisCompletedFn 注册 failure_analysis 完成后的回调。
func (s *Service) SetAnalysisCompletedFn(fn AnalysisCompletedFn) { s.onAnalysisCompleted = fn }

// BindExecutor 为已存在的 workflow 重新绑定执行器。
func (s *Service) BindExecutor(workflowRunID int64) {
	s.scheduler.SetExecutor(s.newDomainTaskExecutor(workflowRunID))
}

// BindExecuteStage 为 execute 阶段重新绑定执行器和完成回调。
func (s *Service) BindExecuteStage(stageRunID, workflowRunID int64) {
	s.BindExecutor(workflowRunID)
	s.scheduler.SetCompletionCallback(s.newCompletionCallback(stageRunID))
}

// InstantiateAndStart 将审核通过的蓝图实例化为领域任务并启动调度。
func (s *Service) InstantiateAndStart(ctx context.Context, stageRunID int64, planVersionID int64) error {
	// 获取 workflow_run_id
	stageRun, err := s.stageRunRepo.GetByIDMap(ctx, stageRunID, "workflow_run_id")
	if err != nil || len(stageRun) == 0 {
		return fmt.Errorf("stage_run(%d) 不存在", stageRunID)
	}
	workflowRunID := g.NewVar(stageRun["workflow_run_id"]).Int64()

	// 1. 实例化蓝图为领域任务
	taskCount, err := s.taskSvc.InstantiateFromBlueprint(ctx, planVersionID, stageRunID, workflowRunID)
	if err != nil {
		return fmt.Errorf("蓝图实例化失败: %w", err)
	}
	if taskCount == 0 {
		return fmt.Errorf("没有生成任何领域任务")
	}

	g.Log().Infof(ctx, "[ExecuteStage] 实例化 %d 个领域任务, stageRunID=%d, planVersionID=%d", taskCount, stageRunID, planVersionID)

	// 2. 注册执行器和 execute 完成回调
	s.BindExecuteStage(stageRunID, workflowRunID)

	// 3. 启动调度
	return s.scheduler.Start(context.Background(), workflowRunID)
}

// AnalysisCompletedFn 分析任务完成回调（路由到 rework service）。
type AnalysisCompletedFn func(ctx context.Context, stageRunID, analysisTaskID int64) error

// domainTaskExecutor 领域任务执行器，通过 executor.Registry 分发到具体执行器实现。
type domainTaskExecutor struct {
	workflowRunID       int64
	scheduler           *scheduler.DomainTaskScheduler
	wsMgr               workspace.Manager
	registry            *executor.Registry
	onAnalysisCompleted AnalysisCompletedFn
	workflowRunRepo     *repo.WorkflowRunRepo
	projectRepo         *repo.ProjectRepo
	taskRepo            *repo.DomainTaskRepo
}

func (s *Service) newDomainTaskExecutor(workflowRunID int64) *domainTaskExecutor {
	return &domainTaskExecutor{
		workflowRunID:       workflowRunID,
		scheduler:           s.scheduler,
		wsMgr:               workspace.NewGitWorktreeManager(),
		registry:            s.executorRegistry,
		onAnalysisCompleted: s.onAnalysisCompleted,
		workflowRunRepo:     s.workflowRunRepo,
		projectRepo:         s.projectRepo,
		taskRepo:            s.taskRepo,
	}
}

func (s *Service) newCompletionCallback(stageRunID int64) scheduler.CompletionCallback {
	finalStageRunID := stageRunID
	cleanupMgr := workspace.NewGitWorktreeManager()
	return func(ctx context.Context, wfRunID int64) {
		g.Log().Infof(ctx, "[ExecuteStage] 所有任务完成, workflowRunID=%d", wfRunID)

		// 压缩上下文
		workflowRun, pidErr := s.workflowRunRepo.GetByIDMap(ctx, wfRunID, "project_id")
		if pidErr != nil {
			g.Log().Warningf(ctx, "[ExecuteStage] 查询 project_id 失败: wfRunID=%d err=%v", wfRunID, pidErr)
		}
		projectID := g.NewVar(workflowRun["project_id"]).Int64()
		if projectID > 0 {
			_ = engine.GetCompressor().CompressProjectContext(ctx, projectID)
		}

		// 延时清理：扫描已超过保留期的 workspace（失败/取消态的 worktree 依赖此机制）。
		// 注意：刚完成的 workspace 不会被清理（尚在保留期内），这是预期行为。
		go func() {
			defer func() {
				if r := recover(); r != nil {
					g.Log().Errorf(context.Background(), "[ExecuteStage] RunCleanup panic: %v", r)
				}
			}()
			workspace.RunCleanup(context.Background(), cleanupMgr, workspace.DefaultCleanupConfig())
		}()

		// 完成 execute stage 并推进工作流到下一阶段（complete）
		if s.stageCompleter != nil {
			if err := s.reconcileCompletionTransition(ctx, wfRunID, finalStageRunID); err != nil {
				g.Log().Errorf(ctx, "[ExecuteStage] reconcile completion failed: workflowRunID=%d stageRunID=%d err=%v",
					wfRunID, finalStageRunID, err)
			}
		}
	}
}

type completionTransitionPlan struct {
	completeStage bool
	transition    bool
}

func buildCompletionTransitionPlan(stageStatus, workflowStatus, workflowStage string, workflowStageRunID, stageRunID int64) completionTransitionPlan {
	if stageRunID == 0 || workflowStageRunID != stageRunID {
		return completionTransitionPlan{}
	}
	if workflowStatus != "executing" || workflowStage != "execute" {
		return completionTransitionPlan{}
	}

	switch stageStatus {
	case "running":
		return completionTransitionPlan{completeStage: true, transition: true}
	case "completed":
		return completionTransitionPlan{transition: true}
	default:
		return completionTransitionPlan{}
	}
}

func (s *Service) reconcileCompletionTransition(ctx context.Context, workflowRunID, stageRunID int64) error {
	if stageRunID == 0 || workflowRunID == 0 {
		return fmt.Errorf("workflowRunID/stageRunID 不能为空")
	}

	stageRun, err := s.stageRunRepo.GetByIDMap(ctx, stageRunID, "status")
	if err != nil {
		return fmt.Errorf("查询 execute stage_run 失败: %w", err)
	}
	if len(stageRun) == 0 {
		return fmt.Errorf("execute stage_run(%d) 不存在", stageRunID)
	}

	workflowRun, err := s.workflowRunRepo.GetByIDMap(ctx, workflowRunID, "status", "current_stage", "current_stage_run_id")
	if err != nil {
		return fmt.Errorf("查询 workflow_run 失败: %w", err)
	}
	if len(workflowRun) == 0 {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}

	plan := buildCompletionTransitionPlan(
		g.NewVar(stageRun["status"]).String(),
		g.NewVar(workflowRun["status"]).String(),
		g.NewVar(workflowRun["current_stage"]).String(),
		g.NewVar(workflowRun["current_stage_run_id"]).Int64(),
		stageRunID,
	)
	if !plan.completeStage && !plan.transition {
		g.Log().Infof(ctx, "[ExecuteStage] completion reconcile skipped: workflowRunID=%d stageRunID=%d stage=%s workflow=%s/%s currentStageRunID=%d",
			workflowRunID,
			stageRunID,
			g.NewVar(stageRun["status"]).String(),
			g.NewVar(workflowRun["status"]).String(),
			g.NewVar(workflowRun["current_stage"]).String(),
			g.NewVar(workflowRun["current_stage_run_id"]).Int64(),
		)
		return nil
	}

	if plan.completeStage {
		if err := s.stageCompleter.CompleteStage(ctx, stageRunID); err != nil {
			return fmt.Errorf("CompleteStage 失败: %w", err)
		}
	}
	if plan.transition {
		if err := s.stageCompleter.TransitionNext(ctx, workflowRunID); err != nil {
			return fmt.Errorf("TransitionNext 失败: %w", err)
		}
	}
	return nil
}

// ExecuteDomainTask 执行单个领域任务。
func (e *domainTaskExecutor) ExecuteDomainTask(ctx context.Context, workflowRunID, taskID int64) {
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[domainTaskExecutor] panic: task=%d, err=%v", taskID, r)
			e.scheduler.OnTaskFailed(ctx, taskID, fmt.Sprintf("panic: %v", r))
		}
	}()

	// 查任务详情
	taskRecord, err := e.taskRepo.GetRecordByID(ctx, taskID)
	if err != nil || taskRecord == nil || taskRecord.IsEmpty() {
		e.scheduler.OnTaskFailed(ctx, taskID, "任务不存在")
		return
	}

	// 查 project_id
	workflowRun, pidErr := e.workflowRunRepo.GetByIDMap(ctx, workflowRunID, "project_id")
	if pidErr != nil {
		e.handleFailure(ctx, taskID, fmt.Sprintf("查询 project_id 失败: %v", pidErr))
		return
	}
	projectID := g.NewVar(workflowRun["project_id"]).Int64()
	if projectID == 0 {
		e.handleFailure(ctx, taskID, "工作流缺少 project_id")
		return
	}

	roleType := g.NewVar(taskRecord["role_type"]).String()
	executionMode := g.NewVar(taskRecord["execution_mode"]).String()
	modelID := g.NewVar(taskRecord["model_id"]).Int64()

	// 从注册表获取执行器
	exec, err := e.registry.MustGet(executionMode)
	if err != nil {
		e.handleFailure(ctx, taskID, err.Error())
		return
	}

	// 获取模型信息
	modelInfo, err := engine.ResolveProjectModelInfo(ctx, projectID, roleType, g.NewVar(taskRecord["role_level"]).String(), modelID)
	if err != nil {
		e.handleFailure(ctx, taskID, err.Error())
		return
	}

	// 如果执行器需要工作空间隔离，准备 worktree
	var ws *workspace.TaskWorkspace
	if exec.NeedsWorkspace() && e.wsMgr != nil {
		project, projErr := e.projectRepo.GetByID(ctx, projectID, "work_dir")
		if projErr != nil {
			g.Log().Warningf(ctx, "[domainTaskExecutor] 查询项目失败: projectID=%d err=%v", projectID, projErr)
		}
		workDir := g.NewVar(project["work_dir"]).String()
		ws, err = e.wsMgr.Prepare(ctx, workspace.PrepareRequest{
			TaskID:        taskID,
			WorkflowRunID: workflowRunID,
			ProjectID:     projectID,
			WorkDir:       workDir,
		})
		if err != nil {
			e.handleFailure(ctx, taskID, fmt.Sprintf("workspace 隔离准备失败: %v", err))
			return
		}
	}

	// 启动心跳
	hbCtx, hbCancel := context.WithCancel(ctx)
	go touchHeartbeatLoop(hbCtx, e.taskRepo, taskID)
	defer hbCancel()

	// 统一调用执行器
	result := exec.Execute(ctx, &executor.Request{
		ProjectID:     projectID,
		WorkflowRunID: workflowRunID,
		TaskID:        taskID,
		TaskRecord:    taskRecord,
		ModelInfo:     modelInfo,
		Workspace:     ws,
	})

	if result.Success {
		e.handleSuccess(ctx, taskID, result.Output)
	} else {
		e.handleFailure(ctx, taskID, result.Error.Error())
	}
}

// handleSuccess 任务成功。
func (e *domainTaskExecutor) handleSuccess(ctx context.Context, taskID int64, result string) {
	now := gtime.Now()
	rows, err := e.taskRepo.UpdateFieldsIfStatus(ctx, taskID, domainTask.StatusRunning, g.Map{
		"status":       domainTask.StatusCompleted,
		"result":       result,
		"completed_at": now,
		"updated_at":   now,
	})
	if err != nil {
		g.Log().Errorf(ctx, "[domainTaskExecutor] handleSuccess 更新失败: task=%d err=%v", taskID, err)
		return
	}
	if rows == 0 {
		// 任务已被并发改状态（failed/canceled/escalated），不推进后续流程
		g.Log().Warningf(ctx, "[domainTaskExecutor] handleSuccess CAS 失败，任务已不在 running: task=%d", taskID)
		return
	}

	// 检查是否为 failure_analysis 任务 → 路由到 rework OnAnalysisCompleted
	if e.onAnalysisCompleted != nil {
		task, taskErr := e.taskRepo.GetByIDMap(ctx, taskID, "task_kind", "stage_run_id")
		if taskErr != nil {
			g.Log().Warningf(ctx, "[domainTaskExecutor] 查询任务类型失败: task=%d err=%v", taskID, taskErr)
		}
		if len(task) > 0 && g.NewVar(task["task_kind"]).String() == "failure_analysis" {
			stageRunID := g.NewVar(task["stage_run_id"]).Int64()
			if err := e.onAnalysisCompleted(ctx, stageRunID, taskID); err != nil {
				// 回调失败：回滚分析任务为 failed，让 watchdog 后续重试
				g.Log().Errorf(ctx, "[domainTaskExecutor] OnAnalysisCompleted 失败，回滚分析任务: task=%d err=%v", taskID, err)
				_, _ = e.taskRepo.UpdateFieldsIfStatus(ctx, taskID, domainTask.StatusCompleted, g.Map{
					"status":     domainTask.StatusFailed,
					"result":     fmt.Sprintf("rework 回调失败: %v", err),
					"updated_at": gtime.Now(),
				})
				e.scheduler.OnTaskFailed(ctx, taskID, "rework callback failed: "+err.Error())
				return
			}
		}
	}

	_ = e.scheduler.OnTaskCompleted(ctx, taskID)
}

// handleFailure 任务失败。
func (e *domainTaskExecutor) handleFailure(ctx context.Context, taskID int64, errMsg string) {
	if _, upErr := e.taskRepo.UpdateFieldsIfStatus(ctx, taskID, domainTask.StatusRunning, g.Map{
		"status":     domainTask.StatusFailed,
		"result":     errMsg,
		"updated_at": gtime.Now(),
	}); upErr != nil {
		g.Log().Errorf(ctx, "[domainTaskExecutor] handleFailure 更新状态失败: task=%d err=%v", taskID, upErr)
	}
	e.scheduler.OnTaskFailed(ctx, taskID, errMsg)
}

// touchHeartbeatLoop 定期更新 domain_task 的 heartbeat_at。
// 启动时立即写一次，之后每 30s 更新。
func touchHeartbeatLoop(ctx context.Context, taskRepo *repo.DomainTaskRepo, taskID int64) {
	// 立即写一次心跳，消除启动空窗期
	if taskRepo != nil {
		_, _ = taskRepo.UpdateFieldsIfStatus(ctx, taskID, domainTask.StatusRunning, g.Map{"heartbeat_at": gtime.Now()})
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if taskRepo != nil {
				_, _ = taskRepo.UpdateFieldsIfStatus(ctx, taskID, domainTask.StatusRunning, g.Map{"heartbeat_at": gtime.Now()})
			}
		}
	}
}
