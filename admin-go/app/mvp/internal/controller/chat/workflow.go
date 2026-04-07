package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/workflow/autonomy"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

var Workflow = cWorkflow{}

type cWorkflow struct{}

const projectRuntimeSnapshotFreshWindow = 2 * time.Minute

type projectRuntimeTaskStat struct {
	WorkflowRunID  int64 `orm:"workflow_run_id"`
	TotalTasks     int   `orm:"total_tasks"`
	CompletedTasks int   `orm:"completed_tasks"`
	FailedTasks    int   `orm:"failed_tasks"`
	RunningTasks   int   `orm:"running_tasks"`
}

type projectRuntimeSnapshot struct {
	CreatedAt *gtime.Time
	Situation autonomy.Situation
}

type projectRuntimeLatestID struct {
	ProjectID int64 `orm:"project_id"`
	ID        int64 `orm:"id"`
}

type workflowRuntimeSnapshotLatestID struct {
	WorkflowRunID int64 `orm:"workflow_run_id"`
	ID            int64 `orm:"id"`
}

// checkProjectOwnership 校验项目访问权限（支持 owner/同部门/超管三级）。
// 兼容别名：旧调用不需要改名。
func checkProjectOwnership(ctx context.Context, projectID int64) error {
	return middleware.CheckProjectAccess(ctx, projectID)
}

func loadLatestWorkflowRuns(ctx context.Context, projectIDs []int64) (map[int64]gdb.Record, error) {
	result := make(map[int64]gdb.Record, len(projectIDs))
	if len(projectIDs) == 0 {
		return result, nil
	}

	// workflow_run 使用雪花 ID 递增写入，MAX(id) 可稳定代表项目下最近一次创建的 run。
	var latestIDs []projectRuntimeLatestID
	if err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		WhereIn("project_id", projectIDs).
		WhereNull("deleted_at").
		Fields("project_id, MAX(id) AS id").
		Group("project_id").
		Scan(&latestIDs); err != nil {
		return nil, err
	}
	if len(latestIDs) == 0 {
		return result, nil
	}

	runIDs := make([]int64, 0, len(latestIDs))
	for _, item := range latestIDs {
		if item.ID > 0 {
			runIDs = append(runIDs, item.ID)
		}
	}
	if len(runIDs) == 0 {
		return result, nil
	}

	runs, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		WhereIn("id", runIDs).
		WhereNull("deleted_at").
		Fields("id, project_id, current_stage, status").
		All()
	if err != nil {
		return nil, err
	}
	for _, run := range runs {
		result[run["project_id"].Int64()] = run
	}
	return result, nil
}

func loadLatestSituationSnapshots(ctx context.Context, workflowRunIDs []int64) (map[int64]*projectRuntimeSnapshot, error) {
	result := make(map[int64]*projectRuntimeSnapshot, len(workflowRunIDs))
	if len(workflowRunIDs) == 0 {
		return result, nil
	}

	var latestIDs []workflowRuntimeSnapshotLatestID
	if err := g.DB().Model("mvp_situation_snapshot").Ctx(ctx).
		WhereIn("workflow_run_id", workflowRunIDs).
		WhereNull("deleted_at").
		Fields("workflow_run_id, MAX(id) AS id").
		Group("workflow_run_id").
		Scan(&latestIDs); err != nil {
		return nil, err
	}
	if len(latestIDs) == 0 {
		return result, nil
	}

	snapshotIDs := make([]int64, 0, len(latestIDs))
	for _, item := range latestIDs {
		if item.ID > 0 {
			snapshotIDs = append(snapshotIDs, item.ID)
		}
	}
	if len(snapshotIDs) == 0 {
		return result, nil
	}

	snapshots, err := g.DB().Model("mvp_situation_snapshot").Ctx(ctx).
		WhereIn("id", snapshotIDs).
		WhereNull("deleted_at").
		Fields("id, workflow_run_id, snapshot_data, created_at").
		All()
	if err != nil {
		return nil, err
	}

	for _, item := range snapshots {
		var sit autonomy.Situation
		if err := json.Unmarshal([]byte(item["snapshot_data"].String()), &sit); err != nil {
			continue
		}
		result[item["workflow_run_id"].Int64()] = &projectRuntimeSnapshot{
			CreatedAt: item["created_at"].GTime(),
			Situation: sit,
		}
	}
	return result, nil
}

func loadTaskStats(ctx context.Context, workflowRunIDs []int64) (map[int64]projectRuntimeTaskStat, error) {
	result := make(map[int64]projectRuntimeTaskStat, len(workflowRunIDs))
	if len(workflowRunIDs) == 0 {
		return result, nil
	}

	var rows []projectRuntimeTaskStat
	if err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		WhereIn("workflow_run_id", workflowRunIDs).
		WhereNull("deleted_at").
		Fields(`
			workflow_run_id,
			COUNT(*) AS total_tasks,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS completed_tasks,
			SUM(CASE WHEN status IN ('failed', 'escalated') THEN 1 ELSE 0 END) AS failed_tasks,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) AS running_tasks`).
		Group("workflow_run_id").
		Scan(&rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.WorkflowRunID] = row
	}
	return result, nil
}

func shouldUseRuntimeSnapshot(snapshot *projectRuntimeSnapshot, workflowStatus string) bool {
	if snapshot == nil || snapshot.Situation.Progress == nil {
		return false
	}
	if snapshot.Situation.WorkflowStatus == workflowStatus {
		switch workflowStatus {
		case "completed", "failed", "canceled", "paused":
			return true
		}
	}
	if snapshot.CreatedAt == nil {
		return false
	}
	ageMillis := gtime.Now().TimestampMilli() - snapshot.CreatedAt.TimestampMilli()
	if ageMillis < 0 {
		ageMillis = 0
	}
	return time.Duration(ageMillis)*time.Millisecond <= projectRuntimeSnapshotFreshWindow
}

