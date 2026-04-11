package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/collab/adapter"
	collabRepo "easymvp/app/mvp/internal/collab/repo"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/provider"
)

// botIntent AI 解析出的用户意图。
type botIntent struct {
	Action             string `json:"action"`               // 意图动作，见下方 intentSystemPrompt
	ProjectName        string `json:"project_name"`         // 项目名称（多数操作需要）
	Category           string `json:"category"`             // 项目分类（create 时有值）
	TaskID             string `json:"task_id"`              // 任务ID（retry_task/skip_task/update_task）
	IssueID            string `json:"issue_id"`             // 审核/验收问题ID
	TargetStage        string `json:"target_stage"`         // 强制切换目标阶段
	TaskName           string `json:"task_name"`            // 更新后的任务名
	TaskDescription    string `json:"task_description"`     // 更新后的任务描述
	RoleType           string `json:"role_type"`            // 更新后的角色类型
	RoleLevel          string `json:"role_level"`           // 更新后的角色等级
	ExecutionMode      string `json:"execution_mode"`       // 更新后的执行方式
	RestartAfterUpdate bool   `json:"restart_after_update"` // 修改后是否重置为 pending
	Reason             string `json:"reason"`               // 人工接管原因
	Reply              string `json:"reply"`                // chat 时 AI 的直接回复文本
}

// intentSystemPrompt AI 意图解析系统提示词。
const intentSystemPrompt = `你是 EasyMVP 的飞书机器人助手。EasyMVP 是 AI 驱动的项目管理平台，支持多角色 AI 团队（架构师/实现者/审计员）自动完成软件项目的需求分析、任务拆解、代码实现、质量审计全流程。

## 支持的 action 完整列表

### 项目管理
- create_project：创建新项目（project_name必填，category可选）
- list_projects：列出我的项目
- project_status：查询项目状态和进度（project_name必填）
- pause_project：暂停项目执行（project_name必填）
- resume_project：继续/恢复项目（project_name必填）
- cancel_project：取消当前工作流（project_name必填）
- force_stage：强制切换到指定阶段（project_name必填，target_stage必填，可选值：design/review/execute/accept/rework；rework 时 task_id 建议填写失败任务ID）
- confirm_plan：确认当前对话中的方案并启动自动执行

### 任务管理
- list_tasks：查看项目任务列表（project_name必填）
- retry_task：重试失败任务（project_name必填，task_id可选，不填则重试所有失败任务）
- skip_task：跳过阻塞任务（project_name必填，task_id必填）
- update_task：人工修改任务（project_name必填，task_id必填；可选 task_name/task_description/role_type/role_level/execution_mode/restart_after_update）

### 审核管理
- review_status：查看项目当前审核状态和问题（project_name必填）
- approve_review：通过人工审核（project_name必填）
- reject_review：驳回人工审核（project_name必填）

### 验收管理
- accept_status：查看项目验收状态（project_name必填）
- approve_accept：验收通过（project_name必填）
- reject_accept：验收驳回/打回返工（project_name必填）

### 验证修复
- verification_start：启动 Docker-first 项目验证（project_name必填）
- verification_status：查看最近一次验证状态（project_name必填）
- verification_repair：基于验证问题触发返工（project_name必填，issue_id可选）

### 自治管理
- autonomy_status：查看自治模式状态和待审检查点（project_name必填）
- approve_checkpoint：批准自治检查点（project_name必填）
- reject_checkpoint：拒绝自治检查点（project_name必填）

### 通用
- chat：普通对话/不确定意图（reply字段填写中文回复）
- help：显示帮助

## 意图识别规则
- "创建/新建/做一个XXX项目" → create_project
- "我的项目/项目列表/有哪些项目" → list_projects
- "XXX进度/状态/怎么样了" → project_status
- "暂停/停止XXX" → pause_project
- "继续/恢复/重启XXX" → resume_project
- "取消/终止XXX项目" → cancel_project
- "回到设计/重开审核/重跑执行/强制验收/进入返工" → force_stage，并填写 target_stage
- "确认方案/开始执行/启动" → confirm_plan
- "任务列表/查看任务/XXX的任务" → list_tasks
- "重试/重新执行XXX失败任务" → retry_task
- "跳过任务/跳过阻塞" → skip_task
- "修改任务/调整任务/重置任务" → update_task
- "审核状态/审核结果/审核通过了吗" → review_status
- "通过审核/审核通过" → approve_review
- "驳回审核/审核不通过" → reject_review
- "验收状态/验收结果/通过验收了吗" → accept_status
- "验收通过/通过了" → approve_accept
- "验收不通过/打回/返工" → reject_accept
- "启动验证/开始验证/跑验证" → verification_start
- "验证状态/验证结果/最近验证" → verification_status
- "修复验证问题/根据验证返工/验证问题返工" → verification_repair
- "自治状态/自治模式/检查点" → autonomy_status
- "批准检查点/同意自治" → approve_checkpoint
- "拒绝检查点/不同意自治" → reject_checkpoint
- 其他：chat（reply填写友好回复）

category 常见值：软件开发、游戏开发、数据分析、内容创作、运营策划。未指定默认"软件开发"。
reason 可选，用户提到“因为.../原因是...”时尽量提取。
update_task 时，若用户提到“重跑/重置后重做/重新开始”，将 restart_after_update 设为 true。

只返回 JSON，格式：{"action":"...","project_name":"...","category":"...","task_id":"...","issue_id":"...","target_stage":"...","task_name":"...","task_description":"...","role_type":"...","role_level":"...","execution_mode":"...","restart_after_update":false,"reason":"...","reply":"..."}`

// feishuBotPlatform 实现 BotPlatform 接口，封装飞书消息回复。
type feishuBotPlatform struct {
	messageID string
	chatID    string
	openID    string // 兜底：用 open_id 发私信
}

func (f *feishuBotPlatform) Reply(ctx context.Context, text string) {
	feishu := adapter.NewFeishuAdapter()
	// 优先回复原消息
	if f.messageID != "" {
		g.Log().Debugf(ctx, "[FeishuBot] 回复消息: messageID=%s text_len=%d", f.messageID, len(text))
		if err := feishu.ReplyMessage(ctx, f.messageID, text); err != nil {
			g.Log().Warningf(ctx, "[FeishuBot] 回复消息失败: %v", err)
		}
		return
	}
	// 其次发到对话（群聊或单聊 chat_id）
	if f.chatID != "" {
		g.Log().Debugf(ctx, "[FeishuBot] 发送消息: chatID=%s text_len=%d", f.chatID, len(text))
		if err := feishu.SendTextToChat(ctx, f.chatID, text); err != nil {
			g.Log().Warningf(ctx, "[FeishuBot] 发送消息失败(chatID): %v", err)
		}
		return
	}
	// 兜底：用 open_id 发私信
	if f.openID != "" {
		g.Log().Debugf(ctx, "[FeishuBot] 发送私信: openID=%s text_len=%d", f.openID, len(text))
		if err := feishu.SendTextMessage(ctx, f.openID, text); err != nil {
			g.Log().Warningf(ctx, "[FeishuBot] 发送私信失败(openID): %v", err)
		}
		return
	}
	g.Log().Warningf(ctx, "[FeishuBot] 无法回复：messageID/chatID/openID 均为空")
}

func (f *feishuBotPlatform) PlatformName() string { return "feishu" }

