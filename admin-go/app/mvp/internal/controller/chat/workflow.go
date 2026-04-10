package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/app/mvp/internal/workflow/resourcepath"
	"easymvp/app/mvp/internal/workspace"
	"easymvp/utility/snowflake"
)

var Workflow = cWorkflow{}

type cWorkflow struct{}

type workflowArtifactResetOptions struct {
	PauseScheduler          bool
	CancelRuntime           bool
	DeleteDomainTasks       bool
	DeleteStageTasks        bool
	DeleteStageRuns         bool
	DeleteReviewIssues      bool
	DeleteAcceptRuns        bool
	DeleteTaskWorkspaces    bool
	CleanupPhysicalWorktree bool
	SupersedePlanVersions   bool
}

type domainTaskUpdateOptions struct {
	TaskID                   int64
	Name                     string
	Description              string
	RoleType                 string
	RoleLevel                string
	ExecutionMode            string
	BatchNo                  *int
	Sort                     *int
	AffectedResources        []string
	ReplaceAffectedResources bool
	RestartAfterUpdate       bool
	Reason                   string
}

// checkProjectOwnership 校验项目访问权限（支持 owner/同部门/超管三级）。
// 兼容别名：旧调用不需要改名。
func checkProjectOwnership(ctx context.Context, projectID int64) error {
	return middleware.CheckProjectAccess(ctx, projectID)
}

func latestWorkflowRunForProject(ctx context.Context, projectID int64) (gdb.Record, error) {
	wfRun, err := repo.NewWorkflowRunRepo().GetLatestByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("查询工作流运行失败: %w", err)
	}
	if wfRun == nil || wfRun.IsEmpty() {
		return nil, fmt.Errorf("项目 %d 没有工作流运行记录", projectID)
	}
	return wfRun, nil
}

func resolvePlanVersionForForceStage(ctx context.Context, projectID, workflowRunID, requestedPlanVersionID int64) (int64, error) {
	planVersionRepo := repo.NewPlanVersionRepo()

	if requestedPlanVersionID > 0 {
		record, err := planVersionRepo.GetByProjectAndIDStatuses(ctx, projectID, requestedPlanVersionID, []string{"draft", "active"}, "id")
		if err != nil {
			return 0, fmt.Errorf("查询方案版本失败: %w", err)
		}
		if len(record) == 0 {
			return 0, fmt.Errorf("方案版本 %d 不存在或不可用于重启阶段", requestedPlanVersionID)
		}
		return requestedPlanVersionID, nil
	}

	wfRun, err := repo.NewWorkflowRunRepo().GetByIDMap(ctx, workflowRunID, "active_plan_version_id")
	if err == nil && len(wfRun) > 0 && g.NewVar(wfRun["active_plan_version_id"]).Int64() > 0 {
		return g.NewVar(wfRun["active_plan_version_id"]).Int64(), nil
	}

	record, err := planVersionRepo.GetLatestByProjectStatuses(ctx, projectID, []string{"active", "draft"}, "id")
	if err != nil {
		return 0, fmt.Errorf("查询最新方案版本失败: %w", err)
	}
	if len(record) == 0 {
		return 0, fmt.Errorf("项目 %d 没有可用于重启的方案版本", projectID)
	}
	return g.NewVar(record["id"]).Int64(), nil
}

func resetWorkflowArtifacts(ctx context.Context, projectID, workflowRunID int64, opts workflowArtifactResetOptions) error {
	if workflowRunID == 0 {
		return fmt.Errorf("workflowRunID 不能为空")
	}

	var (
		acceptRunRepo     = repo.NewAcceptRunRepo()
		blueprintRepo     = repo.NewBlueprintRepo()
		domainTaskRepo    = repo.NewDomainTaskRepo()
		planVersionRepo   = repo.NewPlanVersionRepo()
		reviewIssueRepo   = repo.NewReviewIssueRepo()
		stageRunRepo      = repo.NewStageRunRepo()
		stageTaskRepo     = repo.NewStageTaskRepo()
		taskWorkspaceRepo = repo.NewTaskWorkspaceRepo()
	)

	if opts.PauseScheduler {
		if scheduler := orchestrator.GetTaskScheduler(); scheduler != nil {
			scheduler.Pause(ctx, workflowRunID)
		}
	}

	if opts.CancelRuntime {
		orchestrator.GetRuntimeManager().Cancel(workflowRunID)
	}

	now := gtime.Now()

	if opts.DeleteTaskWorkspaces {
		workspaces, err := taskWorkspaceRepo.ListByWorkflow(ctx, workflowRunID, "task_id")
		if err != nil {
			return fmt.Errorf("查询工作空间记录失败: %w", err)
		}

		if opts.CleanupPhysicalWorktree {
			wsMgr := workspace.NewGitWorktreeManager()
			for _, ws := range workspaces {
				taskID := g.NewVar(ws["task_id"]).Int64()
				if taskID == 0 {
					continue
				}
				if cleanErr := wsMgr.Cleanup(context.Background(), taskID); cleanErr != nil {
					g.Log().Warningf(ctx, "[WorkflowReset] 清理任务工作空间失败: task=%d err=%v", taskID, cleanErr)
				}
			}
		}

		if err := taskWorkspaceRepo.SoftDeleteByWorkflow(ctx, workflowRunID, now); err != nil {
			return fmt.Errorf("归档旧工作空间失败: %w", err)
		}
	}

	if opts.DeleteStageTasks {
		stageRuns, err := stageRunRepo.ListByWorkflow(ctx, workflowRunID)
		if err != nil {
			return fmt.Errorf("查询阶段实例失败: %w", err)
		}
		if len(stageRuns) > 0 {
			idList := make([]int64, 0, len(stageRuns))
			for _, item := range stageRuns {
				idList = append(idList, int64(item.Id))
			}
			if err := stageTaskRepo.SoftDeleteByStageRuns(ctx, idList, now); err != nil {
				return fmt.Errorf("归档旧阶段任务失败: %w", err)
			}
		}
	}

	if opts.DeleteDomainTasks {
		if err := domainTaskRepo.SoftDeleteByWorkflow(ctx, workflowRunID, now); err != nil {
			return fmt.Errorf("归档旧领域任务失败: %w", err)
		}
	}
	if opts.DeleteReviewIssues {
		if err := reviewIssueRepo.SoftDeleteByWorkflow(ctx, workflowRunID, now); err != nil {
			return fmt.Errorf("归档旧审核问题失败: %w", err)
		}
	}
	if opts.DeleteAcceptRuns {
		if err := acceptRunRepo.SoftDeleteByWorkflow(ctx, workflowRunID, now); err != nil {
			return fmt.Errorf("归档旧验收记录失败: %w", err)
		}
	}
	if opts.DeleteStageRuns {
		if err := stageRunRepo.SoftDeleteByWorkflow(ctx, workflowRunID, now); err != nil {
			return fmt.Errorf("归档旧阶段实例失败: %w", err)
		}
	}

	if opts.SupersedePlanVersions {
		planVersions, err := planVersionRepo.ListByProjectStatuses(ctx, projectID, []string{"draft", "active"}, "id")
		if err != nil {
			return fmt.Errorf("查询方案版本失败: %w", err)
		}
		if len(planVersions) > 0 {
			idList := make([]int64, 0, len(planVersions))
			for _, item := range planVersions {
				idList = append(idList, g.NewVar(item["id"]).Int64())
			}
			if err := planVersionRepo.UpdateByIDs(ctx, idList, g.Map{"status": "superseded", "updated_at": now}); err != nil {
				return fmt.Errorf("废弃方案版本失败: %w", err)
			}
			if err := blueprintRepo.UpdateByPlanVersionIDs(ctx, idList, g.Map{"blueprint_status": "superseded", "updated_at": now}); err != nil {
				return fmt.Errorf("废弃任务蓝图失败: %w", err)
			}
		}
	}

	return nil
}