func taskStatFromProgress(progress *autonomy.ProgressMetrics) projectRuntimeTaskStat {
	if progress == nil {
		return projectRuntimeTaskStat{}
	}
	return projectRuntimeTaskStat{
		TotalTasks:     progress.TotalTasks,
		CompletedTasks: progress.CompletedTasks,
		FailedTasks:    progress.FailedTasks,
		RunningTasks:   progress.RunningTasks,
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

// ConfirmPlan 确认实施方案
func (c *cWorkflow) ConfirmPlan(ctx context.Context, req *v1.WorkflowConfirmPlanReq) (res *v1.WorkflowConfirmPlanRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
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

	wfRun, qErr := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNotIn("status", g.Slice{"completed", "canceled", "paused"}).
		WhereNull("deleted_at").OrderDesc("run_no").One()
	if qErr != nil {
		return nil, fmt.Errorf("查询工作流运行失败: %w", qErr)
	}
	if wfRun.IsEmpty() {
		return nil, fmt.Errorf("没有活跃的工作流运行")
	}
	wfRunID := wfRun["id"].Int64()

	wfSvc := orchestrator.GetWorkflowService()
	if err := wfSvc.Pause(ctx, wfRunID, req.PauseReason); err != nil {
		return nil, err
	}
	orchestrator.GetTaskScheduler().Pause(ctx, wfRunID)
	return &v1.WorkflowPauseRes{}, nil
}

// Resume 恢复项目
func (c *cWorkflow) Resume(ctx context.Context, req *v1.WorkflowResumeReq) (res *v1.WorkflowResumeRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, qErr := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		Where("status", "paused").
		WhereNull("deleted_at").OrderDesc("run_no").One()
	if qErr != nil {
		return nil, fmt.Errorf("查询暂停的工作流失败: %w", qErr)
	}
	if wfRun.IsEmpty() {
		return nil, fmt.Errorf("没有暂停的工作流运行")
	}
	wfRunID := wfRun["id"].Int64()

	wfSvc := orchestrator.GetWorkflowService()
	if err := wfSvc.Resume(ctx, wfRunID); err != nil {
		return nil, err
	}
	// 恢复后启动调度器（execute 和 rework 阶段都需要调度任务）
	currentStage := wfRun["current_stage"].String()
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

	result, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).WhereIn("status", g.Slice{"failed", "escalated"}).
		Update(g.Map{
			"status":        "pending",
			"retry_count":   gdb.Raw("retry_count + 1"),
			"result":        nil,
			"error_message": nil,
			"updated_at":    gdb.Raw("NOW()"),
		})
	if err != nil {
		return nil, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("任务(%d)不在 failed/escalated 状态，无法重试", taskID)
	}
	return &v1.WorkflowRetryTaskRes{}, nil
}

// SkipTask 跳过失败任务（防止批次永久阻塞）
func (c *cWorkflow) SkipTask(ctx context.Context, req *v1.WorkflowSkipTaskReq) (res *v1.WorkflowSkipTaskRes, err error) {
	projectID := int64(req.ProjectID)
	taskID := int64(req.TaskID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	result, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).
		WhereIn("status", g.Slice{"pending", "failed", "escalated"}).
		Update(g.Map{
			"status":       "completed",
			"result":       "skipped",
			"completed_at": gdb.Raw("NOW()"),
			"updated_at":   gdb.Raw("NOW()"),
		})
	if err != nil {
		return nil, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("任务不在可跳过的状态")
	}
	if completeErr := orchestrator.GetTaskScheduler().OnTaskCompleted(ctx, taskID); completeErr != nil {
		g.Log().Warningf(ctx, "[SkipTask] 通知调度器任务完成失败: task=%d err=%v", taskID, completeErr)
	}
	return &v1.WorkflowSkipTaskRes{}, nil
}

// ParseTasks 手动解析架构师回复中的任务清单（托底机制）
// dryRun=true 时仅检查不创建，dryRun=false 时实际创建草案任务
func (c *cWorkflow) ParseTasks(ctx context.Context, req *v1.WorkflowParseTasksReq) (res *v1.WorkflowParseTasksRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 查找该项目的架构师对话
	conv, err := g.DB().Model("mvp_conversation").
		Where("project_id", projectID).
		Where("role_type", "architect").
		Where("task_id IS NULL OR task_id = 0").
		Where("deleted_at IS NULL").
		One()
	if err != nil || conv.IsEmpty() {
		return &v1.WorkflowParseTasksRes{HasTasks: false, TaskCount: 0}, nil
	}

	// 查找最新一轮方案：从最后一条 user 消息之后的所有 assistant 回复
	convID := conv["id"].Int64()

	// 先找最后一条 user 消息的时间，作为"最新一轮"的起点
	// 但需要跳过"继续"/"截断"等续写指令，找到真正的需求消息
	userMsgs, err := g.DB().Model("mvp_message").
		Where("conversation_id", convID).
		Where("role", "user").
		Where("status", "completed").
		Where("deleted_at IS NULL").
		OrderDesc("created_at").
		Limit(10).
		All()
	if err != nil || len(userMsgs) == 0 {
		return &v1.WorkflowParseTasksRes{HasTasks: false, TaskCount: 0}, nil
	}

	// 找到真正的需求消息（跳过"继续"、"截断"等短消息）
	var anchorTime string
	for _, um := range userMsgs {
		content := strings.TrimSpace(um["content"].String())
		// 短消息（<10字）且是续写指令，跳过
		if len([]rune(content)) < 10 && isFollowUpMessage(content) {
			continue
		}
		anchorTime = um["created_at"].String()
		break
	}
	if anchorTime == "" {
		// 全是短消息，用最早的那条
		anchorTime = userMsgs[len(userMsgs)-1]["created_at"].String()
	}

	// 取该 user 消息之后的所有 assistant 回复（即最新一轮方案，可能跨多条消息）
	allMsgs, err := g.DB().Model("mvp_message").
		Where("conversation_id", convID).
		Where("role", "assistant").
		Where("status", "completed").
		Where("deleted_at IS NULL").
		Where("created_at >= ?", anchorTime).
		OrderAsc("created_at").
		All()
	if err != nil || len(allMsgs) == 0 {
		return &v1.WorkflowParseTasksRes{HasTasks: false, TaskCount: 0}, nil
	}

	// 拼接最新一轮的 assistant 消息内容
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
	projectCategory, _ := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).Value("project_category")

	g.Log().Infof(ctx, "[ParseTasks] 开始提取: projectID=%d aiReplyLen=%d convID=%d lastMsgID=%d",
		projectID, len([]rune(aiReply)), convID, lastMsgID)

	// 快速正则提取（毫秒级）
	plan, fastErr := engine.GetParser().FastExtract(aiReply)
	if fastErr != nil {
		g.Log().Warningf(ctx, "[ParseTasks] FastExtract 错误: projectID=%d err=%v", projectID, fastErr)
	}
	if plan != nil && len(plan.Tasks) > 0 {
		tasks := engine.GetParser().NormalizeTasks(ctx, plan.Tasks, projectCategory.String())
		g.Log().Infof(ctx, "[ParseTasks] FastExtract 成功: projectID=%d rawTasks=%d normalized=%d",
			projectID, len(plan.Tasks), len(tasks))
		if len(tasks) > 0 {
			count, err := createBlueprints(ctx, projectID, convID, lastMsgID, tasks)
			if err != nil {
				g.Log().Errorf(ctx, "[ParseTasks] createBlueprints 失败: projectID=%d err=%v", projectID, err)
				return nil, err
			}
			return &v1.WorkflowParseTasksRes{HasTasks: count > 0, TaskCount: count}, nil
		}
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

		tasks, err := engine.GetParser().AIExtractTasks(bgCtx, aiReply, projectCategory.String())
		if err != nil || len(tasks) == 0 {
			g.Log().Warningf(bgCtx, "[ParseTasks] AI 异步提取失败或无结果: projectID=%d err=%v", projectID, err)
			engine.NotifyProjectArchitectConversation(bgCtx, projectID,
				"## 任务提取失败\n\n未能从回复中自动提取任务清单。请让架构师用标准 JSON 格式（`{\"tasks\": [...]}`）重新输出任务列表。")
			return
		}

		normalized := engine.GetParser().NormalizeTasks(bgCtx, tasks, projectCategory.String())
		count, createErr := createBlueprints(bgCtx, projectID, convID, lastMsgID, normalized)
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
	records, err := g.DB().Model("mvp_project_category").Ctx(ctx).
		Where("status", 1).
		WhereNull("deleted_at").
		Fields("category_code, display_name, family_code, description").
		OrderAsc("sort").
		All()
	if err != nil {
		return nil, err
	}

	list := make([]v1.CategoryItem, 0, len(records))
	for _, r := range records {
		list = append(list, v1.CategoryItem{
			CategoryCode: r["category_code"].String(),
			DisplayName:  r["display_name"].String(),
			FamilyCode:   r["family_code"].String(),
			Description:  r["description"].String(),
		})
	}
	return &v1.WorkflowCategoriesRes{List: list}, nil
}

// ProjectStatus 获取项目状态
func (c *cWorkflow) ProjectStatus(ctx context.Context, req *v1.WorkflowProjectStatusReq) (res *v1.WorkflowProjectStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	project, err := g.DB().Model("mvp_project").Where("id", projectID).Where("deleted_at IS NULL").One()
	if err != nil {
		return nil, err
	}
	if project.IsEmpty() {
		return nil, fmt.Errorf("项目不存在")
	}

	return projectStatusV2(ctx, project)
}

// projectStatusV2 V2 引擎的项目状态聚合。
func projectStatusV2(ctx context.Context, project gdb.Record) (*v1.WorkflowProjectStatusRes, error) {
	projectID := project["id"].Int64()

	var wfStatus, currentStage string
	var progressPercent int
	var totalTasks, completedTasks, failedTasks, runningTasks int

	wfRuns, err := loadLatestWorkflowRuns(ctx, []int64{projectID})
	if err != nil {
		return nil, err
	}
	wfRun := wfRuns[projectID]
	if !wfRun.IsEmpty() {
		wfRunID := wfRun["id"].Int64()
		wfStatus = wfRun["status"].String()
		currentStage = wfRun["current_stage"].String()

		stats := projectRuntimeTaskStat{}
		snapshots, snapshotErr := loadLatestSituationSnapshots(ctx, []int64{wfRunID})
		if snapshotErr != nil {
			g.Log().Warningf(ctx, "[ProjectStatus] 读取态势快照失败，回退实时聚合: workflowRunID=%d, err=%v", wfRunID, snapshotErr)
		}
		if snapshot := snapshots[wfRunID]; shouldUseRuntimeSnapshot(snapshot, wfStatus) {
			stats = taskStatFromProgress(snapshot.Situation.Progress)
			if currentStage == "" {
				currentStage = snapshot.Situation.ActiveStage
			}
		} else {
			taskStats, taskErr := loadTaskStats(ctx, []int64{wfRunID})
			if taskErr != nil {
				return nil, taskErr
			}
			stats = taskStats[wfRunID]
		}

		totalTasks = stats.TotalTasks
		completedTasks = stats.CompletedTasks
		failedTasks = stats.FailedTasks
		runningTasks = stats.RunningTasks
		if totalTasks > 0 {
			progressPercent = completedTasks * 100 / totalTasks
		}
	}

	// 蓝图状态统计（设计阶段用）
	type StatusCount struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	var counts []StatusCount
	if scanErr := g.DB().Model("mvp_task_blueprint AS bp").
		InnerJoin("mvp_plan_version AS pv", "pv.id = bp.plan_version_id").
		Where("pv.project_id", projectID).
		WhereIn("pv.status", g.Slice{"draft", "active"}).
		WhereNull("bp.deleted_at").
		Fields("bp.blueprint_status AS status, COUNT(*) AS count").
		Group("bp.blueprint_status").
		Scan(&counts); scanErr != nil {
		g.Log().Warningf(ctx, "[ProjectStatus] 蓝图统计查询失败: project=%d err=%v", projectID, scanErr)
	}

	statusCounts := make(map[string]int)
	bpTotal := 0
	for _, sc := range counts {
		statusCounts[sc.Status] = sc.Count
		bpTotal += sc.Count
	}

	if totalTasks > 0 {
		statusCounts["domain_total"] = totalTasks
		statusCounts["domain_completed"] = completedTasks
		statusCounts["domain_failed"] = failedTasks
		statusCounts["domain_running"] = runningTasks
	}

	displayTotal := bpTotal
	if totalTasks > 0 {
		displayTotal = totalTasks
	}

	res := &v1.WorkflowProjectStatusRes{
		Status:          project["status"].String(),
		PauseReason:     project["pause_reason"].String(),
		TotalTasks:      displayTotal,
		StatusCounts:    statusCounts,
		EngineVersion:   "workflow_v2",
		WorkflowStatus:  wfStatus,
		CurrentStage:    currentStage,
		ProgressPercent: progressPercent,
	}

	if wfStatus != "" {
		res.Status = wfStatus
	}

	return res, nil
}

// SystemCheck 系统配置检测
func (c *cWorkflow) SystemCheck(ctx context.Context, req *v1.SystemCheckReq) (res *v1.SystemCheckRes, err error) {
	items := make([]v1.SystemCheckItem, 0, 12)
	allPass := true

	addItem := func(key, name, link, status, message string) {
		if status != "ok" {
			allPass = false
		}
		items = append(items, v1.SystemCheckItem{
			Key: key, Name: name, Status: status, Message: message, Link: link,
		})
	}

	// 1. AI 供应商
	count, e := g.DB().Model("ai_provider").
		Where("status", 1).Where("base_url != ''").Where("deleted_at IS NULL").Count()
	if e != nil {
		addItem("ai_provider", "AI 供应商", "/ai/provider", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_provider", "AI 供应商", "/ai/provider", "error", "未配置启用的 AI 供应商（需要有 base_url）")
	} else {
		addItem("ai_provider", "AI 供应商", "/ai/provider", "ok", fmt.Sprintf("已有 %d 个启用供应商", count))
	}

	// 2. AI 套餐
	count, e = g.DB().Model("ai_plan").
		Where("status", 1).Where("api_key != ''").Where("deleted_at IS NULL").Count()
	if e != nil {
		addItem("ai_plan", "AI 套餐", "/ai/plan", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_plan", "AI 套餐", "/ai/plan", "error", "未配置启用的 AI 套餐（需要有 api_key）")
	} else {
		addItem("ai_plan", "AI 套餐", "/ai/plan", "ok", fmt.Sprintf("已有 %d 个启用套餐", count))
	}

	// 3. 架构师模型
	count, e = g.DB().Model("ai_model").
		Where("capability", "architect").Where("status", 1).Where("deleted_at IS NULL").Count()
	if e != nil {
		addItem("ai_model_architect", "AI 模型（架构师）", "/ai/model", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_model_architect", "AI 模型（架构师）", "/ai/model", "error", "未配置 capability=architect 且启用的 AI 模型")
	} else {
		addItem("ai_model_architect", "AI 模型（架构师）", "/ai/model", "ok", fmt.Sprintf("已有 %d 个架构师模型", count))
	}

	// 4. 实施员模型
	count, e = g.DB().Model("ai_model").
		WhereIn("capability", g.Slice{"implementer", "coding", "chat"}).
		Where("status", 1).Where("deleted_at IS NULL").Count()
	if e != nil {
		addItem("ai_model_implementer", "AI 模型（实施员）", "/ai/model", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_model_implementer", "AI 模型（实施员）", "/ai/model", "error", "未配置 capability 为 implementer/coding/chat 且启用的 AI 模型")
	} else {
		addItem("ai_model_implementer", "AI 模型（实施员）", "/ai/model", "ok", fmt.Sprintf("已有 %d 个实施员模型", count))
	}

	// 5. 角色预设
	architectCount, _ := repo.CountRolePresets(ctx, repo.RolePresetQuery{RoleType: "architect"})
	implementerCount, _ := repo.CountRolePresets(ctx, repo.RolePresetQuery{RoleType: "implementer"})
	if architectCount == 0 || implementerCount == 0 {
		addItem("role_preset", "角色预设", "/mvp/role-preset", "error",
			fmt.Sprintf("缺少角色预设：架构师=%d，实施员=%d（各需至少 1 条）", architectCount, implementerCount))
	} else {
		addItem("role_preset", "角色预设", "/mvp/role-preset", "ok",
			fmt.Sprintf("架构师预设 %d 条，实施员预设 %d 条", architectCount, implementerCount))
	}

	// 6. AI 执行引擎
	count, e = g.DB().Model("ai_engine").
		Where("status", 1).Where("deleted_at IS NULL").Count()
	if e != nil {
		addItem("ai_engine", "AI 执行引擎", "/ai/engine", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_engine", "AI 执行引擎", "/ai/engine", "error", "未配置启用的 AI 执行引擎")
	} else {
		addItem("ai_engine", "AI 执行引擎", "/ai/engine", "ok", fmt.Sprintf("已有 %d 个启用引擎", count))
	}

	// 7. Aider 引擎配置
	aiderCfg, e := g.DB().Model("ai_engine_config").
		Where("engine_code", "aider").Where("deleted_at IS NULL").One()
	if e != nil || aiderCfg.IsEmpty() {
		addItem("ai_engine_config_aider", "Aider 引擎配置", "/ai/engine", "error", "未配置 Aider 引擎参数")
	} else if aiderCfg["workspace_root"].String() == "" {
		addItem("ai_engine_config_aider", "Aider 引擎配置", "/ai/engine", "warning", "Aider 引擎未配置 workspace_root（工作区根目录）")
	} else {
		addItem("ai_engine_config_aider", "Aider 引擎配置", "/ai/engine", "ok",
			"工作区根目录: "+aiderCfg["workspace_root"].String())
	}

	// 8. OpenHands 引擎配置
	ohCfg, e := g.DB().Model("ai_engine_config").
		Where("engine_code", "openhands").Where("deleted_at IS NULL").One()
	if e != nil || ohCfg.IsEmpty() {
		addItem("ai_engine_config_openhands", "OpenHands 引擎配置", "/ai/engine", "warning", "未配置 OpenHands 引擎参数（非必须，仅使用 Aider 可忽略）")
	} else if ohCfg["command_template"].String() == "" {
		addItem("ai_engine_config_openhands", "OpenHands 引擎配置", "/ai/engine", "warning", "OpenHands 未配置 command_template（命令模板）")
	} else {
		addItem("ai_engine_config_openhands", "OpenHands 引擎配置", "/ai/engine", "ok", "命令模板已配置")
	}

	// 9. 角色引擎授权
	count, e = g.DB().Model("system_role_ai_engine").Count()
	if e != nil {
		addItem("role_ai_engine", "角色引擎授权", "", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("role_ai_engine", "角色引擎授权", "", "error", "没有角色被授权使用 AI 引擎，请在角色管理中配置")
	} else {
		addItem("role_ai_engine", "角色引擎授权", "", "ok", fmt.Sprintf("已有 %d 条角色引擎授权", count))
	}

	// 10. Aider 执行环境
	if aiderPath, err := exec.LookPath("aider"); err == nil {
		addItem("aider_installed", "Aider 执行环境", "", "ok", "aider 已安装: "+aiderPath)
	} else if uvPath, uvErr := exec.LookPath("uv"); uvErr == nil {
		addItem("aider_installed", "Aider 执行环境", "", "ok", "本机未安装 aider，将通过 uv 自动安装/执行: "+uvPath)
	} else if dockerPath, dockerErr := exec.LookPath("docker"); dockerErr == nil {
		addItem("aider_installed", "Aider 执行环境", "", "warning", "本机未安装 aider/uv，将回退使用 Docker 执行: "+dockerPath)
	} else {
		addItem("aider_installed", "Aider 执行环境", "", "error", "未找到 aider 可执行文件，且 uv/docker 都不可用")
	}

	// 11. OpenHands 执行环境
	openHandsNeedsDocker := false
	openHandsMessage := "OpenHands 当前可通过 HTTP 接口工作，未强制依赖 Docker。"
	if !ohCfg.IsEmpty() {
		commandTemplate := strings.TrimSpace(ohCfg["command_template"].String())
		if commandTemplate != "" {
			lowerCommand := strings.ToLower(commandTemplate)
			if strings.Contains(lowerCommand, "docker run") || strings.Contains(lowerCommand, " docker ") {
				openHandsNeedsDocker = true
				openHandsMessage = "OpenHands 命令模板依赖 Docker 运行。"
			} else {
				openHandsMessage = "OpenHands 命令模板已配置，当前不依赖 Docker。"
			}
		}
	}

	if dockerPath, err := exec.LookPath("docker"); err == nil {
		addItem("docker_installed", "OpenHands 执行环境", "", "ok", openHandsMessage+" docker 已安装: "+dockerPath)
	} else if _, err := exec.LookPath("openhands"); err == nil {
		addItem("docker_installed", "OpenHands 执行环境", "", "ok", "OpenHands CLI 可用，当前可不依赖 Docker。")
	} else if uvPath, err := exec.LookPath("uv"); err == nil {
		addItem("docker_installed", "OpenHands 执行环境", "", "ok", "OpenHands 将通过 uv 自动安装/执行: "+uvPath)
	} else if openHandsNeedsDocker {
		addItem("docker_installed", "OpenHands 执行环境", "", "warning", "服务器上未找到 docker，当前 OpenHands 命令模板依赖 Docker。")
	} else {
		addItem("docker_installed", "OpenHands 执行环境", "", "warning", "未找到 openhands/uv/docker，OpenHands 相关能力暂不可用。")
	}

	// 12. 引擎核心配置
	requiredKeys := []string{
		"runtime.task_timeout_seconds",
		"runtime.max_steps",
		"watchdog.check_interval",
		"watchdog.max_stale_count",
		"watchdog.max_retries",
		"scheduler.max_concurrent",
		"scheduler.poll_interval",
	}
	count, e = g.DB().Model("mvp_config").
		WhereIn("config_key", requiredKeys).Where("deleted_at IS NULL").Count()
	if e != nil {
		addItem("engine_config", "引擎核心配置", "/mvp/config", "error", "查询失败: "+e.Error())
	} else if count < len(requiredKeys) {
		addItem("engine_config", "引擎核心配置", "/mvp/config", "warning",
			fmt.Sprintf("核心配置仅有 %d/%d 项，缺少的将使用默认值", count, len(requiredKeys)))
	} else {
		addItem("engine_config", "引擎核心配置", "/mvp/config", "ok",
			fmt.Sprintf("全部 %d 项核心配置已就绪", len(requiredKeys)))
	}

	// 13. 飞书协作配置
	feishuEnabled := engine.GetConfigInt(ctx, "workflow.collab.feishu_enabled", "workflow.collab.feishuEnabled", 0)
	feishuAppID := strings.TrimSpace(engine.GetConfigString(ctx, "workflow.collab.feishu_app_id", "workflow.collab.feishuAppId", ""))
	feishuAppSecret := strings.TrimSpace(engine.GetConfigString(ctx, "workflow.collab.feishu_app_secret", "workflow.collab.feishuAppSecret", ""))
	feishuEncryptKey := strings.TrimSpace(engine.GetConfigString(ctx, "workflow.collab.feishu_encrypt_key", "workflow.collab.feishuEncryptKey", ""))
	feishuBindings, _ := g.DB().Model("mvp_user_collab_binding").Ctx(ctx).Where("platform", "feishu").WhereNull("deleted_at").Count()
	switch {
	case feishuEnabled != 1:
		addItem("feishu_collab", "飞书协作", "/mvp/workflow/feishu", "warning", "飞书协作未启用")
	case feishuAppID == "" || feishuAppSecret == "":
		addItem("feishu_collab", "飞书协作", "/mvp/workflow/feishu", "warning", "已启用但缺少 App ID / App Secret")
	case feishuEncryptKey == "":
		addItem("feishu_collab", "飞书协作", "/mvp/workflow/feishu", "warning", "已启用但缺少 Encrypt Key")
	default:
		addItem("feishu_collab", "飞书协作", "/mvp/workflow/feishu", "ok",
			fmt.Sprintf("飞书配置已就绪，当前有效绑定 %d 条", feishuBindings))
	}

	return &v1.SystemCheckRes{Items: items, AllPass: allPass}, nil
}

// buildConfirmPlanResult 查询最新审核结果，组装 ConfirmPlan 响应。
func buildConfirmPlanResult(ctx context.Context, projectID int64) *v1.WorkflowConfirmPlanRes {
	res := &v1.WorkflowConfirmPlanRes{}

	// 查最新的 review stage_run
	stageRun, _ := g.DB().Model("mvp_stage_run").Ctx(ctx).
		InnerJoin("mvp_workflow_run wf", "wf.id = mvp_stage_run.workflow_run_id").
		Where("wf.project_id", projectID).
		Where("mvp_stage_run.stage_type", "review").
		WhereNull("mvp_stage_run.deleted_at").
		Fields("mvp_stage_run.id, mvp_stage_run.status, mvp_stage_run.error_message").
		OrderDesc("mvp_stage_run.stage_no").
		One()
	if stageRun.IsEmpty() {
		return res
	}
	stageRunID := stageRun["id"].Int64()

	// 统计 issue
	res.ErrorCount, _ = g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRunID).Where("severity", "error").Where("status", "open").Count()
	res.WarningCount, _ = g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRunID).Where("severity", "warning").Where("status", "open").Count()

	// 查 issue 列表（error 优先，最多 50 条）
	issues, _ := g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRunID).
		Where("status", "open").
		WhereNull("deleted_at").
		OrderDesc("severity").
		OrderDesc("created_at").
		Limit(50).
		All()

	for _, issue := range issues {
		res.Issues = append(res.Issues, v1.ReviewIssueItem{
			ID:         snowflake.JsonInt64(issue["id"].Int64()),
			Severity:   issue["severity"].String(),
			IssueCode:  issue["issue_code"].String(),
			SourceRole: issue["source_role"].String(),
			TaskName:   issue["task_name"].String(),
			Message:    issue["message"].String(),
			Suggestion: issue["suggestion"].String(),
			Status:     issue["status"].String(),
			CreatedAt:  issue["created_at"].GTime(),
		})
	}

	if stageRun["status"].String() == "failed" {
		res.RejectReason = stageRun["error_message"].String()
	}

	return res
}

// parseAndCreateBlueprints V2 专用：解析 AI 回复并创建蓝图。
func parseAndCreateBlueprints(ctx context.Context, projectID, conversationID, messageID int64, aiReply string) (int, error) {
	projectCategory, _ := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).Value("project_category")

	tasks, err := engine.GetParser().ExtractAndNormalize(ctx, aiReply, projectCategory.String())
	if err != nil || len(tasks) == 0 {
		return 0, err
	}

	return createBlueprints(ctx, projectID, conversationID, messageID, tasks)
}

// createBlueprints 将已提取的任务列表写入 plan_version + task_blueprint。
func createBlueprints(ctx context.Context, projectID, conversationID, messageID int64, tasks []engine.ArchitectTask) (int, error) {
	wfRun, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNotIn("status", g.Slice{"completed", "canceled"}).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	var wfRunID int64
	if !wfRun.IsEmpty() {
		wfRunID = wfRun["id"].Int64()
	}

	pvSvc := orchestrator.GetPlanVersionService()
	_, bpCount, err := pvSvc.CreateFromArchitectReply(ctx, projectID, wfRunID, conversationID, messageID, tasks)
	if err != nil {
		return 0, err
	}
	return bpCount, nil
}

// ReviewStatus 获取项目审核状态（V2 专用）
func (c *cWorkflow) ReviewStatus(ctx context.Context, req *v1.WorkflowReviewStatusReq) (res *v1.WorkflowReviewStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	res = &v1.WorkflowReviewStatusRes{}

	// 查最新的活跃 plan_version
	pv, err := g.DB().Model("mvp_plan_version").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"draft", "active"}).
		WhereNull("deleted_at").
		OrderDesc("version_no").
		One()
	if err != nil || pv.IsEmpty() {
		return res, nil
	}
	pvID := pv["id"].Int64()
	res.PlanVersionID = snowflake.JsonInt64(pvID)
	res.ReviewStatus = pv["review_status"].String()

	// 蓝图数
	bpCount, _ := g.DB().Model("mvp_task_blueprint").Ctx(ctx).
		Where("plan_version_id", pvID).WhereNull("deleted_at").Count()
	res.BlueprintCount = bpCount

	// 查最新的 review stage_run
	stageRun, _ := g.DB().Model("mvp_stage_run").Ctx(ctx).
		InnerJoin("mvp_workflow_run wf", "wf.id = mvp_stage_run.workflow_run_id").
		Where("wf.project_id", projectID).
		Where("mvp_stage_run.stage_type", "review").
		WhereNull("mvp_stage_run.deleted_at").
		Fields("mvp_stage_run.*").
		OrderDesc("mvp_stage_run.stage_no").
		One()
	if !stageRun.IsEmpty() {
		stageRunID := stageRun["id"].Int64()
		res.StageRunID = snowflake.JsonInt64(stageRunID)
		res.StageStatus = stageRun["status"].String()

		// stage_tasks
		var stageTasks []v1.ReviewStageTask
		tasks, _ := g.DB().Model("mvp_stage_task").Ctx(ctx).
			Where("stage_run_id", stageRunID).
			WhereNull("deleted_at").
			OrderAsc("created_at").
			All()
		for _, t := range tasks {
			st := v1.ReviewStageTask{
				ID:       snowflake.JsonInt64(t["id"].Int64()),
				TaskType: t["task_type"].String(),
				RoleType: t["role_type"].String(),
				Status:   t["status"].String(),
			}
			if !t["started_at"].IsEmpty() {
				startedAt := t["started_at"].GTime()
				st.StartedAt = startedAt
			}
			if !t["completed_at"].IsEmpty() {
				completedAt := t["completed_at"].GTime()
				st.CompletedAt = completedAt
			}
			if t["error_message"].String() != "" {
				st.ErrorMessage = t["error_message"].String()
			}
			stageTasks = append(stageTasks, st)
		}
		res.StageTasks = stageTasks

		// issue 统计
		res.ErrorCount, _ = g.DB().Model("mvp_review_issue").Ctx(ctx).
			Where("stage_run_id", stageRunID).Where("severity", "error").Where("status", "open").Count()
		res.WarningCount, _ = g.DB().Model("mvp_review_issue").Ctx(ctx).
			Where("stage_run_id", stageRunID).Where("severity", "warning").Where("status", "open").Count()
	}

	return res, nil
}

// ReviewIssues 获取审核问题列表
func (c *cWorkflow) ReviewIssues(ctx context.Context, req *v1.WorkflowReviewIssuesReq) (res *v1.WorkflowReviewIssuesRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 查最新的 review stage_run
	stageRun, _ := g.DB().Model("mvp_stage_run").Ctx(ctx).
		InnerJoin("mvp_workflow_run wf", "wf.id = mvp_stage_run.workflow_run_id").
		Where("wf.project_id", projectID).
		Where("mvp_stage_run.stage_type", "review").
		WhereNull("mvp_stage_run.deleted_at").
		Fields("mvp_stage_run.id").
		OrderDesc("mvp_stage_run.stage_no").
		One()
	if stageRun.IsEmpty() {
		return &v1.WorkflowReviewIssuesRes{Issues: []v1.ReviewIssueItem{}}, nil
	}

	issues, _ := g.DB().Model("mvp_review_issue").Ctx(ctx).
		Where("stage_run_id", stageRun["id"].Int64()).
		WhereNull("deleted_at").
		OrderDesc("severity").
		OrderDesc("created_at").
		All()

	items := make([]v1.ReviewIssueItem, 0, len(issues))
	for _, issue := range issues {
		items = append(items, v1.ReviewIssueItem{
			ID:         snowflake.JsonInt64(issue["id"].Int64()),
			Severity:   issue["severity"].String(),
			IssueCode:  issue["issue_code"].String(),
			SourceRole: issue["source_role"].String(),
			TaskName:   issue["task_name"].String(),
			Message:    issue["message"].String(),
			Suggestion: issue["suggestion"].String(),
			Status:     issue["status"].String(),
			CreatedAt:  issue["created_at"].GTime(),
		})
	}

	return &v1.WorkflowReviewIssuesRes{Issues: items}, nil
}

// ManualApprove 手动审批通过
func (c *cWorkflow) ManualApprove(ctx context.Context, req *v1.WorkflowManualApproveReq) (res *v1.WorkflowManualApproveRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 查活跃的 plan_version
	pv, err := g.DB().Model("mvp_plan_version").Ctx(ctx).
		Where("project_id", projectID).
		Where("status", "active").
		Where("review_status", "pending").
		WhereNull("deleted_at").
		OrderDesc("version_no").
		One()
	if err != nil || pv.IsEmpty() {
		return nil, fmt.Errorf("没有待审核的方案版本")
	}

	pvSvc := orchestrator.GetPlanVersionService()
	planVersionID := pv["id"].Int64()
	if err := pvSvc.Approve(ctx, planVersionID); err != nil {
		return nil, err
	}

	// 查活跃的 workflow_run，推进到 execute stage
	wfRun, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNotIn("status", g.Slice{"completed", "canceled"}).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	if !wfRun.IsEmpty() {
		wfRunID := wfRun["id"].Int64()
		currentStageRunID := wfRun["current_stage_run_id"].Int64()

		// 完成当前 review stage
		if currentStageRunID > 0 {
			stgSvc := orchestrator.GetStageService()
			_ = stgSvc.CompleteStage(ctx, currentStageRunID)
		}

		// 创建 execute stage + 实例化 + 启动调度
		execSvc := orchestrator.GetExecuteStageService()
		stgSvc := orchestrator.GetStageService()
		execStageRunID, err2 := stgSvc.StartStage(ctx, wfRunID, "execute")
		if err2 != nil {
			return nil, fmt.Errorf("审核已通过，但创建执行阶段失败: %w", err2)
		}
		if err3 := execSvc.InstantiateAndStart(ctx, execStageRunID, planVersionID); err3 != nil {
			_ = stgSvc.FailStage(ctx, execStageRunID, err3.Error())
			return nil, fmt.Errorf("审核已通过，但执行阶段启动失败: %w", err3)
		}
	}

	return &v1.WorkflowManualApproveRes{}, nil
}

// ManualReject 手动驳回
func (c *cWorkflow) ManualReject(ctx context.Context, req *v1.WorkflowManualRejectReq) (res *v1.WorkflowManualRejectRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	pv, err := g.DB().Model("mvp_plan_version").Ctx(ctx).
		Where("project_id", projectID).
		Where("status", "active").
		Where("review_status", "pending").
		WhereNull("deleted_at").
		OrderDesc("version_no").
		One()
	if err != nil || pv.IsEmpty() {
		return nil, fmt.Errorf("没有待审核的方案版本")
	}

	pvSvc := orchestrator.GetPlanVersionService()
	if err := pvSvc.Reject(ctx, pv["id"].Int64()); err != nil {
		return nil, err
	}

	// 项目状态回退 designing
	if _, upErr := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).
		Update(g.Map{"status": "designing", "updated_at": gdb.Raw("NOW()")}); upErr != nil {
		g.Log().Errorf(ctx, "[ManualReject] 项目状态回退失败: project=%d err=%v", projectID, upErr)
	}

	return &v1.WorkflowManualRejectRes{}, nil
}

