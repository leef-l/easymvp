package chat

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	v1 "easymvp/app/mvp/api/mvp/v1"
	collabRepo "easymvp/app/mvp/internal/collab/repo"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/regression"
	"easymvp/app/mvp/internal/workflow/eventstream"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	workspacepkg "easymvp/app/mvp/internal/workspace"
)

var inspectWorkflowEventMetadataColumnsFn = func(ctx context.Context) error {
	return repo.NewWorkflowEventRepo().InspectMetadataColumns(ctx)
}

var inspectWorkflowEventLedgerTableFn = func(ctx context.Context) error {
	return repo.NewWorkflowEventLedgerRepo().InspectDurableColumns(ctx)
}

type experienceReviewerPresetCheck struct {
	CategoryCode string
	ModelID      int64
}

type experienceReviewerModelCheck struct {
	Exists  bool
	Enabled bool
	Name    string
}

// SystemCheck 系统配置检测
func (c *cWorkflow) SystemCheck(ctx context.Context, req *v1.SystemCheckReq) (res *v1.SystemCheckRes, err error) {
	items := make([]v1.SystemCheckItem, 0, 13)
	allPass := true

	addItem := func(key, name, link, status, message string) {
		if status != "ok" {
			allPass = false
		}
		items = append(items, v1.SystemCheckItem{
			Key: key, Name: name, Status: status, Message: message, Link: link,
		})
	}

	count, e := repo.NewAIProviderRepo().CountEnabledWithBaseURL(ctx)
	if e != nil {
		addItem("ai_provider", "AI 供应商", "/ai/provider", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_provider", "AI 供应商", "/ai/provider", "error", "未配置启用的 AI 供应商（需要有 base_url）")
	} else {
		addItem("ai_provider", "AI 供应商", "/ai/provider", "ok", fmt.Sprintf("已有 %d 个启用供应商", count))
	}

	count, e = repo.NewAIPlanRepo().CountEnabledWithAPIKey(ctx)
	if e != nil {
		addItem("ai_plan", "AI 套餐", "/ai/plan", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_plan", "AI 套餐", "/ai/plan", "error", "未配置启用的 AI 套餐（需要有 api_key）")
	} else {
		addItem("ai_plan", "AI 套餐", "/ai/plan", "ok", fmt.Sprintf("已有 %d 个启用套餐", count))
	}

	count, e = repo.NewAIModelRepo().CountEnabledByCapability(ctx, "architect")
	if e != nil {
		addItem("ai_model_architect", "AI 模型（架构师）", "/ai/model", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_model_architect", "AI 模型（架构师）", "/ai/model", "error", "未配置 capability=architect 且启用的 AI 模型")
	} else {
		addItem("ai_model_architect", "AI 模型（架构师）", "/ai/model", "ok", fmt.Sprintf("已有 %d 个架构师模型", count))
	}

	count, e = repo.NewAIModelRepo().CountEnabledByCapabilities(ctx, []string{"implementer", "coding", "chat"})
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

	experienceReviewerPresets, experienceReviewerErr := repo.ListRolePresets(ctx, repo.RolePresetQuery{
		RoleType:    "experience_reviewer",
		RoleLevel:   "max",
		DefaultOnly: true,
	})
	if experienceReviewerErr != nil {
		addItem("experience_reviewer", "体验评审师预设", "/mvp/role-preset", "error", "查询失败: "+experienceReviewerErr.Error())
	} else {
		checks := make([]experienceReviewerPresetCheck, 0, len(experienceReviewerPresets))
		models := make(map[int64]experienceReviewerModelCheck)
		seenModels := make(map[int64]struct{})
		for _, row := range experienceReviewerPresets {
			modelID := row["model_id"].Int64()
			checks = append(checks, experienceReviewerPresetCheck{
				CategoryCode: row["project_category"].String(),
				ModelID:      modelID,
			})
			if modelID <= 0 {
				continue
			}
			if _, seen := seenModels[modelID]; seen {
				continue
			}
			seenModels[modelID] = struct{}{}

			model, modelErr := repo.NewAIModelRepo().GetByID(ctx, modelID, "name", "status")
			if modelErr != nil || len(model) == 0 {
				models[modelID] = experienceReviewerModelCheck{}
				continue
			}
			models[modelID] = experienceReviewerModelCheck{
				Exists:  true,
				Enabled: strings.TrimSpace(mapString(model, "status")) == "1",
				Name:    mapString(model, "name"),
			}
		}
		status, message := inspectExperienceReviewerReadiness(checks, models)
		addItem("experience_reviewer", "体验评审师预设", "/mvp/role-preset", status, message)
	}

	count, e = repo.NewAIEngineRepo().CountEnabled(ctx)
	if e != nil {
		addItem("ai_engine", "AI 执行引擎", "/ai/engine", "error", "查询失败: "+e.Error())
	} else if count == 0 {
		addItem("ai_engine", "AI 执行引擎", "/ai/engine", "error", "未配置启用的 AI 执行引擎")
	} else {
		addItem("ai_engine", "AI 执行引擎", "/ai/engine", "ok", fmt.Sprintf("已有 %d 个启用引擎", count))
	}

	aiderCfg, e := repo.NewAIEngineConfigRepo().GetByCode(ctx, "aider", "workspace_root")
	if e != nil || len(aiderCfg) == 0 {
		addItem("ai_engine_config_aider", "Aider 引擎配置", "/ai/engine", "error", "未配置 Aider 引擎参数")
	} else if strings.TrimSpace(mapString(aiderCfg, "workspace_root")) == "" {
		addItem("ai_engine_config_aider", "Aider 引擎配置", "/ai/engine", "warning", "Aider 引擎未配置 workspace_root（工作区根目录）")
	} else {
		addItem("ai_engine_config_aider", "Aider 引擎配置", "/ai/engine", "ok",
			"工作区根目录: "+mapString(aiderCfg, "workspace_root"))
	}

	ohCfg, e := repo.NewAIEngineConfigRepo().GetByCode(ctx, "openhands", "command_template")
	if e != nil || len(ohCfg) == 0 {
		addItem("ai_engine_config_openhands", "OpenHands 引擎配置", "/ai/engine", "warning", "未配置 OpenHands 引擎参数（非必须，仅使用 Aider 可忽略）")
	} else if strings.TrimSpace(mapString(ohCfg, "command_template")) == "" {
		addItem("ai_engine_config_openhands", "OpenHands 引擎配置", "/ai/engine", "warning", "OpenHands 未配置 command_template（命令模板）")
	} else {
		addItem("ai_engine_config_openhands", "OpenHands 引擎配置", "/ai/engine", "ok", "命令模板已配置")
	}

	count, e = repo.NewSystemRoleAIEngineRepo().CountAll(ctx)
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
	if len(ohCfg) > 0 {
		commandTemplate := strings.TrimSpace(mapString(ohCfg, "command_template"))
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
		"watchdog.heartbeat_timeout_seconds",
		"watchdog.max_stale_count",
		"watchdog.max_retries",
		"scheduler.poll_interval",
	}
	count, e = repo.NewConfigRepo().CountByKeys(ctx, requiredKeys)
	if e != nil {
		addItem("engine_config", "引擎核心配置", "/mvp/config", "error", "查询失败: "+e.Error())
	} else {
		schedulerCount, schedulerErr := repo.NewConfigRepo().CountByKeys(ctx, []string{"scheduler.max_concurrent", "workflow.scheduler.max_concurrency"})
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

	watchdogStatus := "warning"
	watchdogMessage := "watchdog 尚未初始化"
	if wd := orchestrator.GetDomainWatchdog(); wd != nil {
		snapshot := wd.Snapshot()
		watchdogStatus = "ok"
		if snapshot.CheckIntervalSeconds <= 0 || snapshot.HeartbeatTimeoutSeconds <= 0 {
			watchdogStatus = "warning"
		}
		if snapshot.HeartbeatTimeoutSeconds > 90 {
			watchdogStatus = "warning"
		}
		watchdogMessage = fmt.Sprintf(
			"check=%ds lease=%ds max_retries=%d last_running=%s(%d) last_failed=%s(%d) timeout=%d retry=%d escalate=%d",
			snapshot.CheckIntervalSeconds,
			snapshot.HeartbeatTimeoutSeconds,
			snapshot.MaxRetries,
			formatSystemCheckTime(snapshot.LastRunningCheckAt),
			snapshot.LastRunningTaskCount,
			formatSystemCheckTime(snapshot.LastFailedCheckAt),
			snapshot.LastFailedTaskCount,
			snapshot.LeaseTimeoutDetections,
			snapshot.AutoRetrySuccesses,
			snapshot.AutoEscalations,
		)
	}
	addItem("watchdog_runtime", "Watchdog 运行态", "/mvp/config", watchdogStatus, watchdogMessage)

	streamEnabled := engine.GetConfigInt(ctx, "workflow.event_stream.enabled", "workflow.event_stream.enabled", 0) == 1
	streamConsumerEnabled := engine.GetConfigInt(ctx, "workflow.event_stream.consumer_enabled", "workflow.event_stream.consumer_enabled", 0) == 1
	streamRedisRequired := engine.GetConfigInt(ctx, "workflow.event_stream.redis_required", "workflow.event_stream.redis_required", 0) == 1
	streamName := strings.TrimSpace(engine.GetConfigString(ctx, "workflow.event_stream.stream_name", "workflow.event_stream.stream_name", ""))
	streamStatus := "warning"
	streamMessage := "事件流未启用，当前仅依赖本地快路径 + watchdog"
	if publisher := orchestrator.GetEventPublisher(); publisher != nil {
		status := publisher.StreamStatus()
		if strings.TrimSpace(streamName) == "" {
			streamName = strings.TrimSpace(status.StreamName)
		}
		switch {
		case !streamEnabled:
		case status.Enabled && status.Degraded && streamRedisRequired:
			streamStatus = "error"
			streamMessage = fmt.Sprintf("stream=%s 已启用但 Redis 不可用: %s", safeSystemCheckValue(streamName), safeSystemCheckValue(status.LastError))
		case status.Enabled && status.Degraded:
			streamStatus = "warning"
			streamMessage = fmt.Sprintf("stream=%s 已降级到本地快路径: %s", safeSystemCheckValue(streamName), safeSystemCheckValue(status.LastError))
		case status.Enabled:
			streamStatus = "ok"
			streamMessage = fmt.Sprintf("stream=%s healthy consumer_enabled=%t consumer_started=%t updated=%s",
				safeSystemCheckValue(streamName),
				streamConsumerEnabled,
				orchestrator.GetWorkflowEventConsumer() != nil,
				formatSystemCheckTime(status.UpdatedAt),
			)
		default:
			streamMessage = fmt.Sprintf("stream=%s 配置已启用，但桥接器尚未就绪", safeSystemCheckValue(streamName))
		}
	}
	addItem("workflow_event_stream", "Workflow 事件流", "/mvp/config", streamStatus, streamMessage)

	consumerStatus := "warning"
	consumerMessage := "consumer 未创建，当前仅有 producer / local fast path"
	if consumer := orchestrator.GetWorkflowEventConsumer(); consumer != nil {
		snapshot := consumer.Snapshot(ctx)
		consumerStatus, consumerMessage = summarizeWorkflowEventConsumerSnapshot(snapshot)
	}
	addItem("workflow_event_consumer", "Workflow 事件消费", "/mvp/config", consumerStatus, consumerMessage)

	durableStatus, durableMessage := inspectWorkflowEventDurableSchema(ctx)
	addItem("workflow_event_durable", "Workflow 事件幂等账本", "/mvp/config", durableStatus, durableMessage)

	commandResourcePolicy := engine.GetCommandResourcePolicy(ctx)
	commandResourceStatus := "ok"
	if !commandResourcePolicy.Enabled {
		commandResourceStatus = "warning"
	}
	addItem("command_resource", "命令资源限制", "/mvp/config", commandResourceStatus, commandResourcePolicy.Summary())

	categoryRecords, categoryErr := repo.NewProjectCategoryRepo().ListAll(ctx, "verification_profile_json", "verification_gate_json")
	if categoryErr != nil {
		addItem("project_category_verification", "项目分类验证配置", "/mvp/project_category", "error", "查询失败: "+categoryErr.Error())
	} else {
		totalCategories := len(categoryRecords)
		profileConfigured := 0
		gateConfigured := 0
		for _, row := range categoryRecords {
			if strings.TrimSpace(mapString(row, "verification_profile_json")) != "" {
				profileConfigured++
			}
			if strings.TrimSpace(mapString(row, "verification_gate_json")) != "" {
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
	columnErr := repo.NewTaskWorkspaceRepo().InspectColumns(ctx, deliveryColumns)
	if columnErr != nil {
		addItem("workspace_migration", "Workspace 交付 Migration", "", "warning",
			"mvp_task_workspace 交付结果列尚未全部就绪，需执行最新 migration")
	} else {
		addItem("workspace_migration", "Workspace 交付 Migration", "", "ok",
			fmt.Sprintf("mvp_task_workspace 新列已就绪 %d/%d", len(deliveryColumns), len(deliveryColumns)))
	}

	deliveryRuleCount, deliveryRuleErr := repo.NewAcceptRuleRepo().CountByCodeProjectTypes(ctx, "software.delivery_review_required", []string{"software_dev", "coding"})
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

	orphanReport, orphanErr := workspacepkg.RunOrphanSweep(ctx, workspacepkg.NewGitWorktreeManager(), workspacepkg.OrphanSweepConfig{})
	if orphanErr != nil {
		addItem("workspace_orphan", "Workspace Orphan 对账", "", "warning", "对账失败: "+orphanErr.Error())
	} else {
		orphanStatus, orphanMessage := summarizeOrphanSweepReport(orphanReport)
		addItem("workspace_orphan", "Workspace Orphan 对账", "", orphanStatus, orphanMessage)
	}

	feishuEnabled := engine.GetConfigInt(ctx, "workflow.collab.feishu_enabled", "workflow.collab.feishuEnabled", 0)
	feishuAppID := strings.TrimSpace(engine.GetConfigString(ctx, "workflow.collab.feishu_app_id", "workflow.collab.feishuAppId", ""))
	feishuAppSecret := strings.TrimSpace(engine.GetConfigString(ctx, "workflow.collab.feishu_app_secret", "workflow.collab.feishuAppSecret", ""))
	feishuEncryptKey := strings.TrimSpace(engine.GetConfigString(ctx, "workflow.collab.feishu_encrypt_key", "workflow.collab.feishuEncryptKey", ""))
	feishuBindings, _ := collabRepo.NewBindingRepo().CountByPlatform(ctx, "feishu")
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

func summarizeOrphanSweepReport(report *workspacepkg.OrphanSweepReport) (string, string) {
	if report == nil {
		return "warning", "未获取到 orphan 对账结果"
	}
	status := "ok"
	if report.DiskOrphans > 0 || report.MissingOnDisk > 0 || report.RunningMismatch > 0 || report.Errors > 0 {
		status = "warning"
	}
	return status, fmt.Sprintf(
		"roots=%d db=%d disk=%d disk_orphan=%d db_orphan=%d running_mismatch=%d repaired_missing=%d repaired_running=%d cleaned_disk=%d errors=%d",
		report.ScannedRoots,
		report.DBWorkspaces,
		report.DiskWorktrees,
		report.DiskOrphans,
		report.MissingOnDisk,
		report.RunningMismatch,
		report.RepairedMissingOnDisk,
		report.RepairedRunningMismatch,
		report.CleanedDiskOrphans,
		report.Errors,
	)
}

func formatSystemCheckTime(ts interface {
	IsZero() bool
	Format(string) string
}) string {
	if ts.IsZero() {
		return "n/a"
	}
	return ts.Format("2006-01-02 15:04:05")
}

func safeSystemCheckValue(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "n/a"
	}
	return value
}

func summarizeWorkflowEventConsumerSnapshot(snapshot eventstream.RuntimeSnapshot) (string, string) {
	status := "warning"
	if !snapshot.ConsumerCreated {
		return status, "consumer 未创建，当前仅有 producer / local fast path"
	}

	pending := formatSystemCheckInt64(snapshot.PendingKnown, snapshot.Pending)
	lag := formatSystemCheckInt64(snapshot.LagKnown, snapshot.Lag)
	heartbeat := formatSystemCheckTime(snapshot.WorkerHeartbeatAt)

	if !snapshot.ConsumerStarted {
		return status, fmt.Sprintf(
			"consumer 已创建但未启动 stream=%s group=%s consumer=%s pending=%s lag=%s reclaim_attempts=%d reclaimed=%d last_consume=%s last_ack=%s heartbeat=%s",
			safeSystemCheckValue(snapshot.StreamName),
			safeSystemCheckValue(snapshot.ConsumerGroup),
			safeSystemCheckValue(snapshot.ConsumerName),
			pending,
			lag,
			snapshot.ReclaimAttempts,
			snapshot.ReclaimedMessages,
			formatSystemCheckTime(snapshot.LastConsumeAt),
			formatSystemCheckTime(snapshot.LastAckAt),
			heartbeat,
		)
	}

	status = "ok"
	if snapshot.Degraded {
		status = "warning"
	}
	if !snapshot.WorkerHeartbeatAt.IsZero() && time.Since(snapshot.WorkerHeartbeatAt) > 30*time.Second {
		status = "warning"
	}

	message := fmt.Sprintf(
		"stream=%s group=%s consumer=%s pending=%s lag=%s reclaim_attempts=%d reclaimed=%d last_consume=%s last_ack=%s heartbeat=%s started=%s",
		safeSystemCheckValue(snapshot.StreamName),
		safeSystemCheckValue(snapshot.ConsumerGroup),
		safeSystemCheckValue(snapshot.ConsumerName),
		pending,
		lag,
		snapshot.ReclaimAttempts,
		snapshot.ReclaimedMessages,
		formatSystemCheckTime(snapshot.LastConsumeAt),
		formatSystemCheckTime(snapshot.LastAckAt),
		heartbeat,
		formatSystemCheckTime(snapshot.StartedAt),
	)
	if snapshot.Degraded && strings.TrimSpace(snapshot.LastError) != "" {
		message += " degraded=" + safeSystemCheckValue(snapshot.LastError)
	}
	return status, message
}

func formatSystemCheckInt64(known bool, value int64) string {
	if !known {
		return "n/a"
	}
	return fmt.Sprintf("%d", value)
}

func inspectWorkflowEventDurableSchema(ctx context.Context) (string, string) {
	eventErr := inspectWorkflowEventMetadataColumnsFn(ctx)
	ledgerErr := inspectWorkflowEventLedgerTableFn(ctx)

	switch {
	case eventErr == nil && ledgerErr == nil:
		return "ok", "event metadata 列与 durable ledger 表已就绪"
	case eventErr != nil && ledgerErr != nil:
		return "warning", fmt.Sprintf(
			"durable idempotency migration 未完全就绪: workflow_event=%s; ledger=%s",
			safeSystemCheckValue(eventErr.Error()),
			safeSystemCheckValue(ledgerErr.Error()),
		)
	case eventErr != nil:
		return "warning", "workflow_event durable 元数据列未就绪: " + safeSystemCheckValue(eventErr.Error())
	default:
		return "warning", "workflow_event durable ledger 表未就绪: " + safeSystemCheckValue(ledgerErr.Error())
	}
}

func inspectExperienceReviewerReadiness(presets []experienceReviewerPresetCheck, models map[int64]experienceReviewerModelCheck) (string, string) {
	requiredCategories := []string{"game_dev", "software_dev"}
	grouped := make(map[string][]experienceReviewerPresetCheck)
	for _, preset := range presets {
		categoryCode := strings.TrimSpace(preset.CategoryCode)
		if categoryCode == "" {
			continue
		}
		grouped[categoryCode] = append(grouped[categoryCode], preset)
	}

	ready := make([]string, 0, len(requiredCategories))
	issues := make([]string, 0)
	for _, categoryCode := range requiredCategories {
		categoryPresets := grouped[categoryCode]
		if len(categoryPresets) == 0 {
			issues = append(issues, fmt.Sprintf("%s 缺少 experience_reviewer/max 默认预设", categoryCode))
			continue
		}

		var (
			categoryReady bool
			reasons       []string
		)
		for _, preset := range categoryPresets {
			switch {
			case preset.ModelID <= 0:
				reasons = append(reasons, "model_id=0")
			default:
				model, ok := models[preset.ModelID]
				switch {
				case !ok || !model.Exists:
					reasons = append(reasons, fmt.Sprintf("model_id=%d 不存在", preset.ModelID))
				case !model.Enabled:
					reasons = append(reasons, fmt.Sprintf("model_id=%d 已禁用", preset.ModelID))
				default:
					modelLabel := strings.TrimSpace(model.Name)
					if modelLabel == "" {
						modelLabel = fmt.Sprintf("model_id=%d", preset.ModelID)
					} else {
						modelLabel = fmt.Sprintf("%s(%d)", modelLabel, preset.ModelID)
					}
					ready = append(ready, fmt.Sprintf("%s=%s", categoryCode, modelLabel))
					categoryReady = true
				}
			}
			if categoryReady {
				break
			}
		}
		if !categoryReady {
			issues = append(issues, fmt.Sprintf("%s 预设不可用：%s", categoryCode, strings.Join(uniqueSortedSystemCheckStrings(reasons), "，")))
		}
	}

	if len(issues) > 0 {
		return "error", strings.Join(issues, "；")
	}
	sort.Strings(ready)
	return "ok", "已就绪：" + strings.Join(ready, "；")
}

func uniqueSortedSystemCheckStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
