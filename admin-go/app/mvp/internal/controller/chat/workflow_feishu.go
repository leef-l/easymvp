package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/collab"
	"easymvp/app/mvp/internal/collab/adapter"
	collabRepo "easymvp/app/mvp/internal/collab/repo"
	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/workflow/orchestrator"
	"easymvp/utility/snowflake"
)

const feishuCallbackPath = "/api/mvp/collab/feishu/callback"

// FeishuConfig 查询飞书协作配置。
func (c *cWorkflow) FeishuConfig(ctx context.Context, req *v1.WorkflowFeishuConfigReq) (res *v1.WorkflowFeishuConfigRes, err error) {
	mode := engine.GetConfigString(ctx, "workflow.collab.feishu_connection_mode", "", "webhook")
	return &v1.WorkflowFeishuConfigRes{
		Config: v1.FeishuConfigDTO{
			Enabled:              engine.GetConfigInt(ctx, "workflow.collab.feishu_enabled", "workflow.collab.feishuEnabled", 0),
			AppID:                engine.GetConfigString(ctx, "workflow.collab.feishu_app_id", "workflow.collab.feishuAppId", ""),
			AppSecret:            engine.GetConfigString(ctx, "workflow.collab.feishu_app_secret", "workflow.collab.feishuAppSecret", ""),
			VerificationToken:    engine.GetConfigString(ctx, "workflow.collab.feishu_verification_token", "workflow.collab.feishuVerificationToken", ""),
			EncryptKey:           engine.GetConfigString(ctx, "workflow.collab.feishu_encrypt_key", "workflow.collab.feishuEncryptKey", ""),
			DefaultNotifyUserIDs: engine.GetConfigString(ctx, "workflow.collab.feishu_default_notify_user_ids", "workflow.collab.feishuDefaultNotifyUserIds", ""),
			ConnectionMode:       mode,
			CallbackPath:         feishuCallbackPath,
			WSRunning:            collab.GetWSManager().IsRunning(),
		},
	}, nil
}

// SaveFeishuConfig 保存飞书协作配置。
func (c *cWorkflow) SaveFeishuConfig(ctx context.Context, req *v1.WorkflowSaveFeishuConfigReq) (res *v1.WorkflowSaveFeishuConfigRes, err error) {
	mode := strings.TrimSpace(req.ConnectionMode)
	if mode != "websocket" {
		mode = "webhook"
	}
	configs := []struct {
		key         string
		value       string
		configType  string
		description string
	}{
		{"workflow.collab.feishu_enabled", fmt.Sprintf("%d", boolToInt(req.Enabled == 1)), "int", "飞书通知总开关(0关/1开)"},
		{"workflow.collab.feishu_app_id", strings.TrimSpace(req.AppID), "string", "飞书应用 App ID"},
		{"workflow.collab.feishu_app_secret", strings.TrimSpace(req.AppSecret), "string", "飞书应用 App Secret"},
		{"workflow.collab.feishu_verification_token", strings.TrimSpace(req.VerificationToken), "string", "飞书 Verification Token"},
		{"workflow.collab.feishu_encrypt_key", strings.TrimSpace(req.EncryptKey), "string", "飞书事件回调加密 Key(签名验证)"},
		{"workflow.collab.feishu_default_notify_user_ids", strings.TrimSpace(req.DefaultNotifyUserIDs), "string", "降级通知的系统用户ID列表(逗号分隔)"},
		{"workflow.collab.feishu_connection_mode", mode, "string", "飞书连接模式：webhook|websocket"},
	}
	for _, item := range configs {
		if err := saveMvpConfig(ctx, item.key, item.value, item.configType, "collab", item.description); err != nil {
			return nil, err
		}
	}

	// 联动 WebSocket 长连接
	appID := strings.TrimSpace(req.AppID)
	appSecret := strings.TrimSpace(req.AppSecret)
	encryptKey := strings.TrimSpace(req.EncryptKey)
	wsMgr := collab.GetWSManager()
	if req.Enabled == 1 && mode == "websocket" && appID != "" && appSecret != "" {
		wsMgr.StartWS(appID, appSecret, encryptKey, feishuWSEventHandler)
	} else {
		wsMgr.StopWS()
	}

	return &v1.WorkflowSaveFeishuConfigRes{}, nil
}

