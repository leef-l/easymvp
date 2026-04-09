package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/collab/adapter"
	collabRepo "easymvp/app/mvp/internal/collab/repo"
	collabService "easymvp/app/mvp/internal/collab/service"
	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/acceptance"
	"easymvp/app/mvp/internal/workflow/autonomy"
	"easymvp/app/mvp/internal/workflow/domain/plan"
	domainTask "easymvp/app/mvp/internal/workflow/domain/task"
	"easymvp/app/mvp/internal/workflow/event"
	executorPkg "easymvp/app/mvp/internal/workflow/executor"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/app/mvp/internal/workflow/runtime"
	"easymvp/app/mvp/internal/workflow/scheduler"
	acceptStage "easymvp/app/mvp/internal/workflow/stage/accept"
	completeStage "easymvp/app/mvp/internal/workflow/stage/complete"
	executeStage "easymvp/app/mvp/internal/workflow/stage/execute"
	reviewStage "easymvp/app/mvp/internal/workflow/stage/review"
	reworkStage "easymvp/app/mvp/internal/workflow/stage/rework"
	watchdogV2 "easymvp/app/mvp/internal/workflow/watchdog"
	"easymvp/app/mvp/internal/workspace"
)

var (
	once              sync.Once
	workflowSvc       *WorkflowService
	stageSvc          *StageService
	planVersionSvc    *plan.PlanVersionService
	reviewStageSvc    *reviewStage.Service
	executeStageSvc   *executeStage.Service
	taskScheduler     *scheduler.DomainTaskScheduler
	acceptStageSvc    *acceptStage.Service
	reworkStageSvc    *reworkStage.Service
	completeStageSvc  *completeStage.Service
	domainWatchdog    *watchdogV2.DomainTaskWatchdog
	runtimeMgr        *runtime.Manager
	eventBus          *event.Bus
	eventPublisher    *event.Publisher
	execRegistry      *executorPkg.Registry
	decisionCenter    *autonomy.DecisionCenter
	collabBindingRepo *collabRepo.BindingRepo
	metaAssessor      *autonomy.MetaAssessor
	metaTuner         *autonomy.MetaTuner
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
		execRegistry = executorPkg.NewRegistry()
		wsMgr := workspace.NewGitWorktreeManager()
		execRegistry.Register(executorPkg.NewAiderExecutor(wsMgr))
		execRegistry.Register(executorPkg.NewChatExecutor())
		execRegistry.Register(executorPkg.NewOpenHandsExecutor(wsMgr))
		execRegistry.Register(executorPkg.NewClaudeCodeExecutor(wsMgr))
		execRegistry.Register(executorPkg.NewCodexCLIExecutor(wsMgr))
		execRegistry.Register(executorPkg.NewGeminiCLIExecutor(wsMgr))
		execRegistry.Register(executorPkg.NewAutoExecutor(execRegistry))

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
		engine.RegisterBlueprintPatchApplier(func(ctx context.Context, projectID, workflowRunID, conversationID, messageID int64, patches []engine.ArchitectTaskPatch) (int64, int, error) {
			return planVersionSvc.ApplyTaskPatchesFromArchitectReply(ctx, projectID, workflowRunID, conversationID, messageID, patches)
		})
		engine.RegisterArchitectReviewResubmitter(func(ctx context.Context, projectID int64) error {
			return planVersionSvc.SubmitForReviewAsync(ctx, projectID)
		})

		// 完成阶段服务
		completeStageSvc = completeStage.NewService()

		// 验收阶段服务
		acceptRunRepo := repo.NewAcceptRunRepo()
		acceptIssueRepo := repo.NewAcceptIssueRepo()
		acceptRuleRepo := repo.NewAcceptRuleRepo()
		acceptEvidenceRepo := repo.NewAcceptEvidenceRepo()
		evidenceCollector := acceptance.NewEvidenceCollector(acceptEvidenceRepo)
		ruleEngine := acceptance.NewRuleEngine(acceptRuleRepo)
		// LLM Judge：总开关控制是否注入，项目类型灰度在 Judge.Evaluate 内部运行时判断
		var judge *acceptance.Judge
		if engine.GetConfigInt(context.Background(), "accept.llm_judge_enabled", "accept.llmJudgeEnabled", 1) == 1 {
			judge = acceptance.NewJudge()
		}
		decisionReducer := acceptance.NewDecisionReducer(acceptIssueRepo, judge)
		acceptStageSvc = acceptStage.NewService(acceptRunRepo, evidenceCollector, ruleEngine, decisionReducer)
		acceptStageSvc.SetStageCompleter(stageSvc)

		// 注册 accept → complete 回调（决策点 4: accept.passed）
		acceptStageSvc.SetCompleteTrigger(func(ctx context.Context, workflowRunID int64) error {
			if acceptStage.HasManualCompleteBypass(ctx) {
				return stageSvc.completeWorkflow(ctx, workflowRunID)
			}
			if decisionCenter.IsEnabled(ctx) {
				projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", workflowRunID).Value("project_id") //nolint: errcheck — best-effort for autonomy decision
				resp := decisionCenter.Decide(ctx, &autonomy.DecisionRequest{
					WorkflowRunID: workflowRunID,
					ProjectID:     projectID.Int64(),
					TriggerSource: consts.TriggerAcceptPassed,
				})
				if resp.Handled {
					return nil
				}
			}
			return stageSvc.completeWorkflow(ctx, workflowRunID)
		})

		// 注册 accept 阶段触发回调到 StageService
		stageSvc.SetAcceptTriggerFn(func(ctx context.Context, workflowRunID, stageRunID int64) error {
			return acceptStageSvc.Run(ctx, workflowRunID, stageRunID)
		})

		// 返工阶段服务
		reworkStageSvc = reworkStage.NewService()
		reworkStageSvc.SetStageCompleter(stageSvc)

		// 注册 rework 完成后推回 execute 的回调
		reworkStageSvc.SetExecuteTrigger(func(ctx context.Context, workflowRunID, planVersionID int64) error {
			wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
				Where("id", workflowRunID).
				WhereNull("deleted_at").
				Fields("status").
				One()
			if err != nil {
				return fmt.Errorf("查询 workflow_run 失败: %w", err)
			}
			if wfRun.IsEmpty() {
				return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
			}
			if wfRun["status"].String() != consts.WorkflowRunStatusReworking {
				g.Log().Warningf(ctx, "[Registry] rework 完成时 workflow_run(%d) 已不在 reworking 状态，跳过回流 execute", workflowRunID)
				return nil
			}

			stageRunID, err := stageSvc.StartStage(ctx, workflowRunID, consts.StageTypeExecute)
			if err != nil {
				return fmt.Errorf("rework 完成后创建 execute stage 失败: %w", err)
			}
			g.Log().Infof(ctx, "[Registry] rework 完成，已重新进入 execute: workflowRunID=%d stageRunID=%d", workflowRunID, stageRunID)

			if err := PrepareTaskSchedulerForStage(ctx, workflowRunID, consts.StageTypeExecute, stageRunID); err != nil {
				return fmt.Errorf("rework 完成后绑定 execute 调度器失败: %w", err)
			}

			// 重启调度器拾取 pending 任务
			return taskScheduler.Start(context.Background(), workflowRunID)
		})

		// 注册 accept failed → rework 回调（决策点 5: accept.failed, 6: accept.manual_review）
		acceptStageSvc.SetReworkTrigger(func(ctx context.Context, workflowRunID int64, acceptRunID int64, issues []acceptance.RuleHit) error {
			g.Log().Infof(ctx, "[Registry] 验收失败，触发返工: workflowRunID=%d acceptRunID=%d issues=%d",
				workflowRunID, acceptRunID, len(issues))
			// 找到第一个 blocker/error issue 关联的 taskID 作���返工入口
			var failedTaskID int64
			for _, issue := range issues {
				if issue.DomainTaskID > 0 {
					failedTaskID = issue.DomainTaskID
					break
				}
			}
			if failedTaskID == 0 {
				// 决策点 6: accept.manual_review
				if decisionCenter.IsEnabled(ctx) {
					projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", workflowRunID).Value("project_id") //nolint: errcheck — best-effort for autonomy decision
					resp := decisionCenter.Decide(ctx, &autonomy.DecisionRequest{
						WorkflowRunID:  workflowRunID,
						ProjectID:      projectID.Int64(),
						TriggerSource:  consts.TriggerAcceptManualReview,
						TriggerContext: map[string]interface{}{"accept_run_id": acceptRunID},
					})
					if resp.Handled {
						return acceptStage.ErrManualReviewRequired
					}
				}
				g.Log().Warningf(ctx, "[Registry] 验收失败但无关联任务，转 manual_review: workflowRunID=%d", workflowRunID)
				acceptRunRepoLocal := repo.NewAcceptRunRepo()
				_ = acceptRunRepoLocal.UpdateDecision(ctx, acceptRunID, acceptance.DecisionManualReview, 0,
					"自动返工���败：验收问题未关联具体任务，需人工决策")
				return acceptStage.ErrManualReviewRequired
			}
			// 决策��� 5: accept.failed
			if decisionCenter.IsEnabled(ctx) {
				projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", workflowRunID).Value("project_id") //nolint: errcheck — best-effort for autonomy decision
				resp := decisionCenter.Decide(ctx, &autonomy.DecisionRequest{
					WorkflowRunID: workflowRunID,
					ProjectID:     projectID.Int64(),
					DomainTaskID:  failedTaskID,
					TriggerSource: consts.TriggerAcceptFailed,
					TriggerContext: map[string]interface{}{
						"accept_run_id": acceptRunID,
						"task_id":       failedTaskID,
					},
				})
				if resp.Handled {
					return nil
				}
			}
			return triggerReworkStage(ctx, workflowRunID, failedTaskID, "accept")
		})

		// ��册 rework → accept 回调（决策点 7: rework.completed）
		reworkStageSvc.SetAcceptTrigger(func(ctx context.Context, workflowRunID int64) error {
			g.Log().Infof(ctx, "[Registry] rework 完成，恢复验收状态: workflowRunID=%d", workflowRunID)

			if decisionCenter.IsEnabled(ctx) {
				projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", workflowRunID).Value("project_id") //nolint: errcheck — best-effort for autonomy decision
				resp := decisionCenter.Decide(ctx, &autonomy.DecisionRequest{
					WorkflowRunID: workflowRunID,
					ProjectID:     projectID.Int64(),
					TriggerSource: consts.TriggerReworkCompleted,
				})
				if resp.Handled {
					return nil
				}
			}

			// StartStage("accept") 会自动 CAS: reworking→accepting + 同步 project.status
			stageRunID, stageErr := stageSvc.StartStage(ctx, workflowRunID, "accept")
			if stageErr != nil {
				return fmt.Errorf("重建 accept stage 失败: %w", stageErr)
			}
			if runErr := acceptStageSvc.Run(ctx, workflowRunID, stageRunID); runErr != nil {
				g.Log().Errorf(ctx, "[Registry] accept 运行失败，标记 stage 失败: workflowRunID=%d err=%v", workflowRunID, runErr)
				_ = stageSvc.FailStage(ctx, stageRunID, runErr.Error())
				return runErr
			}
			return nil
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

		// 注册工作流暂停后的清理回调
		workflowSvc.SetWorkflowPausedCallback(func(ctx context.Context, workflowRunID int64) {
			g.Log().Infof(ctx, "[Registry] 工作流暂停，停止调度器: workflowRunID=%d", workflowRunID)
			taskScheduler.Stop(workflowRunID)
		})

		// 注册工作流恢复后的回调（重启调度器）
		workflowSvc.SetWorkflowResumedCallback(func(ctx context.Context, workflowRunID int64, resumeStatus string) {
			if resumeStatus == consts.WorkflowRunStatusExecuting || resumeStatus == consts.WorkflowRunStatusReworking {
				g.Log().Infof(ctx, "[Registry] 工作流恢复，重启调度器: workflowRunID=%d status=%s", workflowRunID, resumeStatus)
				stageType := consts.StageTypeExecute
				if resumeStatus == consts.WorkflowRunStatusReworking {
					stageType = consts.StageTypeRework
				}
				if err := PrepareTaskSchedulerForStage(ctx, workflowRunID, stageType, 0); err != nil {
					g.Log().Errorf(ctx, "[Registry] 工作流恢复后绑定调度器失败: workflowRunID=%d err=%v", workflowRunID, err)
					return
				}
				if err := taskScheduler.Start(ctx, workflowRunID); err != nil {
					g.Log().Errorf(ctx, "[Registry] 工作流恢复后调度器启动失败: workflowRunID=%d err=%v", workflowRunID, err)
				}
			}
		})

		// 注册工作流取消后的回调（停止调度器并释放运行中任务）。
		workflowSvc.SetWorkflowCanceledCallback(func(ctx context.Context, workflowRunID int64) {
			g.Log().Infof(ctx, "[Registry] 工作流取消，停止调度器: workflowRunID=%d", workflowRunID)
			taskScheduler.Pause(ctx, workflowRunID)
		})

		taskScheduler.SetFailureCallback(func(ctx context.Context, workflowRunID, taskID int64, errMsg string) bool {
			taskRecord, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
				Where("id", taskID).
				WhereNull("deleted_at").
				Fields("status, retry_count").
				One()
			if err != nil {
				g.Log().Warningf(ctx, "[Registry] 查询失败任务失败: task=%d err=%v", taskID, err)
				return false
			}
			if taskRecord.IsEmpty() || taskRecord["status"].String() != domainTask.StatusFailed {
				return true
			}

			normalizedErr := strings.TrimSpace(errMsg)
			immediateEscalate := strings.Contains(normalizedErr, "检测到可疑文件") ||
				strings.Contains(normalizedErr, "检测到越界修改") ||
				strings.Contains(normalizedErr, "affected_resources 存在歧义")

			maxRetries := engine.GetConfigInt(ctx, "watchdog.max_retries", "engine.watchdog.maxRetries", 3)
			nextRetry := taskRecord["retry_count"].Int() + 1
			if !immediateEscalate && nextRetry < maxRetries {
				now := gtime.Now()
				result, upErr := g.DB().Model("mvp_domain_task").Ctx(ctx).
					Where("id", taskID).
					Where("status", domainTask.StatusFailed).
					Update(g.Map{
						"status":           domainTask.StatusPending,
						"retry_count":      nextRetry,
						"result":           nil,
						"error_message":    nil,
						"started_at":       nil,
						"completed_at":     nil,
						"heartbeat_at":     nil,
						"locked_resources": nil,
						"updated_at":       now,
					})
				if upErr != nil {
					g.Log().Errorf(ctx, "[Registry] 即时重试更新失败: task=%d err=%v", taskID, upErr)
					return false
				}
				rows, _ := result.RowsAffected()
				if rows == 0 {
					return true
				}
				g.Log().Infof(ctx, "[Registry] 任务失败后即时重试: task=%d retry=%d/%d", taskID, nextRetry, maxRetries)
				taskScheduler.Wakeup(ctx, workflowRunID)
				return true
			}

			result, upErr := g.DB().Model("mvp_domain_task").Ctx(ctx).
				Where("id", taskID).
				Where("status", domainTask.StatusFailed).
				Update(g.Map{
					"status":     domainTask.StatusEscalated,
					"updated_at": gtime.Now(),
				})
			if upErr != nil {
				g.Log().Errorf(ctx, "[Registry] 即时升级失败任务状态失败: task=%d err=%v", taskID, upErr)
				return false
			}
			rows, _ := result.RowsAffected()
			if rows == 0 {
				return true
			}

			if decisionCenter.IsEnabled(ctx) {
				projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", workflowRunID).Value("project_id") //nolint: errcheck
				resp := decisionCenter.Decide(ctx, &autonomy.DecisionRequest{
					WorkflowRunID:  workflowRunID,
					ProjectID:      projectID.Int64(),
					DomainTaskID:   taskID,
					TriggerSource:  consts.TriggerTaskRetryExhausted,
					TriggerContext: map[string]interface{}{"task_id": taskID, "reason": normalizedErr},
				})
				if resp.Handled {
					return true
				}
			}

			if err := triggerReworkStage(ctx, workflowRunID, taskID); err != nil {
				g.Log().Errorf(ctx, "[Registry] 即时触发 rework 失败: task=%d err=%v", taskID, err)
				return false
			}
			g.Log().Infof(ctx, "[Registry] 任务失败后已即时触发 rework: workflowRunID=%d task=%d", workflowRunID, taskID)
			return true
		})

		// Watchdog V2: 监控 domain_task 心跳
		domainWatchdog = watchdogV2.New()
		domainWatchdog.SetScheduler(taskScheduler)
		domainWatchdog.SetRetryFn(func(ctx context.Context, taskID int64) error {
			wfRunID, wfErr := g.DB().Model("mvp_domain_task").Ctx(ctx).Where("id", taskID).Value("workflow_run_id")
			if wfErr != nil {
				return fmt.Errorf("查询 domain_task workflow_run_id 失败: %w", wfErr)
			}
			if wfRunID.Int64() == 0 {
				return nil
			}
			// 自治中台包裹
			if decisionCenter.IsEnabled(ctx) {
				projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", wfRunID.Int64()).Value("project_id") //nolint: errcheck — best-effort for autonomy
				resp := decisionCenter.Decide(ctx, &autonomy.DecisionRequest{
					WorkflowRunID:  wfRunID.Int64(),
					ProjectID:      projectID.Int64(),
					DomainTaskID:   taskID,
					TriggerSource:  consts.TriggerTaskFailed,
					TriggerContext: map[string]interface{}{"task_id": taskID},
				})
				if resp.Handled {
					return nil // 中台已接管
				}
				// 降级到原逻辑
			}
			taskScheduler.Wakeup(ctx, wfRunID.Int64())
			return nil
		})
		domainWatchdog.SetEscalateFn(func(ctx context.Context, workflowRunID, taskID int64) error {
			if decisionCenter.IsEnabled(ctx) {
				projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", workflowRunID).Value("project_id") //nolint: errcheck — best-effort for autonomy decision
				resp := decisionCenter.Decide(ctx, &autonomy.DecisionRequest{
					WorkflowRunID:  workflowRunID,
					ProjectID:      projectID.Int64(),
					DomainTaskID:   taskID,
					TriggerSource:  consts.TriggerTaskRetryExhausted,
					TriggerContext: map[string]interface{}{"task_id": taskID},
				})
				if resp.Handled {
					return nil
				}
			}
			return triggerReworkStage(ctx, workflowRunID, taskID)
		})
		// 熔断器：项目级异常检测
		decisionRepo := repo.NewAutonomyDecisionRepo()
		circuitBreaker := autonomy.NewCircuitBreaker(decisionRepo, nil)
		domainWatchdog.SetCircuitBreaker(circuitBreaker)
		domainWatchdog.SetPauseFn(func(ctx context.Context, workflowRunID int64, reason string) error {
			if decisionCenter.IsEnabled(ctx) {
				projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).Where("id", workflowRunID).Value("project_id") //nolint: errcheck — best-effort for autonomy decision
				resp := decisionCenter.Decide(ctx, &autonomy.DecisionRequest{
					WorkflowRunID:  workflowRunID,
					ProjectID:      projectID.Int64(),
					TriggerSource:  consts.TriggerCircuitBreak,
					TriggerContext: map[string]interface{}{"reason": reason},
				})
				if resp.Handled {
					return nil
				}
			}
			return workflowSvc.Pause(ctx, workflowRunID, reason)
		})
		// 熔断后自动重规划评估
		replanner := autonomy.NewReplanner(decisionRepo)
		domainWatchdog.SetReplanFn(func(ctx context.Context, workflowRunID, projectID int64, breakReason string) {
			// 决策点 8: replan.suggested — 自治中台包裹
			if decisionCenter.IsEnabled(ctx) {
				resp := decisionCenter.Decide(ctx, &autonomy.DecisionRequest{
					WorkflowRunID:  workflowRunID,
					ProjectID:      projectID,
					TriggerSource:  consts.TriggerReplanSuggested,
					TriggerContext: map[string]interface{}{"break_reason": breakReason},
				})
				if resp.Handled {
					return
				}
			}
			// 降级到原逻辑
			failedTasks, ftErr := g.DB().Model("mvp_domain_task").Ctx(ctx).
				Where("workflow_run_id", workflowRunID).
				WhereIn("status", g.Slice{"failed", "escalated"}).
				WhereNull("deleted_at").
				Fields("id, name, result, retry_count").All()
			if ftErr != nil {
				g.Log().Warningf(ctx, "[Registry] 查询失败任务列表失败: wfRunID=%d err=%v", workflowRunID, ftErr)
			}
			var failed []autonomy.FailedTaskInfo
			for _, t := range failedTasks {
				failed = append(failed, autonomy.FailedTaskInfo{
					TaskID:       t["id"].Int64(),
					TaskName:     t["name"].String(),
					ErrorMessage: t["result"].String(),
					RetryCount:   t["retry_count"].Int(),
				})
			}
			rec, err := replanner.Evaluate(ctx, &autonomy.ReplanInput{
				WorkflowRunID: workflowRunID,
				ProjectID:     projectID,
				TriggerSource: "circuit_breaker",
				BreakReason:   breakReason,
				FailedTasks:   failed,
			})
			if err != nil {
				g.Log().Warningf(ctx, "[Registry] 熔断后重规划评估失败: wfRun=%d err=%v", workflowRunID, err)
				return
			}
			g.Log().Infof(ctx, "[Registry] 熔断后重规划建议: wfRun=%d action=%s", workflowRunID, rec.Action)
		})
		domainWatchdog.Start(context.Background())

		// 阶段报告回调：每个阶段完成时自动生成报告
		reportRepo := repo.NewProjectReportRepo()
		reporter := autonomy.NewReporter(reportRepo)
		stageSvc.SetStageReportFn(func(ctx context.Context, workflowRunID int64, stageType string) {
			if err := reporter.GenerateStageReport(ctx, workflowRunID, stageType); err != nil {
				g.Log().Warningf(ctx, "[Registry] 阶段报告生成失败: wfRun=%d stage=%s err=%v", workflowRunID, stageType, err)
			}
		})

		// ==================== 自治决策中台初始化 ====================
		policyRuleRepo := repo.NewPolicyRuleRepo()
		riskGateRuleRepo := repo.NewRiskGateRuleRepo()
		decisionActionRepo := repo.NewDecisionActionRepo()
		humanCheckpointRepo := repo.NewHumanCheckpointRepo()
		situationSnapshotRepo := repo.NewSituationSnapshotRepo()

		policyEngine := autonomy.NewPolicyEngine(policyRuleRepo)
		riskGate := autonomy.NewRiskGate(riskGateRuleRepo)
		actionDispatcher := autonomy.NewActionDispatcher(decisionActionRepo)
		decisionCenter = autonomy.NewDecisionCenter(
			policyEngine, riskGate, actionDispatcher,
			decisionActionRepo, humanCheckpointRepo, eventPublisher,
		)
		sensor := autonomy.NewSensor(situationSnapshotRepo)
		objectiveSvc := autonomy.NewObjectiveService()
		decisionCenter.SetPhaseADeps(sensor, objectiveSvc)

		// ==================== Phase B: 策略函数 + Planner + Actuator ====================
		planner := autonomy.NewPlanner()
		planner.Register(autonomy.NewCostGuardStrategy())       // 优先级 100：成本最优先
		planner.Register(autonomy.NewAdaptiveRetryStrategy())   // 优先级 90：失败处理
		planner.Register(autonomy.NewProactiveReplanStrategy()) // 优先级 70：主动重规划
		planner.Register(autonomy.NewEngineSelectionStrategy()) // 优先级 70：执行器选择
		planner.Register(autonomy.NewBatchAdjustStrategy())     // 优先级 60：批次调整
		planner.Register(autonomy.NewQualityGateStrategy())     // 优先级 50：质量门
		actuator := autonomy.NewActuator()
		decisionCenter.SetPhaseBDeps(planner, actuator)

		// ==================== Phase D: 元认知 — Observer + Learner ====================
		metaObserver := autonomy.NewMetaObserver()
		learner := autonomy.NewLearner()
		decisionCenter.SetPhaseDDeps(metaObserver, learner)

		// MetaAssessor 和 MetaTuner 作为服务单例，供 API 层调用
		metaAssessor = autonomy.NewMetaAssessor(metaObserver, learner)
		metaTuner = autonomy.NewMetaTuner(learner)

		// 注册 ActionDispatcher 回调（通过回调注入避免循环依赖）
		actionDispatcher.SetCallback(consts.ActionTypeRetryTask, func(ctx context.Context, req *autonomy.DecisionRequest) error {
			if req.DomainTaskID == 0 {
				return fmt.Errorf("retry_task: 缺少 domain_task_id")
			}
			taskScheduler.Wakeup(ctx, req.WorkflowRunID)
			return nil
		})
		actionDispatcher.SetCallback(consts.ActionTypeTriggerRework, func(ctx context.Context, req *autonomy.DecisionRequest) error {
			return triggerReworkStage(ctx, req.WorkflowRunID, req.DomainTaskID)
		})
		actionDispatcher.SetCallback(consts.ActionTypeRerunAccept, func(ctx context.Context, req *autonomy.DecisionRequest) error {
			stageRunID, err := stageSvc.StartStage(ctx, req.WorkflowRunID, "accept")
			if err != nil {
				return err
			}
			return acceptStageSvc.Run(ctx, req.WorkflowRunID, stageRunID)
		})
		actionDispatcher.SetCallback(consts.ActionTypePauseWorkflow, func(ctx context.Context, req *autonomy.DecisionRequest) error {
			return workflowSvc.Pause(ctx, req.WorkflowRunID, "自治中台: 暂停")
		})
		actionDispatcher.SetCallback(consts.ActionTypeApproveComplete, func(ctx context.Context, req *autonomy.DecisionRequest) error {
			return stageSvc.completeWorkflow(ctx, req.WorkflowRunID)
		})
		actionDispatcher.SetCallback(consts.ActionTypeNotifyHuman, func(ctx context.Context, req *autonomy.DecisionRequest) error {
			g.Log().Infof(ctx, "[DecisionCenter] notify_human: wfRun=%d trigger=%s", req.WorkflowRunID, req.TriggerSource)
			// 通过事件系统触发飞书/协作平台通知（CheckpointNotifier 订阅此事件）
			eventPublisher.Emit(ctx, event.Event{
				WorkflowRunID: req.WorkflowRunID,
				EntityType:    event.EntityDecisionAction,
				EventType:     event.EventAutonomyActionFailed, // 复用告警事件
				Payload: g.Map{
					"trigger_source": req.TriggerSource,
					"project_id":     req.ProjectID,
					"error":          fmt.Sprintf("人工通知: trigger=%s wfRun=%d", req.TriggerSource, req.WorkflowRunID),
				},
			})
			return nil
		})
		actionDispatcher.SetCallback(consts.ActionTypeReplanWorkflow, func(ctx context.Context, req *autonomy.DecisionRequest) error {
			g.Log().Infof(ctx, "[DecisionCenter] replan_workflow: wfRun=%d project=%d", req.WorkflowRunID, req.ProjectID)
			replannerInst := autonomy.NewReplanner(repo.NewAutonomyDecisionRepo())
			rec, err := replannerInst.Evaluate(ctx, &autonomy.ReplanInput{
				WorkflowRunID: req.WorkflowRunID,
				ProjectID:     req.ProjectID,
				TriggerSource: req.TriggerSource,
			})
			if err != nil {
				return fmt.Errorf("重规划评估失败: %w", err)
			}
			g.Log().Infof(ctx, "[DecisionCenter] replan result: action=%s reasoning=%s", rec.Action, rec.Reasoning)
			return nil
		})
		actionDispatcher.SetCallback(consts.ActionTypeSwitchExecutor, func(ctx context.Context, req *autonomy.DecisionRequest) error {
			if req.DomainTaskID == 0 {
				return fmt.Errorf("switch_executor: 缺少 domain_task_id")
			}
			// 从 TriggerContext 获取目标引擎类型
			engineType := ""
			if req.TriggerContext != nil {
				if v, ok := req.TriggerContext["engine_type"].(string); ok {
					engineType = v
				}
			}
			if engineType == "" {
				return fmt.Errorf("switch_executor: 缺少 engine_type")
			}
			g.Log().Infof(ctx, "[DecisionCenter] switch_executor: task=%d → %s", req.DomainTaskID, engineType)
			// 更新任务的执行模式
			_, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
				Where("id", req.DomainTaskID).
				Update(g.Map{"execution_mode": engineType})
			if err != nil {
				return fmt.Errorf("更新执行模式失败: %w", err)
			}
			// 唤醒调度器重新执行
			taskScheduler.Wakeup(ctx, req.WorkflowRunID)
			return nil
		})

		// Legacy → V2 执行器桥接：注入 V2ExecutorFn 到旧引擎，使 legacy 任务也能走 V2 执行器
		legacyScheduler := engine.GetScheduler()
		legacyScheduler.GetExecutor().SetV2Executor(func(ctx context.Context, projectID, taskID int64, task gdb.Record, modelInfo *engine.ModelInfo, executionMode string) bool {
			v2Exec := execRegistry.Get(executionMode)
			if v2Exec == nil {
				return false // V2 不支持，回退 legacy
			}
			g.Log().Infof(ctx, "[Registry] legacy task=%d 桥接到 V2 执行器 mode=%s", taskID, executionMode)
			// 注：此回调已在 legacy executor 的独立 goroutine 中执行，无需再起 goroutine
			defer func() {
				if r := recover(); r != nil {
					g.Log().Errorf(ctx, "[Registry] V2 执行器 panic: task=%d mode=%s err=%v", taskID, executionMode, r)
					legacyScheduler.OnTaskFailed(projectID, taskID, fmt.Sprintf("V2 executor panic: %v", r))
				}
			}()
			result := v2Exec.Execute(ctx, &executorPkg.Request{
				ProjectID:  projectID,
				TaskID:     taskID,
				TaskRecord: task,
				ModelInfo:  modelInfo,
			})
			if result.Error != nil {
				legacyScheduler.OnTaskFailed(projectID, taskID, result.Error.Error())
			} else {
				legacyScheduler.OnTaskCompleted(projectID, taskID)
			}
			return true
		})

		// ==================== 协作平台通知初始化 ====================
		collabBindingRepo = collabRepo.NewBindingRepo()
		if engine.GetConfigInt(context.Background(), "workflow.collab.feishu_enabled", "workflow.collab.feishuEnabled", 0) == 1 {
			feishuAdapter := adapter.NewFeishuAdapter()
			cpNotifier := collabService.NewCheckpointNotifier(feishuAdapter, collabBindingRepo)
			eventBus.Subscribe(event.EventAutonomyCheckpointOpened, cpNotifier.OnCheckpointOpened)
			eventBus.Subscribe(event.EventAutonomyActionFailed, cpNotifier.OnActionFailed)
			eventBus.Subscribe(event.EventAutonomyGateBlocked, cpNotifier.OnGateBlocked)

			reportNotifier := collabService.NewReportNotifier(feishuAdapter, collabBindingRepo)
			eventBus.Subscribe(event.EventStageCompleted, reportNotifier.OnStageCompleted)
			eventBus.Subscribe(event.EventWorkflowCompleted, reportNotifier.OnWorkflowCompleted)

			g.Log().Info(context.Background(), "[Registry] 飞书协作通知已注册")
		}

		// ==================== Phase D: 元认知定时评估任务 ====================
		go startMetaAssessmentLoop(metaAssessor, metaTuner, metaObserver)
	})
}

