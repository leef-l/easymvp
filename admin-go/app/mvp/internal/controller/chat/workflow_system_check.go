package chat

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/regression"
	"easymvp/app/mvp/internal/workflow/repo"
	workspacepkg "easymvp/app/mvp/internal/workspace"
)

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

	count, e := g.DB().Ctx(ctx).Model("ai_provider").
		Where("status", 1).Where("base_url != ''").WhereNull("deleted_at").Count()
	if e != nil {
		addItem("ai_provider", "AI 供应商", "/ai/provider", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_provider", "AI 供应商", "/ai/provider", "error", "未配置启用的 AI 供应商（需要有 base_url）")
	} else {
		addItem("ai_provider", "AI 供应商", "/ai/provider", "ok", fmt.Sprintf("已有 %d 个启用供应商", count))
	}

	count, e = g.DB().Ctx(ctx).Model("ai_plan").
		Where("status", 1).Where("api_key != ''").WhereNull("deleted_at").Count()
	if e != nil {
		addItem("ai_plan", "AI 套餐", "/ai/plan", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_plan", "AI 套餐", "/ai/plan", "error", "未配置启用的 AI 套餐（需要有 api_key）")
	} else {
		addItem("ai_plan", "AI 套餐", "/ai/plan", "ok", fmt.Sprintf("已有 %d 个启用套餐", count))
	}

	count, e = g.DB().Ctx(ctx).Model("ai_model").
		Where("capability", "architect").Where("status", 1).WhereNull("deleted_at").Count()
	if e != nil {
		addItem("ai_model_architect", "AI 模型（架构师）", "/ai/model", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_model_architect", "AI 模型（架构师）", "/ai/model", "error", "未配置 capability=architect 且启用的 AI 模型")
	} else {
		addItem("ai_model_architect", "AI 模型（架构师）", "/ai/model", "ok", fmt.Sprintf("已有 %d 个架构师模型", count))
	}

	count, e = g.DB().Ctx(ctx).Model("ai_model").
		WhereIn("capability", g.Slice{"implementer", "coding", "chat"}).
		Where("status", 1).WhereNull("deleted_at").Count()
	if e != nil {
		addItem("ai_model_implementer", "AI 模型（实施员）", "/ai/model", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_model_implementer", "AI 模型（实施员）", "/ai/model", "error", "未配置 capability 为 implementer/coding/chat 且启用的 AI 模型")
	} else {
		addItem("ai_model_implementer", "AI 模型（实施员）", "/ai/model", "ok", fmt.Sprintf("已有 %d 个实施员模型", count))
	}

	architectCount, _ := repo.CountRolePresets(ctx, repo.RolePresetQuery{RoleType: "architect"})
	implementerCount, _ := repo.CountRolePresets(ctx, repo.RolePresetQuery{RoleType: "implementer"})
	if architectCount == 0 || implementerCount == 0 {
		addItem("role_preset", "角色预设", "/mvp/role-preset", "error",
			fmt.Sprintf("缺少角色预设：架构师=%d，实施员=%d（各需至少 1 条）", architectCount, implementerCount))
	} else {
		addItem("role_preset", "角色预设", "/mvp/role-preset", "ok",
			fmt.Sprintf("架构师预设 %d 条，实施员预设 %d 条", architectCount, implementerCount))
	}

	count, e = g.DB().Ctx(ctx).Model("ai_engine").
		Where("status", 1).WhereNull("deleted_at").Count()
	if e != nil {
		addItem("ai_engine", "AI 执行引擎", "/ai/engine", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_engine", "AI 执行引擎", "/ai/engine", "error", "未配置启用的 AI 执行引擎")
	} else {
		addItem("ai_engine", "AI 执行引擎", "/ai/engine", "ok", fmt.Sprintf("已有 %d 个启用引擎", count))
	}

	aiderCfg, e := g.DB().Ctx(ctx).Model("ai_engine_config").
		Where("engine_code", "aider").WhereNull("deleted_at").One()
	if e != nil || aiderCfg.IsEmpty() {
		addItem("ai_engine_config_aider", "Aider 引擎配置", "/ai/engine", "error", "未配置 Aider 引擎参数")
	} else if aiderCfg["workspace_root"].String() == "" {
		addItem("ai_engine_config_aider", "Aider 引擎配置", "/ai/engine", "warning", "Aider 引擎未配置 workspace_root（工作区根目录）")
	} else {
		addItem("ai_engine_config_aider", "Aider 引擎配置", "/ai/engine", "ok",
			"工作区根目录: "+aiderCfg["workspace_root"].String())
	}

	ohCfg, e := g.DB().Ctx(ctx).Model("ai_engine_config").
		Where("engine_code", "openhands").WhereNull("deleted_at").One()
	if e != nil || ohCfg.IsEmpty() {
		addItem("ai_engine_config_openhands", "OpenHands 引擎配置", "/ai/engine", "warning", "未配置 OpenHands 引擎参数（非必须，仅使用 Aider 可忽略）")
	} else if ohCfg["command_template"].String() == "" {
		addItem("ai_engine_config_openhands", "OpenHands 引擎配置", "/ai/engine", "warning", "OpenHands 未配置 command_template（命令模板）")
	} else {
		addItem("ai_engine_config_openhands", "OpenHands 引擎配置", "/ai/engine", "ok", "命令模板已配置")
	}

	count, e = g.DB().Ctx(ctx).Model("system_role_ai_engine").Count()
	if e != nil {
		addItem("role_ai_engine", "角色引擎授权", "", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("role_ai_engine", "角色引擎授权", "", "error", "没有角色被授权使用 AI 引擎，请在角色管理中配置")
	} else {
		addItem("role_ai_engine", "角色引擎授权", "", "ok", fmt.Sprintf("已有 %d 条角色引擎授权", count))
	}

	if aiderPath, err := exec.LookPath("aider"); err == nil {
		addItem("aider_installed", "Aider 执行环境", "", "ok", "aider 已安装: "+aiderPath)
	} else if uvPath, uvErr := exec.LookPath("uv"); uvErr == nil {
		addItem("aider_installed", "Aider 执行环境", "", "ok", "本机未安装 aider，将通过 uv 自动安装/执行: "+uvPath)
	} else if dockerPath, dockerErr := exec.LookPath("docker"); dockerErr == nil {
		addItem("aider_installed", "Aider 执行环境", "", "warning", "本机未安装 aider/uv，将回退使用 Docker 执行: "+dockerPath)
	} else {
		addItem("aider_installed", "Aider 执行环境", "", "error", "未找到 aider 可执行文件，且 uv/docker 都不可用")
	}

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

	requiredKeys := []string{
		"runtime.task_timeout_seconds",
		"runtime.max_steps",
		"watchdog.check_interval",
		"watchdog.max_stale_count",
		"watchdog.max_retries",
		"scheduler.poll_interval",
	}
	count, e = g.DB().Ctx(ctx).Model("mvp_config").
		WhereIn("config_key", requiredKeys).WhereNull("deleted_at").Count()
	if e != nil {
		addItem("engine_config", "引擎核心配置", "/mvp/config", "error", "查询失败: "+e.Error())
	} else {
		schedulerCount, schedulerErr := g.DB().Ctx(ctx).Model("mvp_config").
			WhereIn("config_key", []string{"scheduler.max_concurrent", "workflow.scheduler.max_concurrency"}).
			WhereNull("deleted_at").Count()
		if schedulerErr != nil {
			addItem("engine_config", "引擎核心配置", "/mvp/config", "error", "查询调度并发配置失败: "+schedulerErr.Error())
		} else if count < len(requiredKeys) || schedulerCount == 0 {
			addItem("engine_config", "引擎核心配置", "/mvp/config", "warning",
				fmt.Sprintf("核心配置仅有 %d/%d 项，且调度并发键需要至少配置 1 个，缺少项将使用默认值", count, len(requiredKeys)))
		} else {
			addItem("engine_config", "引擎核心配置", "/mvp/config", "ok",
				fmt.Sprintf("全部 %d 项核心配置已就绪，调度并发兼容键已配置", len(requiredKeys)+1))
		}
	}

	commandResourcePolicy := engine.GetCommandResourcePolicy(ctx)
	commandResourceStatus := "ok"
	if !commandResourcePolicy.Enabled {
		commandResourceStatus = "warning"
	}
	addItem("command_resource", "命令资源限制", "/mvp/config", commandResourceStatus, commandResourcePolicy.Summary())

	categoryRecords, categoryErr := g.DB().Ctx(ctx).Model("mvp_project_category").
		Where("status", 1).
		WhereNull("deleted_at").
		Fields("verification_profile_json, verification_gate_json").
		All()
	if categoryErr != nil {
		addItem("project_category_verification", "项目分类验证配置", "/mvp/project_category", "error", "查询失败: "+categoryErr.Error())
	} else {
		totalCategories := len(categoryRecords)
		profileConfigured := 0
		gateConfigured := 0
		for _, row := range categoryRecords {
			if strings.TrimSpace(row["verification_profile_json"].String()) != "" {
				profileConfigured++
			}
			if strings.TrimSpace(row["verification_gate_json"].String()) != "" {
				gateConfigured++
			}
		}
		status := "ok"
		message := fmt.Sprintf("已启用分类 %d 个，gate 已配置 %d/%d，profile 已配置 %d/%d", totalCategories, gateConfigured, totalCategories, profileConfigured, totalCategories)
		switch {
		case totalCategories == 0:
			status = "warning"
			message = "尚未配置启用的项目分类"
		case gateConfigured < totalCategories:
			status = "warning"
			message += "；建议为未配置分类补齐 verification_gate_json"
		}
		addItem("project_category_verification", "项目分类验证配置", "/mvp/project_category", status, message)
	}

	deliveryMode := strings.TrimSpace(engine.GetConfigString(ctx,
		"workspace.delivery.default_mode",
		"engine.workspace.delivery.defaultMode",
		workspacepkg.DeliveryModePatch,
	))
	syncStrategy := strings.TrimSpace(engine.GetConfigString(ctx,
		"workspace.delivery.default_sync_strategy",
		"engine.workspace.delivery.defaultSyncStrategy",
		workspacepkg.SyncStrategyAutoApply,
	))
	switch {
	case deliveryMode == workspacepkg.DeliveryModePatch &&
		(syncStrategy == workspacepkg.SyncStrategyAutoApply || syncStrategy == workspacepkg.SyncStrategyManual):
		addItem("workspace_delivery", "Workspace 交付策略", "/mvp/config", "ok",
			fmt.Sprintf("默认结果=%s，默认回写策略=%s", deliveryMode, syncStrategy))
	case deliveryMode == workspacepkg.DeliveryModePR && syncStrategy == workspacepkg.SyncStrategyManual:
		addItem("workspace_delivery", "Workspace 交付策略", "/mvp/config", "ok",
			fmt.Sprintf("默认结果=%s，PR 草稿交付已启用，回写策略=%s", deliveryMode, syncStrategy))
	case deliveryMode == workspacepkg.DeliveryModeManual:
		addItem("workspace_delivery", "Workspace 交付策略", "/mvp/config", "warning",
			fmt.Sprintf("默认结果=%s，建议确认人工审核与落库流程是否已接入", deliveryMode))
	default:
		addItem("workspace_delivery", "Workspace 交付策略", "/mvp/config", "warning",
			fmt.Sprintf("检测到非标准配置：defaultMode=%s，defaultSyncStrategy=%s", deliveryMode, syncStrategy))
	}

	riskPolicies := workspacepkg.GetRiskDeliveryPolicies(ctx)
	riskPolicyStatus, riskPolicyMessage := summarizeRiskDeliveryPolicies(riskPolicies)
	addItem("workspace_delivery_risk", "Workspace 风险交付矩阵", "/mvp/config", riskPolicyStatus, riskPolicyMessage)

	deliveryColumns := []string{
		"delivery_mode",
		"delivery_status",
		"sync_strategy",
		"sync_status",
		"risk_level",
		"patch_ref",
		"delivery_ref",
		"delivery_title",
	}
	_, columnErr := g.DB().Ctx(ctx).Model("mvp_task_workspace").
		Fields(strings.Join(deliveryColumns, ",")).
		Limit(1).
		All()
	if columnErr != nil {
		addItem("workspace_migration", "Workspace 交付 Migration", "", "warning",
			"mvp_task_workspace 交付结果列尚未全部就绪，需执行最新 migration")
	} else {
		addItem("workspace_migration", "Workspace 交付 Migration", "", "ok",
			fmt.Sprintf("mvp_task_workspace 新列已就绪 %d/%d", len(deliveryColumns), len(deliveryColumns)))
	}

	deliveryRuleCount, deliveryRuleErr := g.DB().Ctx(ctx).Model("mvp_accept_rule").
		Where("rule_code", "software.delivery_review_required").
		WhereIn("project_type", []string{"software_dev", "coding"}).
		WhereNull("deleted_at").
		Count()
	if deliveryRuleErr != nil {
		addItem("accept_delivery_rule", "验收交付审核规则", "", "warning", "无法检查 software.delivery_review_required 规则: "+deliveryRuleErr.Error())
	} else if deliveryRuleCount < 2 {
		addItem("accept_delivery_rule", "验收交付审核规则", "", "warning",
			fmt.Sprintf("software.delivery_review_required 规则仅就绪 %d/2，需执行最新 migration", deliveryRuleCount))
	} else {
		addItem("accept_delivery_rule", "验收交付审核规则", "", "ok",
			fmt.Sprintf("software.delivery_review_required 规则已就绪 %d/2", deliveryRuleCount))
	}

	if manifestPath, manifestErr := regression.ResolveManifestPath(); manifestErr != nil {
		addItem("regression_manifest", "回归样例清单", "", "warning", "未找到 regression-manifest.json")
	} else if report, validateErr := regression.ValidateManifest(manifestPath); validateErr != nil {
		addItem("regression_manifest", "回归样例清单", "", "warning",
			fmt.Sprintf("样例清单存在但校验失败: %s", validateErr.Error()))
	} else {
		addItem("regression_manifest", "回归样例清单", "", "ok",
			fmt.Sprintf("已加载并校验样例清单: ready=%d planned=%d (%s)", report.ReadyCount, report.PlannedCount, manifestPath))
	}

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

func summarizeRiskDeliveryPolicies(policies map[string]workspacepkg.RiskDeliveryPolicy) (string, string) {
	report := workspacepkg.InspectRiskDeliveryPoliciesFromMap(policies)
	message := report.Summary()
	if len(report.Warnings) > 0 {
		return "warning", message + "；告警: " + strings.Join(report.Warnings, "；")
	}
	return "ok", message
}