// FeishuBindings 查询飞书绑定列表。
func (c *cWorkflow) FeishuBindings(ctx context.Context, req *v1.WorkflowFeishuBindingsReq) (res *v1.WorkflowFeishuBindingsRes, err error) {
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	bindings, err := bindingRepo.List(ctx, "feishu")
	if err != nil {
		return nil, err
	}
	items := make([]v1.FeishuBindingDTO, 0, len(bindings))
	for _, item := range bindings {
		items = append(items, mapToFeishuBindingDTO(item))
	}
	return &v1.WorkflowFeishuBindingsRes{Bindings: items}, nil
}

// BindFeishuUser 绑定飞书用户。
func (c *cWorkflow) BindFeishuUser(ctx context.Context, req *v1.WorkflowBindFeishuUserReq) (res *v1.WorkflowBindFeishuUserRes, err error) {
	currentUserID := middleware.GetUserID(ctx)
	if currentUserID != 1 && int64(req.UserID) != currentUserID {
		return nil, fmt.Errorf("普通用户只能绑定自己的飞书账号")
	}
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	id, err := bindingRepo.Rebind(ctx, g.Map{
		"user_id":          int64(req.UserID),
		"platform":         "feishu",
		"platform_user_id": strings.TrimSpace(req.PlatformUserID),
		"platform_name":    strings.TrimSpace(req.PlatformName),
		"created_by":       currentUserID,
		"dept_id":          middleware.GetDeptID(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowBindFeishuUserRes{ID: snowflake.JsonInt64(id)}, nil
}

// UnbindFeishuUser 解绑飞书用户。
func (c *cWorkflow) UnbindFeishuUser(ctx context.Context, req *v1.WorkflowUnbindFeishuUserReq) (res *v1.WorkflowUnbindFeishuUserRes, err error) {
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	binding, err := bindingRepo.GetByIDScoped(ctx, int64(req.BindingID))
	if err != nil || binding == nil {
		return nil, fmt.Errorf("绑定记录不存在或无权操作")
	}
	if err := bindingRepo.UnbindByID(ctx, int64(req.BindingID)); err != nil {
		return nil, err
	}
	return &v1.WorkflowUnbindFeishuUserRes{}, nil
}

// TestFeishuMessage 发送飞书测试消息。
func (c *cWorkflow) TestFeishuMessage(ctx context.Context, req *v1.WorkflowTestFeishuMessageReq) (res *v1.WorkflowTestFeishuMessageRes, err error) {
	if engine.GetConfigInt(ctx, "workflow.collab.feishu_enabled", "workflow.collab.feishuEnabled", 0) != 1 {
		return nil, fmt.Errorf("飞书通知总开关未开启，请先保存并启用飞书配置")
	}
	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		bindingRepo = collabRepo.NewBindingRepo()
	}
	binding, err := bindingRepo.GetByIDScoped(ctx, int64(req.BindingID))
	if err != nil || binding == nil {
		return nil, fmt.Errorf("绑定记录不存在或无权操作")
	}

	text := strings.TrimSpace(req.Content)
	if text == "" {
		text = "EasyMVP 飞书联通测试成功。后续审批卡片和阶段报告会通过当前绑定发送。"
	}
	feishu := adapter.NewFeishuAdapter()
	if err := feishu.SendTextMessage(ctx, mapString(binding, "platform_user_id"), text); err != nil {
		return nil, err
	}
	return &v1.WorkflowTestFeishuMessageRes{}, nil
}

// ─── 飞书机器人菜单管理 ─────────────────────────────────────────────────────────

// defaultBotMenuItems 返回默认菜单结构（供前端展示和恢复用）。
func defaultBotMenuItems() []v1.BotMenuItem {
	return []v1.BotMenuItem{
		{
			EventKey: "PARENT_PROJECT",
			Name:     "项目管理",
			Children: []v1.BotMenuItem{
				{EventKey: "list_projects", Name: "我的项目"},
				{EventKey: "create_project_tip", Name: "创建项目"},
				{EventKey: "help", Name: "帮助"},
			},
		},
		{
			EventKey: "PARENT_TASK",
			Name:     "任务管理",
			Children: []v1.BotMenuItem{
				{EventKey: "project_status_tip", Name: "项目进度"},
				{EventKey: "list_tasks_tip", Name: "查看任务"},
				{EventKey: "retry_task_tip", Name: "重试失败任务"},
			},
		},
		{
			EventKey: "PARENT_REVIEW",
			Name:     "审核验收",
			Children: []v1.BotMenuItem{
				{EventKey: "review_status_tip", Name: "审核状态"},
				{EventKey: "accept_status_tip", Name: "验收状态"},
				{EventKey: "autonomy_status_tip", Name: "自治状态"},
			},
		},
	}
}

// GetBotMenu 查询当前飞书机器人菜单配置（从 DB 读取，未设置则返回默认）。
func (c *cWorkflow) GetBotMenu(ctx context.Context, req *v1.WorkflowGetBotMenuReq) (res *v1.WorkflowGetBotMenuRes, err error) {
	defaults := defaultBotMenuItems()

	// 从 mvp_config 读取自定义菜单
	var configVal string
	_ = g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", "workflow.collab.feishu_bot_menu").
		WhereNull("deleted_at").
		Fields("config_value").
		Scan(&configVal)

	if configVal == "" {
		return &v1.WorkflowGetBotMenuRes{
			MenuItems:    defaults,
			IsDefault:    true,
			DefaultItems: defaults,
		}, nil
	}

	var items []v1.BotMenuItem
	if err := json.Unmarshal([]byte(configVal), &items); err != nil || len(items) == 0 {
		return &v1.WorkflowGetBotMenuRes{
			MenuItems:    defaults,
			IsDefault:    true,
			DefaultItems: defaults,
		}, nil
	}

	return &v1.WorkflowGetBotMenuRes{
		MenuItems:    items,
		IsDefault:    false,
		DefaultItems: defaults,
	}, nil
}

// SetBotMenu 保存飞书机器人菜单配置到 DB。
// 注意：飞书机器人自定义菜单（单聊底部菜单）没有服务端 API，
// 需要在飞书开发者后台手动配置，此处仅保存 EventKey 映射供后端事件处理使用。
func (c *cWorkflow) SetBotMenu(ctx context.Context, req *v1.WorkflowSetBotMenuReq) (res *v1.WorkflowSetBotMenuRes, err error) {
	var items []v1.BotMenuItem
	if req.UseDefault || len(req.MenuItems) == 0 {
		items = defaultBotMenuItems()
		_ = saveMvpConfig(ctx, "workflow.collab.feishu_bot_menu", "", "string", "collab", "飞书机器人自定义菜单 JSON")
	} else {
		items = req.MenuItems
		menuJSON, _ := json.Marshal(items)
		_ = saveMvpConfig(ctx, "workflow.collab.feishu_bot_menu", string(menuJSON), "string", "collab", "飞书机器人自定义菜单 JSON")
	}
	_ = items

	return &v1.WorkflowSetBotMenuRes{
		Message: "菜单配置已保存。飞书机器人菜单需在「飞书开发者后台 → 应用功能 → 机器人 → 自定义菜单」中手动配置，将菜单响应动作设为「推送事件」并填写对应 EventKey 即可。",
	}, nil
}

// ─── 飞书群菜单 ───────────────────────────────────────────────────────────────

// defaultChatMenuItems 返回默认群菜单（快捷跳转后台页面）。
// baseURL 为后台访问地址（如 https://easymvp.example.com）。
func defaultChatMenuItems(baseURL string) []v1.ChatMenuItem {
	if baseURL == "" {
		baseURL = "https://easymvp.example.com"
	}
	return []v1.ChatMenuItem{
		{
			Name: "项目管理",
			Children: []v1.ChatMenuItem{
				{Name: "项目列表", URL: baseURL + "/mvp/project/index"},
				{Name: "新建项目", URL: baseURL + "/mvp/workflow/create"},
				{Name: "项目仪表盘", URL: baseURL + "/mvp/workflow/dashboard"},
			},
		},
		{
			Name: "任务管理",
			Children: []v1.ChatMenuItem{
				{Name: "任务列表", URL: baseURL + "/mvp/task/index"},
				{Name: "工作流状态", URL: baseURL + "/mvp/workflow/situation"},
			},
		},
		{
			Name: "系统设置",
			Children: []v1.ChatMenuItem{
				{Name: "飞书配置", URL: baseURL + "/mvp/workflow/feishu"},
				{Name: "AI 配置", URL: baseURL + "/ai/model/index"},
			},
		},
	}
}

// buildFeishuChatMenuBody 将 ChatMenuItem 转换为飞书 API 请求体结构。
func buildFeishuChatMenuBody(items []v1.ChatMenuItem) g.Map {
	topLevels := make([]g.Map, 0, len(items))
	for _, item := range items {
		menuItem := g.Map{
			"action_type": "NONE",
			"name":        item.Name,
			"i18n_names":  g.Map{"zh_cn": item.Name},
		}
		entry := g.Map{"chat_menu_item": menuItem}

		if len(item.Children) > 0 {
			children := make([]g.Map, 0, len(item.Children))
			for _, sub := range item.Children {
				children = append(children, g.Map{
					"chat_menu_item": g.Map{
						"action_type":   "REDIRECT_LINK",
						"name":          sub.Name,
						"i18n_names":    g.Map{"zh_cn": sub.Name},
						"redirect_link": g.Map{"common_url": sub.URL},
					},
				})
			}
			entry["children"] = children
		} else if item.URL != "" {
			menuItem["action_type"] = "REDIRECT_LINK"
			menuItem["redirect_link"] = g.Map{"common_url": item.URL}
		}

		topLevels = append(topLevels, entry)
	}
	return g.Map{
		"menu_tree": g.Map{
			"chat_menu_top_levels": topLevels,
		},
	}
}

// CreateChatMenu 在指定飞书群创建快捷跳转菜单。
func (c *cWorkflow) CreateChatMenu(ctx context.Context, req *v1.WorkflowCreateChatMenuReq) (res *v1.WorkflowCreateChatMenuRes, err error) {
	if engine.GetConfigInt(ctx, "workflow.collab.feishu_enabled", "workflow.collab.feishuEnabled", 0) != 1 {
		return nil, fmt.Errorf("飞书通知总开关未开启，请先启用飞书配置")
	}
	feishu := adapter.NewFeishuAdapter()
	token, err := feishu.GetTenantAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取飞书 token 失败: %w", err)
	}

	items := req.MenuItems
	if len(items) == 0 {
		baseURL := engine.GetConfigString(ctx, "workflow.collab.base_url", "", "")
		items = defaultChatMenuItems(baseURL)
	}

	body := buildFeishuChatMenuBody(items)
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/chats/%s/menu_tree", req.ChatID)
	resp, err := g.Client().
		SetHeaderMap(map[string]string{
			"Authorization": "Bearer " + token,
			"Content-Type":  "application/json; charset=utf-8",
		}).
		Post(ctx, url, body)
	if err != nil {
		return nil, fmt.Errorf("调用飞书群菜单 API 失败: %w", err)
	}
	defer resp.Close()
	respBody := resp.ReadAllString()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("飞书 API 返回 %d: %s", resp.StatusCode, respBody)
	}
	g.Log().Infof(ctx, "[FeishuChatMenu] 创建群菜单成功 chatID=%s: %s", req.ChatID, respBody)

	return &v1.WorkflowCreateChatMenuRes{
		Message: fmt.Sprintf("群菜单创建成功，用户在群 %s 中可看到快捷菜单", req.ChatID),
	}, nil
}