// ==================== Timeline / Rework / Stage History ====================

// eventLabelMap 事件类型 → 可读标签
var eventLabelMap = map[string]string{
	"workflow.created":       "工作流已创建",
	"workflow.paused":        "工作流已暂停",
	"workflow.resumed":       "工作流已恢复",
	"workflow.canceled":      "工作流已取消",
	"workflow.completed":     "工作流已完成",
	"stage.started":          "阶段已启动",
	"stage.completed":        "阶段已完成",
	"stage.failed":           "阶段失败",
	"plan_version.created":   "方案版本已创建",
	"plan_version.submitted": "方案已提交审核",
	"plan_version.approved":  "方案审核通过",
	"plan_version.rejected":  "方案被驳回",
	"review.issue_created":   "发现审核问题",
	"review.decision_ready":  "审核决策就绪",
	"task.created":           "任务已创建",
	"task.started":           "任务已启动",
	"task.completed":         "任务已完成",
	"task.failed":            "任务失败",
	"task.escalated":         "任务已升级",
	"task.retried":           "任务已重试",
	"replan.completed":       "重规划完成",
	"replan.failed":          "重规划失败",
	"replan.aborted":         "重规划中止",
}

// Timeline 工作流事件时间线
func (c *cWorkflow) Timeline(ctx context.Context, req *v1.WorkflowTimelineReq) (res *v1.WorkflowTimelineRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	// 查活跃 workflow_run
	wfRuns, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Fields("id").
		OrderDesc("run_no").
		All()
	if err != nil || len(wfRuns) == 0 {
		return &v1.WorkflowTimelineRes{Events: []v1.TimelineEvent{}}, nil
	}

	wfRunIDs := make([]int64, 0, len(wfRuns))
	for _, r := range wfRuns {
		wfRunIDs = append(wfRunIDs, r["id"].Int64())
	}

	events, err := g.DB().Model("mvp_workflow_event").Ctx(ctx).
		WhereIn("workflow_run_id", wfRunIDs).
		OrderDesc("created_at").
		Limit(limit).
		All()
	if err != nil {
		return nil, err
	}

	list := make([]v1.TimelineEvent, 0, len(events))
	for _, e := range events {
		eventType := e["event_type"].String()
		label := eventLabelMap[eventType]
		if label == "" {
			label = eventType
		}
		// 补充 payload 中的上下文信息到 label
		payload := e["payload"].String()
		if payload != "" && payload != "null" {
			var pm map[string]string
			if json.Unmarshal([]byte(payload), &pm) == nil {
				if st, ok := pm["stage_type"]; ok {
					stageLabel := map[string]string{"design": "设计", "review": "审核", "execute": "执行", "rework": "返工", "complete": "完成"}[st]
					if stageLabel != "" {
						label = stageLabel + label[strings.Index(label, "阶段"):]
						if strings.Index(label, "阶段") < 0 {
							label = stageLabel + "阶段 " + label
						}
					}
				}
				if reason, ok := pm["reason"]; ok && reason != "" {
					label += "：" + reason
				}
			}
		}

		item := v1.TimelineEvent{
			ID:            snowflake.JsonInt64(e["id"].Int64()),
			WorkflowRunID: snowflake.JsonInt64(e["workflow_run_id"].Int64()),
			EntityType:    e["entity_type"].String(),
			EventType:     eventType,
			Label:         label,
			Payload:       payload,
			CreatedAt:     e["created_at"].GTime(),
		}
		if sid := e["stage_run_id"].Int64(); sid > 0 {
			v := snowflake.JsonInt64(sid)
			item.StageRunID = &v
		}
		if eid := e["entity_id"].Int64(); eid > 0 {
			v := snowflake.JsonInt64(eid)
			item.EntityID = &v
		}
		list = append(list, item)
	}

	return &v1.WorkflowTimelineRes{Events: list}, nil
}