// startMetaAssessmentLoop 元认知评估定时循环。
// 每 24 小时执行一次，扫描所有活跃项目，生成评估 + 调参建议。
func startMetaAssessmentLoop(assessor *autonomy.MetaAssessor, tuner *autonomy.MetaTuner, observer *autonomy.MetaObserver) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		func() {
			ctx := context.Background()
			defer func() {
				if r := recover(); r != nil {
					g.Log().Errorf(ctx, "[MetaAssessment] 定时评估 panic: %v", r)
				}
			}()

			// 灰度检查
			if !observer.IsEnabled(ctx) {
				return
			}

			g.Log().Info(ctx, "[MetaAssessment] 开始定时评估...")

			// 扫描有观测记录的项目
			projectIDs, err := g.DB().Model("mvp_observation_record").Ctx(ctx).
				WhereNull("deleted_at").
				Fields("DISTINCT project_id").
				All()
			if err != nil {
				g.Log().Warningf(ctx, "[MetaAssessment] 查询项目列表失败: %v", err)
				return
			}

			now := gtime.Now()
			periodStart := gtime.New(now.AddDate(0, 0, -7))

			for _, row := range projectIDs {
				pid := row["project_id"].Int64()
				if pid == 0 {
					continue
				}

				result, err := assessor.Assess(ctx, pid, periodStart, now)
				if err != nil {
					g.Log().Warningf(ctx, "[MetaAssessment] 评估失败: project=%d err=%v", pid, err)
					continue
				}

				// 生成调参建议
				recommendations := tuner.GenerateRecommendations(ctx, result)
				if len(recommendations) > 0 {
					_ = tuner.SaveAndApply(ctx, recommendations)
					g.Log().Infof(ctx, "[MetaAssessment] project=%d: %d 条建议已生成", pid, len(recommendations))
				}
			}

			g.Log().Info(ctx, "[MetaAssessment] 定时评估完成")
		}()
	}
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

	// 完成 design stage。若当前 design stage 已被人工清理或重建，允许跳过旧实例，避免审核链路被陈旧 stage_run 卡死。
	if currentStageRunID > 0 {
		stageRun, stageErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
			Where("id", currentStageRunID).
			Fields("stage_type, status, deleted_at").
			One()
		if stageErr != nil {
			return fmt.Errorf("查询 design stage 失败，无法进入审核: %w", stageErr)
		}

		switch {
		case stageRun.IsEmpty():
			g.Log().Warningf(ctx, "[triggerReviewStage] design stage_run(%d) 不存在，继续进入 review", currentStageRunID)
		case stageRun["stage_type"].String() != consts.StageTypeDesign:
			g.Log().Warningf(ctx, "[triggerReviewStage] 当前 stage_run(%d) 不是 design，而是 %s，继续进入 review",
				currentStageRunID, stageRun["stage_type"].String())
		case !stageRun["deleted_at"].IsEmpty():
			g.Log().Warningf(ctx, "[triggerReviewStage] design stage_run(%d) 已被软删除，继续进入 review", currentStageRunID)
		case stageRun["status"].String() == consts.StageStatusCompleted:
			// 已完成则直接进入审核
		case stageRun["status"].String() == consts.StageStatusRunning:
			if err := stageSvc.CompleteStage(ctx, currentStageRunID); err != nil {
				return fmt.Errorf("完成 design stage 失败，无法进入审核: %w", err)
			}
		default:
			return fmt.Errorf("design stage_run(%d) 状态异常(%s)，无法进入审核",
				currentStageRunID, stageRun["status"].String())
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

	// 同步运行审核流程（让 ConfirmPlan 能拿到审核结果返回给前端）
	rtCtx := runtimeMgr.GetContext(workflowRunID)
	if err := reviewStageSvc.RunReview(rtCtx, stageRunID, planVersionID); err != nil {
		if rtCtx.Err() != nil {
			g.Log().Infof(rtCtx, "[triggerReviewStage] 审核流程被取消: stageRunID=%d", stageRunID)
			return nil
		}
		g.Log().Errorf(rtCtx, "[triggerReviewStage] 审核流程失败: stageRunID=%d err=%v", stageRunID, err)
		_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
		// 不返回 error — 审核驳回不是系统错误，concludeReview 已处理状态回退
	}

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

// PrepareTaskSchedulerForStage 在恢复/人工接管场景下，为既有任务重新绑定执行器后再启动调度器。
func PrepareTaskSchedulerForStage(ctx context.Context, workflowRunID int64, stageType string, stageRunID int64) error {
	Init()

	switch stageType {
	case consts.StageTypeExecute:
		if stageRunID == 0 {
			val, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
				Where("id", workflowRunID).
				WhereNull("deleted_at").
				Value("current_stage_run_id")
			if err != nil {
				return fmt.Errorf("查询 execute stage_run 失败: %w", err)
			}
			stageRunID = val.Int64()
		}
		if stageRunID == 0 {
			return fmt.Errorf("workflow_run(%d) 当前 execute stage_run 为空", workflowRunID)
		}
		executeStageSvc.BindExecuteStage(stageRunID, workflowRunID)
	case consts.StageTypeRework:
		executeStageSvc.BindExecutor(workflowRunID)
	}

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

// GetAcceptStageService 获取验收阶段服务。
func GetAcceptStageService() *acceptStage.Service {
	Init()
	return acceptStageSvc
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

// GetExecutorRegistry 获取 V2 执行器注册表（供 legacy 引擎统一分发）。
func GetExecutorRegistry() *executorPkg.Registry {
	Init()
	return execRegistry
}

// GetDecisionCenter 获取自治决策中台单例。
func GetDecisionCenter() *autonomy.DecisionCenter {
	Init()
	return decisionCenter
}

// GetCollabBindingRepo 获取协作平台绑定仓储单例。
func GetCollabBindingRepo() *collabRepo.BindingRepo {
	Init()
	return collabBindingRepo
}

// GetMetaAssessor 获取元认知评估器单例。
func GetMetaAssessor() *autonomy.MetaAssessor {
	Init()
	return metaAssessor
}

// GetMetaTuner 获取元认知校准器单例。
func GetMetaTuner() *autonomy.MetaTuner {
	Init()
	return metaTuner
}

// triggerReworkStage 触发返工阶段：创建 rework stage，启动分析流程。
// sourceStage 默认 "execute"，从 accept 触发时传 "accept"。
func triggerReworkStage(ctx context.Context, workflowRunID int64, failedTaskID int64, sourceStages ...string) error {
	sourceStage := "execute"
	if len(sourceStages) > 0 && sourceStages[0] != "" {
		sourceStage = sourceStages[0]
	}

	stageRunID, err := stageSvc.StartStage(ctx, workflowRunID, "rework")
	if err != nil {
		return fmt.Errorf("创建 rework stage 失败: %w", err)
	}

	if err := reworkStageSvc.HandleReworkWithSource(ctx, stageRunID, failedTaskID, sourceStage); err != nil {
		_ = stageSvc.FailStage(ctx, stageRunID, err.Error())
		return fmt.Errorf("返工阶段启动失败: %w", err)
	}

	g.Log().Infof(ctx, "[triggerReworkStage] 返工阶段已启动 workflowRunID=%d stageRunID=%d failedTask=%d",
		workflowRunID, stageRunID, failedTaskID)
	return nil
}
