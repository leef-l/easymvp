package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
	executorPkg "easymvp/app/mvp/internal/workflow/executor"
	"easymvp/app/mvp/internal/workspace"
	"easymvp/app/mvp/internal/workflow/domain/plan"
	domainTask "easymvp/app/mvp/internal/workflow/domain/task"
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/app/mvp/internal/workflow/runtime"
	"easymvp/app/mvp/internal/workflow/scheduler"
	completeStage "easymvp/app/mvp/internal/workflow/stage/complete"
	executeStage "easymvp/app/mvp/internal/workflow/stage/execute"
	reworkStage "easymvp/app/mvp/internal/workflow/stage/rework"
	reviewStage "easymvp/app/mvp/internal/workflow/stage/review"
	watchdogV2 "easymvp/app/mvp/internal/workflow/watchdog"
)

var (
	once            sync.Once
	workflowSvc     *WorkflowService
	stageSvc        *StageService
	planVersionSvc  *plan.PlanVersionService
	reviewStageSvc  *reviewStage.Service
	executeStageSvc *executeStage.Service
	taskScheduler   *scheduler.DomainTaskScheduler
	reworkStageSvc    *reworkStage.Service
	completeStageSvc  *completeStage.Service
	domainWatchdog    *watchdogV2.DomainTaskWatchdog
	runtimeMgr      *runtime.Manager
	eventBus        *event.Bus
	eventPublisher  *event.Publisher
)

// Init 初始化所有工作流服务单例。在应用启动时调用一次。
func Init() {
	once.Do(func() {
		// 基础设施
		runtimeMgr = runtime.NewManager()
		eventBus = event.NewBus()
		eventPublisher = event.NewPublisher(eventBus)

		// 仓储
		wfRepo := repo.NewWorkflowRunRepo()
		stageRepo := repo.NewStageRunRepo()
		planRepo := repo.NewPlanVersionRepo()
		bpRepo := repo.NewBlueprintRepo()

		// 服务
		workflowSvc = NewWorkflowService(runtimeMgr, eventPublisher, wfRepo, stageRepo)
		stageSvc = NewStageService(workflowSvc)
		planVersionSvc = plan.NewPlanVersionService(planRepo, bpRepo)

		// 执行器注册表
		execRegistry := executorPkg.NewRegistry()
		execRegistry.Register(executorPkg.NewAiderExecutor(workspace.NewGitWorktreeManager()))
		execRegistry.Register(executorPkg.NewChatExecutor())

		// 执行阶段服务
		taskRepo := repo.NewDomainTaskRepo()
		taskSvc := domainTask.NewTaskService(taskRepo)
		taskScheduler = scheduler.NewDomainTaskScheduler()
		executeStageSvc = executeStage.NewService(taskSvc, taskScheduler, stageSvc, execRegistry)

		// 审核阶段服务
		issueRepo := repo.NewReviewIssueRepo()
		reviewStageSvc = reviewStage.NewService(stageSvc, issueRepo)

		// 注册审核驳回时回退 design 阶段的回调
		reviewStageSvc.SetDesignRollbackFn(func(ctx context.Context, workflowRunID int64) error {
			_, err := stageSvc.TransitionTo(ctx, workflowRunID, "design")
			return err
		})

		// 注册审核通过后触发执行阶段的回调
		reviewStageSvc.SetExecuteTriggerFn(func(ctx context.Context, workflowRunID, planVersionID int64) error {
			return triggerExecuteStage(ctx, workflowRunID, planVersionID)
		})

		// 注册审核触发回调到 PlanVersionService（避免循环依赖）
		planVersionSvc.SetReviewTrigger(func(ctx context.Context, projectID, planVersionID int64) error {
			return triggerReviewStage(ctx, projectID, planVersionID)
		})

		// 注册 V2 蓝图创建回调到 engine 包（避免循环依赖）
		engine.RegisterBlueprintCreator(func(ctx context.Context, projectID, workflowRunID, conversationID, messageID int64, tasks []engine.ArchitectTask) (int64, int, error) {
			return planVersionSvc.CreateFromArchitectReply(ctx, projectID, workflowRunID, conversationID, messageID, tasks)
		})

		// 完成阶段服务
		completeStageSvc = completeStage.NewService()

		// 返工阶段服务
		reworkStageSvc = reworkStage.NewService()
		reworkStageSvc.SetStageCompleter(stageSvc)

		// 注册 rework 完成后推回 execute 的回调
		reworkStageSvc.SetExecuteTrigger(func(ctx context.Context, workflowRunID, planVersionID int64) error {
			g.Log().Infof(ctx, "[Registry] rework 完成，恢复执行状态: workflowRunID=%d", workflowRunID)

			// CAS: reworking → executing（CompleteStage(rework) 不改 workflow_run 状态，需显式恢复）
			rows, err := workflowSvc.wfRepo.UpdateStatus(ctx, workflowRunID,
				consts.WorkflowRunStatusReworking, consts.WorkflowRunStatusExecuting, g.Map{
					"current_stage": "execute",
					"updated_at":    gtime.Now(),
				})
			if err != nil {
				return fmt.Errorf("rework→executing 状态恢复失败: %w", err)
			}
			if rows == 0 {
				g.Log().Warningf(ctx, "[Registry] rework→executing CAS 失败，workflow_run(%d) 可能已被取消/暂停", workflowRunID)
				return nil
			}

			// 同步 project.status
			projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
				Where("id", workflowRunID).Value("project_id")
			if projectID.Int64() > 0 {
				_, _ = g.DB().Model("mvp_project").Ctx(ctx).
					Where("id", projectID.Int64()).
					Update(g.Map{"status": consts.WorkflowRunStatusExecuting, "updated_at": gtime.Now()})
			}

			// 重启调度器拾取 pending 任务
			return taskScheduler.Start(ctx, workflowRunID)
		})

		// 注册 failure_analysis 完成回调到 execute service
		executeStageSvc.SetAnalysisCompletedFn(func(ctx context.Context, stageRunID, analysisTaskID int64) error {
			return reworkStageSvc.OnAnalysisCompleted(ctx, stageRunID, analysisTaskID)
		})

		// 注册阶段失败后的执行链终止回调
		stageSvc.SetWorkflowFailedCallback(func(ctx context.Context, workflowRunID int64) {
			g.Log().Infof(ctx, "[Registry] 工作流失败，终止执行链: workflowRunID=%d", workflowRunID)
			taskScheduler.Stop(workflowRunID)
			runtimeMgr.Cancel(workflowRunID)
		})

		// Watchdog V2: 监控 domain_task 心跳
		domainWatchdog = watchdogV2.New()
		domainWatchdog.SetScheduler(taskScheduler)
		domainWatchdog.SetRetryFn(func(ctx context.Context, taskID int64) error {
			// 重试后唤醒一次调度扫描（不重建调度循环）
			wfRunID, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).Where("id", taskID).Value("workflow_run_id")
			if wfRunID.Int64() > 0 {
				taskScheduler.Wakeup(ctx, wfRunID.Int64())
			}
			return nil
		})
		domainWatchdog.SetEscalateFn(func(ctx context.Context, workflowRunID, taskID int64) error {
			return triggerReworkStage(ctx, workflowRunID, taskID)
		})
		domainWatchdog.Start(context.Background())
	})
}