// DispatchFeishuCommand 飞书消息入口，解析消息后转发到统一 Bot 调度器。
func DispatchFeishuCommand(ctx context.Context, openID, messageID, chatID, contentStr string) {
	text := extractFeishuText(contentStr)
	text = removeAtMention(text)
	text = strings.TrimSpace(text)

	DispatchBotCommand(ctx, &BotContext{
		OpenID:  openID,
		Content: text,
		Platform: &feishuBotPlatform{
			messageID: messageID,
			chatID:    chatID,
			openID:    openID,
		},
	})
}

// parseIntentWithAI 调用 AI 解析用户消息意图，返回结构化 botIntent。
func parseIntentWithAI(ctx context.Context, userText string, systemUserID int64) (*botIntent, error) {
	// 加载模型（优先用绑定用户的架构师角色模型，否则取全局第一个可用模型）
	modelInfo, err := loadBotModel(ctx, systemUserID)
	if err != nil {
		return nil, fmt.Errorf("加载 Bot 模型失败: %w", err)
	}

	// 构造用户上下文提示（让 AI 了解用户信息，提升理解准确性）
	userCtx := ""
	if systemUserID > 0 {
		userCtx = fmt.Sprintf("（用户已绑定系统账号 ID:%d）", systemUserID)
	} else {
		userCtx = "（用户尚未绑定系统账号）"
	}

	p, err := provider.GetProvider(provider.Config{
		ProviderType:       modelInfo.ProviderType,
		SupportedProtocols: modelInfo.SupportedProtocols,
		BaseURL:            modelInfo.BaseURL,
		APIKey:             modelInfo.APIKey,
		APISecret:          modelInfo.APISecret,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化 provider 失败: %w", err)
	}

	req := &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		SystemPrompt: intentSystemPrompt,
		Messages: []provider.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("用户%s说：%s", userCtx, userText),
			},
		},
		MaxTokens:   512,
		Temperature: 0.1, // 低温度，确保输出稳定的 JSON
	}

	resp, err := p.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AI 调用失败: %w", err)
	}

	// 提取 JSON（AI 可能在 JSON 前后有多余文字）
	raw := extractJSON(resp.Content)
	var intent botIntent
	if err := json.Unmarshal([]byte(raw), &intent); err != nil {
		return nil, fmt.Errorf("解析 AI 响应失败: %w (raw: %s)", err, resp.Content)
	}

	return &intent, nil
}

// loadBotModel 加载 Bot 使用的 AI 模型。
// 优先：用户名下项目的架构师模型 → 全局 sort 最小的可用模型
func loadBotModel(ctx context.Context, systemUserID int64) (*engine.ModelInfo, error) {
	// 先尝试找用户最近项目的架构师模型
	if systemUserID > 0 {
		project, _ := repo.NewProjectRepo().GetLatestByCreator(ctx, systemUserID, "id")
		if len(project) > 0 {
			role, roleErr := engine.ResolveProjectRole(ctx, g.NewVar(project["id"]).Int64(), "architect")
			modelID := int64(0)
			if roleErr == nil && role != nil {
				modelID = role["model_id"].Int64()
			}
			info, err := engine.GetModelInfoByID(ctx, modelID)
			if err == nil && info != nil {
				return info, nil
			}
		}
	}

	// 降级：取全局第一个可用模型（ai_model join ai_plan join ai_provider，按 sort 排序）
	record, err := repo.NewAIModelRepo().GetFirstEnabledWithCredentials(ctx, "m.id")
	if err != nil || len(record) == 0 {
		return nil, fmt.Errorf("系统未配置任何可用的 AI 模型")
	}

	return engine.GetModelInfoByID(ctx, g.NewVar(record["id"]).Int64())
}

// ─── 指令处理器 ───────────────────────────────────────────────────────────────

func handleBotCreateProject(ctx context.Context, projectName, category string, systemUserID, deptID int64, openID, platform string, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定账号，请先在 EasyMVP 管理端完成账号绑定。")
		return
	}
	if projectName == "" {
		reply("❌ 请提供项目名称，例如：帮我创建一个电商后台项目")
		return
	}
	if category == "" {
		category = "软件开发"
	}

	projectID, _, err := engine.CreateProject(ctx, projectName, category, "", "", 0, systemUserID, deptID, nil)
	if err != nil {
		reply(fmt.Sprintf("❌ 创建项目失败：%v", err))
		return
	}

	// 查询项目对应的架构师对话，自动进入对话模式
	extraTip := ""
	if openID != "" {
		var convID int64
		conv, _ := repo.NewConversationRepo().GetFirstByProject(ctx, projectID, "id")
		if len(conv) > 0 {
			convID = g.NewVar(conv["id"]).Int64()
		}
		if convID > 0 {
			setBotSession(platform, openID, convID)
			extraTip = "\n\n💬 已进入对话模式，可以直接和架构师描述需求。\n说「确认方案」可以启动执行，说「退出对话」可以结束对话。"
		}
	}

	reply(fmt.Sprintf(
		"✅ 项目已创建\n"+
			"───────────────\n"+
			"📁 项目名称：%s\n"+
			"🏷️ 分类：%s\n"+
			"🆔 项目ID：%d\n"+
			"───────────────\n"+
			"下一步：直接描述你的需求，架构师AI会为你拆解任务。%s",
		projectName, category, projectID, extraTip,
	))
}

func handleBotListProjects(ctx context.Context, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}

	scope, err := resolveBotProjectScope(ctx, systemUserID)
	if err != nil {
		reply(fmt.Sprintf("❌ 查询权限范围失败：%v", err))
		return
	}

	type projectRow struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	var rows []projectRow
	projectRecords, err := repo.NewProjectRepo().ListRecentByScope(ctx, repo.ProjectScopeFilter{
		All:         scope.All,
		IncludeSelf: scope.IncludeSelf,
		UserID:      systemUserID,
		DeptIDs:     scope.DeptIDs,
	}, 8, "id", "name", "status")
	if err != nil {
		reply(fmt.Sprintf("❌ 查询失败：%v", err))
		return
	}
	for _, record := range projectRecords {
		rows = append(rows, projectRow{
			ID:     g.NewVar(record["id"]).Int64(),
			Name:   mapString(record, "name"),
			Status: mapString(record, "status"),
		})
	}

	if len(rows) == 0 {
		reply("📭 暂无项目，可以说「帮我创建一个XXX项目」来快速创建。")
		return
	}

	statusLabel := map[string]string{
		"designing": "⚙️ 设计中",
		"reviewing": "🔍 审核中",
		"executing": "🚀 执行中",
		"accepting": "🎯 验收中",
		"reworking": "🔁 返工中",
		"running":   "🚀 执行中",
		"paused":    "⏸️ 已暂停",
		"completed": "✅ 已完成",
		"canceled":  "🛑 已取消",
	}
	lines := []string{"📋 您的项目（最近8个）", "───────────────"}
	for i, row := range rows {
		label := statusLabel[row.Status]
		if label == "" {
			label = row.Status
		}
		lines = append(lines, fmt.Sprintf("%d. %s  %s  ID:%d", i+1, row.Name, label, row.ID))
	}
	reply(strings.Join(lines, "\n"))
}