func resetWorkflowExecutionArtifacts(ctx context.Context, projectID, workflowRunID int64) error {
	return resetWorkflowArtifacts(ctx, projectID, workflowRunID, workflowArtifactResetOptions{
		PauseScheduler:          true,
		CancelRuntime:           true,
		DeleteDomainTasks:       true,
		DeleteStageTasks:        true,
		DeleteReviewIssues:      true,
		DeleteAcceptRuns:        true,
		DeleteTaskWorkspaces:    true,
		CleanupPhysicalWorktree: true,
	})
}

func ensureFreshDesignStageRun(ctx context.Context, workflowRunID int64) (int64, error) {
	stageRunRepo := repo.NewStageRunRepo()

	existing, err := stageRunRepo.GetLatestByWorkflowTypeStatuses(ctx, workflowRunID, orchestrator.StageDesign, []string{"pending", "running"}, "id")
	if err != nil {
		return 0, fmt.Errorf("查询设计阶段失败: %w", err)
	}
	if existing != nil && !existing.IsEmpty() {
		return existing["id"].Int64(), nil
	}

	maxNo, err := stageRunRepo.GetMaxStageNoByWorkflow(ctx, workflowRunID)
	if err != nil {
		return 0, fmt.Errorf("查询 stage_no 失败: %w", err)
	}

	scope := repo.GetProjectScopeByWorkflowRun(ctx, workflowRunID)
	now := gtime.Now()
	stageRunID, err := stageRunRepo.Create(ctx, g.Map{
		"workflow_run_id": workflowRunID,
		"stage_type":      orchestrator.StageDesign,
		"stage_no":        maxNo + 1,
		"status":          "running",
		"created_by":      scope.CreatedBy,
		"dept_id":         scope.DeptID,
		"started_at":      now,
		"created_at":      now,
		"updated_at":      now,
	})
	if err != nil {
		return 0, fmt.Errorf("创建设计阶段失败: %w", err)
	}

	return stageRunID, nil
}

func preparePlanVersionForForceStage(ctx context.Context, projectID, workflowRunID, requestedPlanVersionID int64, targetStage string) (int64, error) {
	planVersionID, err := resolvePlanVersionForForceStage(ctx, projectID, workflowRunID, requestedPlanVersionID)
	if err != nil {
		return 0, err
	}
	if err := activatePlanVersionForForceStage(ctx, projectID, workflowRunID, planVersionID, targetStage); err != nil {
		return 0, err
	}
	return planVersionID, nil
}

func activatePlanVersionForForceStage(ctx context.Context, projectID, workflowRunID, planVersionID int64, targetStage string) error {
	now := gtime.Now()

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		otherPlanVersions, err := tx.Model("mvp_plan_version").Ctx(ctx).
			Where("project_id", projectID).
			WhereIn("status", g.Slice{"draft", "active"}).
			Where("id <>", planVersionID).
			WhereNull("deleted_at").
			Fields("id").
			Array()
		if err != nil {
			return fmt.Errorf("查询其他方案版本失败: %w", err)
		}

		record, err := tx.Model("mvp_plan_version").Ctx(ctx).
			Where("id", planVersionID).
			Where("project_id", projectID).
			WhereIn("status", g.Slice{"draft", "active"}).
			WhereNull("deleted_at").
			Fields("id, status").
			One()
		if err != nil {
			return fmt.Errorf("查询方案版本失败: %w", err)
		}
		if record.IsEmpty() {
			return fmt.Errorf("方案版本 %d 不存在或不可用于强制切换", planVersionID)
		}

		if len(otherPlanVersions) > 0 {
			otherIDs := make([]int64, 0, len(otherPlanVersions))
			for _, item := range otherPlanVersions {
				otherIDs = append(otherIDs, item.Int64())
			}
			if _, err := tx.Model("mvp_plan_version").Ctx(ctx).
				WhereIn("id", otherIDs).
				Update(g.Map{"status": "superseded", "updated_at": now}); err != nil {
				return fmt.Errorf("废弃其他方案版本失败: %w", err)
			}
			if _, err := tx.Model("mvp_task_blueprint").Ctx(ctx).
				WhereIn("plan_version_id", otherIDs).
				WhereNull("deleted_at").
				Update(g.Map{"blueprint_status": "superseded", "updated_at": now}); err != nil {
				return fmt.Errorf("废弃其他版本蓝图失败: %w", err)
			}
		}

		if _, err := tx.Model("mvp_task_blueprint").Ctx(ctx).
			Where("plan_version_id", planVersionID).
			Where("blueprint_status", "draft").
			Update(g.Map{"blueprint_status": "confirmed", "updated_at": now}); err != nil {
			return fmt.Errorf("确认任务蓝图失败: %w", err)
		}

		planUpdate := g.Map{
			"status":     "active",
			"updated_at": now,
		}
		switch targetStage {
		case "review":
			planUpdate["review_status"] = "pending"
			planUpdate["approved_at"] = nil
			planUpdate["rejected_at"] = nil
		case "execute":
			planUpdate["review_status"] = "approved"
			planUpdate["approved_at"] = now
			planUpdate["rejected_at"] = nil
		}

		if _, err := tx.Model("mvp_plan_version").Ctx(ctx).
			Where("id", planVersionID).
			Update(planUpdate); err != nil {
			return fmt.Errorf("更新方案版本状态失败: %w", err)
		}

		confirmedCount, err := tx.Model("mvp_task_blueprint").Ctx(ctx).
			Where("plan_version_id", planVersionID).
			Where("blueprint_status", "confirmed").
			WhereNull("deleted_at").
			Count()
		if err != nil {
			return fmt.Errorf("查询确认蓝图失败: %w", err)
		}
		if confirmedCount == 0 {
			return fmt.Errorf("方案版本 %d 没有可执行的确认蓝图", planVersionID)
		}

		if _, err := tx.Model("mvp_workflow_run").Ctx(ctx).
			Where("id", workflowRunID).
			Update(g.Map{"active_plan_version_id": planVersionID, "updated_at": now}); err != nil {
			return fmt.Errorf("回写 active_plan_version_id 失败: %w", err)
		}
		return nil
	})
}

