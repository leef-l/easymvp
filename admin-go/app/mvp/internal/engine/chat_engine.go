package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/activity"
	mvpmodel "easymvp/app/mvp/internal/model"
	"easymvp/utility/provider"
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

// resolveModel 根据项目 ID 和角色类型查找对应的 AI 模型配置
func (e *ChatEngine) resolveModel(ctx context.Context, projectID int64, roleType string) (*ModelInfo, error) {
	// 从 mvp_project_role 查找角色配置
	role, err := g.DB().Model("mvp_project_role").
		Where("project_id", projectID).
		Where("role_type", roleType).
		Where("deleted_at IS NULL").
		Where("status", 1).
		One()
	if err != nil {
		return nil, fmt.Errorf("查询角色配置失败: %w", err)
	}
	if role.IsEmpty() {
		return nil, fmt.Errorf("项目 %d 未配置 %s 角色的 AI 模型", projectID, roleType)
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

// runAICall 异步调用 AI（goroutine 中运行，不依赖前端连接）
func (e *ChatEngine) runAICall(conversationID int64, replyID int64, modelInfo *ModelInfo) {
	ctx := context.Background()

	// 1. 获取对话历史
	messages, err := e.loadHistory(ctx, conversationID, replyID)
	if err != nil {
		e.failMessage(ctx, replyID, err.Error())
		return
	}

	// 2. 创建 Provider
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
	})
	if err != nil {
		e.failMessage(ctx, replyID, err.Error())
		return
	}

	// 3. 构建请求
	req := &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     messages,
		MaxTokens:    modelInfo.MaxTokens,
		Temperature:  0.7,
		Stream:       true,
		SystemPrompt: modelInfo.SystemPrompt,
	}

	// 4. 流式调用 AI
	var fullContent strings.Builder
	chunkIndex := 0

	err = p.ChatStream(ctx, req, func(chunk *provider.StreamChunk) error {
		if chunk.Content != "" {
			fullContent.WriteString(chunk.Content)
			chunkIndex++

			// 写入 message_chunk 表
			_, insertErr := g.DB().Model("mvp_message_chunk").Insert(g.Map{
				"message_id":  replyID,
				"chunk_index": chunkIndex,
				"content":     chunk.Content,
				"created_at":  gtime.Now(),
			})
			if insertErr != nil {
				g.Log().Errorf(ctx, "写入 chunk 失败: %v", insertErr)
			}
			activity.TouchMessageActivity(ctx, replyID)
			activity.TouchConversationActivity(ctx, conversationID)

			// 推送到 SSE Hub
			chunkJSON, _ := json.Marshal(map[string]interface{}{
				"content": chunk.Content,
				"index":   chunkIndex,
			})
			e.hub.Publish(replyID, string(chunkJSON))
		}

		// 最后一个 chunk（有 finish_reason）
		if chunk.FinishReason != "" {
			// 更新 token 用量
			if chunk.Usage != nil {
				usageJSON, _ := json.Marshal(chunk.Usage)
				if _, err := g.DB().Model("mvp_message").Where("id", replyID).Update(g.Map{
					"token_usage": string(usageJSON),
				}); err != nil {
					g.Log().Errorf(ctx, "[ChatEngine] 更新 token_usage 失败: msg=%d, err=%v", replyID, err)
				}
			}
		}

		return nil
	})

	if err != nil {
		e.failMessage(ctx, replyID, err.Error())
		return
	}

	// 5. 更新消息为完成状态
	_, updateErr := g.DB().Model("mvp_message").Where("id", replyID).Update(g.Map{
		"content":      fullContent.String(),
		"status":       "completed",
		"updated_at":   gtime.Now(),
	})
	if updateErr != nil {
		g.Log().Errorf(ctx, "更新消息状态失败: %v", updateErr)
	}

	// 6. 如果是架构师对话，尝试从回复中解析任务清单
	go e.tryParseArchitectTasks(conversationID, fullContent.String())

	// 7. 通知 SSE Hub 流式输出完成
	doneJSON, _ := json.Marshal(map[string]interface{}{
		"done": true,
	})
	e.hub.Publish(replyID, string(doneJSON))

	// 短暂延迟后关闭 channel，让前端有时间接收最后的消息
	time.Sleep(100 * time.Millisecond)
	e.hub.Done(replyID)
}

// tryParseArchitectTasks 尝试从架构师回复中解析任务清单
func (e *ChatEngine) tryParseArchitectTasks(conversationID int64, aiReply string) {
	ctx := context.Background()

	// 查对话的角色类型和项目ID
	conv, err := g.DB().Model("mvp_conversation").Where("id", conversationID).One()
	if err != nil || conv.IsEmpty() {
		return
	}

	// 只有架构师对话且是项目级对话（task_id 为空）才解析
	if conv["role_type"].String() != "architect" || conv["task_id"].Int64() != 0 {
		return
	}

	projectID := conv["project_id"].Int64()

	// 尝试解析
	count, err := GetParser().ParseAndCreateTasks(ctx, projectID, aiReply)
	if err != nil {
		g.Log().Warningf(ctx, "[ChatEngine] 解析任务失败: %v", err)
		return
	}
	if count > 0 {
		g.Log().Infof(ctx, "[ChatEngine] 架构师回复中解析出 %d 个任务（draft），项目 %d", count, projectID)
		// 注：前端在架构师回复结束后会主动调用 loadProjectStatus() 刷新草稿任务数，
		// 因此不再通过 SSE 推送 tasks_parsed 事件（之前推到 conversationID channel，
		// 但前端只订阅 messageID channel，导致这条通知始终无效）。
	}
}

// loadHistory 加载对话历史（排除当前正在 streaming 的消息）
func (e *ChatEngine) loadHistory(ctx context.Context, conversationID int64, excludeID int64) ([]provider.Message, error) {
	var records []gdb.Record
	err := g.DB().Model("mvp_message").
		Where("conversation_id", conversationID).
		Where("deleted_at IS NULL").
		Where("status", "completed").
		Where("(message_type IS NULL OR message_type <> ?)", mvpmodel.MessageTypePoison).
		Where("id != ?", excludeID).
		Order("created_at ASC").
		Scan(&records)
	if err != nil {
		return nil, fmt.Errorf("加载对话历史失败: %w", err)
	}

	messages := make([]provider.Message, 0, len(records))
	for _, r := range records {
		role := provider.Role(r["role"].String())
		messages = append(messages, provider.Message{
			Role:    role,
			Content: r["content"].String(),
		})
	}
	return messages, nil
}

// failMessage 标记消息为失败状态
func (e *ChatEngine) failMessage(ctx context.Context, replyID int64, errMsg string) {
	g.Log().Errorf(ctx, "AI 调用失败 (messageID=%d): %s", replyID, errMsg)

	g.DB().Model("mvp_message").Where("id", replyID).Update(g.Map{
		"message_type": mvpmodel.MessageTypePoison,
		"status":     "failed",
		"content":    "AI 调用失败: " + errMsg,
		"updated_at": gtime.Now(),
	})

	// 通知前端失败
	errJSON, _ := json.Marshal(map[string]interface{}{
		"error": errMsg,
		"done":  true,
	})
	e.hub.Publish(replyID, string(errJSON))
	e.hub.Done(replyID)
}
