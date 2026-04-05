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
)

// DispatchFeishuCommand 解析飞书消息内容，路由到对应指令处理器。
// openID: 飞书用户 open_id
// messageID: 原始消息 ID（用于回复）
// chatID: 会话 ID
// contentStr: 飞书消息 content JSON 字符串
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
	// 去除 @机器人 提及（飞书格式：<at user_id="xxx">name</at> 或 @机器人名称）
	text = removeAtMention(text)
	text = strings.TrimSpace(text)

	if text == "" {
		reply(feishuHelpText())
		return
	}

	// 2. 反查绑定的系统用户
	systemUserID, deptID := lookupSystemUser(ctx, openID)

	// 3. 路由指令
	lower := strings.ToLower(text)
	switch {
	case strings.HasPrefix(lower, "创建项目") || strings.HasPrefix(lower, "new project") || strings.HasPrefix(lower, "create project"):
		handleBotCreateProject(ctx, text, systemUserID, deptID, reply)

	case lower == "项目列表" || lower == "list projects" || lower == "ls":
		handleBotListProjects(ctx, systemUserID, reply)

	case strings.HasPrefix(lower, "项目状态") || strings.HasPrefix(lower, "status"):
		handleBotProjectStatus(ctx, text, systemUserID, reply)

	case strings.HasPrefix(lower, "暂停") || strings.HasPrefix(lower, "pause"):
		handleBotPauseProject(ctx, text, systemUserID, reply)

	case strings.HasPrefix(lower, "继续") || strings.HasPrefix(lower, "resume"):
		handleBotResumeProject(ctx, text, systemUserID, reply)

	case lower == "帮助" || lower == "help" || lower == "?":
		reply(feishuHelpText())

	default:
		reply(fmt.Sprintf("未识别的指令：「%s」\n\n%s", text, feishuHelpText()))
	}
}

// handleBotCreateProject 处理创建项目指令。
// 格式：创建项目 <名称> [分类:<分类>]
func handleBotCreateProject(ctx context.Context, text string, systemUserID, deptID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}

	// 去掉指令前缀
	arg := trimCommandPrefix(text, "创建项目", "new project", "create project")
	arg = strings.TrimSpace(arg)
	if arg == "" {
		reply("❌ 请提供项目名称\n示例：创建项目 电商后台\n或：创建项目 H5游戏 分类:游戏开发")
		return
	}

	// 解析分类
	projectName := arg
	category := "软件开发"
	if idx := strings.Index(arg, "分类:"); idx != -1 {
		projectName = strings.TrimSpace(arg[:idx])
		category = strings.TrimSpace(arg[idx+len("分类:"):])
	} else if idx := strings.Index(arg, "分类："); idx != -1 {
		projectName = strings.TrimSpace(arg[:idx])
		category = strings.TrimSpace(arg[idx+len("分类："):])
	}

	if projectName == "" {
		reply("❌ 项目名称不能为空")
		return
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

// handleBotListProjects 处理项目列表指令。
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
		reply("📭 暂无项目，发送「创建项目 <名称>」快速创建。")
		return
	}

	statusLabel := map[string]string{
		"designing":  "设计中",
		"running":    "执行中",
		"paused":     "已暂停",
		"completed":  "已完成",
	}
	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, "📋 您的项目列表（最近8个）")
	lines = append(lines, "───────────────")
	for i, row := range rows {
		label := statusLabel[row.Status]
		if label == "" {
			label = row.Status
		}
		lines = append(lines, fmt.Sprintf("%d. %s  [%s]  ID:%d", i+1, row.Name, label, row.ID))
	}
	reply(strings.Join(lines, "\n"))
}

// handleBotProjectStatus 处理项目状态查询指令。
// 格式：项目状态 <项目名或ID>
func handleBotProjectStatus(ctx context.Context, text string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}

	keyword := strings.TrimSpace(trimCommandPrefix(text, "项目状态", "status"))
	if keyword == "" {
		reply("❌ 请提供项目名称或ID\n示例：项目状态 电商后台")
		return
	}

	project, err := findProjectByKeyword(ctx, keyword, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", keyword))
		return
	}

	projectID := project["id"].Int64()
	projectName := project["name"].String()
	status := project["status"].String()

	// 统计任务数
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

	pr := project["pause_reason"].String()
	lines := []string{
		fmt.Sprintf("📊 项目状态：%s", projectName),
		"───────────────",
		fmt.Sprintf("状态：%s", label),
		fmt.Sprintf("任务：共%d个 | 完成%d | 运行%d | 失败%d", total, done, running, failed),
	}
	if total > 0 {
		pct := done * 100 / total
		lines = append(lines, fmt.Sprintf("进度：%d%%", pct))
	}
	if pr != "" {
		lines = append(lines, fmt.Sprintf("暂停原因：%s", pr))
	}
	lines = append(lines, fmt.Sprintf("项目ID：%d", projectID))

	reply(strings.Join(lines, "\n"))
}

