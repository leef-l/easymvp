package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/activity"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/workflow/autonomy"
	"easymvp/app/mvp/internal/workflow/compat"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
)

var Workflow = cWorkflow{}

type cWorkflow struct{}

// checkProjectOwnership 校验项目归属（复用 middleware.CheckOwnership，支持超管跳过）
func checkProjectOwnership(ctx context.Context, projectID int64) error {
	return middleware.CheckOwnership(ctx,
		g.DB().Model("mvp_project").WhereNull("deleted_at"),
		projectID, "id", "created_by",
	)
}

// CreateProject 创建项目
func (c *cWorkflow) CreateProject(ctx context.Context, req *v1.WorkflowCreateProjectReq) (res *v1.WorkflowCreateProjectRes, err error) {
	userID := middleware.GetUserID(ctx)
	deptID := middleware.GetDeptID(ctx)

	projectID, convID, err := engine.CreateProject(ctx, req.Name, req.ProjectCategory, req.Description, req.WorkDir, int64(req.ArchitectModelID), userID, deptID, req.EngineVersion)
	if err != nil {
		return nil, err
	}

	var wfRunID int64
	// 创建 WorkflowRun + 启动 design 阶段（默认 V2，仅显式 legacy 跳过）
	if req.EngineVersion != "legacy" {
		wfSvc := orchestrator.GetWorkflowService()
		wfRunID, err = wfSvc.CreateRun(ctx, projectID)
		if err != nil {
			g.Log().Warningf(ctx, "[CreateProject] CreateRun 失败: %v", err)
		} else {
			if err2 := wfSvc.StartDesign(ctx, wfRunID); err2 != nil {
				g.Log().Warningf(ctx, "[CreateProject] StartDesign 失败: %v", err2)
			}
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

	gw := compat.NewLegacyGateway()
	isV2, _ := gw.IsWorkflowV2(ctx, projectID)

	// Legacy 兼容分支
	if !isV2 {
		if err := engine.GetScheduler().ConfirmPlan(ctx, projectID); err != nil {
			return nil, err
		}
		return &v1.WorkflowConfirmPlanRes{}, nil
	}

	// V2 主路径
	if err := orchestrator.GetPlanVersionService().SubmitForReview(ctx, projectID); err != nil {
		return nil, err
	}
	return &v1.WorkflowConfirmPlanRes{}, nil
}

// Pause 暂停项目
func (c *cWorkflow) Pause(ctx context.Context, req *v1.WorkflowPauseReq) (res *v1.WorkflowPauseRes, err error) {
	projectID := int64(req.ProjectID)
	if err := checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	gw := compat.NewLegacyGateway()
	isV2, _ := gw.IsWorkflowV2(ctx, projectID)

	// Legacy 兼容分支
	if !isV2 {
		if err := engine.GetScheduler().Pause(ctx, projectID, req.PauseReason); err != nil {
			return nil, err
		}
		return &v1.WorkflowPauseRes{}, nil
	}

	// V2 主路径
	wfRun, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNotIn("status", g.Slice{"completed", "canceled", "paused"}).
		WhereNull("deleted_at").OrderDesc("run_no").One()
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

	gw := compat.NewLegacyGateway()
	isV2, _ := gw.IsWorkflowV2(ctx, projectID)

	// Legacy 兼容分支
	if !isV2 {
		if err := engine.GetScheduler().Resume(ctx, projectID); err != nil {
			return nil, err
		}
		return &v1.WorkflowResumeRes{}, nil
	}

	// V2 主路径
	wfRun, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		Where("status", "paused").
		WhereNull("deleted_at").OrderDesc("run_no").One()
	if wfRun.IsEmpty() {
		return nil, fmt.Errorf("没有暂停的工作流运行")
	}
	wfRunID := wfRun["id"].Int64()

	wfSvc := orchestrator.GetWorkflowService()
	if err := wfSvc.Resume(ctx, wfRunID); err != nil {
		return nil, err
	}
	if wfRun["current_stage"].String() == "execute" {
		_ = orchestrator.GetTaskScheduler().Start(ctx, wfRunID)
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

	gw := compat.NewLegacyGateway()
	isV2, _ := gw.IsWorkflowV2(ctx, projectID)

	// Legacy 兼容分支
	if !isV2 {
		engine.GetWatchdog().ResetRetryCount(taskID)
		if err := engine.GetScheduler().RetryTask(projectID, taskID); err != nil {
			return nil, err
		}
		return &v1.WorkflowRetryTaskRes{}, nil
	}

	// V2 主路径：domain_task failed → pending
	result, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).Where("status", "failed").
		Update(g.Map{
			"status":      "pending",
			"retry_count": g.Map{"retry_count": "retry_count + 1"},
			"result":      nil,
			"updated_at":  g.Map{"updated_at": "NOW()"},
		})
	if err != nil {
		return nil, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("任务(%d)不在 failed 状态，无法重试", taskID)
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

	gw := compat.NewLegacyGateway()
	isV2, _ := gw.IsWorkflowV2(ctx, projectID)

	// Legacy 兼容分支
	if !isV2 {
		if err := engine.GetScheduler().SkipTask(ctx, projectID, taskID, req.Reason); err != nil {
			return nil, err
		}
		return &v1.WorkflowSkipTaskRes{}, nil
	}

	// V2 主路径：跳过领域任务 (pending/failed → completed, result=skipped)
	result, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).
		WhereIn("status", g.Slice{"pending", "failed"}).
		Update(g.Map{
			"status":       "completed",
			"result":       "skipped",
			"completed_at": g.Map{"completed_at": "NOW()"},
			"updated_at":   g.Map{"updated_at": "NOW()"},
		})
	if err != nil {
		return nil, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("任务不在可跳过的状态")
	}
	_ = orchestrator.GetTaskScheduler().OnTaskCompleted(ctx, taskID)
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
	lastUserMsg, err := g.DB().Model("mvp_message").
		Where("conversation_id", convID).
		Where("role", "user").
		Where("status", "completed").
		Where("deleted_at IS NULL").
		OrderDesc("created_at").
		One()
	if err != nil || lastUserMsg.IsEmpty() {
		return &v1.WorkflowParseTasksRes{HasTasks: false, TaskCount: 0}, nil
	}
	lastUserTime := lastUserMsg["created_at"].String()

	// 取该 user 消息之后的所有 assistant 回复（即最新一轮方案）
	allMsgs, err := g.DB().Model("mvp_message").
		Where("conversation_id", convID).
		Where("role", "assistant").
		Where("status", "completed").
		Where("deleted_at IS NULL").
		Where("created_at >= ?", lastUserTime).
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

	// 判断引擎版本
	gw := compat.NewLegacyGateway()
	isV2, _ := gw.IsWorkflowV2(ctx, projectID)

	if req.DryRun {
		count := engine.GetParser().DryParseTaskCount(aiReply)
		return &v1.WorkflowParseTasksRes{
			HasTasks:  count > 0,
			TaskCount: count,
		}, nil
	}

	// Legacy 兼容分支
	if !isV2 {
		count, err := engine.GetParser().ParseAndCreateTasks(ctx, projectID, aiReply)
		if err != nil {
			return nil, err
		}
		return &v1.WorkflowParseTasksRes{HasTasks: count > 0, TaskCount: count}, nil
	}

	// V2 主路径：解析为蓝图写入 plan_version + task_blueprint
	count, err := parseAndCreateBlueprints(ctx, projectID, convID, lastMsgID, aiReply)
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowParseTasksRes{HasTasks: count > 0, TaskCount: count}, nil
}

// RolePresets 获取角色预设列表（前端创建项目时读取默认模型）
func (c *cWorkflow) RolePresets(ctx context.Context, req *v1.WorkflowRolePresetsReq) (res *v1.WorkflowRolePresetsRes, err error) {
	m := g.DB().Model("mvp_role_preset AS p").
		LeftJoin("ai_model AS m", "m.id = p.model_id").
		Fields("p.role_type, p.role_level, p.model_id, m.name AS model_name, p.system_prompt, p.execution_mode, p.is_default").
		Where("p.status", 1).
		Where("p.deleted_at IS NULL")
	if req.ProjectCategory != "" {
		m = m.Where("p.project_category", req.ProjectCategory)
	}
	if !req.All {
		m = m.Where("p.is_default", 1)
	}
	presets, err := m.OrderAsc("p.sort").All()
	if err != nil {
		return nil, err
	}

	list := make([]v1.RolePresetItem, 0, len(presets))
	for _, p := range presets {
		list = append(list, v1.RolePresetItem{
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

	// 查项目状态
	project, err := g.DB().Model("mvp_project").Where("id", projectID).Where("deleted_at IS NULL").One()
	if err != nil {
		return nil, err
	}
	if project.IsEmpty() {
		return nil, fmt.Errorf("项目不存在")
	}

	gw := compat.NewLegacyGateway()
	isV2, _ := gw.IsWorkflowV2(ctx, projectID)

	// Legacy 兼容分支
	if !isV2 {
		type StatusCount struct {
			Status string `json:"status"`
			Count  int    `json:"count"`
		}
		var counts []StatusCount
		if err := g.DB().Model("mvp_task").
			Where("project_id", projectID).
			Where("deleted_at IS NULL").
			Fields("status, COUNT(*) as count").
			Group("status").
			Scan(&counts); err != nil {
			return nil, err
		}

		statusCounts := make(map[string]int)
		total := 0
		for _, sc := range counts {
			statusCounts[sc.Status] = sc.Count
			total += sc.Count
		}

		activitySummary, err := activity.LoadProjectSummary(ctx, projectID)
		if err != nil {
			return nil, err
		}

		return &v1.WorkflowProjectStatusRes{
			Status:             project["status"].String(),
			PauseReason:        project["pause_reason"].String(),
			ActiveBatch:        engine.GetScheduler().GetActiveBatch(projectID),
			TotalTasks:         total,
			StatusCounts:       statusCounts,
			LastActiveAt:       activitySummary.LastActiveAt,
			IsActuallyWorking:  activitySummary.IsActuallyWorking,
			ActiveRunningTasks: activitySummary.ActiveRunningTasks,
			StalledTaskCount:   activitySummary.StalledTaskCount,
		}, nil
	}

	// V2 主路径
	return projectStatusV2(ctx, project)
}

// projectStatusV2 V2 引擎的项目状态聚合。
func projectStatusV2(ctx context.Context, project gdb.Record) (*v1.WorkflowProjectStatusRes, error) {
	projectID := project["id"].Int64()

	// 使用 ProjectStatusAdapter 获取聚合状态
	adapter := compat.NewProjectStatusAdapter()
	dto, err := adapter.GetProjectStatus(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 蓝图状态统计（设计阶段用）
	type StatusCount struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	var counts []StatusCount
	_ = g.DB().Model("mvp_task_blueprint AS bp").
		InnerJoin("mvp_plan_version AS pv", "pv.id = bp.plan_version_id").
		Where("pv.project_id", projectID).
		WhereIn("pv.status", g.Slice{"draft", "active"}).
		WhereNull("bp.deleted_at").
		Fields("bp.blueprint_status AS status, COUNT(*) AS count").
		Group("bp.blueprint_status").
		Scan(&counts)

	statusCounts := make(map[string]int)
	bpTotal := 0
	for _, sc := range counts {
		statusCounts[sc.Status] = sc.Count
		bpTotal += sc.Count
	}

	// 合并领域任务统计到 statusCounts
	if dto.TotalTasks > 0 {
		statusCounts["domain_total"] = dto.TotalTasks
		statusCounts["domain_completed"] = dto.CompletedTasks
		statusCounts["domain_failed"] = dto.FailedTasks
		statusCounts["domain_running"] = dto.RunningTasks
	}

	totalTasks := bpTotal
	if dto.TotalTasks > 0 {
		totalTasks = dto.TotalTasks
	}

	res := &v1.WorkflowProjectStatusRes{
		Status:          project["status"].String(),
		PauseReason:     project["pause_reason"].String(),
		TotalTasks:      totalTasks,
		StatusCounts:    statusCounts,
		EngineVersion:   dto.EngineVersion,
		WorkflowStatus:  dto.WorkflowStatus,
		CurrentStage:    dto.CurrentStage,
		ProgressPercent: dto.ProgressPercent,
	}

	// V2 项目状态以 workflow 状态为准
	if dto.WorkflowStatus != "" {
		res.Status = dto.WorkflowStatus
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
	architectCount, _ := g.DB().Model("mvp_role_preset").
		Where("role_type", "architect").Where("status", 1).Where("deleted_at IS NULL").Count()
	implementerCount, _ := g.DB().Model("mvp_role_preset").
		Where("role_type", "implementer").Where("status", 1).Where("deleted_at IS NULL").Count()
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

	return &v1.SystemCheckRes{Items: items, AllPass: allPass}, nil
}

// parseAndCreateBlueprints V2 专用：解析 AI 回复并创建蓝图。
func parseAndCreateBlueprints(ctx context.Context, projectID, conversationID, messageID int64, aiReply string) (int, error) {
	// 获取项目分类
	projectCategory, _ := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).Value("project_category")

	// 复用旧 TaskParser 提取结构化任务
	tasks, err := engine.GetParser().ExtractAndNormalize(ctx, aiReply, projectCategory.String())
	if err != nil || len(tasks) == 0 {
		return 0, err
	}

	// 查当前活跃的 workflow_run
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

	// 创建 plan_version + blueprints
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
	_, _ = g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).
		Update(g.Map{"status": "designing", "updated_at": g.Map{"updated_at": "NOW()"}})

	return &v1.WorkflowManualRejectRes{}, nil
}

// ==================== Timeline / Rework / Stage History ====================

// eventLabelMap 事件类型 → 可读标签
var eventLabelMap = map[string]string{
	"workflow.created":        "工作流已创建",
	"workflow.paused":         "工作流已暂停",
	"workflow.resumed":        "工作流已恢复",
	"workflow.canceled":       "工作流已取消",
	"workflow.completed":      "工作流已完成",
	"stage.started":           "阶段已启动",
	"stage.completed":         "阶段已完成",
	"stage.failed":            "阶段失败",
	"plan_version.created":    "方案版本已创建",
	"plan_version.submitted":  "方案已提交审核",
	"plan_version.approved":   "方案审核通过",
	"plan_version.rejected":   "方案被驳回",
	"review.issue_created":    "发现审核问题",
	"review.decision_ready":   "审核决策就绪",
	"task.created":            "任务已创建",
	"task.started":            "任务已启动",
	"task.completed":          "任务已完成",
	"task.failed":             "任务失败",
	"task.escalated":          "任务已升级",
	"task.retried":            "任务已重试",
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
	stageRun, _ := g.DB().Model("mvp_stage_run").Ctx(ctx).
		Where("workflow_run_id", wfRunID).
		Where("stage_type", "execute").
		WhereNull("deleted_at").
		OrderDesc("stage_no").
		One()
	if !stageRun.IsEmpty() {
		res.StageRunID = snowflake.JsonInt64(stageRun["id"].Int64())
		res.StageStatus = stageRun["status"].String()
	}

	// 查领域任务
	tasks, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", wfRunID).
		WhereNull("deleted_at").
		OrderAsc("batch_no").
		OrderAsc("sort").
		All()

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
	wfRun, _ := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("run_no").
		One()
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

	_, replanErr := replanner.Evaluate(ctx, &autonomy.ReplanInput{
		WorkflowRunID: wfRun["id"].Int64(),
		ProjectID:     projectID,
		TriggerSource: "manual",
		FailedTasks:   failed,
	})
	if replanErr != nil {
		return nil, replanErr
	}
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

	// 权限过滤：只保留用户有权访问的项目
	userID := middleware.GetUserID(ctx)
	if userID != 1 { // 超管跳过
		ownedProjects, _ := g.DB().Model("mvp_project").Ctx(ctx).
			WhereIn("id", ids).
			Where("created_by", userID).
			WhereNull("deleted_at").
			Fields("id").All()
		allowedIDs := make(map[int64]bool)
		for _, p := range ownedProjects {
			allowedIDs[p["id"].Int64()] = true
		}
		filtered := ids[:0]
		for _, id := range ids {
			if allowedIDs[id] {
				filtered = append(filtered, id)
			}
		}
		ids = filtered
	}
	if len(ids) == 0 {
		return &v1.WorkflowBatchProjectStatsRes{Stats: []v1.ProjectRuntimeStat{}}, nil
	}

	// 批量查 workflow_run（排除已完成/已取消的旧 run，只取活跃或最新的）
	wfRuns, err := g.DB().Model("mvp_workflow_run").Ctx(ctx).
		WhereIn("project_id", ids).
		WhereNull("deleted_at").
		Fields("id, project_id, current_stage, status").
		OrderDesc("run_no").
		All()
	if err != nil {
		return nil, err
	}

	// project_id → workflow_run 映射（取最新的）
	wfMap := make(map[int64]gdb.Record)
	for _, r := range wfRuns {
		pid := r["project_id"].Int64()
		if _, exists := wfMap[pid]; !exists {
			wfMap[pid] = r
		}
	}

	// 收集所有 workflow_run_id
	wfRunIDs := make([]int64, 0, len(wfMap))
	for _, r := range wfMap {
		wfRunIDs = append(wfRunIDs, r["id"].Int64())
	}

	// 批量统计 domain_task（按 workflow_run_id 分组）
	type taskStat struct {
		total, completed, failed, running int
	}
	taskStats := make(map[int64]*taskStat)

	if len(wfRunIDs) > 0 {
		tasks, _ := g.DB().Model("mvp_domain_task").Ctx(ctx).
			WhereIn("workflow_run_id", wfRunIDs).
			WhereNull("deleted_at").
			Fields("workflow_run_id, status").
			All()
		for _, t := range tasks {
			wfID := t["workflow_run_id"].Int64()
			if taskStats[wfID] == nil {
				taskStats[wfID] = &taskStat{}
			}
			s := taskStats[wfID]
			s.total++
			switch t["status"].String() {
			case "completed":
				s.completed++
			case "failed", "escalated":
				s.failed++
			case "running":
				s.running++
			}
		}
	}

	// legacy 项目统计（独立 map，避免 key 冲突）
	legacyStats := make(map[int64]*taskStat)
	for _, pid := range ids {
		if _, exists := wfMap[pid]; exists {
			continue
		}
		tasks, _ := g.DB().Model("mvp_task").Ctx(ctx).
			Where("project_id", pid).
			WhereNull("deleted_at").
			WhereNotIn("status", g.Slice{"draft"}).
			Fields("status").All()
		if len(tasks) > 0 {
			st := &taskStat{}
			for _, t := range tasks {
				st.total++
				switch t["status"].String() {
				case "completed":
					st.completed++
				case "failed":
					st.failed++
				case "running":
					st.running++
				}
			}
			legacyStats[pid] = st
		}
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
			if ts, exists := taskStats[wfID]; exists {
				stat.TotalTasks = ts.total
				stat.CompletedTasks = ts.completed
				stat.FailedTasks = ts.failed
				stat.RunningTasks = ts.running
			}
		} else if ts, exists := legacyStats[pid]; exists {
			stat.TotalTasks = ts.total
			stat.CompletedTasks = ts.completed
			stat.FailedTasks = ts.failed
			stat.RunningTasks = ts.running
		}
		stats = append(stats, stat)
	}

	return &v1.WorkflowBatchProjectStatsRes{Stats: stats}, nil
}