func handleBotProjectStatus(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}
	if projectName == "" {
		reply("❌ 请告诉我是哪个项目，例如：电商后台的进度怎么样")
		return
	}

	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」，可以发送「项目列表」查看你的所有项目", projectName))
		return
	}

	projectID := project["id"].Int64()
	name := project["name"].String()
	status := project["status"].String()

	total, done, running, failed := 0, 0, 0, 0
	if isWorkflowV2Project(project) {
		wfRun, wfErr := latestWorkflowRunForProject(ctx, projectID)
		if wfErr == nil && !wfRun.IsEmpty() {
			status = wfRun["status"].String()
			counts, _ := repo.NewDomainTaskRepo().ListStatusRowsByWorkflow(ctx, wfRun["id"].Int64())
			for _, c := range counts {
				count := mapInt(c, "count")
				total += count
				switch mapString(c, "status") {
				case "completed":
					done += count
				case "running":
					running += count
				case "failed", "escalated":
					failed += count
				}
			}
		}
	} else {
		counts, _ := repo.NewTaskRepo().CountStatusRowsByProject(ctx, projectID)
		for _, c := range counts {
			count := mapInt(c, "count")
			total += count
			switch mapString(c, "status") {
			case "completed":
				done += count
			case "running":
				running += count
			case "failed":
				failed += count
			}
		}
	}

	statusLabel := map[string]string{
		"designing": "⚙️ 设计中",
		"reviewing": "🔍 审核中",
		"executing": "🚀 执行中",
		"accepting": "🎯 验收中",
		"reworking": "🔁 返工中",
		"running":   "🚀 执行中",
		"paused":    "⏸️ 已暂停",
		"completed": "✅ 已完成",
		"canceled":  "🛑 已取消",
	}
	label := statusLabel[status]
	if label == "" {
		label = status
	}

	lines := []string{
		fmt.Sprintf("📊 %s", name),
		"───────────────",
		fmt.Sprintf("状态：%s", label),
		fmt.Sprintf("任务：共%d个 | 完成%d | 运行%d | 失败%d", total, done, running, failed),
	}
	if total > 0 {
		lines = append(lines, fmt.Sprintf("进度：%d%%", done*100/total))
	}
	pr := project["pause_reason"].String()
	if pr != "" {
		lines = append(lines, fmt.Sprintf("暂停原因：%s", pr))
	}
	lines = append(lines, fmt.Sprintf("ID：%d", projectID))
	reply(strings.Join(lines, "\n"))
}

func handleBotPauseProject(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}
	if projectName == "" {
		reply("❌ 请告诉我要暂停哪个项目")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	if isWorkflowV2Project(project) {
		wfRun, qErr := repo.NewWorkflowRunRepo().GetLatestByProjectExcludingStatuses(ctx, project["id"].Int64(), []string{"completed", "canceled", "paused"}, "id")
		if qErr != nil {
			reply(fmt.Sprintf("❌ 查询工作流失败：%v", qErr))
			return
		}
		if len(wfRun) == 0 {
			reply(fmt.Sprintf("❌ 项目「%s」当前没有可暂停的工作流", project["name"].String()))
			return
		}
		if err := orchestrator.GetWorkflowService().Pause(ctx, g.NewVar(wfRun["id"]).Int64(), "飞书机器人指令暂停"); err != nil {
			reply(fmt.Sprintf("❌ 暂停失败：%v", err))
			return
		}
	} else {
		if err := engine.GetScheduler().Pause(ctx, project["id"].Int64(), "飞书机器人指令暂停"); err != nil {
			reply(fmt.Sprintf("❌ 暂停失败：%v", err))
			return
		}
	}
	reply(fmt.Sprintf("⏸️ 项目「%s」已暂停", project["name"].String()))
}

func handleBotResumeProject(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}
	if projectName == "" {
		reply("❌ 请告诉我要继续哪个项目")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	if isWorkflowV2Project(project) {
		wfRun, qErr := repo.NewWorkflowRunRepo().GetLatestByProjectStatuses(ctx, project["id"].Int64(), []string{"paused"}, "id")
		if qErr != nil {
			reply(fmt.Sprintf("❌ 查询暂停工作流失败：%v", qErr))
			return
		}
		if len(wfRun) == 0 {
			reply(fmt.Sprintf("❌ 项目「%s」当前没有暂停的工作流", project["name"].String()))
			return
		}
		if err := orchestrator.GetWorkflowService().Resume(ctx, g.NewVar(wfRun["id"]).Int64()); err != nil {
			reply(fmt.Sprintf("❌ 恢复失败：%v", err))
			return
		}
	} else {
		if err := engine.GetScheduler().Resume(ctx, project["id"].Int64()); err != nil {
			reply(fmt.Sprintf("❌ 恢复失败：%v", err))
			return
		}
	}
	reply(fmt.Sprintf("▶️ 项目「%s」已继续执行", project["name"].String()))
}

func handleBotCancelProject(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}
	if projectName == "" {
		reply("❌ 请告诉我要取消哪个项目")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}

	wfRun, err := repo.NewWorkflowRunRepo().GetLatestByProjectExcludingStatuses(ctx, project["id"].Int64(), []string{"completed", "canceled"}, "id")
	if err != nil {
		reply(fmt.Sprintf("❌ 查询工作流失败：%v", err))
		return
	}
	if len(wfRun) == 0 {
		reply(fmt.Sprintf("❌ 项目「%s」当前没有可取消的工作流", project["name"].String()))
		return
	}

	reason := "机器人人工取消"
	workflowRunID := g.NewVar(wfRun["id"]).Int64()
	if err := orchestrator.GetWorkflowService().Cancel(ctx, workflowRunID, reason); err != nil {
		reply(fmt.Sprintf("❌ 取消工作流失败：%v", err))
		return
	}

	recordWorkflowEvent(ctx, workflowRunID, "workflow", "workflow.canceled", &workflowRunID, nil, map[string]interface{}{
		"project_id": project["id"].Int64(),
		"reason":     reason,
		"source":     "bot",
	})
	reply(fmt.Sprintf("🛑 项目「%s」当前工作流已取消", project["name"].String()))
}