func reopenWorkflowForTaskRestart(ctx context.Context, workflowRunID, taskID int64) error {
	wfRun, err := repo.NewWorkflowRunRepo().GetByIDMap(ctx, workflowRunID, "status", "current_stage", "current_stage_run_id")
	if err != nil {
		return fmt.Errorf("查询工作流状态失败: %w", err)
	}
	if len(wfRun) == 0 {
		return fmt.Errorf("workflow_run(%d) 不存在", workflowRunID)
	}

	status := g.NewVar(wfRun["status"]).String()
	currentStage := g.NewVar(wfRun["current_stage"]).String()
	if status == orchestrator.WorkflowExecuting && currentStage == orchestrator.StageExecute {
		if err := orchestrator.PrepareTaskSchedulerForStage(ctx, workflowRunID, orchestrator.StageExecute, g.NewVar(wfRun["current_stage_run_id"]).Int64()); err != nil {
			return fmt.Errorf("恢复 execute 调度器失败: %w", err)
		}
		if scheduler := orchestrator.GetTaskScheduler(); scheduler != nil {
			return scheduler.Start(context.Background(), workflowRunID)
		}
		return nil
	}

	reason := fmt.Sprintf("manual restart task %d", taskID)
	stageRunID, err := orchestrator.GetStageService().ForceStartStage(ctx, workflowRunID, orchestrator.StageExecute, reason)
	if err != nil {
		return fmt.Errorf("重开执行阶段失败: %w", err)
	}
	if err := orchestrator.PrepareTaskSchedulerForStage(ctx, workflowRunID, orchestrator.StageExecute, stageRunID); err != nil {
		return fmt.Errorf("绑定重启后的 execute 调度器失败: %w", err)
	}
	if scheduler := orchestrator.GetTaskScheduler(); scheduler != nil {
		return scheduler.Start(context.Background(), workflowRunID)
	}
	return nil
}

func resetDomainTaskExecutionArtifacts(ctx context.Context, taskID int64) error {
	now := gtime.Now()
	wsMgr := workspace.NewGitWorktreeManager()
	if cleanErr := wsMgr.Cleanup(context.Background(), taskID); cleanErr != nil {
		g.Log().Warningf(ctx, "[TaskReset] 清理旧工作空间失败: task=%d err=%v", taskID, cleanErr)
	}
	if err := repo.NewTaskWorkspaceRepo().SoftDeleteByTask(ctx, taskID, now); err != nil {
		return fmt.Errorf("归档旧工作空间失败: %w", err)
	}
	if err := repo.NewTaskResourceLockRepo().ReleaseHeldByTask(ctx, taskID, now); err != nil {
		return fmt.Errorf("释放旧资源锁失败: %w", err)
	}
	return nil
}

func updateDomainTaskInternal(ctx context.Context, projectID int64, opts domainTaskUpdateOptions) (res *v1.WorkflowUpdateDomainTaskRes, err error) {
	taskID := opts.TaskID
	taskRepo := repo.NewDomainTaskRepo()
	task, err := taskRepo.GetByProjectAndID(ctx, projectID, taskID, "t.id", "t.workflow_run_id", "t.status", "t.affected_resources")
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	if len(task) == 0 {
		return nil, fmt.Errorf("任务不存在")
	}

	workflowRunID := g.NewVar(task["workflow_run_id"]).Int64()
	currentStatus := g.NewVar(task["status"]).String()
	pausedForRestart := false
	defer func() {
		if err != nil && pausedForRestart {
			if scheduler := orchestrator.GetTaskScheduler(); scheduler != nil {
				_ = scheduler.Start(context.Background(), workflowRunID)
			}
		}
	}()
	if currentStatus == "running" && !opts.RestartAfterUpdate {
		return nil, fmt.Errorf("运行中的任务必须配合 restartAfterUpdate=true 一起修改")
	}
	if currentStatus == "running" {
		if scheduler := orchestrator.GetTaskScheduler(); scheduler != nil {
			scheduler.Pause(ctx, workflowRunID)
			pausedForRestart = true
		}
		currentStatus = "pending"
	}

	updateData := g.Map{
		"updated_at": gtime.Now(),
	}
	changed := false

	if name := strings.TrimSpace(opts.Name); name != "" {
		updateData["name"] = name
		changed = true
	}
	if desc := strings.TrimSpace(opts.Description); desc != "" {
		updateData["description"] = desc
		changed = true
	}
	if roleType := strings.TrimSpace(opts.RoleType); roleType != "" {
		updateData["role_type"] = roleType
		changed = true
	}
	if roleLevel := strings.TrimSpace(opts.RoleLevel); roleLevel != "" {
		updateData["role_level"] = roleLevel
		changed = true
	}
	if executionMode := strings.TrimSpace(opts.ExecutionMode); executionMode != "" {
		updateData["execution_mode"] = executionMode
		changed = true
	}
	if opts.BatchNo != nil {
		updateData["batch_no"] = *opts.BatchNo
		changed = true
	}
	if opts.Sort != nil {
		updateData["sort"] = *opts.Sort
		changed = true
	}
	if opts.ReplaceAffectedResources {
		currentResources, decodeErr := decodeAffectedResourcesJSON(g.NewVar(task["affected_resources"]).String())
		if decodeErr != nil {
			return nil, fmt.Errorf("解析任务现有 affected_resources 失败: %w", decodeErr)
		}
		if introduced := resourcepath.FindNewlyIntroducedGovernedRootFiles(currentResources, opts.AffectedResources); len(introduced) > 0 {
			return nil, fmt.Errorf("affectedResources 不允许新增受治理仓库文件: %s", strings.Join(introduced, ", "))
		}
		resJSON, jsonErr := json.Marshal(opts.AffectedResources)
		if jsonErr != nil {
			return nil, fmt.Errorf("序列化 affectedResources 失败: %w", jsonErr)
		}
		updateData["affected_resources"] = string(resJSON)
		changed = true
	}

	message := "任务已更新"
	if opts.RestartAfterUpdate {
		if err := resetDomainTaskExecutionArtifacts(ctx, taskID); err != nil {
			return nil, err
		}
		now := gtime.Now()
		updateData["status"] = "pending"
		updateData["result"] = nil
		updateData["error_message"] = nil
		updateData["started_at"] = nil
		updateData["completed_at"] = nil
		updateData["heartbeat_at"] = nil
		updateData["locked_resources"] = nil
		updateData["updated_at"] = now
		changed = true
		currentStatus = "pending"
		message = "任务已更新并重置为 pending"
	}

	if !changed {
		return nil, fmt.Errorf("没有可更新的字段")
	}

	if err := taskRepo.UpdateFields(ctx, taskID, updateData); err != nil {
		return nil, fmt.Errorf("更新任务失败: %w", err)
	}

	recordWorkflowEvent(ctx, workflowRunID, "task", "task.manual_updated", &taskID, nil, map[string]interface{}{
		"project_id":           projectID,
		"task_id":              taskID,
		"restart_after_update": opts.RestartAfterUpdate,
		"replace_resources":    opts.ReplaceAffectedResources,
		"reason":               opts.Reason,
	})

	if opts.RestartAfterUpdate {
		if err := reopenWorkflowForTaskRestart(ctx, workflowRunID, taskID); err != nil {
			return nil, err
		}
	}

	res = &v1.WorkflowUpdateDomainTaskRes{
		TaskID:  snowflake.JsonInt64(taskID),
		Status:  currentStatus,
		Message: message,
	}
	return res, nil
}