// GetWorkflowService 获取工作流服务单例。
func GetWorkflowService() *WorkflowService {
	Init()
	return workflowSvc
}

// GetStageService 获取阶段服务单例。
func GetStageService() *StageService {
	Init()
	return stageSvc
}

// GetPlanVersionService 获取计划版本服务单例。
func GetPlanVersionService() *plan.PlanVersionService {
	Init()
	return planVersionSvc
}

// GetRuntimeManager 获取运行时管理器。
func GetRuntimeManager() *runtime.Manager {
	Init()
	return runtimeMgr
}

// GetEventBus 获取事件总线。
func GetEventBus() *event.Bus {
	Init()
	return eventBus
}

// GetEventPublisher 获取事件发布器。
func GetEventPublisher() *event.Publisher {
	Init()
	return eventPublisher
}

// GetReviewStageService 获取审核阶段服务。
func GetReviewStageService() *reviewStage.Service {
	Init()
	return reviewStageSvc
}

// triggerReviewStage 触发审核阶段：完成当前 design stage，创建 review stage，运行审核。
func triggerReviewStage(ctx context.Context, projectID, planVersionID int64) error {
	// 查活跃的 workflow_run
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{consts.WorkflowRunStatusDesigning, consts.WorkflowRunStatusReviewing, consts.WorkflowRunStatusExecuting, consts.WorkflowRunStatusReworking}).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	if err != nil || wfRun.IsEmpty() {
		return fmt.Errorf("未找到活跃的 workflow_run: projectID=%d", projectID)
	}
	workflowRunID := wfRun["id"].Int64()
	currentStageRunID := wfRun["current_stage_run_id"].Int64()

	// 完成 design stage（必须成功，否则阻止进入 review）
	if currentStageRunID > 0 {
		if err := stageSvc.CompleteStage(ctx, currentStageRunID); err != nil {
			return fmt.Errorf("完成 design stage 失败，无法进入审核: %w", err)
		}
	}

	// 创建并启动 review stage
	stageRunID, err := stageSvc.StartStage(ctx, workflowRunID, "review")
	if err != nil {
		// 回滚 design stage: completed → running（否则下次提审会因 CompleteStage 报错永久卡住）
		if currentStageRunID > 0 {
			_, rollbackErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
				Where("id", currentStageRunID).
				Where("status", consts.StageStatusCompleted).
				Update(g.Map{"status": consts.StageStatusRunning, "finished_at": nil, "updated_at": gtime.Now()})
			if rollbackErr != nil {
				g.Log().Errorf(ctx, "[triggerReviewStage] 回滚 design stage 失败: %v", rollbackErr)
			}
		}
		return fmt.Errorf("创建 review stage 失败: %w", err)
	}

	// 异步运行审核流程（使用 runtime context 响应工作流级取消/暂停）
	go func() {
		rtCtx := runtimeMgr.GetContext(workflowRunID)
		if err := reviewStageSvc.RunReview(rtCtx, stageRunID, planVersionID); err != nil {
			// 如果是 context 取消导致的错误，不标记为失败
			if rtCtx.Err() != nil {
				g.Log().Infof(rtCtx, "[triggerReviewStage] 审核流程被取消: stageRunID=%d", stageRunID)
				return
			}
			g.Log().Errorf(rtCtx, "[triggerReviewStage] 审核流程失败: stageRunID=%d err=%v", stageRunID, err)
			_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
		}
	}()

	return nil
}

