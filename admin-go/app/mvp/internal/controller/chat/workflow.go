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
	"easymvp/app/mvp/internal/workflow/compat"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/utility/snowflake"
)

var Workflow = cWorkflow{}

type cWorkflow struct{}

// checkProjectOwnership 校验项目归属
func checkProjectOwnership(ctx context.Context, projectID int64) error {
	userID := middleware.GetUserID(ctx)
	if userID == 1 {
		return nil
	}
	project, err := g.DB().Model("mvp_project").Where("id", projectID).Where("deleted_at IS NULL").One()
	if err != nil || project.IsEmpty() {
		return fmt.Errorf("项目不存在")
	}
	if project["created_by"].Int64() != userID {
		return fmt.Errorf("无权操作该项目")
	}
	return nil
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

	// 查找最新的 assistant 消息
	msg, err := g.DB().Model("mvp_message").
		Where("conversation_id", conv["id"].Int64()).
		Where("role", "assistant").
		Where("deleted_at IS NULL").
		OrderDesc("created_at").
		One()
	if err != nil || msg.IsEmpty() {
		return &v1.WorkflowParseTasksRes{HasTasks: false, TaskCount: 0}, nil
	}

	aiReply := msg["content"].String()

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
	count, err := parseAndCreateBlueprints(ctx, projectID, conv["id"].Int64(), msg["id"].Int64(), aiReply)
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