// ReworkStatus 返工阶段状态
func (c *cWorkflow) ReworkStatus(ctx context.Context, req *v1.WorkflowReworkStatusReq) (res *v1.WorkflowReworkStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 查活跃 workflow_run
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	if err != nil || wfRun.IsEmpty() {
		return &v1.WorkflowReworkStatusRes{HasRework: false, History: []v1.ReworkRoundInfo{}}, nil
	}
	wfRunID := wfRun["id"].Int64()

	// 查 handoff_record
	handoffs, err := g.DB().Model("mvp_handoff_record").Ctx(ctx).
		Where("workflow_run_id", wfRunID).
		Where("handoff_type", "failure_escalation").
		OrderAsc("created_at").
		All()
	if err != nil {
		return nil, err
	}

	if len(handoffs) == 0 {
		return &v1.WorkflowReworkStatusRes{HasRework: false, History: []v1.ReworkRoundInfo{}}, nil
	}

	// 查当前 rework stage
	var currentStage *v1.ReworkStageInfo
	reworkStage, _ := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", wfRunID).
		Where("stage_type", "rework").
		WhereNull("deleted_at").
		OrderDesc("stage_no").
		One()
	if !reworkStage.IsEmpty() {
		currentStage = &v1.ReworkStageInfo{
			StageRunID: snowflake.JsonInt64(reworkStage["id"].Int64()),
			Status:     reworkStage["status"].String(),
			StartedAt:  reworkStage["started_at"].GTime(),
		}
	}

	// 构建轮次历史
	history := make([]v1.ReworkRoundInfo, 0, len(handoffs))
	for i, h := range handoffs {
		fromTaskID := h["from_task_id"].Int64()
		toTaskID := h["to_task_id"].Int64()

		// 查失败任务名称和原因
		failedTask, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
			Where("id", fromTaskID).Fields("name, result").One()
		failedName := ""
		failedReason := h["reason"].String()
		if !failedTask.IsEmpty() {
			failedName = failedTask["name"].String()
		}

		// 查分析任务结果
		var analysisID *snowflake.JsonInt64
		analysisResult := ""
		if toTaskID > 0 {
			v := snowflake.JsonInt64(toTaskID)
			analysisID = &v
			analysisTask, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
				Where("id", toTaskID).Fields("result").One()
			if !analysisTask.IsEmpty() {
				analysisResult = analysisTask["result"].String()
			}
		}

		history = append(history, v1.ReworkRoundInfo{
			Round:          i + 1,
			FailedTaskID:   snowflake.JsonInt64(fromTaskID),
			FailedTaskName: failedName,
			FailedReason:   failedReason,
			AnalysisTaskID: analysisID,
			AnalysisResult: analysisResult,
			HandoffType:    h["handoff_type"].String(),
			CreatedAt:      h["created_at"].GTime(),
		})
	}

	return &v1.WorkflowReworkStatusRes{
		HasRework:    true,
		ReworkRounds: len(history),
		CurrentStage: currentStage,
		History:      history,
	}, nil
}

