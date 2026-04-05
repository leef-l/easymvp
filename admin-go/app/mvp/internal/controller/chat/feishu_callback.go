package chat

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/orchestrator"
)

// FeishuCallback 飞书回调网关控制器（不走 JWT 认证，走飞书签名校验）。
var FeishuCallback = cFeishuCallback{}

type cFeishuCallback struct{}

// Handle 处理飞书交互卡片回调。
func (c *cFeishuCallback) Handle(r *ghttp.Request) {
	ctx := r.GetCtx()
	body := r.GetBody()

	// 1. 签名验证
	timestamp := r.GetHeader("X-Lark-Request-Timestamp")
	nonce := r.GetHeader("X-Lark-Request-Nonce")
	signature := r.GetHeader("X-Lark-Signature")

	encryptKey := engine.GetConfigString(ctx,
		"workflow.collab.feishu_encrypt_key",
		"workflow.collab.feishuEncryptKey", "")

	if encryptKey != "" && signature != "" {
		expected := sha256Hex(timestamp + nonce + encryptKey + string(body))
		if expected != signature {
			g.Log().Warningf(ctx, "[FeishuCallback] 签名校验失败: expected=%s got=%s", expected, signature)
			r.Response.WriteStatus(403)
			r.Response.WriteJson(g.Map{"msg": "signature mismatch"})
			return
		}
	}

	// 2. 解析事件类型
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		r.Response.WriteStatus(400)
		r.Response.WriteJson(g.Map{"msg": "invalid json"})
		return
	}

	// 2a. URL 验证请求
	if challenge, ok := raw["challenge"]; ok {
		r.Response.WriteJson(g.Map{"challenge": challenge})
		return
	}

	// 2b. 消息事件（im.message.receive_v1）
	if header, ok := raw["header"].(map[string]interface{}); ok {
		eventType, _ := header["event_type"].(string)
		if eventType == "im.message.receive_v1" {
			if event, ok := raw["event"].(map[string]interface{}); ok {
				handleFeishuMessageEvent(r, event)
			} else {
				r.Response.WriteJson(g.Map{"msg": "ok"})
			}
			return
		}
	}

	// 2c. 交互卡片回调 (v2 schema: action 在顶层)
	actionMap, _ := raw["action"].(map[string]interface{})
	if actionMap == nil {
		r.Response.WriteJson(g.Map{"msg": "ok"})
		return
	}

	valueMap, _ := actionMap["value"].(map[string]interface{})
	if valueMap == nil {
		r.Response.WriteJson(g.Map{"msg": "ok"})
		return
	}

	action := fmt.Sprintf("%v", valueMap["action"])
	actionIDStr := fmt.Sprintf("%v", valueMap["action_id"])
	actionID, _ := strconv.ParseInt(actionIDStr, 10, 64)

	if actionID == 0 || (action != "approve" && action != "reject") {
		g.Log().Warningf(ctx, "[FeishuCallback] 无效的回调参数: action=%s actionID=%s", action, actionIDStr)
		r.Response.WriteJson(g.Map{"msg": "invalid action"})
		return
	}

	// 3. 用户反查
	openID := ""
	if operatorMap, ok := raw["operator"].(map[string]interface{}); ok {
		openID = fmt.Sprintf("%v", operatorMap["open_id"])
	}

	bindingRepo := orchestrator.GetCollabBindingRepo()
	if bindingRepo == nil {
		g.Log().Warning(ctx, "[FeishuCallback] 协作绑定仓储未初始化")
		r.Response.WriteStatus(500)
		r.Response.WriteJson(g.Map{"msg": "service not ready"})
		return
	}

	var systemUserID int64
	if openID != "" {
		binding, _ := bindingRepo.GetByPlatformUserID(ctx, "feishu", openID)
		if binding != nil {
			systemUserID = toCallbackInt64(binding["user_id"])
		}
	}

	if systemUserID == 0 {
		g.Log().Warningf(ctx, "[FeishuCallback] 未找到绑定用户: openID=%s", openID)
		r.Response.WriteJson(g.Map{"msg": "user not bound"})
		return
	}

	// 4. 执行审批
	dc := orchestrator.GetDecisionCenter()
	if dc == nil {
		r.Response.WriteStatus(500)
		r.Response.WriteJson(g.Map{"msg": "decision center not ready"})
		return
	}

	g.Log().Infof(ctx, "[FeishuCallback] 收到审批: action=%s actionID=%d userID=%d openID=%s",
		action, actionID, systemUserID, openID)

	var execErr error
	switch action {
	case "approve":
		execErr = dc.ApproveAction(ctx, actionID)
	case "reject":
		execErr = dc.RejectAction(ctx, actionID, "飞书驳回")
	}

	if execErr != nil {
		g.Log().Warningf(ctx, "[FeishuCallback] 审批执行失败: actionID=%d err=%v", actionID, execErr)
		r.Response.WriteJson(g.Map{"msg": execErr.Error()})
		return
	}

	r.Response.WriteJson(g.Map{"msg": "ok"})
}

// handleFeishuMessageEvent 将消息事件路由到 Bot 指令处理器。
func handleFeishuMessageEvent(r *ghttp.Request, event map[string]interface{}) {
	ctx := r.GetCtx()
	sender, _ := event["sender"].(map[string]interface{})
	messageMap, _ := event["message"].(map[string]interface{})

	if sender == nil || messageMap == nil {
		r.Response.WriteJson(g.Map{"msg": "ok"})
		return
	}

	senderID, _ := sender["sender_id"].(map[string]interface{})
	openID := ""
	if senderID != nil {
		openID, _ = senderID["open_id"].(string)
	}

	messageID, _ := messageMap["message_id"].(string)
	chatID, _ := messageMap["chat_id"].(string)
	contentStr, _ := messageMap["content"].(string)

	g.Log().Infof(ctx, "[FeishuBot] 收到消息: openID=%s messageID=%s chatID=%s", openID, messageID, chatID)

	DispatchFeishuCommand(ctx, openID, messageID, chatID, contentStr)
	r.Response.WriteJson(g.Map{"msg": "ok"})
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

func toCallbackInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	case int:
		return int64(n)
	default:
		i, _ := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		return i
	}
}