func decodeAffectedResourcesJSON(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var resources []string
	if err := json.Unmarshal([]byte(raw), &resources); err != nil {
		return nil, err
	}
	return resources, nil
}

func recordWorkflowEvent(ctx context.Context, workflowRunID int64, entityType, eventType string, entityID, stageRunID *int64, payload map[string]interface{}) {
	if err := event.PersistRecord(ctx, event.Event{
		WorkflowRunID: workflowRunID,
		StageRunID:    stageRunID,
		EntityType:    entityType,
		EntityID:      entityID,
		EventType:     eventType,
		Payload:       payload,
	}); err != nil {
		g.Log().Warningf(ctx, "[WorkflowEvent] 写入失败: event=%s workflowRunID=%d err=%v", eventType, workflowRunID, err)
	}
}

// CreateProject 创建项目
func (c *cWorkflow) CreateProject(ctx context.Context, req *v1.WorkflowCreateProjectReq) (res *v1.WorkflowCreateProjectRes, err error) {
	userID := middleware.GetUserID(ctx)
	deptID := middleware.GetDeptID(ctx)

	// 优先使用 categoryCode，通过 CategoryResolver 获取展示名
	projectCategory := req.ProjectCategory
	if req.CategoryCode != "" {
		resolver := engine.GetCategoryResolver()
		catInfo, _ := resolver.ResolveByCode(ctx, req.CategoryCode)
		if catInfo != nil {
			projectCategory = catInfo.DisplayName
		}
	}
	if projectCategory == "" {
		projectCategory = "软件开发"
	}

	// 提取用户选择的预设 ID 列表
	var selectedPresetIDs []int64
	for _, sr := range req.SelectedRoles {
		if int64(sr.PresetID) > 0 {
			selectedPresetIDs = append(selectedPresetIDs, int64(sr.PresetID))
		}
	}

	projectID, convID, err := engine.CreateProject(ctx, req.Name, projectCategory, req.Description, req.WorkDir, int64(req.ArchitectModelID), userID, deptID, selectedPresetIDs, req.EngineVersion)
	if err != nil {
		return nil, err
	}

	wfSvc := orchestrator.GetWorkflowService()
	wfRunID, err := wfSvc.CreateRun(ctx, projectID)
	if err != nil {
		g.Log().Warningf(ctx, "[CreateProject] CreateRun 失败: %v", err)
	} else {
		if err2 := wfSvc.StartDesign(ctx, wfRunID); err2 != nil {
			g.Log().Warningf(ctx, "[CreateProject] StartDesign 失败: %v", err2)
		}
	}

	return &v1.WorkflowCreateProjectRes{
		ProjectID:      snowflake.JsonInt64(projectID),
		ConversationID: snowflake.JsonInt64(convID),
		WorkflowRunID:  snowflake.JsonInt64(wfRunID),
	}, nil
}

// ConfirmPlan 确认实施方案。
// 每次确认前先清理该项目所有旧的执行数据（domain_task、stage_run 等），
// 确保每次确认都像第一次一样干净。
func (c *cWorkflow) ConfirmPlan(ctx context.Context, req *v1.WorkflowConfirmPlanReq) (res *v1.WorkflowConfirmPlanRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	now := gtime.Now()
	workflowRunRepo := repo.NewWorkflowRunRepo()
	projectRepo := repo.NewProjectRepo()

	// 清理旧的执行数据，让每次确认方案都从干净状态开始
	wfRun, wfErr := workflowRunRepo.GetLatestByProjectExcludingStatuses(ctx, projectID, []string{"completed", "canceled"}, "id")
	if wfErr != nil {
		g.Log().Warningf(ctx, "[ConfirmPlan] 查询活跃 workflow_run 失败: projectID=%d err=%v", projectID, wfErr)
	}
	if len(wfRun) > 0 {
		wfRunID := g.NewVar(wfRun["id"]).Int64()
		if resetErr := resetWorkflowArtifacts(ctx, projectID, wfRunID, workflowArtifactResetOptions{
			PauseScheduler:          true,
			CancelRuntime:           true,
			DeleteDomainTasks:       true,
			DeleteStageTasks:        true,
			DeleteStageRuns:         true,
			DeleteReviewIssues:      true,
			DeleteAcceptRuns:        true,
			DeleteTaskWorkspaces:    true,
			CleanupPhysicalWorktree: true,
		}); resetErr != nil {
			g.Log().Errorf(ctx, "[ConfirmPlan] 清理工作流执行数据失败: projectID=%d wfRunID=%d err=%v", projectID, wfRunID, resetErr)
		}

		stageRunID, stageErr := ensureFreshDesignStageRun(ctx, wfRunID)
		if stageErr != nil {
			return nil, stageErr
		}

		// workflow_run 回到 designing（SubmitForReviewAsync 会再改成 reviewing）
		if wfErr := workflowRunRepo.UpdateFields(ctx, wfRunID, g.Map{
			"status":               "designing",
			"current_stage":        "design",
			"current_stage_run_id": stageRunID,
			"pause_reason":         nil,
			"status_before_pause":  nil,
			"finished_at":          nil,
			"updated_at":           now,
		}); wfErr != nil {
			g.Log().Errorf(ctx, "[ConfirmPlan] 回退 workflow_run 状态失败: wfRun=%d err=%v", wfRunID, wfErr)
		}
		// project 状态也先回到 designing
		if pErr := projectRepo.UpdateFields(ctx, projectID, g.Map{"status": "designing", "pause_reason": nil, "updated_at": now}); pErr != nil {
			g.Log().Errorf(ctx, "[ConfirmPlan] 回退 project 状态失败: project=%d err=%v", projectID, pErr)
		}

		orchestrator.GetRuntimeManager().Create(wfRunID, projectID)

		g.Log().Infof(ctx, "[ConfirmPlan] 已清理旧执行数据: projectID=%d wfRunID=%d", projectID, wfRunID)
	}

	submitErr := orchestrator.GetPlanVersionService().SubmitForReviewAsync(ctx, projectID)
	if submitErr != nil {
		return nil, submitErr
	}

	return &v1.WorkflowConfirmPlanRes{
		Submitted:    true,
		ReviewStatus: "pending",
		StageStatus:  "pending",
		Message:      "方案已提交审核，请稍候查看审核进度",
	}, nil
}

