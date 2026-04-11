package orchestrator

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/consts"
	"easymvp/app/mvp/internal/workflow/repo"
)

type recoverableWorkflowRun struct {
	WorkflowRunID int64
	ProjectID     int64
	Status        string
	Stage         string
	StageRunID    int64
}

var (
	createRecoveredRuntime = func(workflowRunID, projectID int64) {
		if runtimeMgr != nil {
			runtimeMgr.Create(workflowRunID, projectID)
		}
	}
	prepareRecoveredTaskScheduler = func(ctx context.Context, workflowRunID int64, stage string, stageRunID int64) error {
		return PrepareTaskSchedulerForStage(ctx, workflowRunID, stage, stageRunID)
	}
	hasRecoveredWorkflowUnfinishedTasks = func(ctx context.Context, workflowRunID int64) bool {
		if taskScheduler == nil {
			return false
		}
		return taskScheduler.HasUnfinished(ctx, workflowRunID)
	}
	startRecoveredTaskScheduler = func(ctx context.Context, workflowRunID int64) error {
		if taskScheduler == nil {
			return nil
		}
		return taskScheduler.Start(ctx, workflowRunID)
	}
)

// RecoverActiveWorkflows 在服务启动时恢复活跃 workflow 的内存态。
// 当前优先恢复 runtime，并重启 execute/rework 阶段的调度循环。
func RecoverActiveWorkflows(ctx context.Context) error {
	Init()

	runs, err := repo.NewWorkflowRunRepo().ListByStatuses(ctx, []string{
		consts.WorkflowRunStatusDesigning,
		consts.WorkflowRunStatusReviewing,
		consts.WorkflowRunStatusExecuting,
		consts.WorkflowRunStatusAccepting,
		consts.WorkflowRunStatusReworking,
	}, "id", "project_id", "status", "current_stage", "current_stage_run_id")
	if err != nil {
		return err
	}
	if len(runs) == 0 {
		return nil
	}

	items := make([]recoverableWorkflowRun, 0, len(runs))
	for _, run := range runs {
		items = append(items, recoverableWorkflowRun{
			WorkflowRunID: g.NewVar(run["id"]).Int64(),
			ProjectID:     g.NewVar(run["project_id"]).Int64(),
			Status:        g.NewVar(run["status"]).String(),
			Stage:         g.NewVar(run["current_stage"]).String(),
			StageRunID:    g.NewVar(run["current_stage_run_id"]).Int64(),
		})
	}

	restartedSchedulers := 0
	var failedWorkflows []int64
	for _, run := range items {
		// 检查启动 ctx 是否已取消，防止无限恢复
		if ctx.Err() != nil {
			g.Log().Warningf(ctx, "[WorkflowRecovery] 启动上下文已取消，停止恢复")
			break
		}

		restarted, failed := recoverWorkflowRun(ctx, run)
		if failed {
			failedWorkflows = append(failedWorkflows, run.WorkflowRunID)
			continue
		}
		if restarted {
			restartedSchedulers++
		}
	}

	if len(failedWorkflows) > 0 {
		g.Log().Errorf(ctx, "[WorkflowRecovery] 以下工作流调度器恢复失败，需要人工检查: %v", failedWorkflows)
	}
	g.Log().Infof(ctx, "[WorkflowRecovery] 恢复活跃工作流完成: total=%d restartedSchedulers=%d failedSchedulers=%d",
		len(runs), restartedSchedulers, len(failedWorkflows))
	return nil
}

func recoverWorkflowRun(ctx context.Context, run recoverableWorkflowRun) (bool, bool) {
	workflowRunID := run.WorkflowRunID
	projectID := run.ProjectID

	if projectID == 0 {
		g.Log().Errorf(ctx, "[WorkflowRecovery] workflow_run(%d) project_id 为 0，跳过恢复", workflowRunID)
		return false, true
	}

	createRecoveredRuntime(workflowRunID, projectID)

	switch run.Stage {
	case consts.StageTypeExecute, consts.StageTypeRework:
		if err := prepareRecoveredTaskScheduler(ctx, workflowRunID, run.Stage, run.StageRunID); err != nil {
			g.Log().Errorf(ctx, "[WorkflowRecovery] 绑定调度器执行器失败: workflowRunID=%d stage=%s err=%v",
				workflowRunID, run.Stage, err)
			return false, true
		}
		if run.Stage == consts.StageTypeExecute && !hasRecoveredWorkflowUnfinishedTasks(ctx, workflowRunID) {
			reconcileWorkflowProgressFn(ctx, workflowRunID)
			if !hasRecoveredWorkflowUnfinishedTasks(ctx, workflowRunID) {
				g.Log().Infof(ctx, "[WorkflowRecovery] execute 已全部完成，跳过调度器重启: workflowRunID=%d", workflowRunID)
				return false, false
			}
		}
		// 调度器用独立 ctx，生命周期不跟随启动流程
		if err := startRecoveredTaskScheduler(context.Background(), workflowRunID); err != nil {
			g.Log().Errorf(ctx, "[WorkflowRecovery] 重启调度失败: workflowRunID=%d stage=%s status=%s err=%v",
				workflowRunID, run.Stage, run.Status, err)
			return false, true
		}
		if run.Stage == consts.StageTypeExecute {
			reconcileWorkflowProgressFn(context.Background(), workflowRunID)
		}
		return true, false
	default:
		g.Log().Infof(ctx, "[WorkflowRecovery] 恢复 runtime: workflowRunID=%d stage=%s status=%s",
			workflowRunID, run.Stage, run.Status)
		return false, false
	}
}
