package chat

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/middleware"
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

	projectID, convID, err := engine.CreateProject(ctx, req.Name, req.Description, req.WorkDir, int64(req.ArchitectModelID), userID, deptID)
	if err != nil {
		return nil, err
	}

	return &v1.WorkflowCreateProjectRes{
		ProjectID:      snowflake.JsonInt64(projectID),
		ConversationID: snowflake.JsonInt64(convID),
	}, nil
}

// ConfirmPlan 确认实施方案
func (c *cWorkflow) ConfirmPlan(ctx context.Context, req *v1.WorkflowConfirmPlanReq) (res *v1.WorkflowConfirmPlanRes, err error) {
	if err := checkProjectOwnership(ctx, int64(req.ProjectID)); err != nil {
		return nil, err
	}
	err = engine.GetScheduler().ConfirmPlan(ctx, int64(req.ProjectID))
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowConfirmPlanRes{}, nil
}

// Pause 暂停项目
func (c *cWorkflow) Pause(ctx context.Context, req *v1.WorkflowPauseReq) (res *v1.WorkflowPauseRes, err error) {
	if err := checkProjectOwnership(ctx, int64(req.ProjectID)); err != nil {
		return nil, err
	}
	err = engine.GetScheduler().Pause(ctx, int64(req.ProjectID), req.PauseReason)
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowPauseRes{}, nil
}

// Resume 恢复项目
func (c *cWorkflow) Resume(ctx context.Context, req *v1.WorkflowResumeReq) (res *v1.WorkflowResumeRes, err error) {
	if err := checkProjectOwnership(ctx, int64(req.ProjectID)); err != nil {
		return nil, err
	}
	err = engine.GetScheduler().Resume(ctx, int64(req.ProjectID))
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowResumeRes{}, nil
}

// RetryTask 重新执行失败任务
func (c *cWorkflow) RetryTask(ctx context.Context, req *v1.WorkflowRetryTaskReq) (res *v1.WorkflowRetryTaskRes, err error) {
	if err := checkProjectOwnership(ctx, int64(req.ProjectID)); err != nil {
		return nil, err
	}
	engine.GetWatchdog().ResetRetryCount(int64(req.TaskID))
	err = engine.GetScheduler().RetryTask(int64(req.ProjectID), int64(req.TaskID))
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowRetryTaskRes{}, nil
}

// SkipTask 跳过失败任务（防止批次永久阻塞）
func (c *cWorkflow) SkipTask(ctx context.Context, req *v1.WorkflowSkipTaskReq) (res *v1.WorkflowSkipTaskRes, err error) {
	if err := checkProjectOwnership(ctx, int64(req.ProjectID)); err != nil {
		return nil, err
	}
	err = engine.GetScheduler().SkipTask(ctx, int64(req.ProjectID), int64(req.TaskID), req.Reason)
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowSkipTaskRes{}, nil
}

// ParseTasks 手动解析架构师回复中的任务清单（托底机制）
// dryRun=true 时仅检查不创建，dryRun=false 时实际创建草案任务
func (c *cWorkflow) ParseTasks(ctx context.Context, req *v1.WorkflowParseTasksReq) (res *v1.WorkflowParseTasksRes, err error) {
	if err := checkProjectOwnership(ctx, int64(req.ProjectID)); err != nil {
		return nil, err
	}
	projectID := int64(req.ProjectID)

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

	if req.DryRun {
		// 仅检查，不创建
		count := engine.GetParser().DryParseTaskCount(aiReply)
		return &v1.WorkflowParseTasksRes{
			HasTasks:  count > 0,
			TaskCount: count,
		}, nil
	}

	// 实际创建草案任务
	count, err := engine.GetParser().ParseAndCreateTasks(ctx, projectID, aiReply)
	if err != nil {
		return nil, err
	}

	return &v1.WorkflowParseTasksRes{
		HasTasks:  count > 0,
		TaskCount: count,
	}, nil
}

// RolePresets 获取角色预设列表（前端创建项目时读取默认模型）
func (c *cWorkflow) RolePresets(ctx context.Context, req *v1.WorkflowRolePresetsReq) (res *v1.WorkflowRolePresetsRes, err error) {
	presets, err := g.DB().Model("mvp_role_preset AS p").
		LeftJoin("ai_model AS m", "m.id = p.model_id").
		Fields("p.role_type, p.role_level, p.model_id, m.name AS model_name, p.system_prompt").
		Where("p.status", 1).
		Where("p.deleted_at IS NULL").
		OrderAsc("p.sort").
		All()
	if err != nil {
		return nil, err
	}

	list := make([]v1.RolePresetItem, 0, len(presets))
	for _, p := range presets {
		list = append(list, v1.RolePresetItem{
			RoleType:     p["role_type"].String(),
			RoleLevel:    p["role_level"].String(),
			ModelID:      snowflake.JsonInt64(p["model_id"].Int64()),
			ModelName:    p["model_name"].String(),
			SystemPrompt: p["system_prompt"].String(),
		})
	}

	return &v1.WorkflowRolePresetsRes{List: list}, nil
}

// ProjectStatus 获取项目状态
func (c *cWorkflow) ProjectStatus(ctx context.Context, req *v1.WorkflowProjectStatusReq) (res *v1.WorkflowProjectStatusRes, err error) {
	if err := checkProjectOwnership(ctx, int64(req.ProjectID)); err != nil {
		return nil, err
	}
	projectID := int64(req.ProjectID)

	// 查项目状态
	project, err := g.DB().Model("mvp_project").Where("id", projectID).Where("deleted_at IS NULL").One()
	if err != nil {
		return nil, err
	}
	if project.IsEmpty() {
		return nil, fmt.Errorf("项目不存在")
	}

	// 统计各状态任务数
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

	return &v1.WorkflowProjectStatusRes{
		Status:       project["status"].String(),
		PauseReason:  project["pause_reason"].String(),
		ActiveBatch:  engine.GetScheduler().GetActiveBatch(projectID),
		TotalTasks:   total,
		StatusCounts: statusCounts,
	}, nil
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