// Pause 暂停项目
func (c *cWorkflow) Pause(ctx context.Context, req *v1.WorkflowPauseReq) (res *v1.WorkflowPauseRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, qErr := repo.NewWorkflowRunRepo().GetLatestByProjectExcludingStatuses(ctx, projectID, []string{"completed", "canceled", "paused"}, "id")
	if qErr != nil {
		return nil, fmt.Errorf("查询工作流运行失败: %w", qErr)
	}
	if len(wfRun) == 0 {
		return nil, fmt.Errorf("没有活跃的工作流运行")
	}
	wfRunID := g.NewVar(wfRun["id"]).Int64()

	wfSvc := orchestrator.GetWorkflowService()
	if err := wfSvc.Pause(ctx, wfRunID, req.PauseReason); err != nil {
		return nil, err
	}
	orchestrator.GetTaskScheduler().Pause(ctx, wfRunID)
	return &v1.WorkflowPauseRes{}, nil
}

// ResetToDesign 回到设计阶段（暂停状态下可用）。
// 清理已有的方案、蓝图、领域任务、阶段实例、worktree，项目回到 designing 状态。
func (c *cWorkflow) ResetToDesign(ctx context.Context, req *v1.WorkflowResetToDesignReq) (res *v1.WorkflowResetToDesignRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	projectRepo := repo.NewProjectRepo()
	workflowRunRepo := repo.NewWorkflowRunRepo()

	// 只允许 paused 或 designing 状态
	project, err := projectRepo.GetByID(ctx, projectID, "status")
	if err != nil || len(project) == 0 {
		return nil, fmt.Errorf("项目不存在")
	}
	status := g.NewVar(project["status"]).String()
	if status != "paused" && status != "designing" {
		return nil, fmt.Errorf("当前状态(%s)不允许回到设计阶段，请先暂停项目", status)
	}

	now := gtime.Now()

	// 查活跃 workflow_run
	wfRun, wfErr := workflowRunRepo.GetLatestByProjectExcludingStatuses(ctx, projectID, []string{"completed", "canceled"}, "id")
	if wfErr != nil {
		return nil, fmt.Errorf("查询活跃 workflow_run 失败: %w", wfErr)
	}

	if len(wfRun) == 0 {
		wfSvc := orchestrator.GetWorkflowService()
		newWorkflowRunID, createErr := wfSvc.CreateRun(ctx, projectID)
		if createErr != nil {
			return nil, fmt.Errorf("创建新的设计工作流失败: %w", createErr)
		}
		if startErr := wfSvc.StartDesign(ctx, newWorkflowRunID); startErr != nil {
			return nil, fmt.Errorf("启动新的设计阶段失败: %w", startErr)
		}
		if pErr := projectRepo.UpdateFields(ctx, projectID, g.Map{"status": "designing", "pause_reason": nil, "updated_at": now}); pErr != nil {
			g.Log().Errorf(ctx, "[ResetToDesign] project 重置失败: projectID=%d err=%v", projectID, pErr)
		}
		g.Log().Infof(ctx, "[ResetToDesign] 项目已回到设计阶段并创建新 workflow_run: projectID=%d wfRunID=%d", projectID, newWorkflowRunID)
		return &v1.WorkflowResetToDesignRes{Message: "已回到设计阶段，可重新拆分方案"}, nil
	}

	if len(wfRun) > 0 {
		wfRunID := g.NewVar(wfRun["id"]).Int64()
		if resetErr := resetWorkflowArtifacts(ctx, projectID, wfRunID, workflowArtifactResetOptions{
			PauseScheduler:          true,
			CancelRuntime:           true,
			DeleteDomainTasks:       true,
			DeleteStageTasks:        true,
			DeleteStageRuns:         true,
			DeleteReviewIssues:      true,
			DeleteAcceptRuns:        true,
			DeleteTaskWorkspaces:    true,
			CleanupPhysicalWorktree: true,
			SupersedePlanVersions:   true,
		}); resetErr != nil {
			g.Log().Errorf(ctx, "[ResetToDesign] 重置工作流执行数据失败: projectID=%d wfRunID=%d err=%v", projectID, wfRunID, resetErr)
		}

		// 7. 重建 design stage_run，确保回到 designing 后链路仍可继续运行
		newStageRunID, stageErr := ensureFreshDesignStageRun(ctx, wfRunID)
		if stageErr != nil {
			return nil, stageErr
		}

		// 8. workflow_run 回到 designing
		if wfUpErr := workflowRunRepo.UpdateFields(ctx, wfRunID, g.Map{
			"status":                 "designing",
			"current_stage":          "design",
			"current_stage_run_id":   newStageRunID,
			"active_plan_version_id": nil,
			"pause_reason":           nil,
			"status_before_pause":    nil,
			"finished_at":            nil,
			"updated_at":             now,
		}); wfUpErr != nil {
			g.Log().Errorf(ctx, "[ResetToDesign] workflow_run 重置失败: wfRunID=%d err=%v", wfRunID, wfUpErr)
		}
		orchestrator.GetRuntimeManager().Create(wfRunID, projectID)
	}

	// 9. project 回到 designing
	if pErr := projectRepo.UpdateFields(ctx, projectID, g.Map{"status": "designing", "pause_reason": nil, "updated_at": now}); pErr != nil {
		g.Log().Errorf(ctx, "[ResetToDesign] project 重置失败: projectID=%d err=%v", projectID, pErr)
	}

	g.Log().Infof(ctx, "[ResetToDesign] 项目已回到设计阶段: projectID=%d", projectID)
	return &v1.WorkflowResetToDesignRes{Message: "已回到设计阶段，可重新拆分方案"}, nil
}