func handleBotForceStage(ctx context.Context, projectName, targetStage, taskIDStr string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}
	if projectName == "" {
		reply("❌ 请告诉我要操作哪个项目")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}

	projectID := project["id"].Int64()
	wfRun, err := latestWorkflowRunForProject(ctx, projectID)
	if err != nil {
		reply(fmt.Sprintf("❌ 查询工作流失败：%v", err))
		return
	}
	workflowRunID := wfRun["id"].Int64()
	reason := "机器人人工切换阶段"

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
			reply(fmt.Sprintf("❌ 回到设计阶段失败：%v", err))
			return
		}
	case "review", "execute":
		if err := resetWorkflowExecutionArtifacts(ctx, projectID, workflowRunID); err != nil {
			reply(fmt.Sprintf("❌ 清理旧执行数据失败：%v", err))
			return
		}
	case "accept", "rework":
		if scheduler := orchestrator.GetTaskScheduler(); scheduler != nil {
			scheduler.Pause(ctx, workflowRunID)
		}
		orchestrator.GetRuntimeManager().Cancel(workflowRunID)
	default:
		reply("❌ 不支持的目标阶段")
		return
	}

	stageSvc := orchestrator.GetStageService()
	stageRunID, err := stageSvc.ForceStartStage(ctx, workflowRunID, targetStage, reason)
	if err != nil {
		reply(fmt.Sprintf("❌ 强制切换阶段失败：%v", err))
		return
	}

	switch targetStage {
	case "review":
		planVersionID, err := preparePlanVersionForForceStage(ctx, projectID, workflowRunID, 0, targetStage)
		if err != nil {
			_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
			reply(fmt.Sprintf("❌ 准备审核阶段失败：%v", err))
			return
		}
		go func() {
			bgCtx := context.Background()
			if runErr := orchestrator.GetReviewStageService().RunReview(bgCtx, stageRunID, planVersionID); runErr != nil {
				g.Log().Errorf(bgCtx, "[BotForceStage] review 重启失败: workflowRunID=%d stageRunID=%d err=%v", workflowRunID, stageRunID, runErr)
				_ = stageSvc.FailStage(bgCtx, stageRunID, runErr.Error())
			}
		}()
	case "execute":
		planVersionID, err := preparePlanVersionForForceStage(ctx, projectID, workflowRunID, 0, targetStage)
		if err != nil {
			_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
			reply(fmt.Sprintf("❌ 准备执行阶段失败：%v", err))
			return
		}
		if err := orchestrator.GetExecuteStageService().InstantiateAndStart(ctx, stageRunID, planVersionID); err != nil {
			_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
			reply(fmt.Sprintf("❌ 启动执行阶段失败：%v", err))
			return
		}
	case "design":
		if err := repo.NewWorkflowRunRepo().UpdateFields(ctx, workflowRunID, g.Map{"active_plan_version_id": nil, "updated_at": gtime.Now()}); err != nil {
			g.Log().Warningf(ctx, "[BotForceStage] 清空 active_plan_version_id 失败: workflowRunID=%d err=%v", workflowRunID, err)
		}
	case "accept":
		go func() {
			bgCtx := context.Background()
			if runErr := orchestrator.GetAcceptStageService().Run(bgCtx, workflowRunID, stageRunID); runErr != nil {
				g.Log().Errorf(bgCtx, "[BotForceStage] accept 重启失败: workflowRunID=%d stageRunID=%d err=%v", workflowRunID, stageRunID, runErr)
				_ = stageSvc.FailStage(bgCtx, stageRunID, runErr.Error())
			}
		}()
	case "rework":
		var failedTaskID int64
		if taskIDStr != "" {
			fmt.Sscanf(taskIDStr, "%d", &failedTaskID)
		}
		if failedTaskID == 0 {
			_ = stageSvc.FailStage(context.Background(), stageRunID, "机器人返工未提供失败任务ID")
			reply("❌ 强制返工需要提供失败任务ID，例如：让项目X进入返工，任务123")
			return
		}
		if err := orchestrator.GetReworkStageService().HandleReworkWithSource(ctx, stageRunID, failedTaskID, "execute"); err != nil {
			_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
			reply(fmt.Sprintf("❌ 启动返工阶段失败：%v", err))
			return
		}
		if err := orchestrator.ActivateReworkStage(ctx, workflowRunID, stageRunID); err != nil {
			_ = stageSvc.FailStage(context.Background(), stageRunID, err.Error())
			reply(fmt.Sprintf("❌ 启动返工调度器失败：%v", err))
			return
		}
	}

	recordWorkflowEvent(ctx, workflowRunID, "workflow", "workflow.force_stage", &workflowRunID, &stageRunID, map[string]interface{}{
		"project_id":     projectID,
		"target_stage":   targetStage,
		"failed_task_id": taskIDStr,
		"reason":         reason,
		"source":         "bot",
	})
	reply(fmt.Sprintf("🧭 项目「%s」已强制切换到%s阶段", project["name"].String(), botStageLabel(targetStage)))
}

func handleBotUpdateTask(ctx context.Context, intent *botIntent, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}
	if intent == nil {
		reply("❌ 指令解析失败")
		return
	}
	if strings.TrimSpace(intent.ProjectName) == "" {
		reply("❌ 请告诉我要操作哪个项目")
		return
	}
	if strings.TrimSpace(intent.TaskID) == "" {
		reply("❌ 修改任务必须提供任务ID")
		return
	}

	project, err := findProjectByKeyword(ctx, intent.ProjectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", intent.ProjectName))
		return
	}

	var taskID int64
	if _, scanErr := fmt.Sscanf(strings.TrimSpace(intent.TaskID), "%d", &taskID); scanErr != nil || taskID == 0 {
		reply("❌ 任务ID格式不正确")
		return
	}

	res, err := updateDomainTaskInternal(ctx, project["id"].Int64(), domainTaskUpdateOptions{
		TaskID:             taskID,
		Name:               intent.TaskName,
		Description:        intent.TaskDescription,
		RoleType:           intent.RoleType,
		RoleLevel:          intent.RoleLevel,
		ExecutionMode:      intent.ExecutionMode,
		RestartAfterUpdate: intent.RestartAfterUpdate,
		Reason:             intent.Reason,
	})
	if err != nil {
		reply(fmt.Sprintf("❌ 修改任务失败：%v", err))
		return
	}

	reply(fmt.Sprintf("✏️ 项目「%s」任务 %d 已更新，当前状态：%s。%s",
		project["name"].String(),
		taskID,
		res.Status,
		res.Message,
	))
}

func botStageLabel(stage string) string {
	switch stage {
	case "design":
		return "设计"
	case "review":
		return "审核"
	case "execute":
		return "执行"
	case "accept":
		return "验收"
	case "rework":
		return "返工"
	default:
		return stage
	}
}

func isWorkflowV2Project(project gdb.Record) bool {
	if project == nil || project.IsEmpty() {
		return false
	}
	return project["engine_version"].String() != "legacy"
}

// ─── 工具函数 ────────────────────────────────────────────────────────────────

// fallbackParseIntent 关键词降级解析（AI 不可用时使用）。
func fallbackParseIntent(text string) *botIntent {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "创建") || strings.Contains(lower, "新建") || strings.Contains(lower, "create"):
		name := extractProjectName(text)
		category := extractCategory(text)
		return &botIntent{Action: "create_project", ProjectName: name, Category: category}
	case strings.Contains(lower, "列表") || strings.Contains(lower, "list") || lower == "ls":
		return &botIntent{Action: "list_projects"}
	case strings.Contains(lower, "状态") || strings.Contains(lower, "进度") || strings.Contains(lower, "status"):
		return &botIntent{Action: "project_status", ProjectName: extractKeyword(text, "状态", "进度", "status")}
	case strings.Contains(lower, "暂停") || strings.Contains(lower, "pause"):
		return &botIntent{Action: "pause_project", ProjectName: extractKeyword(text, "暂停", "pause")}
	case strings.Contains(lower, "继续") || strings.Contains(lower, "恢复") || strings.Contains(lower, "resume"):
		return &botIntent{Action: "resume_project", ProjectName: extractKeyword(text, "继续", "恢复", "resume")}
	case strings.Contains(lower, "取消") || strings.Contains(lower, "终止") || strings.Contains(lower, "cancel"):
		return &botIntent{Action: "cancel_project", ProjectName: extractKeyword(text, "取消", "终止", "cancel")}
	case strings.Contains(lower, "回到设计") || strings.Contains(lower, "重新设计"):
		return &botIntent{Action: "force_stage", ProjectName: text, TargetStage: "design"}
	case strings.Contains(lower, "重开审核") || strings.Contains(lower, "重新审核") || strings.Contains(lower, "强制审核"):
		return &botIntent{Action: "force_stage", ProjectName: text, TargetStage: "review"}
	case strings.Contains(lower, "重开执行") || strings.Contains(lower, "重新执行阶段") || strings.Contains(lower, "强制执行"):
		return &botIntent{Action: "force_stage", ProjectName: text, TargetStage: "execute"}
	case strings.Contains(lower, "强制验收") || strings.Contains(lower, "重开验收"):
		return &botIntent{Action: "force_stage", ProjectName: text, TargetStage: "accept"}
	case strings.Contains(lower, "强制返工") || strings.Contains(lower, "进入返工"):
		return &botIntent{Action: "force_stage", ProjectName: text, TargetStage: "rework", TaskID: extractTaskID(text)}
	case strings.Contains(lower, "启动验证") || strings.Contains(lower, "开始验证") || strings.Contains(lower, "跑验证"):
		return &botIntent{Action: "verification_start", ProjectName: extractKeyword(text, "启动验证", "开始验证", "跑验证")}
	case strings.Contains(lower, "验证状态") || strings.Contains(lower, "验证结果") || strings.Contains(lower, "最近验证"):
		return &botIntent{Action: "verification_status", ProjectName: extractKeyword(text, "验证状态", "验证结果", "最近验证")}
	case strings.Contains(lower, "修复验证问题") || strings.Contains(lower, "根据验证返工") || strings.Contains(lower, "验证问题返工"):
		return &botIntent{Action: "verification_repair", ProjectName: text, IssueID: extractTaskID(text)}
	case strings.Contains(lower, "修改任务") || strings.Contains(lower, "调整任务") || strings.Contains(lower, "更新任务"):
		return &botIntent{Action: "update_task", ProjectName: text, TaskID: extractTaskID(text)}
	case strings.Contains(lower, "帮助") || lower == "help" || lower == "?":
		return &botIntent{Action: "help"}
	default:
		return &botIntent{Action: "chat", Reply: "您好，我是 EasyMVP 机器人，可以帮您管理项目。\n\n" + feishuHelpText()}
	}
}