// StageHistory 工作流阶段历史
func (c *cWorkflow) StageHistory(ctx context.Context, req *v1.WorkflowStageHistoryReq) (res *v1.WorkflowStageHistoryRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	if err != nil || wfRun.IsEmpty() {
		return &v1.WorkflowStageHistoryRes{Stages: []v1.StageHistoryItem{}}, nil
	}

	stages, err := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", wfRun["id"].Int64()).
		WhereNull("deleted_at").
		Fields("id, stage_type, stage_no, status, started_at, finished_at, error_message").
		OrderAsc("stage_no").
		All()
	if err != nil {
		return nil, err
	}

	list := make([]v1.StageHistoryItem, 0, len(stages))
	for _, s := range stages {
		list = append(list, v1.StageHistoryItem{
			ID:         snowflake.JsonInt64(s["id"].Int64()),
			StageType:  s["stage_type"].String(),
			StageNo:    s["stage_no"].Int(),
			Status:     s["status"].String(),
			StartedAt:  s["started_at"].GTime(),
			FinishedAt: s["finished_at"].GTime(),
			Error:      s["error_message"].String(),
		})
	}

	return &v1.WorkflowStageHistoryRes{Stages: list}, nil
}

// CompletionSummary 获取项目完成总结
func (c *cWorkflow) CompletionSummary(ctx context.Context, req *v1.WorkflowCompletionSummaryReq) (res *v1.WorkflowCompletionSummaryRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	svc := orchestrator.GetCompleteStageService()
	summary, err := svc.GetSummary(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &v1.WorkflowCompletionSummaryRes{
		WorkflowRunID:   snowflake.JsonInt64(summary.WorkflowRunID),
		ProjectID:       snowflake.JsonInt64(summary.ProjectID),
		TotalTasks:      summary.TotalTasks,
		CompletedTasks:  summary.CompletedTasks,
		FailedTasks:     summary.FailedTasks,
		EscalatedTasks:  summary.EscalatedTasks,
		SkippedTasks:    summary.SkippedTasks,
		SuccessRate:     summary.SuccessRate,
		TotalDuration:   summary.TotalDuration,
		AvgTaskDuration: summary.AvgTaskDuration,
		StageDurations:  summary.StageDurations,
		ReworkRounds:    summary.ReworkRounds,
		HandoffCount:    summary.HandoffCount,
		StartedAt:       summary.StartedAt,
		FinishedAt:      summary.FinishedAt,
	}, nil
}

// ==================== 执行控制台 ====================

// ExecutionStatus 执行阶段实时状态
func (c *cWorkflow) ExecutionStatus(ctx context.Context, req *v1.WorkflowExecutionStatusReq) (res *v1.WorkflowExecutionStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	res = &v1.WorkflowExecutionStatusRes{
		Tasks:         []v1.DomainTaskItem{},
		ResourceLocks: []v1.ResourceLockItem{},
	}

	// 查活跃 workflow_run
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	if err != nil || wfRun.IsEmpty() {
		return res, nil
	}
	wfRunID := wfRun["id"].Int64()
	res.WorkflowRunID = snowflake.JsonInt64(wfRunID)

	// 查 execute stage_run
	stageRun, stageErr := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", wfRunID).
		Where("stage_type", "execute").
		WhereNull("deleted_at").
		OrderDesc("stage_no").
		One()
	if stageErr != nil {
		g.Log().Warningf(ctx, "[ExecutionStatus] 查询 stage_run 失败: wfRun=%d err=%v", wfRunID, stageErr)
	}
	if !stageRun.IsEmpty() {
		res.StageRunID = snowflake.JsonInt64(stageRun["id"].Int64())
		res.StageStatus = stageRun["status"].String()
	}

	// 查领域任务
	tasks, taskErr := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", wfRunID).
		WhereNull("deleted_at").
		OrderAsc("batch_no").
		OrderAsc("sort").
		All()
	if taskErr != nil {
		g.Log().Warningf(ctx, "[ExecutionStatus] 查询领域任务失败: wfRun=%d err=%v", wfRunID, taskErr)
	}

	for _, t := range tasks {
		res.Tasks = append(res.Tasks, buildDomainTaskItem(t))
	}

	// 统计
	for _, t := range res.Tasks {
		res.TotalTasks++
		switch t.Status {
		case "completed":
			res.CompletedTasks++
		case "running":
			res.RunningTasks++
		case "failed":
			res.FailedTasks++
		case "pending":
			res.PendingTasks++
		case "escalated":
			res.EscalatedTasks++
		}
	}

	// 活跃批次
	scheduler := orchestrator.GetTaskScheduler()
	if scheduler != nil {
		lockedRes := scheduler.GetLockedResources()
		for resource, taskID := range lockedRes {
			taskName := ""
			for _, t := range res.Tasks {
				if int64(t.ID) == taskID {
					taskName = t.Name
					break
				}
			}
			res.ResourceLocks = append(res.ResourceLocks, v1.ResourceLockItem{
				Resource: resource,
				TaskID:   snowflake.JsonInt64(taskID),
				TaskName: taskName,
			})
		}
	}

	// 计算活跃批次号
	for _, t := range res.Tasks {
		if t.Status == "running" || t.Status == "pending" {
			if t.BatchNo > 0 && (res.ActiveBatch == 0 || t.BatchNo < res.ActiveBatch) {
				res.ActiveBatch = t.BatchNo
			}
		}
	}

	return res, nil
}

// DomainTasks 领域任务列表
func (c *cWorkflow) DomainTasks(ctx context.Context, req *v1.WorkflowDomainTasksReq) (res *v1.WorkflowDomainTasksRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 查 workflow_run
	wfRun, wfErr := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
	if wfErr != nil {
		return nil, fmt.Errorf("查询工作流运行失败: %w", wfErr)
	}
	if wfRun.IsEmpty() {
		return &v1.WorkflowDomainTasksRes{Tasks: []v1.DomainTaskItem{}}, nil
	}

	query := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", wfRun["id"].Int64()).
		WhereNull("deleted_at")

	if req.Status != "" {
		query = query.Where("status", req.Status)
	}
	if req.BatchNo > 0 {
		query = query.Where("batch_no", req.BatchNo)
	}

	tasks, err := query.OrderAsc("batch_no").OrderAsc("sort").All()
	if err != nil {
		return nil, err
	}

	items := make([]v1.DomainTaskItem, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, buildDomainTaskItem(t))
	}

	return &v1.WorkflowDomainTasksRes{Tasks: items, Total: len(items)}, nil
}

// ResourceLocks 资源锁列表
func (c *cWorkflow) ResourceLocks(ctx context.Context, req *v1.WorkflowResourceLocksReq) (res *v1.WorkflowResourceLocksRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	res = &v1.WorkflowResourceLocksRes{Locks: []v1.ResourceLockItem{}}

	scheduler := orchestrator.GetTaskScheduler()
	if scheduler == nil {
		return res, nil
	}

	lockedRes := scheduler.GetLockedResources()
	if len(lockedRes) == 0 {
		return res, nil
	}

	// 查任务名称
	taskIDs := make([]int64, 0, len(lockedRes))
	for _, tid := range lockedRes {
		taskIDs = append(taskIDs, tid)
	}
	taskNames := make(map[int64]string)
	tasks, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
		WhereIn("id", taskIDs).Fields("id, name").All()
	for _, t := range tasks {
		taskNames[t["id"].Int64()] = t["name"].String()
	}

	for resource, taskID := range lockedRes {
		res.Locks = append(res.Locks, v1.ResourceLockItem{
			Resource: resource,
			TaskID:   snowflake.JsonInt64(taskID),
			TaskName: taskNames[taskID],
		})
	}

	return res, nil
}