// Resume 恢复项目
func (c *cWorkflow) Resume(ctx context.Context, req *v1.WorkflowResumeReq) (res *v1.WorkflowResumeRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, qErr := repo.NewWorkflowRunRepo().GetLatestByProjectStatuses(ctx, projectID, []string{"paused"}, "id", "current_stage")
	if qErr != nil {
		return nil, fmt.Errorf("查询暂停的工作流失败: %w", qErr)
	}
	if len(wfRun) == 0 {
		return nil, fmt.Errorf("没有暂停的工作流运行")
	}
	wfRunID := g.NewVar(wfRun["id"]).Int64()

	wfSvc := orchestrator.GetWorkflowService()
	if err := wfSvc.Resume(ctx, wfRunID); err != nil {
		return nil, err
	}
	// 恢复后启动调度器（execute 和 rework 阶段都需要调度任务）
	currentStage := g.NewVar(wfRun["current_stage"]).String()
	if currentStage == "execute" || currentStage == "rework" {
		_ = orchestrator.GetTaskScheduler().Start(context.Background(), wfRunID)
	}
	return &v1.WorkflowResumeRes{}, nil
}

// RetryTask 重新执行失败任务
func (c *cWorkflow) RetryTask(ctx context.Context, req *v1.WorkflowRetryTaskReq) (res *v1.WorkflowRetryTaskRes, err error) {
	projectID := int64(req.ProjectID)
	taskID := int64(req.TaskID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	taskRepo := repo.NewDomainTaskRepo()
	task, err := taskRepo.GetByProjectAndID(ctx, projectID, taskID, "t.workflow_run_id")
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	if len(task) == 0 {
		return nil, fmt.Errorf("任务不存在")
	}
	if err := resetDomainTaskExecutionArtifacts(ctx, taskID); err != nil {
		return nil, err
	}

	rows, err := taskRepo.ResetForRetry(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, fmt.Errorf("任务(%d)不在 failed/escalated 状态，无法重试", taskID)
	}
	if err := reopenWorkflowForTaskRestart(ctx, g.NewVar(task["workflow_run_id"]).Int64(), taskID); err != nil {
		return nil, err
	}
	return &v1.WorkflowRetryTaskRes{}, nil
}

// ForceStage 强制跳转/重启到指定阶段。
func (c *cWorkflow) ForceStage(ctx context.Context, req *v1.WorkflowForceStageReq) (res *v1.WorkflowForceStageRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	workflowRunID := wfRun["id"].Int64()
	targetStage := strings.TrimSpace(req.TargetStage)
	if targetStage == "rework" && int64(req.FailedTaskID) == 0 {
		return nil, fmt.Errorf("rework 阶段必须提供 failedTaskID")
	}

	switch targetStage {
	case "design":
		if err := resetWorkflowArtifacts(ctx, projectID, workflowRunID, workflowArtifactResetOptions{
			PauseScheduler:          true,
			CancelRuntime:           true,
			DeleteDomainTasks:       true,
			DeleteStageTasks:        true,
			DeleteStageRuns:         true,
			DeleteReviewIssues:      true,
			DeleteAcceptRuns:        true,
			DeleteTaskWorkspaces:    true,
			CleanupPhysicalWorktree: true,
			SupersedePlanVersions:   true,
		}); err != nil {
			return nil, err
		}
	case "review", "execute":
		if err := resetWorkflowExecutionArtifacts(ctx, projectID, workflowRunID); err != nil {
			return nil, err
		}
	case "accept", "rework":
		if scheduler := orchestrator.GetTaskScheduler(); scheduler != nil {
			scheduler.Pause(ctx, workflowRunID)
		}
		orchestrator.GetRuntimeManager().Cancel(workflowRunID)
	}

	stageSvc := orchestrator.GetStageService()
	stageRunID, err := stageSvc.ForceStartStage(ctx, workflowRunID, targetStage, strings.TrimSpace(req.Reason))
	if err != nil {
		return nil, err
	}

	switch targetStage {
	case "review":
		planVersionID, err := preparePlanVersionForForceStage(ctx, projectID, workflowRunID, int64(req.PlanVersionID), targetStage)
		if err != nil {
			_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
			return nil, err
		}
		go func() {
			bgCtx := context.Background()
			if runErr := orchestrator.GetReviewStageService().RunReview(bgCtx, stageRunID, planVersionID); runErr != nil {
				g.Log().Errorf(bgCtx, "[ForceStage] review 重启失败: workflowRunID=%d stageRunID=%d err=%v", workflowRunID, stageRunID, runErr)
				_ = stageSvc.FailStage(bgCtx, stageRunID, runErr.Error())
			}
		}()
	case "execute":
		planVersionID, err := preparePlanVersionForForceStage(ctx, projectID, workflowRunID, int64(req.PlanVersionID), targetStage)
		if err != nil {
			_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
			return nil, err
		}
		if err := orchestrator.GetExecuteStageService().InstantiateAndStart(ctx, stageRunID, planVersionID); err != nil {
			_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
			return nil, fmt.Errorf("重启执行阶段失败: %w", err)
		}
	case "design":
		if err := repo.NewWorkflowRunRepo().UpdateFields(ctx, workflowRunID, g.Map{"active_plan_version_id": nil, "updated_at": gtime.Now()}); err != nil {
			g.Log().Warningf(ctx, "[ForceStage] 清空 active_plan_version_id 失败: workflowRunID=%d err=%v", workflowRunID, err)
		}
	case "accept":
		go func() {
			bgCtx := context.Background()
			if runErr := orchestrator.GetAcceptStageService().Run(bgCtx, workflowRunID, stageRunID); runErr != nil {
				g.Log().Errorf(bgCtx, "[ForceStage] accept 重启失败: workflowRunID=%d stageRunID=%d err=%v", workflowRunID, stageRunID, runErr)
				_ = stageSvc.FailStage(bgCtx, stageRunID, runErr.Error())
			}
		}()
	case "rework":
		failedTaskID := int64(req.FailedTaskID)
		if failedTaskID > 0 {
			sourceStage := strings.TrimSpace(req.SourceStage)
			if sourceStage == "" {
				sourceStage = "execute"
			}
			if err := orchestrator.GetReworkStageService().HandleReworkWithSource(ctx, stageRunID, failedTaskID, sourceStage); err != nil {
				_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
				return nil, fmt.Errorf("重启返工阶段失败: %w", err)
			}
			if err := orchestrator.ActivateReworkStage(ctx, workflowRunID, stageRunID); err != nil {
				_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
				return nil, fmt.Errorf("重启返工调度器失败: %w", err)
			}
		}
	}

	recordWorkflowEvent(ctx, workflowRunID, "workflow", "workflow.force_stage", &workflowRunID, &stageRunID, map[string]interface{}{
		"project_id":      projectID,
		"target_stage":    targetStage,
		"plan_version_id": int64(req.PlanVersionID),
		"failed_task_id":  int64(req.FailedTaskID),
		"reason":          req.Reason,
	})

	return &v1.WorkflowForceStageRes{
		WorkflowRunID: snowflake.JsonInt64(workflowRunID),
		StageRunID:    snowflake.JsonInt64(stageRunID),
		CurrentStage:  targetStage,
	}, nil
}

// Cancel 人工取消工作流。
func (c *cWorkflow) Cancel(ctx context.Context, req *v1.WorkflowCancelReq) (res *v1.WorkflowCancelRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := repo.NewWorkflowRunRepo().GetLatestByProjectExcludingStatuses(ctx, projectID, []string{"completed", "canceled"}, "id")
	if err != nil {
		return nil, fmt.Errorf("查询可取消的工作流失败: %w", err)
	}
	if len(wfRun) == 0 {
		return nil, fmt.Errorf("项目 %d 没有可取消的工作流", projectID)
	}

	workflowRunID := g.NewVar(wfRun["id"]).Int64()
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "manual cancel"
	}

	if err := orchestrator.GetWorkflowService().Cancel(ctx, workflowRunID, reason); err != nil {
		return nil, err
	}

	recordWorkflowEvent(ctx, workflowRunID, "workflow", "workflow.canceled", &workflowRunID, nil, map[string]interface{}{
		"project_id": projectID,
		"reason":     reason,
	})

	return &v1.WorkflowCancelRes{
		WorkflowRunID: snowflake.JsonInt64(workflowRunID),
		Message:       "工作流已取消",
	}, nil
}