func extractProjectName(text string) string {
	for _, kw := range []string{"创建项目", "新建项目", "创建一个", "新建一个", "create project"} {
		if idx := strings.Index(strings.ToLower(text), strings.ToLower(kw)); idx != -1 {
			after := strings.TrimSpace(text[idx+len(kw):])
			if i := strings.Index(after, "分类"); i != -1 {
				after = strings.TrimSpace(after[:i])
			}
			return after
		}
	}
	return text
}

func extractCategory(text string) string {
	for _, sep := range []string{"分类:", "分类："} {
		if idx := strings.Index(text, sep); idx != -1 {
			return strings.TrimSpace(text[idx+len(sep):])
		}
	}
	return "软件开发"
}

func extractKeyword(text string, prefixes ...string) string {
	lower := strings.ToLower(text)
	for _, p := range prefixes {
		if idx := strings.Index(lower, strings.ToLower(p)); idx != -1 {
			return strings.TrimSpace(text[idx+len(p):])
		}
	}
	return text
}

func extractTaskID(text string) string {
	var taskID int64
	if _, err := fmt.Sscanf(text, "%d", &taskID); err == nil && taskID > 0 {
		return fmt.Sprintf("%d", taskID)
	}
	parts := strings.FieldsFunc(text, func(r rune) bool { return r < '0' || r > '9' })
	for _, part := range parts {
		if part != "" {
			return part
		}
	}
	return ""
}

// lookupSystemUser 根据飞书 openID 查找绑定的系统用户 ID 和部门 ID。
func lookupSystemUser(ctx context.Context, openID string) (userID int64, deptID int64) {
	if openID == "" {
		return 0, 0
	}
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	binding, err := bindingRepo.GetByPlatformUserID(ctx, "feishu", openID)
	if err != nil || binding == nil {
		return 0, 0
	}
	userID = toCallbackInt64(binding["user_id"])
	deptID = toCallbackInt64(binding["dept_id"])
	return
}

// lookupBotUser 按平台+openID 查找绑定的系统用户，平台无关版本。
func lookupBotUser(ctx context.Context, platform, openID string) (userID int64, deptID int64) {
	if openID == "" {
		return 0, 0
	}
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	binding, err := bindingRepo.GetByPlatformUserID(ctx, platform, openID)
	if err != nil || binding == nil {
		return 0, 0
	}
	userID = toCallbackInt64(binding["user_id"])
	deptID = toCallbackInt64(binding["dept_id"])
	return
}

// getBotSession 按平台获取会话，平台无关版本。
func getBotSession(platform, openID string) (int64, bool) {
	if platform == "telegram" {
		return getTGSession(openID)
	}
	return getFeishuSession(openID)
}

// setBotSession 按平台设置会话。
func setBotSession(platform, openID string, convID int64) {
	if platform == "telegram" {
		setTGSession(openID, convID)
	} else {
		setFeishuSession(openID, convID)
	}
}

// clearBotSession 按平台清除会话。
func clearBotSession(platform, openID string) {
	if platform == "telegram" {
		clearTGSession(openID)
	} else {
		clearFeishuSession(openID)
	}
}

type botProjectScope struct {
	All         bool
	IncludeSelf bool
	DeptIDs     []int64
}

func resolveBotProjectScope(ctx context.Context, userID int64) (*botProjectScope, error) {
	scope := &botProjectScope{DeptIDs: make([]int64, 0)}
	if userID == 0 {
		return scope, nil
	}

	scopeRepo := repo.NewSystemScopeRepo()
	currentDeptID, err := scopeRepo.GetUserDeptID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles, err := scopeRepo.ListRoleScopesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		scope.IncludeSelf = true
		return scope, nil
	}

	deptSet := make(map[int64]struct{})
	customRoleIDs := make([]int64, 0)
	for _, role := range roles {
		if g.NewVar(role["is_admin"]).Int() == 1 || g.NewVar(role["data_scope"]).Int() == 1 {
			scope.All = true
			return scope, nil
		}
		switch g.NewVar(role["data_scope"]).Int() {
		case 2:
			for _, deptID := range loadBotDeptSubtreeIDs(ctx, currentDeptID) {
				deptSet[deptID] = struct{}{}
			}
		case 3:
			if currentDeptID > 0 {
				deptSet[currentDeptID] = struct{}{}
			}
		case 4:
			scope.IncludeSelf = true
		case 5:
			customRoleIDs = append(customRoleIDs, g.NewVar(role["id"]).Int64())
		}
	}

	if len(customRoleIDs) > 0 {
		customDepts, deptErr := scopeRepo.ListDeptIDsByRoleIDs(ctx, customRoleIDs)
		if deptErr != nil {
			return nil, deptErr
		}
		for _, deptID := range customDepts {
			deptSet[deptID] = struct{}{}
		}
	}

	for deptID := range deptSet {
		scope.DeptIDs = append(scope.DeptIDs, deptID)
	}
	return scope, nil
}

func loadBotDeptSubtreeIDs(ctx context.Context, rootDeptID int64) []int64 {
	if rootDeptID == 0 {
		return nil
	}
	result := []int64{rootDeptID}
	children, err := repo.NewSystemScopeRepo().ListChildDeptIDs(ctx, rootDeptID)
	if err != nil || len(children) == 0 {
		return result
	}
	for _, childID := range children {
		result = append(result, loadBotDeptSubtreeIDs(ctx, childID)...)
	}
	return result
}

