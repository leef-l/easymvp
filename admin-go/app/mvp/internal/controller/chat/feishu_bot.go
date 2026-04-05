package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	Action      string `json:"action"`       // create_project | list_projects | project_status | pause_project | resume_project | chat | help
	ProjectName string `json:"project_name"` // 项目名称（create/status/pause/resume 时有值）
	Category    string `json:"category"`     // 项目分类（create 时有值，默认"软件开发"）
	Reply       string `json:"reply"`        // chat 时 AI 的直接回复文本
}

// intentSystemPrompt AI 意图解析系统提示词。
const intentSystemPrompt = `你是 EasyMVP 的飞书机器人助手，负责理解用户的自然语言消息并转换为结构化指令。

用户可能用各种表达方式描述意图，你需要准确识别并返回 JSON。

支持的 action：
- create_project：创建新项目（需要 project_name，可选 category）
- list_projects：列出项目
- project_status：查询项目状态（需要 project_name）
- pause_project：暂停项目（需要 project_name）
- resume_project：继续/恢复项目（需要 project_name）
- chat：普通对话或不确定的意图（在 reply 字段填写你的回复）
- help：显示帮助

category 常见值：软件开发、游戏开发、数据分析、内容创作、运营策划。未指定时默认"软件开发"。

只返回 JSON，不要有其他文字。格式：
{"action":"...","project_name":"...","category":"...","reply":"..."}`

// DispatchFeishuCommand 用 AI 解析用户意图后路由到对应处理器。
func DispatchFeishuCommand(ctx context.Context, openID, messageID, chatID, contentStr string) {
	feishu := adapter.NewFeishuAdapter()
	reply := func(text string) {
		if messageID != "" {
			if err := feishu.ReplyMessage(ctx, messageID, text); err != nil {
				g.Log().Warningf(ctx, "[FeishuBot] 回复失败: %v", err)
			}
			return
		}
		if chatID != "" {
			if err := feishu.SendTextToChat(ctx, chatID, text); err != nil {
				g.Log().Warningf(ctx, "[FeishuBot] 发送群消息失败: %v", err)
			}
		}
	}

	// 1. 解析消息文本
	text := extractFeishuText(contentStr)
	text = removeAtMention(text)
	text = strings.TrimSpace(text)
	if text == "" {
		reply(feishuHelpText())
		return
	}

	// 2. 反查绑定的系统用户
	systemUserID, deptID := lookupSystemUser(ctx, openID)

	// 3. AI 解析意图
	intent, err := parseIntentWithAI(ctx, text, systemUserID)
	if err != nil {
		g.Log().Warningf(ctx, "[FeishuBot] AI 意图解析失败，降级到关键词匹配: %v", err)
		// 降级到简单关键词匹配
		intent = fallbackParseIntent(text)
	}

	g.Log().Infof(ctx, "[FeishuBot] 意图: action=%s project=%s category=%s", intent.Action, intent.ProjectName, intent.Category)

	// 4. 路由执行
	switch intent.Action {
	case "create_project":
		handleBotCreateProject(ctx, intent.ProjectName, intent.Category, systemUserID, deptID, reply)
	case "list_projects":
		handleBotListProjects(ctx, systemUserID, reply)
	case "project_status":
		handleBotProjectStatus(ctx, intent.ProjectName, systemUserID, reply)
	case "pause_project":
		handleBotPauseProject(ctx, intent.ProjectName, systemUserID, reply)
	case "resume_project":
		handleBotResumeProject(ctx, intent.ProjectName, systemUserID, reply)
	case "help":
		reply(feishuHelpText())
	case "chat":
		if intent.Reply != "" {
			reply(intent.Reply)
		} else {
			reply(feishuHelpText())
		}
	default:
		reply(feishuHelpText())
	}
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

func handleBotCreateProject(ctx context.Context, projectName, category string, systemUserID, deptID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
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

	reply(fmt.Sprintf(
		"✅ 项目已创建\n"+
			"───────────────\n"+
			"📁 项目名称：%s\n"+
			"🏷️ 分类：%s\n"+
			"🆔 项目ID：%d\n"+
			"───────────────\n"+
			"下一步：在 EasyMVP 管理端与架构师对话，确认方案后自动执行。",
		projectName, category, projectID,
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

// feishuHelpText 返回帮助文本。
func feishuHelpText() string {
	return `🤖 EasyMVP 机器人
───────────────
我能理解自然语言，直接说需求即可：

📁 "帮我创建一个电商后台项目"
📁 "新建游戏开发类型的H5小游戏"
📋 "列出我的项目"
📊 "电商后台进度怎么样了"
⏸️ "暂停电商后台"
▶️ "继续执行电商后台"
❓ "帮助"
───────────────
提示：在飞书中 @EasyMVP 后说话`
}