// SkipTask 跳过失败任务（防止批次永久阻塞）
func (c *cWorkflow) SkipTask(ctx context.Context, req *v1.WorkflowSkipTaskReq) (res *v1.WorkflowSkipTaskRes, err error) {
	projectID := int64(req.ProjectID)
	taskID := int64(req.TaskID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	now := gtime.Now()

	rows, err := repo.NewDomainTaskRepo().CompleteAsSkipped(ctx, taskID, now)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, fmt.Errorf("任务不在可跳过的状态")
	}
	if completeErr := orchestrator.GetTaskScheduler().OnTaskCompleted(ctx, taskID); completeErr != nil {
		g.Log().Warningf(ctx, "[SkipTask] 通知调度器任务完成失败: task=%d err=%v", taskID, completeErr)
	}
	return &v1.WorkflowSkipTaskRes{}, nil
}

// UpdateDomainTask 人工修改领域任务，必要时可直接重置为 pending 并重新调度。
func (c *cWorkflow) UpdateDomainTask(ctx context.Context, req *v1.WorkflowUpdateDomainTaskReq) (res *v1.WorkflowUpdateDomainTaskRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	return updateDomainTaskInternal(ctx, projectID, domainTaskUpdateOptions{
		TaskID:                   int64(req.TaskID),
		Name:                     req.Name,
		Description:              req.Description,
		RoleType:                 req.RoleType,
		RoleLevel:                req.RoleLevel,
		ExecutionMode:            req.ExecutionMode,
		BatchNo:                  req.BatchNo,
		Sort:                     req.Sort,
		AffectedResources:        req.AffectedResources,
		ReplaceAffectedResources: req.ReplaceAffectedResources,
		RestartAfterUpdate:       req.RestartAfterUpdate,
		Reason:                   req.Reason,
	})
}

// ParseTasks 手动解析架构师回复中的任务清单（托底机制）
// dryRun=true 时仅检查不创建，dryRun=false 时实际创建草案任务
func (c *cWorkflow) ParseTasks(ctx context.Context, req *v1.WorkflowParseTasksReq) (res *v1.WorkflowParseTasksRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 查找该项目的架构师对话
	conv, err := g.DB().Ctx(ctx).Model("mvp_conversation").
		Where("project_id", projectID).
		Where("role_type", "architect").
		Where("task_id IS NULL OR task_id = 0").
		WhereNull("deleted_at").
		One()
	if err != nil || conv.IsEmpty() {
		return &v1.WorkflowParseTasksRes{HasTasks: false, TaskCount: 0}, nil
	}

	// 收集对话中所有 completed 的 assistant 回复（任务可能分散在多轮"继续"对话中）
	convID := conv["id"].Int64()

	allMsgs, err := g.DB().Ctx(ctx).Model("mvp_message").
		Where("conversation_id", convID).
		Where("role", "assistant").
		Where("status", "completed").
		WhereNull("deleted_at").
		OrderAsc("created_at").
		All()
	if err != nil || len(allMsgs) == 0 {
		return &v1.WorkflowParseTasksRes{HasTasks: false, TaskCount: 0}, nil
	}

	// 拼接所有 assistant 消息内容
	var allReplies strings.Builder
	var lastMsgID int64
	for i, m := range allMsgs {
		content := m["content"].String()
		if strings.TrimSpace(content) == "" {
			continue
		}
		if i > 0 {
			allReplies.WriteString("\n\n---\n\n")
		}
		allReplies.WriteString(content)
		lastMsgID = m["id"].Int64()
	}
	aiReply := allReplies.String()
	_ = lastMsgID

	if req.DryRun {
		count := engine.GetParser().DryParseTaskCount(aiReply)
		// count > 0: 正则提取成功，精确数量
		// count == -1: 有内容但需要 AI 提取，前端显示为"检测到任务内容"
		// count == 0: 确实没有任务
		return &v1.WorkflowParseTasksRes{
			HasTasks:  count != 0,
			TaskCount: count,
		}, nil
	}

	// V2 主路径：先正则快速提取，失败则异步走 AI 二次提取
	projectRecord, projErr := repo.NewProjectRepo().GetByID(ctx, projectID, "project_category", "name", "description")
	if projErr != nil {
		g.Log().Warningf(ctx, "[ParseTasks] 查询项目信息失败: projectID=%d err=%v", projectID, projErr)
	}
	projectCategory := g.NewVar(projectRecord["project_category"]).String()
	projectName := g.NewVar(projectRecord["name"]).String()
	projectDesc := g.NewVar(projectRecord["description"]).String()
	latestUserMsg, userMsgErr := g.DB().Ctx(ctx).Model("mvp_message").
		Where("conversation_id", convID).
		Where("role", "user").
		Where("status", "completed").
		WhereNull("deleted_at").
		OrderDesc("created_at").
		One()
	if userMsgErr != nil {
		g.Log().Warningf(ctx, "[ParseTasks] 查询最近用户消息失败: convID=%d err=%v", convID, userMsgErr)
	}
	extractionInput := buildTaskExtractionInput(
		projectName,
		projectDesc,
		latestUserMsg["content"].String(),
		aiReply,
	)

	g.Log().Infof(ctx, "[ParseTasks] 开始提取: projectID=%d aiReplyLen=%d convID=%d lastMsgID=%d",
		projectID, len([]rune(aiReply)), convID, lastMsgID)

	// 快速正则提取（毫秒级）
	fastTasks, fastReport, fastErr := engine.GetParser().FastExtractWithReport(ctx, aiReply, projectCategory)
	if fastErr != nil {
		g.Log().Warningf(ctx, "[ParseTasks] FastExtractWithReport 错误: projectID=%d err=%v", projectID, fastErr)
	}
	if fastReport != nil && fastReport.HasBlockingIssue() {
		engine.NotifyProjectArchitectConversation(ctx, projectID, fastReport.BuildContinuationPrompt())
		return &v1.WorkflowParseTasksRes{
			HasTasks:  false,
			TaskCount: 0,
			Message:   "任务清单分段/数量不一致，已自动要求架构师补齐缺失块并重发修正块",
		}, nil
	}
	if len(fastTasks) > 0 {
		g.Log().Infof(ctx, "[ParseTasks] FastExtract 成功: projectID=%d summary=%s", projectID, fastReport.Summary())
		count, err := createBlueprints(ctx, projectID, convID, lastMsgID, fastTasks)
		if err != nil {
			g.Log().Errorf(ctx, "[ParseTasks] createBlueprints 失败: projectID=%d err=%v", projectID, err)
			return &v1.WorkflowParseTasksRes{
				HasTasks: false, TaskCount: 0,
				Message: fmt.Sprintf("创建蓝图失败: %v", err),
			}, nil
		}
		g.Log().Infof(ctx, "[ParseTasks] 创建蓝图成功: projectID=%d count=%d", projectID, count)
		return &v1.WorkflowParseTasksRes{HasTasks: count > 0, TaskCount: count}, nil
	}

	g.Log().Infof(ctx, "[ParseTasks] FastExtract 无结果，启动异步 AI 提取: projectID=%d", projectID)

	// 正则提取失败，异步走 AI 二次提取
	go func() {
		bgCtx := context.Background()
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(bgCtx, "[ParseTasks] AI 异步提取 panic: projectID=%d err=%v", projectID, r)
			}
		}()

		tasks, report, err := engine.GetParser().ExtractAndNormalizeWithReport(bgCtx, extractionInput, projectCategory)
		if report != nil && report.HasBlockingIssue() {
			g.Log().Warningf(bgCtx, "[ParseTasks] AI 异步提取发现阻断问题: projectID=%d summary=%s", projectID, report.Summary())
			engine.NotifyProjectArchitectConversation(bgCtx, projectID, report.BuildContinuationPrompt())
			return
		}
		if err != nil || len(tasks) == 0 {
			g.Log().Warningf(bgCtx, "[ParseTasks] AI 异步提取失败或无结果: projectID=%d err=%v", projectID, err)
			engine.NotifyProjectArchitectConversation(bgCtx, projectID,
				"## 任务提取失败\n\n未能从回复中自动提取任务清单。请让架构师用标准 JSON 格式（`{\"tasks\": [...]}`）重新输出任务列表。")
			return
		}

		count, createErr := createBlueprints(bgCtx, projectID, convID, lastMsgID, tasks)
		if createErr != nil {
			g.Log().Errorf(bgCtx, "[ParseTasks] AI 提取后创建蓝图失败: projectID=%d err=%v", projectID, createErr)
			return
		}
		g.Log().Infof(bgCtx, "[ParseTasks] AI 异步提取成功: projectID=%d count=%d", projectID, count)
		engine.NotifyProjectArchitectConversation(bgCtx, projectID,
			fmt.Sprintf("## 任务提取完成\n\n已从回复中提取 %d 个任务蓝图，请检查后确认方案。", count))
	}()

	return &v1.WorkflowParseTasksRes{
		HasTasks:  true,
		TaskCount: 0,
		Message:   "正在通过 AI 提取任务，请稍候刷新查看",
	}, nil
}