// handleBotPauseProject 处理暂停项目指令。
func handleBotPauseProject(ctx context.Context, text string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}

	keyword := strings.TrimSpace(trimCommandPrefix(text, "暂停", "pause"))
	if keyword == "" {
		reply("❌ 请提供项目名称或ID\n示例：暂停 电商后台")
		return
	}

	project, err := findProjectByKeyword(ctx, keyword, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", keyword))
		return
	}

	projectID := project["id"].Int64()
	pName := project["name"].String()
	if err := engine.GetScheduler().Pause(ctx, projectID, "飞书机器人指令暂停"); err != nil {
		reply(fmt.Sprintf("❌ 暂停失败：%v", err))
		return
	}
	reply(fmt.Sprintf("⏸️ 项目「%s」已暂停", pName))
}

// handleBotResumeProject 处理继续项目指令。
func handleBotResumeProject(ctx context.Context, text string, systemUserID int64, reply func(string)) {
	if systemUserID == 0 {
		reply("❌ 您尚未绑定飞书账号，请先在 EasyMVP 管理端完成飞书绑定。")
		return
	}

	keyword := strings.TrimSpace(trimCommandPrefix(text, "继续", "resume"))
	if keyword == "" {
		reply("❌ 请提供项目名称或ID\n示例：继续 电商后台")
		return
	}

	project, err := findProjectByKeyword(ctx, keyword, systemUserID)
	if err != nil || project == nil {
		reply(fmt.Sprintf("❌ 未找到项目「%s」", keyword))
		return
	}

	projectID := project["id"].Int64()
	rName := project["name"].String()
	if err := engine.GetScheduler().Resume(ctx, projectID); err != nil {
		reply(fmt.Sprintf("❌ 恢复失败：%v", err))
		return
	}
	reply(fmt.Sprintf("▶️ 项目「%s」已继续执行", rName))
}

// ─── 工具函数 ────────────────────────────────────────────────────────────────

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

	// 尝试按 ID 精确查找
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

	// 按名称模糊查找（取最新一条）
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
	// 文本消息
	if text, ok := content["text"].(string); ok {
		return text
	}
	return ""
}

// removeAtMention 去除飞书消息中的 @机器人 提及标记。
func removeAtMention(text string) string {
	// 飞书富文本 @: <at id="xxx">name</at>
	for strings.Contains(text, "<at ") {
		start := strings.Index(text, "<at ")
		end := strings.Index(text, "</at>")
		if end == -1 {
			break
		}
		text = text[:start] + text[end+5:]
	}
	// 普通 @机器人 前缀
	text = strings.TrimPrefix(text, "@EasyMVP")
	text = strings.TrimPrefix(text, "@easymvp")
	return strings.TrimSpace(text)
}

// trimCommandPrefix 去除指令前缀，返回参数部分。
func trimCommandPrefix(text string, prefixes ...string) string {
	lower := strings.ToLower(text)
	for _, p := range prefixes {
		if strings.HasPrefix(lower, strings.ToLower(p)) {
			return text[len(p):]
		}
	}
	return text
}

// feishuHelpText 返回帮助文本。
func feishuHelpText() string {
	return `🤖 EasyMVP 机器人指令列表
───────────────
📁 创建项目 <名称>
   创建项目 H5游戏 分类:游戏开发
📋 项目列表
   查看我的项目（最近8个）
📊 项目状态 <名称或ID>
   查看任务进度和状态
⏸️ 暂停 <名称或ID>
   暂停项目执行
▶️ 继续 <名称或ID>
   恢复项目执行
❓ 帮助
   显示此帮助
───────────────
提示：在飞书中 @EasyMVP 后发送以上指令`
}
