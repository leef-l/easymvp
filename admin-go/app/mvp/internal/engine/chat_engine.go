package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	mvpmodel "easymvp/app/mvp/internal/model"
	"easymvp/utility/snowflake"
)

// ChatEngine 对话引擎，后端驱动 AI 调用
type ChatEngine struct {
	hub *SSEHub
}

// NewChatEngine 创建对话引擎
func NewChatEngine() *ChatEngine {
	return &ChatEngine{
		hub: GetHub(),
	}
}

// 全局引擎实例
var defaultEngine = NewChatEngine()

// GetEngine 获取全局引擎实例
func GetEngine() *ChatEngine {
	return defaultEngine
}

// ModelInfo 模型信息
type ModelInfo struct {
	ModelID            int64
	ModelCode          string
	ProviderType       string
	SupportedProtocols []string
	BaseURL            string
	APIKey             string
	APISecret          string
	SystemPrompt       string
	MaxTokens          int
}

// SendMessage 发送用户消息并触发 AI 回复
// 返回用户消息 ID 和 AI 回复消息 ID
func (e *ChatEngine) SendMessage(ctx context.Context, conversationID int64, content string, userID int64, deptID int64) (msgID int64, replyID int64, err error) {
	conv, err := g.DB().Model("mvp_conversation").Ctx(ctx).Where("id", conversationID).WhereNull("deleted_at").One()
	if err != nil {
		return 0, 0, fmt.Errorf("查询对话失败: %w", err)
	}
	if conv.IsEmpty() {
		return 0, 0, fmt.Errorf("对话不存在")
	}

	projectID := conv["project_id"].Int64()

	// 清理卡住的 streaming 消息（AI 调用中断导致）
	if _, cleanErr := g.DB().Model("mvp_message").Ctx(ctx).
		Where("conversation_id", conversationID).
		Where("status", "streaming").
		Where("updated_at < ?", gtime.Now().Add(-5*time.Minute)).
		Update(g.Map{"status": "failed", "content": "AI 调用中断，消息未完成", "updated_at": gtime.Now()}); cleanErr != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 清理卡住的 streaming 消息失败: conv=%d err=%v", conversationID, cleanErr)
	}

	msgID, replyID, modelInfo, err := e.prepareAndSave(ctx, conversationID, projectID, conv["role_type"].String(), content, userID, deptID)
	if err != nil {
		return 0, 0, err
	}

	go e.runAICall(conversationID, replyID, modelInfo)
	return msgID, replyID, nil
}

// SendFeishuMessage 供飞书 Bot 调用：发送用户消息到对话并触发 AI 回复。
// 返回 AI 回复消息的 ID（用于轮询结果）。
func (e *ChatEngine) SendFeishuMessage(ctx context.Context, conversationID, projectID int64, content string, userID, deptID int64) (int64, error) {
	conv, err := g.DB().Model("mvp_conversation").Ctx(ctx).Where("id", conversationID).WhereNull("deleted_at").One()
	if err != nil {
		return 0, fmt.Errorf("查询对话失败: %w", err)
	}
	if conv.IsEmpty() {
		return 0, fmt.Errorf("对话不存在")
	}

	_, replyID, modelInfo, err := e.prepareAndSave(ctx, conversationID, projectID, conv["role_type"].String(), content, userID, deptID)
	if err != nil {
		return 0, err
	}

	go e.runAICall(conversationID, replyID, modelInfo)
	return replyID, nil
}

// prepareAndSave 公共流程：解析模型 → 展开指令 → 保存用户消息 → 创建 AI 回复占位。
func (e *ChatEngine) prepareAndSave(ctx context.Context, conversationID, projectID int64, roleType, content string, userID, deptID int64) (msgID, replyID int64, modelInfo *ModelInfo, err error) {
	modelInfo, err = e.resolveModel(ctx, projectID, roleType)
	if err != nil {
		return 0, 0, nil, err
	}

	workDir := GetProjectWorkDir(ctx, projectID)
	expandedContent := ExpandFileReads(content, workDir)

	msgID = int64(snowflake.Generate())
	if _, err = g.DB().Model("mvp_message").Insert(g.Map{
		"id":              msgID,
		"conversation_id": conversationID,
		"role":            "user",
		"message_type":    mvpmodel.MessageTypeChatUser,
		"content":         expandedContent,
		"status":          "completed",
		"created_by":      userID,
		"dept_id":         deptID,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	}); err != nil {
		return 0, 0, nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	replyID = int64(snowflake.Generate())
	if _, err = g.DB().Model("mvp_message").Insert(g.Map{
		"id":              replyID,
		"conversation_id": conversationID,
		"role":            "assistant",
		"message_type":    mvpmodel.MessageTypeChatReply,
		"content":         "",
		"model_id":        modelInfo.ModelID,
		"status":          "streaming",
		"created_by":      userID,
		"dept_id":         deptID,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	}); err != nil {
		return 0, 0, nil, fmt.Errorf("创建AI回复消息失败: %w", err)
	}

	return msgID, replyID, modelInfo, nil
}

// resolveModel 根据项目 ID 和角色类型查找对应的 AI 模型配置
// 如果项目未配置该角色，自动从默认预设创建。
func (e *ChatEngine) resolveModel(ctx context.Context, projectID int64, roleType string) (*ModelInfo, error) {
	role, err := ResolveProjectRole(ctx, projectID, roleType)
	if err != nil {
		return nil, err
	}

	modelID := role["model_id"].Int64()
	systemPrompt := role["system_prompt"].String()

	// 查询模型详情（关联 plan 和 provider）
	model, err := g.DB().Model("ai_model m").
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.model_code, m.max_tokens, pv.provider_type, pv.supported_protocols, pv.base_url, p.api_key, p.api_secret").
		Where("m.id", modelID).
		Where("m.deleted_at IS NULL").
		One()
	if err != nil {
		return nil, fmt.Errorf("查询模型信息失败: %w", err)
	}
	if model.IsEmpty() {
		return nil, fmt.Errorf("AI 模型 %d 不存在", modelID)
	}

	return &ModelInfo{
		ModelID:            modelID,
		ModelCode:          model["model_code"].String(),
		ProviderType:       model["provider_type"].String(),
		SupportedProtocols: decodeProviderProtocols(model["supported_protocols"].String(), model["provider_type"].String()),
		BaseURL:            model["base_url"].String(),
		APIKey:             model["api_key"].String(),
		APISecret:          model["api_secret"].String(),
		SystemPrompt:       systemPrompt,
		MaxTokens:          model["max_tokens"].Int(),
	}, nil
}

// failMessage 标记消息为失败状态
func (e *ChatEngine) failMessage(ctx context.Context, replyID int64, errMsg string) {
	g.Log().Errorf(ctx, "AI 调用失败 (messageID=%d): %s", replyID, errMsg)

	if _, err := g.DB().Model("mvp_message").Ctx(ctx).Where("id", replyID).Update(g.Map{
		"message_type": mvpmodel.MessageTypePoison,
		"status":       "failed",
		"content":      "AI 调用失败: " + errMsg,
		"updated_at":   gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[ChatEngine] 标记消息失败状态写入DB失败: msgID=%d err=%v", replyID, err)
	}

	// 通知前端失败
	errJSON, mErr := json.Marshal(map[string]interface{}{
		"error": errMsg,
		"done":  true,
	})
	if mErr != nil {
		errJSON = []byte(`{"error":"internal error","done":true}`)
	}
	e.hub.Publish(replyID, string(errJSON))
	e.hub.Done(replyID)
}
