package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/collab"
	"easymvp/app/mvp/internal/engine"
)

// FeishuAdapter 飞书消息适配层，通过 HTTP 调用飞书开放平台 API。
type FeishuAdapter struct {
	mu          sync.Mutex
	tokenCache  string
	tokenExpire time.Time
}

// NewFeishuAdapter 创建飞书适配器。
func NewFeishuAdapter() *FeishuAdapter {
	return &FeishuAdapter{}
}

func (a *FeishuAdapter) GetPlatform() collab.Platform {
	return collab.PlatformFeishu
}

func (a *FeishuAdapter) IsEnabled(ctx context.Context) bool {
	return engine.GetConfigInt(ctx, "workflow.collab.feishu_enabled", "workflow.collab.feishuEnabled", 0) == 1
}

// SendTextMessage 发送纯文本消息。
func (a *FeishuAdapter) SendTextMessage(ctx context.Context, openID string, text string) error {
	if !a.IsEnabled(ctx) {
		return nil
	}
	token, err := a.GetTenantAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取飞书 token 失败: %w", err)
	}

	contentJSON, _ := json.Marshal(map[string]string{"text": text})
	body := g.Map{
		"receive_id": openID,
		"msg_type":   "text",
		"content":    string(contentJSON),
	}
	resp, err := g.Client().
		SetHeaderMap(map[string]string{
			"Authorization": "Bearer " + token,
			"Content-Type":  "application/json; charset=utf-8",
		}).
		Post(ctx, "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=open_id", body)
	if err != nil {
		return fmt.Errorf("飞书发送文本消息失败: %w", err)
	}
	defer resp.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("飞书 API 返回 %d: %s", resp.StatusCode, resp.ReadAllString())
	}
	return nil
}

// SendCardMessage 发送交互卡片消息。
func (a *FeishuAdapter) SendCardMessage(ctx context.Context, openID string, card *collab.InteractiveCard) error {
	if !a.IsEnabled(ctx) {
		return nil
	}
	token, err := a.GetTenantAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取飞书 token 失败: %w", err)
	}

	cardJSON := a.buildMessageCard(card)
	cardBytes, _ := json.Marshal(cardJSON)

	body := g.Map{
		"receive_id": openID,
		"msg_type":   "interactive",
		"content":    string(cardBytes),
	}
	resp, err := g.Client().
		SetHeaderMap(map[string]string{
			"Authorization": "Bearer " + token,
			"Content-Type":  "application/json; charset=utf-8",
		}).
		Post(ctx, "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=open_id", body)
	if err != nil {
		return fmt.Errorf("飞书发送卡片消息失败: %w", err)
	}
	defer resp.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("飞书 API 返回 %d: %s", resp.StatusCode, resp.ReadAllString())
	}
	return nil
}

// ReplyMessage 回复一条飞书消息（reply to message_id）。
func (a *FeishuAdapter) ReplyMessage(ctx context.Context, messageID string, text string) error {
	token, err := a.GetTenantAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取飞书 token 失败: %w", err)
	}
	body := g.Map{
		"msg_type": "text",
		"content":  fmt.Sprintf(`{"text":%q}`, text),
	}
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/messages/%s/reply", messageID)
	resp, err := g.Client().
		SetHeaderMap(map[string]string{
			"Authorization": "Bearer " + token,
			"Content-Type":  "application/json; charset=utf-8",
		}).
		Post(ctx, url, body)
	if err != nil {
		return fmt.Errorf("飞书回复消息失败: %w", err)
	}
	defer resp.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("飞书 API 返回 %d: %s", resp.StatusCode, resp.ReadAllString())
	}
	return nil
}

// SendTextToChat 向群/会话发送文本消息（receive_id 为 chat_id）。
func (a *FeishuAdapter) SendTextToChat(ctx context.Context, chatID string, text string) error {
	token, err := a.GetTenantAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取飞书 token 失败: %w", err)
	}
	body := g.Map{
		"receive_id": chatID,
		"msg_type":   "text",
		"content":    fmt.Sprintf(`{"text":%q}`, text),
	}
	resp, err := g.Client().
		SetHeaderMap(map[string]string{
			"Authorization": "Bearer " + token,
			"Content-Type":  "application/json; charset=utf-8",
		}).
		Post(ctx, "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=chat_id", body)
	if err != nil {
		return fmt.Errorf("飞书发送群消息失败: %w", err)
	}
	defer resp.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("飞书 API 返回 %d: %s", resp.StatusCode, resp.ReadAllString())
	}
	return nil
}

// GetTenantAccessToken 获取 tenant_access_token，内存缓存 110 分钟。
func (a *FeishuAdapter) GetTenantAccessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.tokenCache != "" && time.Now().Before(a.tokenExpire) {
		return a.tokenCache, nil
	}

	appID := engine.GetConfigString(ctx, "workflow.collab.feishu_app_id", "workflow.collab.feishuAppId", "")
	appSecret := engine.GetConfigString(ctx, "workflow.collab.feishu_app_secret", "workflow.collab.feishuAppSecret", "")
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("飞书 app_id 或 app_secret 未配置")
	}

	resp, err := g.Client().Post(ctx,
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal",
		g.Map{"app_id": appID, "app_secret": appSecret})
	if err != nil {
		return "", err
	}
	defer resp.Close()

	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	if err := json.Unmarshal(resp.ReadAll(), &result); err != nil {
		return "", fmt.Errorf("解析飞书 token 响应失败: %w", err)
	}
	if result.Code != 0 {
		return "", fmt.Errorf("飞书 token 获取失败: code=%d msg=%s", result.Code, result.Msg)
	}

	a.tokenCache = result.TenantAccessToken
	a.tokenExpire = time.Now().Add(110 * time.Minute)
	return a.tokenCache, nil
}

// buildMessageCard 将 InteractiveCard 转换为飞书 MessageCard JSON 结构。
func (a *FeishuAdapter) buildMessageCard(card *collab.InteractiveCard) g.Map {
	headerColor := card.HeaderColor
	if headerColor == "" {
		headerColor = "orange"
	}

	// 构建内容区
	content := fmt.Sprintf("**项目**: %s\n**等级**: %s\n**动作**: %s\n**触发**: %s",
		card.ProjectName, card.Level, card.ActionType, card.TriggerSource)
	if card.Details != "" {
		content += fmt.Sprintf("\n**详情**: %s", card.Details)
	}

	elements := []g.Map{
		{
			"tag":  "div",
			"text": g.Map{"tag": "lark_md", "content": content},
		},
	}

	// 构建按钮区
	if len(card.Buttons) > 0 {
		actions := make([]g.Map, 0, len(card.Buttons))
		for _, btn := range card.Buttons {
			actions = append(actions, g.Map{
				"tag":  "button",
				"text": g.Map{"tag": "plain_text", "content": btn.Label},
				"type": btn.Style,
				"value": g.Map{
					"action":    btn.Action,
					"action_id": fmt.Sprintf("%d", card.ActionID),
				},
			})
		}
		elements = append(elements, g.Map{
			"tag":     "action",
			"actions": actions,
		})
	}

	return g.Map{
		"config": g.Map{"wide_screen_mode": true},
		"header": g.Map{
			"title":    g.Map{"tag": "plain_text", "content": card.Title},
			"template": headerColor,
		},
		"elements": elements,
	}
}
