package orchestrator

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/consts"
)

// RecoverActiveWorkflows 在服务启动时恢复活跃 workflow 的内存态。
// 当前优先恢复 runtime，并重启 execute/rework 阶段的调度循环。
func RecoverActiveWorkflows(ctx context.Context) error {
	Init()

	runs, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		WhereIn("status", g.Slice{
			consts.WorkflowRunStatusDesigning,
			consts.WorkflowRunStatusReviewing,
			consts.WorkflowRunStatusExecuting,
			consts.WorkflowRunStatusAccepting,
			consts.WorkflowRunStatusReworking,
		}).
		WhereNull("deleted_at").
		Fields("id, project_id, status, current_stage").
		OrderAsc("created_at").
		All()
	if err != nil {
		return err
	}
	if len(runs) == 0 {
		return nil
	}

	restartedSchedulers := 0
	for _, run := range runs {
		// 检查启动 ctx 是否已取消，防止无限恢复
		if ctx.Err() != nil {
			g.Log().Warningf(ctx, "[WorkflowRecovery] 启动上下文已取消，停止恢复")
			break
		}

		workflowRunID := run["id"].Int64()
		projectID := run["project_id"].Int64()
		status := run["status"].String()
		stage := run["current_stage"].String()

		runtimeMgr.Create(workflowRunID, projectID)

		switch stage {
		case consts.StageTypeExecute, consts.StageTypeRework:
			// 调度器用独立 ctx，生命周期不跟随启动流程
			if err := taskScheduler.Start(context.Background(), workflowRunID); err != nil {
				g.Log().Warningf(ctx, "[WorkflowRecovery] 重启调度失败: workflowRunID=%d stage=%s status=%s err=%v",
					workflowRunID, stage, status, err)
				continue
			}
			restartedSchedulers++
		default:
			g.Log().Infof(ctx, "[WorkflowRecovery] 恢复 runtime: workflowRunID=%d stage=%s status=%s",
				workflowRunID, stage, status)
		}
	}

	g.Log().Infof(ctx, "[WorkflowRecovery] 恢复活跃工作流完成: total=%d restartedSchedulers=%d",
		len(runs), restartedSchedulers)
	return nil
}