// buildDomainTaskItem 构建领域任务响应项。
func buildDomainTaskItem(t gdb.Record) v1.DomainTaskItem {
	var resources []string
	resJSON := t["affected_resources"].String()
	if resJSON != "" && resJSON != "[]" && resJSON != "null" {
		_ = json.Unmarshal([]byte(resJSON), &resources)
	}
	return v1.DomainTaskItem{
		ID:                snowflake.JsonInt64(t["id"].Int64()),
		Name:              t["name"].String(),
		Description:       t["description"].String(),
		Status:            t["status"].String(),
		RoleType:          t["role_type"].String(),
		RoleLevel:         t["role_level"].String(),
		BatchNo:           t["batch_no"].Int(),
		Sort:              t["sort"].Int(),
		ExecutionMode:     t["execution_mode"].String(),
		AffectedResources: resources,
		StartedAt:         t["started_at"].GTime(),
		CompletedAt:       t["completed_at"].GTime(),
		ErrorMessage:      t["error_message"].String(),
		Result:            t["result"].String(),
		RetryCount:        t["retry_count"].Int(),
	}
}

// ==================== 验收控制台 Controller ====================

// AcceptStatus 验收状态总览
func (c *cWorkflow) AcceptStatus(ctx context.Context, req *v1.WorkflowAcceptStatusReq) (res *v1.WorkflowAcceptStatusRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 查活跃 workflow_run
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"accepting", "running", "paused", "completed"}).
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if err != nil || wfRun.IsEmpty() {
		return nil, fmt.Errorf("项目无工作流运行")
	}
	workflowRunID := wfRun["id"].Int64()

	// 查最新 accept_run
	acceptRunRepo := repo.NewAcceptRunRepo()
	acceptRun, err := acceptRunRepo.GetLatestByWorkflow(ctx, workflowRunID)
	if err != nil || len(acceptRun) == 0 {
		return &v1.WorkflowAcceptStatusRes{Status: "none"}, nil
	}

	acceptRunID := acceptRun["id"]
	acceptRunIDInt := g.NewVar(acceptRunID).Int64()

	// 统计各级别 issue 数量
	issueRepo := repo.NewAcceptIssueRepo()
	issues, _ := issueRepo.ListByAcceptRun(ctx, acceptRunIDInt)
	var blockers, errors, warns, infos int
	for _, issue := range issues {
		switch g.NewVar(issue["severity"]).String() {
		case "blocker":
			blockers++
		case "error":
			errors++
		case "warn":
			warns++
		case "info":
			infos++
		}
	}

	// 统计证据数量
	evidenceRepo := repo.NewAcceptEvidenceRepo()
	evidenceList, _ := evidenceRepo.ListByAcceptRun(ctx, acceptRunIDInt)

	res = &v1.WorkflowAcceptStatusRes{
		AcceptRunID:   snowflake.JsonInt64(acceptRunIDInt),
		WorkflowRunID: snowflake.JsonInt64(workflowRunID),
		AcceptRound:   g.NewVar(acceptRun["accept_round"]).Int(),
		Status:        g.NewVar(acceptRun["status"]).String(),
		Decision:      g.NewVar(acceptRun["decision"]).String(),
		Score:         g.NewVar(acceptRun["score"]).Float64(),
		Summary:       g.NewVar(acceptRun["summary"]).String(),
		RulesSnapshot: g.NewVar(acceptRun["rules_snapshot_ref"]).String(),
		StartedAt:     g.NewVar(acceptRun["started_at"]).GTime(),
		FinishedAt:    g.NewVar(acceptRun["finished_at"]).GTime(),
		BlockerCount:  blockers,
		ErrorCount:    errors,
		WarnCount:     warns,
		InfoCount:     infos,
		EvidenceCount: len(evidenceList),
	}
	return res, nil
}

// AcceptIssues 验收问题列表
func (c *cWorkflow) AcceptIssues(ctx context.Context, req *v1.WorkflowAcceptIssuesReq) (res *v1.WorkflowAcceptIssuesRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 查活跃 workflow_run → 最新 accept_run
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"accepting", "running", "paused", "completed"}).
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if err != nil || wfRun.IsEmpty() {
		return &v1.WorkflowAcceptIssuesRes{Issues: []v1.AcceptIssueItem{}}, nil
	}

	acceptRunRepo := repo.NewAcceptRunRepo()
	acceptRun, err := acceptRunRepo.GetLatestByWorkflow(ctx, wfRun["id"].Int64())
	if err != nil || len(acceptRun) == 0 {
		return &v1.WorkflowAcceptIssuesRes{Issues: []v1.AcceptIssueItem{}}, nil
	}

	issueRepo := repo.NewAcceptIssueRepo()
	issues, err := issueRepo.ListByAcceptRun(ctx, g.NewVar(acceptRun["id"]).Int64())
	if err != nil {
		return nil, err
	}

	var items []v1.AcceptIssueItem
	for _, issue := range issues {
		severity := g.NewVar(issue["severity"]).String()
		if req.Severity != "" && severity != req.Severity {
			continue
		}
		items = append(items, v1.AcceptIssueItem{
			ID:              snowflake.JsonInt64(g.NewVar(issue["id"]).Int64()),
			IssueType:       g.NewVar(issue["issue_type"]).String(),
			RuleCode:        g.NewVar(issue["rule_code"]).String(),
			Severity:        severity,
			Title:           g.NewVar(issue["title"]).String(),
			Detail:          g.NewVar(issue["detail"]).String(),
			ExpectedValue:   g.NewVar(issue["expected_value"]).String(),
			ActualValue:     g.NewVar(issue["actual_value"]).String(),
			SuggestedAction: g.NewVar(issue["suggested_action"]).String(),
			DomainTaskID:    snowflake.JsonInt64(g.NewVar(issue["domain_task_id"]).Int64()),
			ResourceRef:     g.NewVar(issue["resource_ref"]).String(),
			Status:          g.NewVar(issue["status"]).String(),
			CreatedAt:       g.NewVar(issue["created_at"]).GTime(),
		})
	}
	if items == nil {
		items = []v1.AcceptIssueItem{}
	}
	return &v1.WorkflowAcceptIssuesRes{Issues: items}, nil
}

// AcceptEvidence 验收证据列表
func (c *cWorkflow) AcceptEvidence(ctx context.Context, req *v1.WorkflowAcceptEvidenceReq) (res *v1.WorkflowAcceptEvidenceRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"accepting", "running", "paused", "completed"}).
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if err != nil || wfRun.IsEmpty() {
		return &v1.WorkflowAcceptEvidenceRes{Evidence: []v1.AcceptEvidenceItem{}}, nil
	}

	acceptRunRepo := repo.NewAcceptRunRepo()
	acceptRun, err := acceptRunRepo.GetLatestByWorkflow(ctx, wfRun["id"].Int64())
	if err != nil || len(acceptRun) == 0 {
		return &v1.WorkflowAcceptEvidenceRes{Evidence: []v1.AcceptEvidenceItem{}}, nil
	}

	evidenceRepo := repo.NewAcceptEvidenceRepo()
	evidenceList, err := evidenceRepo.ListByAcceptRun(ctx, g.NewVar(acceptRun["id"]).Int64())
	if err != nil {
		return nil, err
	}

	var items []v1.AcceptEvidenceItem
	for _, e := range evidenceList {
		items = append(items, v1.AcceptEvidenceItem{
			ID:           snowflake.JsonInt64(g.NewVar(e["id"]).Int64()),
			EvidenceType: g.NewVar(e["evidence_type"]).String(),
			SourceType:   g.NewVar(e["source_type"]).String(),
			SourceID:     snowflake.JsonInt64(g.NewVar(e["source_id"]).Int64()),
			ContentRef:   g.NewVar(e["content_ref"]).String(),
			Summary:      g.NewVar(e["summary"]).String(),
			CreatedAt:    g.NewVar(e["created_at"]).GTime(),
		})
	}
	if items == nil {
		items = []v1.AcceptEvidenceItem{}
	}
	return &v1.WorkflowAcceptEvidenceRes{Evidence: items}, nil
}

// AcceptApprove 人工放行
func (c *cWorkflow) AcceptApprove(ctx context.Context, req *v1.WorkflowAcceptApproveReq) (res *v1.WorkflowAcceptApproveRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	svc := orchestrator.GetAcceptStageService()
	if svc == nil {
		return nil, fmt.Errorf("验收服务未初始化")
	}
	if err := svc.ManualApprove(ctx, projectID, req.Reason); err != nil {
		return nil, err
	}
	return &v1.WorkflowAcceptApproveRes{}, nil
}

// AcceptReject 驳回验收
func (c *cWorkflow) AcceptReject(ctx context.Context, req *v1.WorkflowAcceptRejectReq) (res *v1.WorkflowAcceptRejectRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	svc := orchestrator.GetAcceptStageService()
	if svc == nil {
		return nil, fmt.Errorf("验收服务未初始化")
	}
	if err := svc.ManualReject(ctx, projectID, req.Reason); err != nil {
		return nil, err
	}
	return &v1.WorkflowAcceptRejectRes{}, nil
}

// AcceptRerun 重新验收
func (c *cWorkflow) AcceptRerun(ctx context.Context, req *v1.WorkflowAcceptRerunReq) (res *v1.WorkflowAcceptRerunRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	svc := orchestrator.GetAcceptStageService()
	if svc == nil {
		return nil, fmt.Errorf("验收服务未初始化")
	}
	if err := svc.Rerun(ctx, projectID); err != nil {
		return nil, err
	}
	return &v1.WorkflowAcceptRerunRes{}, nil
}

// AcceptRework 驳回并返工
func (c *cWorkflow) AcceptRework(ctx context.Context, req *v1.WorkflowAcceptReworkReq) (res *v1.WorkflowAcceptReworkRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}
	svc := orchestrator.GetAcceptStageService()
	if svc == nil {
		return nil, fmt.Errorf("验收服务未初始化")
	}
	if err := svc.ManualRework(ctx, projectID, req.Reason); err != nil {
		return nil, err
	}
	return &v1.WorkflowAcceptReworkRes{}, nil
}

// ==================== 自治管理 Controller ====================

// AutonomyDecisions 自治决策列表
func (c *cWorkflow) AutonomyDecisions(ctx context.Context, req *v1.WorkflowAutonomyDecisionsReq) (res *v1.WorkflowAutonomyDecisionsRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	decisionRepo := repo.NewAutonomyDecisionRepo()
	records, err := decisionRepo.ListByProject(ctx, projectID, req.DecisionType)
	if err != nil {
		return nil, err
	}

	var items []v1.AutonomyDecisionItem
	for _, r := range records {
		items = append(items, v1.AutonomyDecisionItem{
			ID:             snowflake.JsonInt64(g.NewVar(r["id"]).Int64()),
			DecisionType:   g.NewVar(r["decision_type"]).String(),
			TriggerSource:  g.NewVar(r["trigger_source"]).String(),
			TriggerContext: g.NewVar(r["trigger_context"]).String(),
			Recommendation: g.NewVar(r["recommendation"]).String(),
			DecisionMode:   g.NewVar(r["decision_mode"]).String(),
			HumanAction:    g.NewVar(r["human_action"]).String(),
			ExecutedAt:     g.NewVar(r["executed_at"]).GTime(),
			Result:         g.NewVar(r["result"]).String(),
			CreatedAt:      g.NewVar(r["created_at"]).GTime(),
		})
	}
	if items == nil {
		items = []v1.AutonomyDecisionItem{}
	}
	return &v1.WorkflowAutonomyDecisionsRes{Decisions: items}, nil
}