// findProjectByKeyword 按项目名或 ID 查找项目（限当前用户）。
func findProjectByKeyword(ctx context.Context, keyword string, userID int64) (gdb.Record, error) {
	scope, err := resolveBotProjectScope(ctx, userID)
	if err != nil {
		return nil, err
	}

	record, err := repo.NewProjectRepo().FindByKeywordWithScope(ctx, keyword, repo.ProjectScopeFilter{
		All:         scope.All,
		IncludeSelf: scope.IncludeSelf,
		UserID:      userID,
		DeptIDs:     scope.DeptIDs,
	}, "id", "name", "status", "pause_reason", "engine_version")
	if err != nil {
		return nil, err
	}
	if len(record) == 0 {
		return nil, nil
	}
	return mapToDBRecord(record), nil
}

// extractFeishuText 从飞书消息 content JSON 中提取文本。
func extractFeishuText(contentStr string) string {
	if contentStr == "" {
		return ""
	}
	var content map[string]interface{}
	if err := json.Unmarshal([]byte(contentStr), &content); err != nil {
		return contentStr
	}
	if text, ok := content["text"].(string); ok {
		return text
	}
	return ""
}

// removeAtMention 去除飞书消息中的 @机器人 提及标记。
func removeAtMention(text string) string {
	for strings.Contains(text, "<at ") {
		start := strings.Index(text, "<at ")
		end := strings.Index(text, "</at>")
		if end == -1 {
			break
		}
		text = text[:start] + text[end+5:]
	}
	text = strings.TrimPrefix(text, "@EasyMVP")
	text = strings.TrimPrefix(text, "@easymvp")
	return strings.TrimSpace(text)
}

// extractJSON 从字符串中提取第一个完整的 JSON 对象。
func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 || end < start {
		return s
	}
	return s[start : end+1]
}

// handleBotConfirmPlan 确认当前活跃对话的方案并启动执行。
func handleBotConfirmPlan(ctx context.Context, openID, platform string, systemUserID int64, reply func(string)) {
	convID, ok := getBotSession(platform, openID)
	if !ok {
		reply("❌ 当前没有活跃的项目对话，请先创建项目")
		return
	}
	conv, err := repo.NewConversationRepo().GetByID(ctx, convID, "project_id")
	if err != nil || len(conv) == 0 {
		reply("❌ 对话不存在")
		return
	}
	projectID := g.NewVar(conv["project_id"]).Int64()
	if err := engine.GetScheduler().ConfirmPlan(ctx, projectID); err != nil {
		reply(fmt.Sprintf("❌ 确认方案失败：%v", err))
		return
	}
	clearBotSession(platform, openID)
	reply(fmt.Sprintf("🚀 方案已确认！项目 ID:%d 开始自动执行。\n发送「项目状态 %d」可查看执行进度", projectID, projectID))
}

// waitForAIReply 轮询等待 AI 回复消息完成（status=completed）。
func waitForAIReply(ctx context.Context, replyMsgID int64, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		record, err := repo.NewMessageRepo().GetByID(ctx, replyMsgID, "content", "status")
		if err == nil && len(record) > 0 {
			status := mapString(record, "status")
			if status == "completed" || status == "done" {
				return mapString(record, "content")
			}
			if status == "failed" {
				return ""
			}
		}
		time.Sleep(1 * time.Second)
	}
	return ""
}

// ─── 任务管理处理器 ─────────────────────────────────────────────────────────

func handleBotListTasks(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()
	type taskRow struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	var tasks []taskRow
	if isWorkflowV2Project(project) {
		wfRun, wfErr := latestWorkflowRunForProject(ctx, projectID)
		if wfErr != nil {
			reply(fmt.Sprintf("❌ 查询工作流失败：%v", wfErr))
			return
		}
		records, _ := repo.NewDomainTaskRepo().ListByWorkflowOrdered(ctx, wfRun["id"].Int64(), "id", "name", "status")
		for i, record := range records {
			if i >= 20 {
				break
			}
			tasks = append(tasks, taskRow{
				ID:     g.NewVar(record["id"]).Int64(),
				Name:   mapString(record, "name"),
				Status: mapString(record, "status"),
			})
		}
	} else {
		records, _ := repo.NewTaskRepo().ListByProject(ctx, projectID, 20, "id", "name", "status")
		for _, record := range records {
			tasks = append(tasks, taskRow{
				ID:     g.NewVar(record["id"]).Int64(),
				Name:   mapString(record, "name"),
				Status: mapString(record, "status"),
			})
		}
	}

	if len(tasks) == 0 {
		reply(fmt.Sprintf("📭 项目「%s」暂无任务", project["name"].String()))
		return
	}

	statusIcon := map[string]string{
		"pending":   "⏳",
		"running":   "🔄",
		"completed": "✅",
		"failed":    "❌",
		"escalated": "🧠",
		"skipped":   "⏭️",
		"draft":     "📝",
	}
	lines := []string{fmt.Sprintf("📋 %s 的任务（最近20条）", project["name"].String()), "───────────────"}
	for _, t := range tasks {
		icon := statusIcon[t.Status]
		if icon == "" {
			icon = "•"
		}
		lines = append(lines, fmt.Sprintf("%s [%d] %s（%s）", icon, t.ID, t.Name, t.Status))
	}
	reply(strings.Join(lines, "\n"))
}

func handleBotRetryTask(ctx context.Context, projectName, taskIDStr string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()

	if isWorkflowV2Project(project) {
		wfRun, wfErr := latestWorkflowRunForProject(ctx, projectID)
		if wfErr != nil {
			reply(fmt.Sprintf("❌ 查询工作流失败：%v", wfErr))
			return
		}

		if taskIDStr != "" {
			var taskID int64
			fmt.Sscanf(taskIDStr, "%d", &taskID)
			if taskID > 0 {
				taskRecord, getErr := repo.NewDomainTaskRepo().GetByWorkflowAndID(ctx, wfRun["id"].Int64(), taskID, "id")
				if getErr != nil || len(taskRecord) == 0 {
					reply(fmt.Sprintf("❌ 任务 %d 不属于当前工作流", taskID))
					return
				}
				rows, err := repo.NewDomainTaskRepo().ResetForRetry(ctx, taskID)
				if err != nil {
					reply(fmt.Sprintf("❌ 重试任务 %d 失败：%v", taskID, err))
					return
				}
				if rows == 0 {
					reply(fmt.Sprintf("❌ 任务 %d 当前不在 failed/escalated 状态", taskID))
					return
				}
				reply(fmt.Sprintf("🔄 任务 %d 已重新加入 V2 队列", taskID))
				return
			}
		}

		rows, _ := repo.NewDomainTaskRepo().ListByWorkflowAndStatuses(ctx, wfRun["id"].Int64(), []string{"failed", "escalated"}, "id")
		if len(rows) == 0 {
			reply(fmt.Sprintf("✅ 项目「%s」没有失败的 V2 任务", project["name"].String()))
			return
		}
		errCount := 0
		for _, r := range rows {
			rowsAffected, err := repo.NewDomainTaskRepo().ResetForRetry(ctx, g.NewVar(r["id"]).Int64())
			if err != nil {
				errCount++
				continue
			}
			if rowsAffected == 0 {
				errCount++
			}
		}
		reply(fmt.Sprintf("🔄 已重试 %d 个 V2 失败任务（失败 %d 个）", len(rows)-errCount, errCount))
		return
	}

	// Legacy: 如果指定了 task_id，重试单个
	if taskIDStr != "" {
		var taskID int64
		fmt.Sscanf(taskIDStr, "%d", &taskID)
		if taskID > 0 {
			if err := engine.GetScheduler().RetryTask(projectID, taskID); err != nil {
				reply(fmt.Sprintf("❌ 重试任务 %d 失败：%v", taskID, err))
				return
			}
			reply(fmt.Sprintf("🔄 任务 %d 已重新加入队列", taskID))
			return
		}
	}

	taskIDs, _ := repo.NewTaskRepo().ListIDsByProjectStatus(ctx, projectID, "failed")
	if len(taskIDs) == 0 {
		reply(fmt.Sprintf("✅ 项目「%s」没有失败的任务", project["name"].String()))
		return
	}
	errCount := 0
	for _, taskID := range taskIDs {
		if err := engine.GetScheduler().RetryTask(projectID, taskID); err != nil {
			errCount++
		}
	}
	reply(fmt.Sprintf("🔄 已重试 %d 个失败任务（失败 %d 个）", len(taskIDs)-errCount, errCount))
}

