package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/domain/plan"
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/app/mvp/internal/workflow/runtime"
	reviewStage "easymvp/app/mvp/internal/workflow/stage/review"
)

var (
	once            sync.Once
	workflowSvc     *WorkflowService
	stageSvc        *StageService
	planVersionSvc  *plan.PlanVersionService
	reviewStageSvc  *reviewStage.Service
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

		// 审核阶段服务
		issueRepo := repo.NewReviewIssueRepo()
		reviewStageSvc = reviewStage.NewService(stageSvc, issueRepo)

		// 注册审核驳回时回退 design 阶段的回调
		reviewStageSvc.SetDesignRollbackFn(func(ctx context.Context, workflowRunID int64) error {
			_, err := stageSvc.TransitionTo(ctx, workflowRunID, "design")
			return err
		})

		// 注册审核触发回调到 PlanVersionService（避免循环依赖）
		planVersionSvc.SetReviewTrigger(func(ctx context.Context, projectID, planVersionID int64) error {
			return triggerReviewStage(ctx, projectID, planVersionID)
		})

		// 注册 V2 蓝图创建回调到 engine 包（避免循环依赖）
		engine.RegisterBlueprintCreator(func(ctx context.Context, projectID, workflowRunID, conversationID, messageID int64, tasks []engine.ArchitectTask) (int64, int, error) {
			return planVersionSvc.CreateFromArchitectReply(ctx, projectID, workflowRunID, conversationID, messageID, tasks)
		})
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

	// 完成 design stage
	if currentStageRunID > 0 {
		_ = stageSvc.CompleteStage(ctx, currentStageRunID)
	}

	// 创建并启动 review stage
	stageRunID, err := stageSvc.StartStage(ctx, workflowRunID, "review")
	if err != nil {
		return fmt.Errorf("创建 review stage 失败: %w", err)
	}

	// 异步运行审核流程
	go func() {
		bgCtx := context.Background()
		if err := reviewStageSvc.RunReview(bgCtx, stageRunID, planVersionID); err != nil {
			g.Log().Errorf(bgCtx, "[triggerReviewStage] 审核流程失败: stageRunID=%d err=%v", stageRunID, err)
			_ = stageSvc.FailStage(bgCtx, stageRunID, err.Error())
		}
	}()

	return nil
}
