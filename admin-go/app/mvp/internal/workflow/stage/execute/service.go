// Package execute 管理执行阶段：蓝图实例化 → domain_task → 调度执行。
package execute

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workspace"
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
	taskSvc             *domainTask.TaskService
	scheduler           *scheduler.DomainTaskScheduler
	stageCompleter      StageCompleter
	onAnalysisCompleted AnalysisCompletedFn
}

// NewService 创建执行阶段服务。
func NewService(ts *domainTask.TaskService, sched *scheduler.DomainTaskScheduler, sc StageCompleter) *Service {
	return &Service{
		taskSvc:        ts,
		scheduler:      sched,
		stageCompleter: sc,
	}
}

// SetAnalysisCompletedFn 注册 failure_analysis 完成后的回调。
func (s *Service) SetAnalysisCompletedFn(fn AnalysisCompletedFn) { s.onAnalysisCompleted = fn }

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
	s.scheduler.SetExecutor(&domainTaskExecutor{
		workflowRunID:       workflowRunID,
		scheduler:           s.scheduler,
		wsMgr:               workspace.NewGitWorktreeManager(),
		onAnalysisCompleted: s.onAnalysisCompleted,
	})

	// 3. 注册完成回调
	finalStageRunID := stageRunID
	wsMgr := workspace.NewGitWorktreeManager()
	s.scheduler.SetCompletionCallback(func(ctx context.Context, wfRunID int64) {
		g.Log().Infof(ctx, "[ExecuteStage] 所有任务完成, workflowRunID=%d", wfRunID)

		// 压缩上下文
		projectID, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
			Where("id", wfRunID).Value("project_id")
		if projectID.Int64() > 0 {
			_ = engine.GetCompressor().CompressProjectContext(ctx, projectID.Int64())
		}

		// 批量清理该 workflow 下所有已完成的 worktree
		go workspace.RunCleanup(context.Background(), wsMgr, workspace.DefaultCleanupConfig())

		// 完成 execute stage
		if s.stageCompleter != nil {
			_ = s.stageCompleter.CompleteStage(ctx, finalStageRunID)
		}
	})

	// 4. 启动调度
	return s.scheduler.Start(ctx, workflowRunID)
}

// AnalysisCompletedFn 分析任务完成回调（路由到 rework service）。
type AnalysisCompletedFn func(ctx context.Context, stageRunID, analysisTaskID int64) error

// domainTaskExecutor 领域任务执行器，桥接旧 engine.Executor。
type domainTaskExecutor struct {
	workflowRunID       int64
	scheduler           *scheduler.DomainTaskScheduler
	wsMgr               workspace.Manager
	onAnalysisCompleted AnalysisCompletedFn
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

	// 如果需要工作空间隔离，准备 worktree
	var ws *workspace.TaskWorkspace
	if workspace.NeedsIsolation(executionMode) && e.wsMgr != nil {
		project, _ := g.DB().Model("mvp_project").Ctx(ctx).Where("id", projectID.Int64()).One()
		workDir := project["work_dir"].String()
		ws, err = e.wsMgr.Prepare(ctx, workspace.PrepareRequest{
			TaskID:        taskID,
			WorkflowRunID: workflowRunID,
			ProjectID:     projectID.Int64(),
			WorkDir:       workDir,
		})
		if err != nil {
			// 隔离失败不降级，直接中断任务，防止污染主工作区
			e.handleFailure(ctx, taskID, fmt.Sprintf("workspace 隔离准备失败: %v", err))
			return
		}
	}

	switch executionMode {
	case "aider":
		e.executeWithAider(ctx, projectID.Int64(), taskID, taskRecord, modelInfo, ws)
	default:
		e.executeWithChat(ctx, projectID.Int64(), taskID, taskRecord, modelInfo)
	}
}