// ApproveDecision 批准自治决策
func (c *cWorkflow) ApproveDecision(ctx context.Context, req *v1.WorkflowApproveDecisionReq) (res *v1.WorkflowApproveDecisionRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	decisionRepo := repo.NewAutonomyDecisionRepo()
	if err := decisionRepo.UpdateHumanAction(ctx, int64(req.DecisionID), "approved"); err != nil {
		return nil, err
	}
	return &v1.WorkflowApproveDecisionRes{}, nil
}

// RejectDecision 拒绝自治决策
func (c *cWorkflow) RejectDecision(ctx context.Context, req *v1.WorkflowRejectDecisionReq) (res *v1.WorkflowRejectDecisionRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	decisionRepo := repo.NewAutonomyDecisionRepo()
	if err := decisionRepo.UpdateHumanAction(ctx, int64(req.DecisionID), "rejected"); err != nil {
		return nil, err
	}
	return &v1.WorkflowRejectDecisionRes{}, nil
}

// TriggerReplan 手动触发重规划
func (c *cWorkflow) TriggerReplan(ctx context.Context, req *v1.WorkflowTriggerReplanReq) (res *v1.WorkflowTriggerReplanRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 查活跃 workflow_run
	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereIn("status", g.Slice{"executing", "reworking", "accepting", "paused"}).
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if err != nil || wfRun.IsEmpty() {
		return nil, fmt.Errorf("无活跃的工作流运行")
	}

	decisionRepo := repo.NewAutonomyDecisionRepo()
	replanner := autonomy.NewReplanner(decisionRepo)

	// 收集失败任务信息
	failedTasks, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", wfRun["id"].Int64()).
		WhereIn("status", g.Slice{"failed", "escalated"}).
		WhereNull("deleted_at").
		Fields("id, name, result, retry_count").All()

	var failed []autonomy.FailedTaskInfo
	for _, t := range failedTasks {
		failed = append(failed, autonomy.FailedTaskInfo{
			TaskID:       t["id"].Int64(),
			TaskName:     t["name"].String(),
			ErrorMessage: t["result"].String(),
			RetryCount:   t["retry_count"].Int(),
		})
	}

	input := &autonomy.ReplanInput{
		WorkflowRunID: wfRun["id"].Int64(),
		ProjectID:     projectID,
		TriggerSource: "manual",
		FailedTasks:   failed,
	}

	wfRunID := wfRun["id"].Int64()

	// 异步执行重规划（LLM 调用耗时长，避免 HTTP 超时）
	go func() {
		bgCtx := context.Background()
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(bgCtx, "[TriggerReplan] panic: %v", r)
			}
		}()

		result, err := replanner.Evaluate(bgCtx, input)

		// 写入时间线事件（不管成功失败都记录）
		eventType := "replan.completed"
		payloadMap := map[string]interface{}{
			"trigger": "manual",
		}
		if err != nil {
			eventType = "replan.failed"
			payloadMap["error"] = err.Error()
			g.Log().Errorf(bgCtx, "[TriggerReplan] 重规划失败: projectID=%d err=%v", projectID, err)
		} else if result != nil {
			payloadMap["action"] = result.Action
			payloadMap["reasoning"] = result.Reasoning
			if result.Action == autonomy.ReplanAbort {
				eventType = "replan.aborted"
				g.Log().Warningf(bgCtx, "[TriggerReplan] 重规划中止: projectID=%d reason=%s", projectID, result.Reasoning)
			} else {
				g.Log().Infof(bgCtx, "[TriggerReplan] 重规划完成: projectID=%d action=%s", projectID, result.Action)
			}
		}
		payloadJSON, jsonErr := json.Marshal(payloadMap)
		if jsonErr != nil {
			g.Log().Warningf(bgCtx, "[TriggerReplan] 序列化事件 payload 失败: %v", jsonErr)
		}
		if _, insErr := g.DB().Model("mvp_workflow_event").Ctx(bgCtx).Insert(g.Map{
			"id":              int64(snowflake.Generate()),
			"workflow_run_id": wfRunID,
			"entity_type":     "workflow",
			"event_type":      eventType,
			"payload":         string(payloadJSON),
			"created_at":      gtime.Now(),
		}); insErr != nil {
			g.Log().Warningf(bgCtx, "[TriggerReplan] 记录重规划事件失败: wfRun=%d err=%v", wfRunID, insErr)
		}
	}()

	return &v1.WorkflowTriggerReplanRes{}, nil
}

// ProjectReports 项目报告列表
func (c *cWorkflow) ProjectReports(ctx context.Context, req *v1.WorkflowProjectReportsReq) (res *v1.WorkflowProjectReportsRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	reportRepo := repo.NewProjectReportRepo()
	records, err := reportRepo.ListByProject(ctx, projectID, req.ReportType)
	if err != nil {
		return nil, err
	}

	var items []v1.ProjectReportItem
	for _, r := range records {
		items = append(items, v1.ProjectReportItem{
			ID:         snowflake.JsonInt64(g.NewVar(r["id"]).Int64()),
			ReportType: g.NewVar(r["report_type"]).String(),
			StageType:  g.NewVar(r["stage_type"]).String(),
			Title:      g.NewVar(r["title"]).String(),
			Content:    g.NewVar(r["content"]).String(),
			Metrics:    g.NewVar(r["metrics"]).String(),
			CreatedAt:  g.NewVar(r["created_at"]).GTime(),
		})
	}
	if items == nil {
		items = []v1.ProjectReportItem{}
	}
	return &v1.WorkflowProjectReportsRes{Reports: items}, nil
}

// TriggerReport 手动生成报告
func (c *cWorkflow) TriggerReport(ctx context.Context, req *v1.WorkflowTriggerReportReq) (res *v1.WorkflowTriggerReportRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	wfRun, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if err != nil || wfRun.IsEmpty() {
		return nil, fmt.Errorf("无工作流运行记录")
	}

	reportRepo := repo.NewProjectReportRepo()
	reporter := autonomy.NewReporter(reportRepo)

	stageType := req.StageType
	if stageType == "" {
		stageType = "complete"
	}

	if err := reporter.GenerateStageReport(ctx, wfRun["id"].Int64(), stageType); err != nil {
		return nil, err
	}
	return &v1.WorkflowTriggerReportRes{}, nil
}

// AutonomyMode 查询当前自治模式
func (c *cWorkflow) AutonomyMode(ctx context.Context, req *v1.WorkflowAutonomyModeReq) (res *v1.WorkflowAutonomyModeRes, err error) {
	return &v1.WorkflowAutonomyModeRes{Mode: autonomy.GetAutonomyMode(ctx)}, nil
}

// SetAutonomyMode 设置自治模式（写入 mvp_config）
func (c *cWorkflow) SetAutonomyMode(ctx context.Context, req *v1.WorkflowSetAutonomyModeReq) (res *v1.WorkflowSetAutonomyModeRes, err error) {
	// 检查是否已有记录
	count, _ := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", "autonomy.mode").
		WhereNull("deleted_at").Count()
	if count > 0 {
		_, err = g.DB().Model("mvp_config").Ctx(ctx).
			Where("config_key", "autonomy.mode").
			Update(g.Map{"config_value": req.Mode})
	} else {
		_, err = g.DB().Model("mvp_config").Ctx(ctx).Insert(g.Map{
			"config_key":   "autonomy.mode",
			"config_value": req.Mode,
			"category":     "autonomy",
			"description":  "自治模式：suggest=建议型 auto=全自动",
		})
	}
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowSetAutonomyModeRes{}, nil
}

// BatchProjectStats 批量查询项目运行时统计（为项目列表页提供进度数据）
func (c *cWorkflow) BatchProjectStats(ctx context.Context, req *v1.WorkflowBatchProjectStatsReq) (res *v1.WorkflowBatchProjectStatsRes, err error) {
	if len(req.ProjectIDs) > 50 {
		return nil, fmt.Errorf("单次最多查询 50 个项目")
	}

	ids := make([]int64, 0, len(req.ProjectIDs))
	for _, id := range req.ProjectIDs {
		ids = append(ids, int64(id))
	}

	// 权限过滤：通过 ApplyDataScope 五级数据权限过滤，只保留用户有权访问的项目
	scopedQuery := middleware.ApplyDataScope(ctx,
		g.DB().Model("mvp_project").Ctx(ctx).
			WhereIn("id", ids).
			WhereNull("deleted_at").
			Fields("id"),
		"created_by", "dept_id",
	)
	allowedRecords, err := scopedQuery.All()
	if err != nil {
		return nil, fmt.Errorf("权限过滤查询失败: %w", err)
	}
	allowedIDs := make(map[int64]bool, len(allowedRecords))
	for _, p := range allowedRecords {
		allowedIDs[p["id"].Int64()] = true
	}
	filtered := ids[:0]
	for _, id := range ids {
		if allowedIDs[id] {
			filtered = append(filtered, id)
		}
	}
	ids = filtered
	if len(ids) == 0 {
		return &v1.WorkflowBatchProjectStatsRes{Stats: []v1.ProjectRuntimeStat{}}, nil
	}

	// 批量查每个项目最新的 workflow_run
	wfMap, err := loadLatestWorkflowRuns(ctx, ids)
	if err != nil {
		return nil, err
	}

	// 收集所有 workflow_run_id
	wfRunIDs := make([]int64, 0, len(wfMap))
	wfStatusByRunID := make(map[int64]string, len(wfMap))
	for _, r := range wfMap {
		wfID := r["id"].Int64()
		wfRunIDs = append(wfRunIDs, wfID)
		wfStatusByRunID[wfID] = r["status"].String()
	}

	// 优先读取最新态势快照；快照缺失或过旧时再回退到实时聚合。
	snapshotMap, snapshotErr := loadLatestSituationSnapshots(ctx, wfRunIDs)
	if snapshotErr != nil {
		g.Log().Warningf(ctx, "[BatchProjectStats] 读取态势快照失败，回退实时聚合: err=%v", snapshotErr)
	}

	fallbackRunIDs := make([]int64, 0, len(wfRunIDs))
	for _, wfID := range wfRunIDs {
		if !shouldUseRuntimeSnapshot(snapshotMap[wfID], wfStatusByRunID[wfID]) {
			fallbackRunIDs = append(fallbackRunIDs, wfID)
		}
	}

	taskStats, err := loadTaskStats(ctx, fallbackRunIDs)
	if err != nil {
		return nil, err
	}

	// 组装结果
	stats := make([]v1.ProjectRuntimeStat, 0, len(ids))
	for _, pid := range ids {
		stat := v1.ProjectRuntimeStat{
			ProjectID: snowflake.JsonInt64(pid),
		}
		if wf, ok := wfMap[pid]; ok {
			stat.CurrentStage = wf["current_stage"].String()
			wfID := wf["id"].Int64()
			if snapshot := snapshotMap[wfID]; shouldUseRuntimeSnapshot(snapshot, wf["status"].String()) {
				if stat.CurrentStage == "" {
					stat.CurrentStage = snapshot.Situation.ActiveStage
				}
				ts := taskStatFromProgress(snapshot.Situation.Progress)
				stat.TotalTasks = ts.TotalTasks
				stat.CompletedTasks = ts.CompletedTasks
				stat.FailedTasks = ts.FailedTasks
				stat.RunningTasks = ts.RunningTasks
			} else if ts, exists := taskStats[wfID]; exists {
				stat.TotalTasks = ts.TotalTasks
				stat.CompletedTasks = ts.CompletedTasks
				stat.FailedTasks = ts.FailedTasks
				stat.RunningTasks = ts.RunningTasks
			}
		}
		stats = append(stats, stat)
	}

	return &v1.WorkflowBatchProjectStatsRes{Stats: stats}, nil
}

