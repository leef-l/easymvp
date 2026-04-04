// Package execute 管理执行阶段：蓝图实例化 → domain_task → 调度执行。
package execute

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/engine"
	domainTask "easymvp/app/mvp/internal/workflow/domain/task"
	"easymvp/app/mvp/internal/workflow/scheduler"
)

// StageCompleter 阶段完成回调（避免循环依赖）。
type StageCompleter interface {
	CompleteStage(ctx context.Context, stageRunID int64) error
	FailStage(ctx context.Context, stageRunID int64, reason string) error
}

// Service 执行阶段服务。
type Service struct {
	taskSvc        *domainTask.TaskService
	scheduler      *scheduler.DomainTaskScheduler
	stageCompleter StageCompleter
}

// NewService 创建执行阶段服务。
func NewService(ts *domainTask.TaskService, sched *scheduler.DomainTaskScheduler, sc StageCompleter) *Service {
	return &Service{
		taskSvc:        ts,
		scheduler:      sched,
		stageCompleter: sc,
	}
}

// InstantiateAndStart 将审核通过的蓝图实例化为领域任务并启动调度。
func (s *Service) InstantiateAndStart(ctx context.Context, stageRunID int64, planVersionID int64) error {
	// 获取 workflow_run_id
	stageRun, err := g.DB().Model("mvp_stage_run").Ctx(ctx).Where("id", stageRunID).One()
	if err != nil || stageRun.IsEmpty() {
		return fmt.Errorf("stage_run(%d) 不存在", stageRunID)
	}
	workflowRunID := stageRun["workflow_run_id"].Int64()

	// 1. 实例化蓝图为领域任务
	taskCount, err := s.taskSvc.InstantiateFromBlueprint(ctx, planVersionID, stageRunID, workflowRunID)
	if err != nil {
		return fmt.Errorf("蓝图实例化失败: %w", err)
	}
	if taskCount == 0 {
		return fmt.Errorf("没有生成任何领域任务")
	}

	g.Log().Infof(ctx, "[ExecuteStage] 实例化 %d 个领域任务, stageRunID=%d, planVersionID=%d", taskCount, stageRunID, planVersionID)

	// 2. 注册执行器
	s.scheduler.SetExecutor(&domainTaskExecutor{workflowRunID: workflowRunID, scheduler: s.scheduler})

	// 3. 注册完成回调
	finalStageRunID := stageRunID
	s.scheduler.SetCompletionCallback(func(ctx context.Context, wfRunID int64) {
		g.Log().Infof(ctx, "[ExecuteStage] 所有任务完成, workflowRunID=%d", wfRunID)

		// 压缩上下文
		projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
			Where("id", wfRunID).Value("project_id")
		if projectID.Int64() > 0 {
			_ = engine.GetCompressor().CompressProjectContext(ctx, projectID.Int64())
		}

		// 完成 execute stage
		if s.stageCompleter != nil {
			_ = s.stageCompleter.CompleteStage(ctx, finalStageRunID)
		}
	})

	// 4. 启动调度
	return s.scheduler.Start(ctx, workflowRunID)
}

// domainTaskExecutor 领域任务执行器，桥接旧 engine.Executor。
type domainTaskExecutor struct {
	workflowRunID int64
	scheduler     *scheduler.DomainTaskScheduler
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
	taskRecord, err := g.DB().Model("mvp_domain_task").Ctx(ctx).Where("id", taskID).One()
	if err != nil || taskRecord.IsEmpty() {
		e.scheduler.OnTaskFailed(ctx, taskID, "任务不存在")
		return
	}

	// 查 project_id
	projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("id", workflowRunID).Value("project_id")

	roleType := taskRecord["role_type"].String()
	executionMode := taskRecord["execution_mode"].String()
	modelID := taskRecord["model_id"].Int64()

	// 获取模型信息
	modelInfo, err := engine.ResolveModelInfo(ctx, projectID.Int64(), roleType, modelID)
	if err != nil {
		e.handleFailure(ctx, taskID, err.Error())
		return
	}

	switch executionMode {
	case "aider":
		e.executeWithAider(ctx, projectID.Int64(), taskID, taskRecord, modelInfo)
	default:
		e.executeWithChat(ctx, projectID.Int64(), taskID, taskRecord, modelInfo)
	}
}

// executeWithAider Aider 执行。
func (e *domainTaskExecutor) executeWithAider(ctx context.Context, projectID, taskID int64, task gdb.Record, modelInfo *engine.ModelInfo) {
	project, _ := g.DB().Model("mvp_project").Ctx(ctx).Where("id", projectID).One()
	workDir := project["work_dir"].String()

	// 解析 affected_resources 作为文件列表
	var files []string
	resJSON := task["affected_resources"].String()
	if resJSON != "" && resJSON != "[]" && resJSON != "null" {
		json.Unmarshal([]byte(resJSON), &files)
	}

	runner := engine.GetAiderRunner()
	aiderCfg := runner.BuildConfigFromModel(ctx, modelInfo, workDir)
	aiderResult := runner.RunTask(ctx, projectID, taskID, modelInfo, task["description"].String(), workDir, files, nil)
	if aiderResult.Error != nil {
		e.handleFailure(ctx, taskID, aiderResult.Error.Error())
		return
	}
	_ = aiderCfg // 配置已在 RunTask 中使用
	e.handleSuccess(ctx, taskID, aiderResult.Output)
}

// executeWithChat ChatStream 执行。
func (e *domainTaskExecutor) executeWithChat(ctx context.Context, projectID, taskID int64, task gdb.Record, modelInfo *engine.ModelInfo) {
	// 创建或获取任务对话
	convID, err := engine.EnsureDomainTaskConversation(ctx, projectID, taskID, task["role_type"].String(), task["name"].String())
	_ = modelInfo // chat 模式由 ChatEngine 内部解析模型
	if err != nil {
		e.handleFailure(ctx, taskID, err.Error())
		return
	}

	// 发送任务描述到对话
	_, _, err = engine.GetEngine().SendMessage(ctx, convID, task["description"].String(), 0, 0)
	if err != nil {
		e.handleFailure(ctx, taskID, err.Error())
		return
	}

	e.handleSuccess(ctx, taskID, "chat execution completed")
}

// handleSuccess 任务成功。
func (e *domainTaskExecutor) handleSuccess(ctx context.Context, taskID int64, result string) {
	now := g.Map{
		"status":       "completed",
		"result":       result,
		"completed_at": g.Map{"completed_at": "NOW()"},
		"updated_at":   g.Map{"updated_at": "NOW()"},
	}
	_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).Where("status", "running").
		Update(now)
	_ = e.scheduler.OnTaskCompleted(ctx, taskID)
}

// handleFailure 任务失败。
func (e *domainTaskExecutor) handleFailure(ctx context.Context, taskID int64, errMsg string) {
	_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).Where("status", "running").
		Update(g.Map{
			"status":     "failed",
			"result":     errMsg,
			"updated_at": g.Map{"updated_at": "NOW()"},
		})
	e.scheduler.OnTaskFailed(ctx, taskID, errMsg)
}