// GetChatMenu 获取指定飞书群的菜单。
func (c *cWorkflow) GetChatMenu(ctx context.Context, req *v1.WorkflowGetChatMenuReq) (res *v1.WorkflowGetChatMenuRes, err error) {
	feishu := adapter.NewFeishuAdapter()
	token, err := feishu.GetTenantAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取飞书 token 失败: %w", err)
	}

	url := fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/chats/%s/menu_tree", req.ChatID)
	resp, err := g.Client().
		SetHeaderMap(map[string]string{
			"Authorization": "Bearer " + token,
		}).
		Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("获取群菜单失败: %w", err)
	}
	defer resp.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			MenuTree struct {
				ChatMenuTopLevels []g.Map `json:"chat_menu_top_levels"`
			} `json:"menu_tree"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.ReadAll(), &result); err != nil {
		return nil, fmt.Errorf("解析群菜单响应失败: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("飞书 API 错误: code=%d msg=%s", result.Code, result.Msg)
	}

	return &v1.WorkflowGetChatMenuRes{
		MenuItems: result.Data.MenuTree.ChatMenuTopLevels,
	}, nil
}

// DeleteChatMenu 删除指定飞书群的菜单项。
func (c *cWorkflow) DeleteChatMenu(ctx context.Context, req *v1.WorkflowDeleteChatMenuReq) (res *v1.WorkflowDeleteChatMenuRes, err error) {
	feishu := adapter.NewFeishuAdapter()
	token, err := feishu.GetTenantAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取飞书 token 失败: %w", err)
	}

	url := fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/chats/%s/menu_tree", req.ChatID)
	resp, err := g.Client().
		SetHeaderMap(map[string]string{
			"Authorization": "Bearer " + token,
			"Content-Type":  "application/json; charset=utf-8",
		}).
		Delete(ctx, url, g.Map{"chat_menu_top_level_ids": req.MenuIDs})
	if err != nil {
		return nil, fmt.Errorf("删除群菜单失败: %w", err)
	}
	defer resp.Close()
	respBody := resp.ReadAllString()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("飞书 API 返回 %d: %s", resp.StatusCode, respBody)
	}

	return &v1.WorkflowDeleteChatMenuRes{}, nil
}

func saveMvpConfig(ctx context.Context, key, value, configType, category, description string) error {
	now := gtime.Now()
	userID := middleware.GetUserID(ctx)
	deptID := middleware.GetDeptID(ctx)

	count, err := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		_, err = g.DB().Model("mvp_config").Ctx(ctx).
			Where("config_key", key).
			WhereNull("deleted_at").
			Update(g.Map{
				"config_value": value,
				"config_type":  configType,
				"category":     category,
				"description":  description,
				"updated_at":   now,
			})
		return err
	}
	_, err = g.DB().Model("mvp_config").Ctx(ctx).Insert(g.Map{
		"id":           snowflake.Generate(),
		"config_key":   key,
		"config_value": value,
		"config_type":  configType,
		"category":     category,
		"description":  description,
		"created_by":   userID,
		"dept_id":      deptID,
		"created_at":   now,
		"updated_at":   now,
	})
	return err
}

func mapToFeishuBindingDTO(m g.Map) v1.FeishuBindingDTO {
	return v1.FeishuBindingDTO{
		ID:             mapJsonInt64(m, "id"),
		UserID:         mapJsonInt64(m, "user_id"),
		Platform:       mapString(m, "platform"),
		PlatformUserID: mapString(m, "platform_user_id"),
		PlatformName:   mapString(m, "platform_name"),
		CreatedBy:      mapJsonInt64(m, "created_by"),
		DeptID:         mapJsonInt64(m, "dept_id"),
		CreatedAt:      mapGTime(m, "created_at"),
		UpdatedAt:      mapGTime(m, "updated_at"),
	}
}

// botMenuKeyToCommand 将飞书菜单 event_key 转换为 Bot 指令文本。
func botMenuKeyToCommand(key string) string {
	switch key {
	case "list_projects":
		return "我的项目列表"
	case "create_project_tip":
		return "我想创建一个新项目，请告诉我需要提供哪些信息？"
	case "help":
		return "help"
	case "project_status_tip":
		return "请告诉我某个项目的进度，你需要知道项目名称"
	case "list_tasks_tip":
		return "查看任务列表，你需要知道项目名称"
	case "retry_task_tip":
		return "重试失败任务，你需要知道项目名称"
	case "review_status_tip":
		return "查看审核状态，你需要知道项目名称"
	case "accept_status_tip":
		return "查看验收状态，你需要知道项目名称"
	case "autonomy_status_tip":
		return "查看自治状态，你需要知道项目名称"
	default:
		return key
	}
}

func boolToInt(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

// FeishuWSEventHandlerExport 导出的 WS 事件处理器，供启动时自动恢复连接使用。
var FeishuWSEventHandlerExport = feishuWSEventHandler

// feishuWSEventHandler WebSocket 模式的事件处理器（与 Webhook 模式共用逻辑）。
func feishuWSEventHandler(ctx context.Context, header map[string]interface{}, event map[string]interface{}) {
	if header == nil || event == nil {
		return
	}
	eventType, _ := header["event_type"].(string)

	switch eventType {
	case "im.message.receive_v1":
		sender, _ := event["sender"].(map[string]interface{})
		messageMap, _ := event["message"].(map[string]interface{})
		if sender == nil || messageMap == nil {
			return
		}
		senderID, _ := sender["sender_id"].(map[string]interface{})
		openID := ""
		if senderID != nil {
			openID, _ = senderID["open_id"].(string)
		}
		messageID, _ := messageMap["message_id"].(string)
		chatID, _ := messageMap["chat_id"].(string)
		msgType, _ := messageMap["message_type"].(string)
		contentStr, _ := messageMap["content"].(string)
		g.Log().Infof(ctx, "[FeishuWSBot] 收到消息: openID=%s messageID=%s type=%s", openID, messageID, msgType)

		// 非文本消息：直接回复提示
		if msgType != "" && msgType != "text" {
			platform := &feishuBotPlatform{messageID: messageID, chatID: chatID}
			switch msgType {
			case "audio":
				platform.Reply(ctx, "收到语音消息，暂不支持语音识别。请直接发文字，我会立即处理 😊")
			case "image":
				platform.Reply(ctx, "收到图片消息，暂不支持图片识别。请用文字描述你的需求 📝")
			case "file":
				platform.Reply(ctx, "收到文件消息，暂不支持文件处理。请用文字说明你的需求 📄")
			case "sticker":
				platform.Reply(ctx, "收到表情包 😄 如果你想操作项目，请发送文字指令。发送「帮助」可查看所有功能。")
			case "video":
				platform.Reply(ctx, "收到视频消息，暂不支持视频处理。请用文字描述你的需求 🎬")
			default:
				platform.Reply(ctx, fmt.Sprintf("收到 %s 类型消息，暂不支持此类型。请发送文字消息 🚀", msgType))
			}
			return
		}
		DispatchFeishuCommand(ctx, openID, messageID, chatID, contentStr)

	case "application.bot.menu_v6":
		// 机器人菜单点击事件
		operatorMap, _ := event["operator"].(map[string]interface{})
		openID := ""
		if operatorMap != nil {
			operatorID, _ := operatorMap["operator_id"].(map[string]interface{})
			if operatorID != nil {
				openID, _ = operatorID["open_id"].(string)
			}
		}
		eventKey, _ := event["event_key"].(string)
		g.Log().Infof(ctx, "[FeishuMenuClick] openID=%s key=%s", openID, eventKey)
		if openID != "" && eventKey != "" {
			cmdText := botMenuKeyToCommand(eventKey)
			DispatchFeishuCommand(ctx, openID, "", "", `{"text":"`+cmdText+`"}`)
		}
	}
}