// executeWithAider Aider 执行。
func (e *domainTaskExecutor) executeWithAider(ctx context.Context, projectID, taskID int64, task gdb.Record, modelInfo *engine.ModelInfo, ws *workspace.TaskWorkspace) {
	project, _ := g.DB().Model("mvp_project").Ctx(ctx).Where("id", projectID).One()
	workDir := project["work_dir"].String()

	// 如果有 workspace 隔离，使用 worktree 路径
	if ws != nil {
		workDir = ws.WorkspacePath
		_ = e.wsMgr.MarkRunning(ctx, taskID)
		g.Log().Infof(ctx, "[domainTaskExecutor] 使用 worktree 隔离: task=%d path=%s", taskID, workDir)
	}

	// 解析 affected_resources 作为文件列表
	var files []string
	resJSON := task["affected_resources"].String()
	if resJSON != "" && resJSON != "[]" && resJSON != "null" {
		json.Unmarshal([]byte(resJSON), &files)
	}

	// 启动心跳 goroutine：定期更新 heartbeat_at，让 watchdog 知道任务还活着
	hbCtx, hbCancel := context.WithCancel(ctx)
	go touchHeartbeatLoop(hbCtx, taskID)
	defer hbCancel()

	runner := engine.GetAiderRunner()
	aiderCfg := runner.BuildConfigFromModel(ctx, modelInfo, workDir)
	aiderResult := runner.RunTask(ctx, projectID, taskID, modelInfo, task["description"].String(), workDir, files, nil)
	_ = aiderCfg // 配置已在 RunTask 中使用

	if aiderResult.Error != nil {
		// workspace finalize: 标记失败
		if ws != nil && e.wsMgr != nil {
			_ = e.wsMgr.Finalize(ctx, taskID, workspace.FinalizeRequest{
				Success: false,
				Error:   aiderResult.Error.Error(),
			})
		}
		e.handleFailure(ctx, taskID, aiderResult.Error.Error())
		return
	}

	// workspace finalize: 标记成功，然后异步清理
	if ws != nil && e.wsMgr != nil {
		if err := e.wsMgr.Finalize(ctx, taskID, workspace.FinalizeRequest{Success: true}); err != nil {
			g.Log().Warningf(ctx, "[domainTaskExecutor] workspace finalize 失败: task=%d err=%v", taskID, err)
		} else {
			go func() {
				if cleanErr := e.wsMgr.Cleanup(context.Background(), taskID); cleanErr != nil {
					g.Log().Warningf(ctx, "[domainTaskExecutor] workspace cleanup 失败: task=%d err=%v", taskID, cleanErr)
				}
			}()
		}
	}
	e.handleSuccess(ctx, taskID, aiderResult.Output)
}

// executeWithChat ChatStream 执行。
func (e *domainTaskExecutor) executeWithChat(ctx context.Context, projectID, taskID int64, task gdb.Record, modelInfo *engine.ModelInfo) {
	// 启动心跳
	hbCtx, hbCancel := context.WithCancel(ctx)
	go touchHeartbeatLoop(hbCtx, taskID)
	defer hbCancel()

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
	now := gtime.Now()
	_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).Where("status", "running").
		Update(g.Map{
			"status":       domainTask.StatusCompleted,
			"result":       result,
			"completed_at": now,
			"updated_at":   now,
		})

	// 检查是否为 failure_analysis 任务 → 路由到 rework OnAnalysisCompleted
	isAnalysis := false
	if e.onAnalysisCompleted != nil {
		task, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("id", taskID).Fields("task_kind, stage_run_id").One()
		if !task.IsEmpty() && task["task_kind"].String() == "failure_analysis" {
			isAnalysis = true
			stageRunID := task["stage_run_id"].Int64()
			if err := e.onAnalysisCompleted(ctx, stageRunID, taskID); err != nil {
				// 回调失败：回滚分析任务为 failed，让 watchdog 后续重试
				g.Log().Errorf(ctx, "[domainTaskExecutor] OnAnalysisCompleted 失败，回滚分析任务: task=%d err=%v", taskID, err)
				_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
					Where("id", taskID).
					Where("status", domainTask.StatusCompleted).
					Update(g.Map{
						"status":     domainTask.StatusFailed,
						"result":     fmt.Sprintf("rework 回调失败: %v", err),
						"updated_at": gtime.Now(),
					})
				e.scheduler.OnTaskFailed(ctx, taskID, "rework callback failed: "+err.Error())
				return
			}
		}
	}

	// 非 analysis 任务 或 analysis 回调成功：正常推进调度
	_ = isAnalysis
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

// touchHeartbeatLoop 定期更新 domain_task 的 heartbeat_at。
// 启动时立即写一次，之后每 30s 更新。
func touchHeartbeatLoop(ctx context.Context, taskID int64) {
	// 立即写一次心跳，消除启动空窗期
	_, _ = g.DB().Model("mvp_domain_task").
		Where("id", taskID).
		Where("status", domainTask.StatusRunning).
		Update(g.Map{"heartbeat_at": gtime.Now()})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = g.DB().Model("mvp_domain_task").
				Where("id", taskID).
				Where("status", domainTask.StatusRunning).
				Update(g.Map{"heartbeat_at": gtime.Now()})
		}
	}
}