func buildTaskExtractionInput(projectName, projectDesc, latestUserPrompt, aiReply string) string {
	parts := make([]string, 0, 4)
	if strings.TrimSpace(projectName) != "" || strings.TrimSpace(projectDesc) != "" {
		parts = append(parts, fmt.Sprintf("=== 项目目标 ===\n项目名称：%s\n项目简介：%s", strings.TrimSpace(projectName), strings.TrimSpace(projectDesc)))
	}
	if strings.TrimSpace(latestUserPrompt) != "" {
		parts = append(parts, "=== 最近用户要求 ===\n"+strings.TrimSpace(latestUserPrompt))
	}
	if strings.TrimSpace(aiReply) != "" {
		parts = append(parts, "=== 架构师回复 ===\n"+strings.TrimSpace(aiReply))
	}
	return strings.Join(parts, "\n\n")
}

// RolePresets 获取角色预设列表（前端创建项目时读取默认模型）
func (c *cWorkflow) RolePresets(ctx context.Context, req *v1.WorkflowRolePresetsReq) (res *v1.WorkflowRolePresetsRes, err error) {
	presets, err := repo.ListRolePresets(ctx, repo.RolePresetQuery{
		CategoryCode:     req.CategoryCode,
		ProjectCategory:  req.ProjectCategory,
		DefaultOnly:      !req.All,
		IncludeModelName: true,
	})
	if err != nil {
		return nil, err
	}

	list := make([]v1.RolePresetItem, 0, len(presets))
	for _, p := range presets {
		list = append(list, v1.RolePresetItem{
			ID:            snowflake.JsonInt64(p["id"].Int64()),
			RoleType:      p["role_type"].String(),
			RoleLevel:     p["role_level"].String(),
			ModelID:       snowflake.JsonInt64(p["model_id"].Int64()),
			ModelName:     p["model_name"].String(),
			ExecutionMode: p["execution_mode"].String(),
			SystemPrompt:  p["system_prompt"].String(),
			IsDefault:     p["is_default"].Bool(),
		})
	}

	return &v1.WorkflowRolePresetsRes{List: list}, nil
}

// Categories 获取项目分类列表（前端创建项目时选择分类）
func (c *cWorkflow) Categories(ctx context.Context, req *v1.WorkflowCategoriesReq) (res *v1.WorkflowCategoriesRes, err error) {
	records, err := repo.NewProjectCategoryRepo().ListAll(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]v1.CategoryItem, 0, len(records))
	for _, r := range records {
		list = append(list, v1.CategoryItem{
			CategoryCode: g.NewVar(r["category_code"]).String(),
			DisplayName:  g.NewVar(r["display_name"]).String(),
			FamilyCode:   g.NewVar(r["family_code"]).String(),
			Description:  g.NewVar(r["description"]).String(),
		})
	}
	return &v1.WorkflowCategoriesRes{List: list}, nil
}

// isFollowUpMessage 判断是否为续写/跟进类短消息（"继续"、"截断了"等）。
func isFollowUpMessage(content string) bool {
	followUps := []string{"继续", "接着", "下一部分", "截断", "断了", "go on", "continue", "next"}
	lower := strings.ToLower(content)
	for _, kw := range followUps {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