// triggerExecuteStage 审核通过后触发执行阶段：创建 execute stage + 实例化蓝图 + 启动调度。
func triggerExecuteStage(ctx context.Context, workflowRunID, planVersionID int64) error {
	// 创建并启动 execute stage
	stageRunID, err := stageSvc.StartStage(ctx, workflowRunID, "execute")
	if err != nil {
		return fmt.Errorf("创建 execute stage 失败: %w", err)
	}

	// 实例化蓝图为领域任务并启动调度
	if err := executeStageSvc.InstantiateAndStart(ctx, stageRunID, planVersionID); err != nil {
		// 只标记 execute stage_run 自身为 failed，不级联到 workflow_run。
		// 调用方（review concludeReview）负责完整回滚工作流状态。
		// 使用 FailStageOnly 避免 FailStage 将 workflow_run 也打成 failed 导致回滚困难。
		stageSvc.FailStageOnly(ctx, stageRunID, err.Error())
		return fmt.Errorf("执行阶段启动失败: %w", err)
	}

	g.Log().Infof(ctx, "[triggerExecuteStage] 执行阶段已启动 workflowRunID=%d stageRunID=%d planVersionID=%d", workflowRunID, stageRunID, planVersionID)
	return nil
}

// GetExecuteStageService 获取执行阶段服务。
func GetExecuteStageService() *executeStage.Service {
	Init()
	return executeStageSvc
}

// GetTaskScheduler 获取领域任务调度器。
func GetTaskScheduler() *scheduler.DomainTaskScheduler {
	Init()
	return taskScheduler
}

// GetReworkStageService 获取返工阶段服务。
func GetReworkStageService() *reworkStage.Service {
	Init()
	return reworkStageSvc
}

// GetCompleteStageService 获取完成阶段服务。
func GetCompleteStageService() *completeStage.Service {
	Init()
	return completeStageSvc
}

// triggerReworkStage 触发返工阶段：创建 rework stage，启动分析流程。
func triggerReworkStage(ctx context.Context, workflowRunID int64, failedTaskID int64) error {
	stageRunID, err := stageSvc.StartStage(ctx, workflowRunID, "rework")
	if err != nil {
		return fmt.Errorf("创建 rework stage 失败: %w", err)
	}

	if err := reworkStageSvc.HandleRework(ctx, stageRunID, failedTaskID); err != nil {
		_ = stageSvc.FailStage(ctx, stageRunID, err.Error())
		return fmt.Errorf("返工阶段启动失败: %w", err)
	}

	g.Log().Infof(ctx, "[triggerReworkStage] 返工阶段已启动 workflowRunID=%d stageRunID=%d failedTask=%d",
		workflowRunID, stageRunID, failedTaskID)
	return nil
}