func handleBotSkipTask(ctx context.Context, projectName, taskIDStr string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	if taskIDStr == "" {
		reply("❌ 请提供任务ID，例如：跳过任务 123456")
		return
	}
	var taskID int64
	fmt.Sscanf(taskIDStr, "%d", &taskID)
	if taskID == 0 {
		reply("❌ 任务ID格式不正确")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	if isWorkflowV2Project(project) {
		wfRun, wfErr := latestWorkflowRunForProject(ctx, project["id"].Int64())
		if wfErr != nil {
			reply(fmt.Sprintf("❌ 查询工作流失败：%v", wfErr))
			return
		}
		taskRecord, getErr := repo.NewDomainTaskRepo().GetByWorkflowAndID(ctx, wfRun["id"].Int64(), taskID, "id")
		if getErr != nil || len(taskRecord) == 0 {
			reply("❌ 任务不属于当前工作流")
			return
		}
		rows, err := repo.NewDomainTaskRepo().CompleteAsSkipped(ctx, taskID, gtime.Now())
		if err != nil {
			reply(fmt.Sprintf("❌ 跳过任务失败：%v", err))
			return
		}
		if rows == 0 {
			reply("❌ 任务当前不在可跳过状态")
			return
		}
		if completeErr := orchestrator.GetTaskScheduler().OnTaskCompleted(ctx, taskID); completeErr != nil {
			g.Log().Warningf(ctx, "[BotSkipTask] 通知调度器任务完成失败: task=%d err=%v", taskID, completeErr)
		}
		reply(fmt.Sprintf("⏭️ V2 任务 %d 已跳过", taskID))
		return
	}

	taskRecord, _ := repo.NewTaskRepo().GetByID(ctx, taskID, "project_id")
	var skipProjectID int64
	if len(taskRecord) > 0 {
		skipProjectID = g.NewVar(taskRecord["project_id"]).Int64()
	}
	if err := engine.GetScheduler().SkipTask(ctx, skipProjectID, taskID, "飞书机器人指令跳过"); err != nil {
		reply(fmt.Sprintf("❌ 跳过任务失败：%v", err))
		return
	}
	reply(fmt.Sprintf("⏭️ 任务 %d 已跳过", taskID))
}

// ─── 审核管理处理器 ─────────────────────────────────────────────────────────

func handleBotReviewStatus(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()

	// 查最新的 workflow_run
	run, _ := repo.NewWorkflowRunRepo().GetLatestByProject(ctx, projectID)
	if run.IsEmpty() {
		reply(fmt.Sprintf("📭 项目「%s」暂无工作流记录，请先确认方案", project["name"].String()))
		return
	}

	// 查审核问题数量
	issueCount, _ := repo.NewReviewIssueRepo().CountOpenByWorkflow(ctx, run["id"].Int64())

	lines := []string{
		fmt.Sprintf("🔍 %s 审核状态", project["name"].String()),
		"───────────────",
		fmt.Sprintf("工作流状态：%s", run["status"].String()),
		fmt.Sprintf("审核状态：%s", run["review_status"].String()),
		fmt.Sprintf("问题数量：%d", issueCount),
	}
	if run["review_status"].String() == "waiting" {
		lines = append(lines, "\n💡 发送「通过审核 "+project["name"].String()+"」或「驳回审核 "+project["name"].String()+"」")
	}
	reply(strings.Join(lines, "\n"))
}

func handleBotApproveReview(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()

	run, _ := repo.NewWorkflowRunRepo().GetLatestByProjectReviewStatus(ctx, projectID, "waiting", "id")
	if len(run) == 0 {
		reply(fmt.Sprintf("❌ 项目「%s」当前没有等待审核的工作流", project["name"].String()))
		return
	}

	err = repo.NewWorkflowRunRepo().UpdateFields(ctx, g.NewVar(run["id"]).Int64(), g.Map{"review_status": "approved", "reviewed_by": systemUserID})
	if err != nil {
		reply(fmt.Sprintf("❌ 审核操作失败：%v", err))
		return
	}
	reply(fmt.Sprintf("✅ 项目「%s」审核已通过，系统将继续执行", project["name"].String()))
}

func handleBotRejectReview(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()

	run, _ := repo.NewWorkflowRunRepo().GetLatestByProjectReviewStatus(ctx, projectID, "waiting", "id")
	if len(run) == 0 {
		reply(fmt.Sprintf("❌ 项目「%s」当前没有等待审核的工作流", project["name"].String()))
		return
	}

	err = repo.NewWorkflowRunRepo().UpdateFields(ctx, g.NewVar(run["id"]).Int64(), g.Map{"review_status": "rejected", "reviewed_by": systemUserID})
	if err != nil {
		reply(fmt.Sprintf("❌ 审核操作失败：%v", err))
		return
	}
	reply(fmt.Sprintf("🚫 项目「%s」审核已驳回，项目已暂停", project["name"].String()))
}

// ─── 验收管理处理器 ─────────────────────────────────────────────────────────

func handleBotAcceptStatus(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()

	run, _ := repo.NewAcceptRunRepo().GetLatestByProject(ctx, projectID, "id", "status", "passed_count", "failed_count", "created_at")
	if len(run) == 0 {
		reply(fmt.Sprintf("📭 项目「%s」暂无验收记录", project["name"].String()))
		return
	}

	lines := []string{
		fmt.Sprintf("🎯 %s 验收状态", project["name"].String()),
		"───────────────",
		fmt.Sprintf("验收状态：%s", mapString(run, "status")),
		fmt.Sprintf("通过：%d  失败：%d", mapInt(run, "passed_count"), mapInt(run, "failed_count")),
	}
	if mapString(run, "status") == "pending_human" {
		lines = append(lines, "\n💡 发送「验收通过 "+project["name"].String()+"」或「验收驳回 "+project["name"].String()+"」")
	}
	reply(strings.Join(lines, "\n"))
}

func handleBotApproveAccept(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()

	run, _ := repo.NewAcceptRunRepo().GetLatestByProjectStatus(ctx, projectID, "pending_human", "id")
	if len(run) == 0 {
		reply(fmt.Sprintf("❌ 项目「%s」当前没有待人工验收的记录", project["name"].String()))
		return
	}

	err = repo.NewAcceptRunRepo().UpdateFields(ctx, g.NewVar(run["id"]).Int64(), g.Map{"status": "passed", "reviewed_by": systemUserID})
	if err != nil {
		reply(fmt.Sprintf("❌ 验收操作失败：%v", err))
		return
	}
	reply(fmt.Sprintf("🎉 项目「%s」验收通过！项目即将完成", project["name"].String()))
}

func handleBotRejectAccept(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()

	run, _ := repo.NewAcceptRunRepo().GetLatestByProjectStatus(ctx, projectID, "pending_human", "id")
	if len(run) == 0 {
		reply(fmt.Sprintf("❌ 项目「%s」当前没有待人工验收的记录", project["name"].String()))
		return
	}

	err = repo.NewAcceptRunRepo().UpdateFields(ctx, g.NewVar(run["id"]).Int64(), g.Map{"status": "rework", "reviewed_by": systemUserID})
	if err != nil {
		reply(fmt.Sprintf("❌ 验收操作失败：%v", err))
		return
	}
	reply(fmt.Sprintf("🔁 项目「%s」验收驳回，已打回返工", project["name"].String()))
}

func handleBotVerificationStart(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}

	runID, workflowRunID, err := startVerificationRun(ctx, project["id"].Int64(), systemUserID, 0, "feishu", "飞书手动触发验证")
	if err != nil {
		reply(fmt.Sprintf("❌ 启动验证失败：%v", err))
		return
	}
	reply(fmt.Sprintf("🧪 项目「%s」验证已启动，verificationRun=%d，workflowRun=%d", project["name"].String(), runID, workflowRunID))
}