// AutonomyCheckpoints 查询项目待处理的人工节点和决策动作。
func (c *cWorkflow) AutonomyCheckpoints(ctx context.Context, req *v1.WorkflowAutonomyCheckpointsReq) (res *v1.WorkflowAutonomyCheckpointsRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	rawCheckpoints, cpErr := dc.ListOpenCheckpoints(ctx, projectID)
	if cpErr != nil {
		return nil, cpErr
	}
	rawActions, acErr := dc.ListPendingActions(ctx, projectID)
	if acErr != nil {
		return nil, acErr
	}

	// snake_case g.Map → camelCase DTO
	checkpoints := make([]v1.CheckpointDTO, 0, len(rawCheckpoints))
	for _, m := range rawCheckpoints {
		checkpoints = append(checkpoints, mapToCheckpointDTO(m))
	}
	actions := make([]v1.DecisionActionDTO, 0, len(rawActions))
	for _, m := range rawActions {
		actions = append(actions, mapToDecisionActionDTO(m))
	}

	return &v1.WorkflowAutonomyCheckpointsRes{
		Checkpoints: checkpoints,
		Actions:     actions,
	}, nil
}

// mapToCheckpointDTO 将 g.Map (snake_case) 映射到 CheckpointDTO (camelCase)。
func mapToCheckpointDTO(m g.Map) v1.CheckpointDTO {
	return v1.CheckpointDTO{
		ID:               mapJsonInt64(m, "id"),
		WorkflowRunID:    mapJsonInt64(m, "workflow_run_id"),
		ProjectID:        mapJsonInt64(m, "project_id"),
		DecisionActionID: mapJsonInt64(m, "decision_action_id"),
		CheckpointType:   mapString(m, "checkpoint_type"),
		Title:            mapString(m, "title"),
		Description:      mapString(m, "description"),
		Status:           mapString(m, "status"),
		AssignedTo:       mapJsonInt64(m, "assigned_to"),
		HandledBy:        mapJsonInt64(m, "handled_by"),
		HandleAction:     mapString(m, "handle_action"),
		HandleReason:     mapString(m, "handle_reason"),
		HandledAt:        mapGTime(m, "handled_at"),
		ExpiresAt:        mapGTime(m, "expires_at"),
		CreatedAt:        mapGTime(m, "created_at"),
	}
}

// mapToDecisionActionDTO 将 g.Map (snake_case) 映射到 DecisionActionDTO (camelCase)。
func mapToDecisionActionDTO(m g.Map) v1.DecisionActionDTO {
	return v1.DecisionActionDTO{
		ID:             mapJsonInt64(m, "id"),
		WorkflowRunID:  mapJsonInt64(m, "workflow_run_id"),
		ProjectID:      mapJsonInt64(m, "project_id"),
		StageRunID:     mapJsonInt64(m, "stage_run_id"),
		DomainTaskID:   mapJsonInt64(m, "domain_task_id"),
		DecisionType:   mapString(m, "decision_type"),
		DecisionLevel:  mapString(m, "decision_level"),
		TriggerSource:  mapString(m, "trigger_source"),
		TriggerContext: mapJSONString(m, "trigger_context"),
		MatchedRuleID:  mapJsonInt64(m, "matched_rule_id"),
		MatchedGateIDs: mapJSONString(m, "matched_gate_ids"),
		ActionType:     mapString(m, "action_type"),
		Recommendation: mapJSONString(m, "recommendation"),
		FinalAction:    mapString(m, "final_action"),
		ActionStatus:   mapString(m, "action_status"),
		AutoExecutable: mapInt(m, "auto_executable"),
		HumanRequired:  mapInt(m, "human_required"),
		ExecutedAt:     mapGTime(m, "executed_at"),
		Result:         mapJSONString(m, "result"),
		CreatedAt:      mapGTime(m, "created_at"),
	}
}

// AutonomyApprove 审批通过决策动作。
func (c *cWorkflow) AutonomyApprove(ctx context.Context, req *v1.WorkflowAutonomyApproveReq) (res *v1.WorkflowAutonomyApproveRes, err error) {
	actionID := int64(req.ActionID)
	// 从 action 记录中获取 project_id 做权限校验
	projectID, lookupErr := autonomyActionProjectID(ctx, actionID)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	if err := dc.ApproveAction(ctx, actionID); err != nil {
		return nil, err
	}
	return &v1.WorkflowAutonomyApproveRes{}, nil
}

// AutonomyReject 驳回决策动作。
func (c *cWorkflow) AutonomyReject(ctx context.Context, req *v1.WorkflowAutonomyRejectReq) (res *v1.WorkflowAutonomyRejectRes, err error) {
	actionID := int64(req.ActionID)
	projectID, lookupErr := autonomyActionProjectID(ctx, actionID)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	if err := dc.RejectAction(ctx, actionID, req.Reason); err != nil {
		return nil, err
	}
	return &v1.WorkflowAutonomyRejectRes{}, nil
}

// autonomyActionProjectID 从决策动作记录中查找关联的项目ID。
func autonomyActionProjectID(ctx context.Context, actionID int64) (int64, error) {
	val, err := g.DB().Model("mvp_decision_action").Ctx(ctx).
		Where("id", actionID).WhereNull("deleted_at").
		Value("project_id")
	if err != nil {
		return 0, fmt.Errorf("查询决策记录失败: %w", err)
	}
	if val.Int64() == 0 {
		return 0, fmt.Errorf("决策记录不存在: %d", actionID)
	}
	return val.Int64(), nil
}

// AutonomyActions 查询项目全量决策记录。
func (c *cWorkflow) AutonomyActions(ctx context.Context, req *v1.WorkflowAutonomyActionsReq) (res *v1.WorkflowAutonomyActionsRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	rawActions, queryErr := dc.ListAllActions(ctx, projectID, req.ActionStatus, req.DecisionType)
	if queryErr != nil {
		return nil, queryErr
	}

	actions := make([]v1.DecisionActionDTO, 0, len(rawActions))
	for _, m := range rawActions {
		actions = append(actions, mapToDecisionActionDTO(m))
	}
	return &v1.WorkflowAutonomyActionsRes{Actions: actions}, nil
}

// AutonomyGateRules 查询项目适用的风险闸门规则。
func (c *cWorkflow) AutonomyGateRules(ctx context.Context, req *v1.WorkflowAutonomyGateRulesReq) (res *v1.WorkflowAutonomyGateRulesRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	rawRules, queryErr := dc.ListGateRules(ctx, projectID)
	if queryErr != nil {
		return nil, queryErr
	}

	rules := make([]v1.RiskGateRuleDTO, 0, len(rawRules))
	for _, m := range rawRules {
		rules = append(rules, mapToRiskGateRuleDTO(m))
	}
	return &v1.WorkflowAutonomyGateRulesRes{Rules: rules}, nil
}

// AutonomyPolicyRules 查询项目适用的策略规则。
func (c *cWorkflow) AutonomyPolicyRules(ctx context.Context, req *v1.WorkflowAutonomyPolicyRulesReq) (res *v1.WorkflowAutonomyPolicyRulesRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	dc := orchestrator.GetDecisionCenter()
	rawRules, queryErr := dc.ListPolicyRules(ctx, projectID)
	if queryErr != nil {
		return nil, queryErr
	}

	rules := make([]v1.PolicyRuleDTO, 0, len(rawRules))
	for _, m := range rawRules {
		rules = append(rules, mapToPolicyRuleDTO(m))
	}
	return &v1.WorkflowAutonomyPolicyRulesRes{Rules: rules}, nil
}

// mapToRiskGateRuleDTO 将 g.Map (snake_case) 映射到 RiskGateRuleDTO (camelCase)。
func mapToRiskGateRuleDTO(m g.Map) v1.RiskGateRuleDTO {
	return v1.RiskGateRuleDTO{
		ID:                  mapJsonInt64(m, "id"),
		GateCode:            mapString(m, "gate_code"),
		GateName:            mapString(m, "gate_name"),
		GateType:            mapString(m, "gate_type"),
		ProjectFamily:       mapString(m, "project_family"),
		ProjectCategoryCode: mapString(m, "project_category_code"),
		TriggerExpression:   mapJSONString(m, "trigger_expression"),
		BlockAction:         mapString(m, "block_action"),
		FallbackAction:      mapString(m, "fallback_action"),
		Enabled:             mapInt(m, "enabled"),
		Priority:            mapInt(m, "priority"),
		CreatedAt:           mapGTime(m, "created_at"),
	}
}

// mapToPolicyRuleDTO 将 g.Map (snake_case) 映射到 PolicyRuleDTO (camelCase)。
func mapToPolicyRuleDTO(m g.Map) v1.PolicyRuleDTO {
	return v1.PolicyRuleDTO{
		ID:                  mapJsonInt64(m, "id"),
		RuleCode:            mapString(m, "rule_code"),
		RuleName:            mapString(m, "rule_name"),
		DecisionType:        mapString(m, "decision_type"),
		DecisionLevel:       mapString(m, "decision_level"),
		TriggerSource:       mapString(m, "trigger_source"),
		ProjectFamily:       mapString(m, "project_family"),
		ProjectCategoryCode: mapString(m, "project_category_code"),
		ConfigJSON:          mapJSONString(m, "config_json"),
		Enabled:             mapInt(m, "enabled"),
		Priority:            mapInt(m, "priority"),
		CreatedAt:           mapGTime(m, "created_at"),
	}
}

// ---- g.Map → DTO 映射辅助函数 ----

func mapString(m g.Map, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func mapInt(m g.Map, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return gconv.Int(v)
	}
}

func mapJsonInt64(m g.Map, key string) snowflake.JsonInt64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return snowflake.JsonInt64(n)
	case float64:
		return snowflake.JsonInt64(int64(n))
	case json.Number:
		i, _ := n.Int64()
		return snowflake.JsonInt64(i)
	default:
		return snowflake.JsonInt64(gconv.Int64(v))
	}
}

func mapGTime(m g.Map, key string) *gtime.Time {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case *gtime.Time:
		return t
	default:
		s := fmt.Sprintf("%v", v)
		if s == "" || s == "<nil>" {
			return nil
		}
		return gtime.New(s)
	}
}

func mapJSONString(m g.Map, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
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
