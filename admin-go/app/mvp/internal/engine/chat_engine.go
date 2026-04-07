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
	ModelID      int64
	ModelCode    string
	ProviderType string
	BaseURL      string
	APIKey       string
	APISecret    string
	SystemPrompt string
	MaxTokens    int
}

// SendMessage 发送用户消息并触发 AI 回复
// 返回用户消息 ID 和 AI 回复消息 ID
func (e *ChatEngine) SendMessage(ctx context.Context, conversationID int64, content string, userID int64, deptID int64) (msgID int64, replyID int64, err error) {
	// 1. 查询对话信息
	conv, err := g.DB().Model("mvp_conversation").Where("id", conversationID).Where("deleted_at IS NULL").One()
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
		Where("updated_at < ?", gtime.Now().Add(-5*time.Minute)). // 超过5分钟的 streaming 视为卡住
		Update(g.Map{"status": "failed", "content": "AI 调用中断，消息未完成", "updated_at": gtime.Now()}); cleanErr != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 清理卡住的 streaming 消息失败: conv=%d err=%v", conversationID, cleanErr)
	}

	// 2. 查找该对话角色对应的 AI 模型配置
	modelInfo, err := e.resolveModel(ctx, projectID, conv["role_type"].String())
	if err != nil {
		return 0, 0, err
	}

	// 3. 展开消息中的 "读取：路径" 指令（限制在项目工作目录内）
	project, _ := g.DB().Model("mvp_project").Where("id", projectID).Fields("work_dir").One()
	workDir := project["work_dir"].String()
	if workDir == "" {
		workDir = "/www/wwwroot/project/easymvp"
	}
	expandedContent := ExpandFileReads(content, workDir)

	// 4. 保存用户消息
	msgID = int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_message").Insert(g.Map{
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
	})
	if err != nil {
		return 0, 0, fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 5. 创建 AI 回复消息（status=streaming）
	replyID = int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_message").Insert(g.Map{
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
	})
	if err != nil {
		return 0, 0, fmt.Errorf("创建AI回复消息失败: %w", err)
	}

	// 6. 启动 goroutine 异步调用 AI（不依赖当前 HTTP 请求的 context）
	go e.runAICall(conversationID, replyID, modelInfo)

	return msgID, replyID, nil
}

// SendFeishuMessage 供飞书 Bot 调用：发送用户消息到对话并触发 AI 回复。
// 返回 AI 回复消息的 ID（用于轮询结果）。
func (e *ChatEngine) SendFeishuMessage(ctx context.Context, conversationID, projectID int64, content string, userID, deptID int64) (int64, error) {
	// 1. 查询对话信息（校验存在）
	conv, err := g.DB().Model("mvp_conversation").Where("id", conversationID).Where("deleted_at IS NULL").One()
	if err != nil {
		return 0, fmt.Errorf("查询对话失败: %w", err)
	}
	if conv.IsEmpty() {
		return 0, fmt.Errorf("对话不存在")
	}

	// 2. 查找该对话角色对应的 AI 模型配置
	modelInfo, err := e.resolveModel(ctx, projectID, conv["role_type"].String())
	if err != nil {
		return 0, err
	}

	// 3. 展开消息中的 "读取：路径" 指令
	project, _ := g.DB().Model("mvp_project").Where("id", projectID).Fields("work_dir").One()
	workDir := project["work_dir"].String()
	if workDir == "" {
		workDir = "/www/wwwroot/project/easymvp"
	}
	expandedContent := ExpandFileReads(content, workDir)

	// 4. 保存用户消息
	msgID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_message").Insert(g.Map{
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
	})
	if err != nil {
		return 0, fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 5. 创建 AI 回复消息（status=streaming）
	replyID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_message").Insert(g.Map{
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
	})
	if err != nil {
		return 0, fmt.Errorf("创建AI回复消息失败: %w", err)
	}

	// 6. 启动 goroutine 异步调用 AI
	go e.runAICall(conversationID, replyID, modelInfo)

	return replyID, nil
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
		Fields("m.model_code, m.max_tokens, pv.provider_type, pv.base_url, p.api_key, p.api_secret").
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
		ModelID:      modelID,
		ModelCode:    model["model_code"].String(),
		ProviderType: model["provider_type"].String(),
		BaseURL:      model["base_url"].String(),
		APIKey:       model["api_key"].String(),
		APISecret:    model["api_secret"].String(),
		SystemPrompt: systemPrompt,
		MaxTokens:    model["max_tokens"].Int(),
	}, nil
}

// failMessage 标记消息为失败状态
func (e *ChatEngine) failMessage(ctx context.Context, replyID int64, errMsg string) {
	g.Log().Errorf(ctx, "AI 调用失败 (messageID=%d): %s", replyID, errMsg)

	g.DB().Model("mvp_message").Where("id", replyID).Update(g.Map{
		"message_type": mvpmodel.MessageTypePoison,
		"status":       "failed",
		"content":      "AI 调用失败: " + errMsg,
		"updated_at":   gtime.Now(),
	})

	// 通知前端失败
	errJSON, _ := json.Marshal(map[string]interface{}{
		"error": errMsg,
		"done":  true,
	})
	e.hub.Publish(replyID, string(errJSON))
	e.hub.Done(replyID)
}
