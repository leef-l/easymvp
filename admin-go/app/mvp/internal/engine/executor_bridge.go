package engine

// executor_bridge.go 导出执行器相关函数，供新 workflow/stage/execute 包调用。

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

// GetModelInfoByID 根据模型 ID 获取模型信息（供 controller 层调用）。
func GetModelInfoByID(ctx context.Context, modelID int64) (*ModelInfo, error) {
	return getModelInfoStatic(ctx, modelID, "")
}

// ResolveModelInfo 根据 projectID + roleType + modelID 解析模型信息。
// 如果 modelID > 0，直接查模型；否则从角色配置中查找。
func ResolveModelInfo(ctx context.Context, projectID int64, roleType string, modelID int64) (*ModelInfo, error) {
	return ResolveProjectModelInfo(ctx, projectID, roleType, "", modelID)
}

// getModelInfoStatic 静态版模型查询（不依赖 Executor 实例）。
func getModelInfoStatic(ctx context.Context, modelID int64, systemPrompt string) (*ModelInfo, error) {
	model, err := g.DB().Model("ai_model m").
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.model_code, m.max_tokens, pv.provider_type, pv.supported_protocols, pv.base_url, p.api_key, p.api_secret, m.role_prompt").
		Where("m.id", modelID).
		WhereNull("m.deleted_at").
		One()
	if err != nil || model.IsEmpty() {
		return nil, fmt.Errorf("AI模型 %d 不存在", modelID)
	}

	prompt := systemPrompt
	if prompt == "" {
		prompt = model["role_prompt"].String()
	}

	return &ModelInfo{
		ModelID:            modelID,
		ModelCode:          model["model_code"].String(),
		ProviderType:       model["provider_type"].String(),
		SupportedProtocols: decodeProviderProtocols(model["supported_protocols"].String(), model["provider_type"].String()),
		BaseURL:            model["base_url"].String(),
		APIKey:             model["api_key"].String(),
		APISecret:          model["api_secret"].String(),
		SystemPrompt:       prompt,
		MaxTokens:          model["max_tokens"].Int(),
	}, nil
}

// EnsureDomainTaskConversation 为领域任务确保对话存在（创建 mvp_conversation）。
func EnsureDomainTaskConversation(ctx context.Context, projectID, taskID int64, roleType, taskName string) (int64, error) {
	// 查找已有对话
	conv, err := g.DB().Model("mvp_conversation").Ctx(ctx).
		Where("project_id", projectID).
		Where("task_id", taskID).
		WhereNull("deleted_at").
		One()
	if err != nil {
		return 0, err
	}
	if !conv.IsEmpty() {
		convID := conv["id"].Int64()
		if convID > 0 {
			_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
				Where("id", taskID).
				WhereNull("deleted_at").
				Update(g.Map{"conversation_id": convID, "updated_at": gtime.Now()})
		}
		return convID, nil
	}

	// 查项目所有者
	project, projErr := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", projectID).Fields("created_by, dept_id").One()
	if projErr != nil {
		return 0, fmt.Errorf("查询项目 %d 失败: %w", projectID, projErr)
	}
	if project.IsEmpty() {
		return 0, fmt.Errorf("项目 %d 不存在", projectID)
	}

	convID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_conversation").Ctx(ctx).Insert(g.Map{
		"id":         convID,
		"project_id": projectID,
		"task_id":    taskID,
		"title":      fmt.Sprintf("任务: %s", taskName),
		"role_type":  roleType,
		"status":     "active",
		"created_by": project["created_by"].Int64(),
		"dept_id":    project["dept_id"].Int64(),
		"created_at": gtime.Now(),
		"updated_at": gtime.Now(),
	})
	if err != nil {
		return 0, err
	}
	_, _ = g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("id", taskID).
		WhereNull("deleted_at").
		Update(g.Map{"conversation_id": convID, "updated_at": gtime.Now()})
	return convID, nil
}

// NewAiderRunner 导出 AiderRunner 构造函数（已存在但确保可见性）。
// GetAiderRunner 已经导出，此处只是别名确认。
var _ = GetAiderRunner

// AiderTaskConfig 在 execute service 中使用的 Aider 配置（简化版）。
// 实际调用走 AiderRunner.RunTask。
type AiderTaskConfig struct {
	WorkDir     string
	TaskName    string
	Description string
	ModelCode   string
	APIKey      string
	APISecret   string
	BaseURL     string
}

// ─── 飞书通知钩子注册 ─────────────────────────────────────────────────────────
// 用函数变量注入避免循环引用（engine 不 import collab/notifier）

// RegisterFeishuNotifyAIReply 注册 AI 回复完成后的飞书推送钩子。
func RegisterFeishuNotifyAIReply(fn func(ctx context.Context, conversationID int64, content string)) {
	feishuNotifyAIReply = fn
}

// RegisterFeishuNotifyTaskFailed 注册任务失败飞书推送钩子。
func RegisterFeishuNotifyTaskFailed(fn func(ctx context.Context, projectID, taskID int64, taskName, errMsg string)) {
	feishuNotifyTaskFailed = fn
}

// RegisterFeishuNotifyProjectCompleted 注册项目完成飞书推送钩子。
func RegisterFeishuNotifyProjectCompleted(fn func(ctx context.Context, projectID int64)) {
	feishuNotifyProjectCompleted = fn
}
