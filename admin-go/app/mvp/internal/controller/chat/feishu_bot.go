package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/collab/adapter"
	collabRepo "easymvp/app/mvp/internal/collab/repo"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/utility/provider"
)

// botIntent AI 解析出的用户意图。
type botIntent struct {
	Action      string `json:"action"`       // 意图动作，见下方 intentSystemPrompt
	ProjectName string `json:"project_name"` // 项目名称（多数操作需要）
	Category    string `json:"category"`     // 项目分类（create 时有值）
	TaskID      string `json:"task_id"`      // 任务ID（retry_task/skip_task）
	IssueID     string `json:"issue_id"`     // 审核/验收问题ID
	Reply       string `json:"reply"`        // chat 时 AI 的直接回复文本
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
- confirm_plan：确认当前对话中的方案并启动自动执行

### 任务管理
- list_tasks：查看项目任务列表（project_name必填）
- retry_task：重试失败任务（project_name必填，task_id可选，不填则重试所有失败任务）
- skip_task：跳过阻塞任务（project_name必填，task_id必填）

### 审核管理
- review_status：查看项目当前审核状态和问题（project_name必填）
- approve_review：通过人工审核（project_name必填）
- reject_review：驳回人工审核（project_name必填）

### 验收管理
- accept_status：查看项目验收状态（project_name必填）
- approve_accept：验收通过（project_name必填）
- reject_accept：验收驳回/打回返工（project_name必填）

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
- "确认方案/开始执行/启动" → confirm_plan
- "任务列表/查看任务/XXX的任务" → list_tasks
- "重试/重新执行XXX失败任务" → retry_task
- "跳过任务/跳过阻塞" → skip_task
- "审核状态/审核结果/审核通过了吗" → review_status
- "通过审核/审核通过" → approve_review
- "驳回审核/审核不通过" → reject_review
- "验收状态/验收结果/通过验收了吗" → accept_status
- "验收通过/通过了" → approve_accept
- "验收不通过/打回/返工" → reject_accept
- "自治状态/自治模式/检查点" → autonomy_status
- "批准检查点/同意自治" → approve_checkpoint
- "拒绝检查点/不同意自治" → reject_checkpoint
- 其他：chat（reply填写友好回复）

category 常见值：软件开发、游戏开发、数据分析、内容创作、运营策划。未指定默认"软件开发"。

只返回 JSON，格式：{"action":"...","project_name":"...","category":"...","task_id":"...","issue_id":"...","reply":"..."}`

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
		g.Log().Infof(ctx, "[FeishuBot] 回复消息: messageID=%s text=%q", f.messageID, text)
		if err := feishu.ReplyMessage(ctx, f.messageID, text); err != nil {
			g.Log().Warningf(ctx, "[FeishuBot] 回复消息失败: %v", err)
		}
		return
	}
	// 其次发到对话（群聊或单聊 chat_id）
	if f.chatID != "" {
		g.Log().Infof(ctx, "[FeishuBot] 发送消息: chatID=%s text=%q", f.chatID, text)
		if err := feishu.SendTextToChat(ctx, f.chatID, text); err != nil {
			g.Log().Warningf(ctx, "[FeishuBot] 发送消息失败(chatID): %v", err)
		}
		return
	}
	// 兜底：用 open_id 发私信
	if f.openID != "" {
		g.Log().Infof(ctx, "[FeishuBot] 发送私信: openID=%s text=%q", f.openID, text)
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
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
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
		var modelID int64
		v, _ := g.DB().Ctx(ctx).Model("mvp_project p").
			LeftJoin("mvp_project_role r", "r.project_id = p.id AND r.role_type = 'architect' AND r.status = 1 AND r.deleted_at IS NULL").
			Where("p.created_by", systemUserID).
			WhereNull("p.deleted_at").
			Where("r.model_id > 0").
			Fields("r.model_id").
			OrderDesc("p.created_at").
			Value()
		if v != nil {
			modelID = v.Int64()
		}
		if modelID > 0 {
			info, err := engine.GetModelInfoByID(ctx, modelID)
			if err == nil && info != nil {
				return info, nil
			}
		}
	}

	// 降级：取全局第一个可用模型（ai_model join ai_plan join ai_provider，按 sort 排序）
	record, err := g.DB().Ctx(ctx).Model("ai_model m").
		LeftJoin("ai_plan p", "p.id = m.plan_id AND p.deleted_at IS NULL").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id AND pv.deleted_at IS NULL AND pv.status = 1").
		Fields("m.id").
		Where("m.deleted_at IS NULL").
		Where("m.status", 1).
		Where("p.api_key != ''").
		OrderAsc("m.sort").
		One()
	if err != nil || record.IsEmpty() {
		return nil, fmt.Errorf("系统未配置任何可用的 AI 模型")
	}

	return engine.GetModelInfoByID(ctx, record["id"].Int64())
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

	projectID, _, err := engine.CreateProject(ctx, projectName, category, "", "", 0, systemUserID, deptID)
	if err != nil {
		reply(fmt.Sprintf("❌ 创建项目失败：%v", err))
		return
	}

	// 查询项目对应的架构师对话，自动进入对话模式
	extraTip := ""
	if openID != "" {
		var convID int64
		val, _ := g.DB().Ctx(ctx).Model("mvp_conversation").
			Where("project_id", projectID).
			WhereNull("deleted_at").
			OrderAsc("created_at").
			Value("id")
		if val != nil {
			convID = val.Int64()
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

	type projectRow struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	var rows []projectRow
	if err := g.DB().Ctx(ctx).Model("mvp_project").
		Where("created_by", systemUserID).
		WhereNull("deleted_at").
		Fields("id, name, status").
		OrderDesc("created_at").
		Limit(8).
		Scan(&rows); err != nil {
		reply(fmt.Sprintf("❌ 查询失败：%v", err))
		return
	}

	if len(rows) == 0 {
		reply("📭 暂无项目，可以说「帮我创建一个XXX项目」来快速创建。")
		return
	}

	statusLabel := map[string]string{
		"designing": "⚙️ 设计中",
		"running":   "🚀 执行中",
		"paused":    "⏸️ 已暂停",
		"completed": "✅ 已完成",
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

	type sc struct {
		Status string
		Count  int
	}
	var counts []sc
	_ = g.DB().Ctx(ctx).Model("mvp_task").
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Fields("status, COUNT(*) as count").
		Group("status").
		Scan(&counts)

	total, done, running, failed := 0, 0, 0, 0
	for _, c := range counts {
		total += c.Count
		switch c.Status {
		case "completed":
			done += c.Count
		case "running":
			running += c.Count
		case "failed":
			failed += c.Count
		}
	}

	statusLabel := map[string]string{
		"designing": "⚙️ 设计中",
		"running":   "🚀 执行中",
		"paused":    "⏸️ 已暂停",
		"completed": "✅ 已完成",
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
	if err := engine.GetScheduler().Pause(ctx, project["id"].Int64(), "飞书机器人指令暂停"); err != nil {
		reply(fmt.Sprintf("❌ 暂停失败：%v", err))
		return
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
	if err := engine.GetScheduler().Resume(ctx, project["id"].Int64()); err != nil {
		reply(fmt.Sprintf("❌ 恢复失败：%v", err))
		return
	}
	reply(fmt.Sprintf("▶️ 项目「%s」已继续执行", project["name"].String()))
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

// findProjectByKeyword 按项目名或 ID 查找项目（限当前用户）。
func findProjectByKeyword(ctx context.Context, keyword string, userID int64) (gdb.Record, error) {
	m := g.DB().Ctx(ctx).Model("mvp_project").
		WhereNull("deleted_at").
		Where("created_by", userID).
		Fields("id, name, status, pause_reason")

	var numID int64
	if _, err := fmt.Sscanf(keyword, "%d", &numID); err == nil && numID > 0 {
		record, err := m.Where("id", numID).One()
		if err != nil {
			return nil, err
		}
		if !record.IsEmpty() {
			return record, nil
		}
	}

	record, err := m.WhereLike("name", "%"+keyword+"%").OrderDesc("created_at").One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, nil
	}
	return record, nil
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
	conv, err := g.DB().Ctx(ctx).Model("mvp_conversation").
		Where("id", convID).WhereNull("deleted_at").One()
	if err != nil || conv.IsEmpty() {
		reply("❌ 对话不存在")
		return
	}
	projectID := conv["project_id"].Int64()
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
		record, err := g.DB().Model("mvp_message").Ctx(ctx).
			Where("id", replyMsgID).
			Fields("content, status").One()
		if err == nil && !record.IsEmpty() {
			status := record["status"].String()
			if status == "completed" || status == "done" {
				return record["content"].String()
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
	_ = g.DB().Ctx(ctx).Model("mvp_task").
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Fields("id, name, status").
		OrderAsc("batch_no").
		Limit(20).
		Scan(&tasks)

	if len(tasks) == 0 {
		reply(fmt.Sprintf("📭 项目「%s」暂无任务", project["name"].String()))
		return
	}

	statusIcon := map[string]string{
		"pending":   "⏳",
		"running":   "🔄",
		"completed": "✅",
		"failed":    "❌",
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

	// 如果指定了 task_id，重试单个
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

	// 重试所有失败任务
	type taskIDRow struct{ ID int64 }
	var rows []taskIDRow
	_ = g.DB().Ctx(ctx).Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "failed").
		WhereNull("deleted_at").
		Fields("id").
		Scan(&rows)

	if len(rows) == 0 {
		reply(fmt.Sprintf("✅ 项目「%s」没有失败的任务", project["name"].String()))
		return
	}
	errCount := 0
	for _, r := range rows {
		if err := engine.GetScheduler().RetryTask(projectID, r.ID); err != nil {
			errCount++
		}
	}
	reply(fmt.Sprintf("🔄 已重试 %d 个失败任务（失败 %d 个）", len(rows)-errCount, errCount))
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
	// 需要 projectID，先查任务所属项目
	taskRecord, _ := g.DB().Ctx(ctx).Model("mvp_task").Where("id", taskID).WhereNull("deleted_at").Fields("project_id").One()
	var skipProjectID int64
	if !taskRecord.IsEmpty() {
		skipProjectID = taskRecord["project_id"].Int64()
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
	run, _ := g.DB().Ctx(ctx).Model("mvp_workflow_run").
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Fields("id, status, review_status, created_at").
		OrderDesc("created_at").One()

	if run.IsEmpty() {
		reply(fmt.Sprintf("📭 项目「%s」暂无工作流记录，请先确认方案", project["name"].String()))
		return
	}

	// 查审核问题数量
	issueCount, _ := g.DB().Ctx(ctx).Model("mvp_review_issue").
		Where("workflow_run_id", run["id"].Int64()).
		WhereNull("deleted_at").
		Count()

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

	run, _ := g.DB().Ctx(ctx).Model("mvp_workflow_run").
		Where("project_id", projectID).
		Where("review_status", "waiting").
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if run.IsEmpty() {
		reply(fmt.Sprintf("❌ 项目「%s」当前没有等待审核的工作流", project["name"].String()))
		return
	}

	_, err = g.DB().Ctx(ctx).Model("mvp_workflow_run").
		Where("id", run["id"].Int64()).
		Update(g.Map{"review_status": "approved", "reviewed_by": systemUserID})
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

	run, _ := g.DB().Ctx(ctx).Model("mvp_workflow_run").
		Where("project_id", projectID).
		Where("review_status", "waiting").
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if run.IsEmpty() {
		reply(fmt.Sprintf("❌ 项目「%s」当前没有等待审核的工作流", project["name"].String()))
		return
	}

	_, err = g.DB().Ctx(ctx).Model("mvp_workflow_run").
		Where("id", run["id"].Int64()).
		Update(g.Map{"review_status": "rejected", "reviewed_by": systemUserID})
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

	run, _ := g.DB().Ctx(ctx).Model("mvp_accept_run").
		Where("project_id", projectID).
		WhereNull("deleted_at").
		Fields("id, status, passed_count, failed_count, created_at").
		OrderDesc("created_at").One()

	if run.IsEmpty() {
		reply(fmt.Sprintf("📭 项目「%s」暂无验收记录", project["name"].String()))
		return
	}

	lines := []string{
		fmt.Sprintf("🎯 %s 验收状态", project["name"].String()),
		"───────────────",
		fmt.Sprintf("验收状态：%s", run["status"].String()),
		fmt.Sprintf("通过：%d  失败：%d", run["passed_count"].Int(), run["failed_count"].Int()),
	}
	if run["status"].String() == "pending_human" {
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

	run, _ := g.DB().Ctx(ctx).Model("mvp_accept_run").
		Where("project_id", projectID).
		Where("status", "pending_human").
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if run.IsEmpty() {
		reply(fmt.Sprintf("❌ 项目「%s」当前没有待人工验收的记录", project["name"].String()))
		return
	}

	_, err = g.DB().Ctx(ctx).Model("mvp_accept_run").
		Where("id", run["id"].Int64()).
		Update(g.Map{"status": "passed", "reviewed_by": systemUserID})
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

	run, _ := g.DB().Ctx(ctx).Model("mvp_accept_run").
		Where("project_id", projectID).
		Where("status", "pending_human").
		WhereNull("deleted_at").
		OrderDesc("created_at").One()
	if run.IsEmpty() {
		reply(fmt.Sprintf("❌ 项目「%s」当前没有待人工验收的记录", project["name"].String()))
		return
	}

	_, err = g.DB().Ctx(ctx).Model("mvp_accept_run").
		Where("id", run["id"].Int64()).
		Update(g.Map{"status": "rework", "reviewed_by": systemUserID})
	if err != nil {
		reply(fmt.Sprintf("❌ 验收操作失败：%v", err))
		return
	}
	reply(fmt.Sprintf("🔁 项目「%s」验收驳回，已打回返工", project["name"].String()))
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
	type cpRow struct {
		ID          int64  `json:"id"`
		CheckType   string `json:"check_type"`
		Description string `json:"description"`
	}
	var checkpoints []cpRow
	_ = g.DB().Ctx(ctx).Model("mvp_human_checkpoint").
		Where("project_id", projectID).
		Where("status", "pending").
		WhereNull("deleted_at").
		Fields("id, check_type, description").
		Limit(5).
		Scan(&checkpoints)

	// 查最新自治决策
	decisionCount, _ := g.DB().Ctx(ctx).Model("mvp_autonomy_decision").
		Where("project_id", projectID).
		WhereNull("deleted_at").Count()

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

	result, err := g.DB().Ctx(ctx).Model("mvp_human_checkpoint").
		Where("project_id", projectID).
		Where("status", "pending").
		WhereNull("deleted_at").
		Update(g.Map{"status": "approved", "reviewed_by": systemUserID})
	if err != nil {
		reply(fmt.Sprintf("❌ 操作失败：%v", err))
		return
	}
	n, _ := result.RowsAffected()
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

	result, err := g.DB().Ctx(ctx).Model("mvp_human_checkpoint").
		Where("project_id", projectID).
		Where("status", "pending").
		WhereNull("deleted_at").
		Update(g.Map{"status": "rejected", "reviewed_by": systemUserID})
	if err != nil {
		reply(fmt.Sprintf("❌ 操作失败：%v", err))
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		reply(fmt.Sprintf("❌ 项目「%s」没有待审的检查点", project["name"].String()))
		return
	}
	reply(fmt.Sprintf("🚫 已拒绝 %d 个自治检查点，项目已暂停等待人工处理", n))
}

// feishuHelpText 飞书平台帮助文本（复用通用版本）。
func feishuHelpText() string { return botHelpText() }