func handleBotVerificationStatus(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}

	_, verificationRun, issues, _, err := loadLatestVerificationBundle(ctx, project["id"].Int64())
	if err != nil {
		reply(fmt.Sprintf("❌ 查询验证状态失败：%v", err))
		return
	}
	if len(verificationRun) == 0 {
		reply(fmt.Sprintf("📭 项目「%s」暂无验证记录", project["name"].String()))
		return
	}

	blockers, errorsCount, warns, infos := countVerificationSeverities(issues)
	lines := []string{
		fmt.Sprintf("🧪 %s 验证状态", project["name"].String()),
		"───────────────",
		fmt.Sprintf("状态：%s", g.NewVar(verificationRun["status"]).String()),
		fmt.Sprintf("结论：%s", g.NewVar(verificationRun["decision"]).String()),
		fmt.Sprintf("执行器：%s", g.NewVar(verificationRun["runner_type"]).String()),
		fmt.Sprintf("问题：blocker=%d error=%d warn=%d info=%d", blockers, errorsCount, warns, infos),
	}
	if summary := strings.TrimSpace(g.NewVar(verificationRun["summary"]).String()); summary != "" {
		lines = append(lines, "摘要："+summary)
	}
	reply(strings.Join(lines, "\n"))
}

func handleBotVerificationRepair(ctx context.Context, projectName, issueID string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}

	var issueIDs []int64
	if strings.TrimSpace(issueID) != "" {
		var parsed int64
		if _, scanErr := fmt.Sscanf(strings.TrimSpace(issueID), "%d", &parsed); scanErr != nil || parsed == 0 {
			reply("❌ issueID 格式不正确")
			return
		}
		issueIDs = []int64{parsed}
	} else {
		_, verificationRun, issues, _, loadErr := loadLatestVerificationBundle(ctx, project["id"].Int64())
		if loadErr != nil {
			reply(fmt.Sprintf("❌ 查询验证问题失败：%v", loadErr))
			return
		}
		if len(verificationRun) == 0 || len(issues) == 0 {
			reply(fmt.Sprintf("❌ 项目「%s」当前没有可返工的验证问题", project["name"].String()))
			return
		}
		for _, item := range issues {
			if g.NewVar(item["status"]).String() == "open" && g.NewVar(item["domain_task_id"]).Int64() > 0 {
				issueIDs = []int64{g.NewVar(item["id"]).Int64()}
				break
			}
		}
		if len(issueIDs) == 0 {
			reply(fmt.Sprintf("❌ 项目「%s」的验证问题没有关联可返工任务", project["name"].String()))
			return
		}
	}

	message, err := requestVerificationRepair(ctx, project["id"].Int64(), issueIDs, "飞书手动触发验证返工")
	if err != nil {
		reply(fmt.Sprintf("❌ 启动验证返工失败：%v", err))
		return
	}
	reply("🔧 " + message)
}

// ─── 自治管理处理器 ─────────────────────────────────────────────────────────

func handleBotAutonomyStatus(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()

	// 查待审检查点
	checkpointRecords, _ := repo.NewHumanCheckpointRepo().ListByProjectStatus(ctx, projectID, "pending", 5, "id", "check_type", "description")
	type cpRow struct {
		ID          int64
		CheckType   string
		Description string
	}
	checkpoints := make([]cpRow, 0, len(checkpointRecords))
	for _, checkpoint := range checkpointRecords {
		checkpoints = append(checkpoints, cpRow{
			ID:          g.NewVar(checkpoint["id"]).Int64(),
			CheckType:   mapString(checkpoint, "check_type"),
			Description: mapString(checkpoint, "description"),
		})
	}

	// 查最新自治决策
	decisionCount, _ := repo.NewAutonomyDecisionRepo().CountByProject(ctx, projectID)

	lines := []string{
		fmt.Sprintf("🤖 %s 自治状态", project["name"].String()),
		"───────────────",
		fmt.Sprintf("累计自治决策：%d 次", decisionCount),
		fmt.Sprintf("待审检查点：%d 个", len(checkpoints)),
	}
	for _, cp := range checkpoints {
		lines = append(lines, fmt.Sprintf("  📌 [%d] %s：%s", cp.ID, cp.CheckType, cp.Description))
	}
	if len(checkpoints) > 0 {
		lines = append(lines, "\n💡 发送「批准检查点 "+project["name"].String()+"」或「拒绝检查点 "+project["name"].String()+"」")
	}
	reply(strings.Join(lines, "\n"))
}

func handleBotApproveCheckpoint(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()

	n, err := repo.NewHumanCheckpointRepo().UpdateStatusByProject(ctx, projectID, "pending", g.Map{"status": "approved", "reviewed_by": systemUserID})
	if err != nil {
		reply(fmt.Sprintf("❌ 操作失败：%v", err))
		return
	}
	if n == 0 {
		reply(fmt.Sprintf("❌ 项目「%s」没有待审的检查点", project["name"].String()))
		return
	}
	reply(fmt.Sprintf("✅ 已批准 %d 个自治检查点，系统继续执行", n))
}

func handleBotRejectCheckpoint(ctx context.Context, projectName string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 请先在 EasyMVP 管理端绑定飞书账号")
		return
	}
	project, err := findProjectByKeyword(ctx, projectName, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", projectName))
		return
	}
	projectID := project["id"].Int64()

	n, err := repo.NewHumanCheckpointRepo().UpdateStatusByProject(ctx, projectID, "pending", g.Map{"status": "rejected", "reviewed_by": systemUserID})
	if err != nil {
		reply(fmt.Sprintf("❌ 操作失败：%v", err))
		return
	}
	if n == 0 {
		reply(fmt.Sprintf("❌ 项目「%s」没有待审的检查点", project["name"].String()))
		return
	}
	reply(fmt.Sprintf("🚫 已拒绝 %d 个自治检查点，项目已暂停等待人工处理", n))
}

// feishuHelpText 飞书平台帮助文本（复用通用版本）。
func feishuHelpText() string { return botHelpText() }
