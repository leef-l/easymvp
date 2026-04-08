package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

var jsonCodeBlockRe = regexp.MustCompile("(?s)```json\\s*\\n?(\\{[\\s\\S]*?\\})\\s*```")

// getReviewRoleModel 获取审核角色的 AI 模型信息
func getReviewRoleModel(ctx context.Context, projectID int64, roleType string) (*ModelInfo, error) {
	role, err := ResolveProjectRole(ctx, projectID, roleType)
	if err != nil {
		return nil, fmt.Errorf("项目未配置 %s 角色: %w", roleType, err)
	}

	modelID := role["model_id"].Int64()
	model, err := g.DB().Ctx(ctx).Model("ai_model m").
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.model_code, m.max_tokens, pv.provider_type, pv.base_url, p.api_key, p.api_secret, m.role_prompt").
		Where("m.id", modelID).
		Where("m.deleted_at IS NULL").
		One()
	if err != nil || model.IsEmpty() {
		return nil, fmt.Errorf("AI模型 %d 不存在", modelID)
	}

	systemPrompt := role["system_prompt"].String()
	if systemPrompt == "" {
		systemPrompt = model["role_prompt"].String()
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

// HandleReviewFailure 审核失败时的处理：退回 designing + 通知架构师
func HandleReviewFailure(ctx context.Context, projectID int64, result *ReviewResult) error {
	// 1. 退回项目状态
	_, err := g.DB().Ctx(ctx).Model("mvp_project").Where("id", projectID).Update(g.Map{
		"status":     "designing",
		"updated_at": gtime.Now(),
	})
	if err != nil {
		return err
	}

	// 2. 将 issues 汇总为通知消息
	var msg strings.Builder
	msg.WriteString("## 方案审核未通过\n\n")

	if len(result.Errors) > 0 {
		msg.WriteString("### 错误（必须修复）\n")
		for i, issue := range result.Errors {
			taskRef := ""
			if issue.TaskName != "" {
				taskRef = fmt.Sprintf("[%s] ", issue.TaskName)
			}
			msg.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, taskRef, issue.Message))
		}
		msg.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		msg.WriteString("### 警告（建议修复）\n")
		for i, issue := range result.Warnings {
			taskRef := ""
			if issue.TaskName != "" {
				taskRef = fmt.Sprintf("[%s] ", issue.TaskName)
			}
			msg.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, taskRef, issue.Message))
		}
		msg.WriteString("\n")
	}

	if len(result.AutoFixes) > 0 {
		msg.WriteString("### 自动修正\n")
		for _, fix := range result.AutoFixes {
			msg.WriteString("- " + fix + "\n")
		}
	}

	msg.WriteString("\n请修正上述问题后重新确认方案。")

	// 3. 在架构师对话中发送通知
	notifyProjectArchitectConversation(ctx, projectID, msg.String())

	return nil
}

// HandleReviewSuccess 审核通过：draft → pending，项目进入 running
// 事务边界：确认任务 + 追加 warning + 项目状态改 running 在同一事务中
func HandleReviewSuccess(ctx context.Context, projectID int64, result *ReviewResult) error {
	// 1. 确认 draft → pending（自带全量确认或回滚保证）
	confirmedCount, err := GetParser().ConfirmDraftTasks(ctx, projectID)
	if err != nil {
		return fmt.Errorf("确认草稿任务失败: %w", err)
	}
	if confirmedCount == 0 {
		return fmt.Errorf("没有任务可确认")
	}

	// 2. 事务内：追加 warning + 项目状态改 running
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 2a. 附加 warnings 到对应任务的描述中（按任务名聚合）
		warningsByTask := make(map[string][]string)
		for _, w := range result.Warnings {
			if w.TaskName != "" {
				warningsByTask[w.TaskName] = append(warningsByTask[w.TaskName], w.Message)
			}
		}
		for taskName, msgs := range warningsByTask {
			suffix := ""
			for _, msg := range msgs {
				suffix += fmt.Sprintf("\n\n⚠️ 审核警告: %s", msg)
			}
			task, qErr := tx.Model("mvp_task").
				Where("project_id", projectID).
				Where("name", taskName).
				Where("status", "pending").
				WhereNull("deleted_at").
				Fields("id,description").One()
			if qErr != nil || task.IsEmpty() {
				continue
			}
			newDesc := task["description"].String() + suffix
			if _, err := tx.Model("mvp_task").
				Where("id", task["id"].Int64()).
				Update(g.Map{
					"description": newDesc,
					"updated_at":  gtime.Now(),
				}); err != nil {
				return fmt.Errorf("附加审核警告失败: task=%s, err=%w", taskName, err)
			}
		}

		// 2b. 更新项目状态为 running
		if _, err := tx.Model("mvp_project").Where("id", projectID).Update(g.Map{
			"status":       "running",
			"pause_reason": nil,
			"updated_at":   gtime.Now(),
		}); err != nil {
			return fmt.Errorf("项目状态更新失败: %w", err)
		}

		return nil
	})
	if err != nil {
		// 事务失败：回滚已确认的任务（pending → draft）
		g.Log().Errorf(ctx, "[Review] 事务失败，回滚 %d 个已确认任务: project=%d, err=%v", confirmedCount, projectID, err)
		rollbackConfirmedTasks(ctx, projectID)
		return fmt.Errorf("审核通过处理失败: %w", err)
	}

	// 4. 压缩架构师对话为全局上下文
	if compErr := GetCompressor().CompressProjectContext(context.Background(), projectID); compErr != nil {
		g.Log().Errorf(ctx, "[Review] 压缩项目上下文失败（非致命）: project=%d, err=%v", projectID, compErr)
	}

	// 5. 启动调度器
	GetScheduler().StartProject(projectID)

	// 6. 如果有 warnings，在架构师对话中通知
	if len(result.Warnings) > 0 || len(result.AutoFixes) > 0 {
		var msg strings.Builder
		msg.WriteString("## 方案审核通过\n\n")
		if len(result.AutoFixes) > 0 {
			msg.WriteString("### 自动修正\n")
			for _, fix := range result.AutoFixes {
				msg.WriteString("- " + fix + "\n")
			}
			msg.WriteString("\n")
		}
		if len(result.Warnings) > 0 {
			msg.WriteString("### 注意事项\n")
			for _, w := range result.Warnings {
				msg.WriteString(fmt.Sprintf("- [%s] %s\n", w.TaskName, w.Message))
			}
		}
		msg.WriteString("\n项目已开始执行。")
		notifyProjectArchitectConversation(ctx, projectID, msg.String())
	}

	return nil
}

// rollbackConfirmedTasks 将项目中已确认的 pending 任务回退为 draft
func rollbackConfirmedTasks(ctx context.Context, projectID int64) {
	taskIDs, err := g.DB().Ctx(ctx).Model("mvp_task").
		Where("project_id", projectID).
		Where("status", "pending").
		WhereNull("deleted_at").
		Fields("id").
		Array()
	if err != nil {
		g.Log().Errorf(ctx, "[Review] 回滚查询失败: project=%d, err=%v", projectID, err)
		return
	}
	for _, idVal := range taskIDs {
		if _, err := updateTaskStatus(ctx, idVal.Int64(), "pending", "draft", nil); err != nil {
			g.Log().Errorf(ctx, "[Review] 回滚 task=%d pending→draft 失败: %v", idVal.Int64(), err)
		}
	}
}

// --- 共享辅助函数 ---

// parseJSONFromAI 从 AI 回复中提取 JSON 并解析
func parseJSONFromAI(content string, v interface{}) error {
	content = strings.TrimSpace(content)

	// 尝试直接解析
	if err := json.Unmarshal([]byte(content), v); err == nil {
		return nil
	}

	// 从 ```json 代码块中提取
	if match := jsonCodeBlockRe.FindStringSubmatch(content); len(match) == 2 {
		if err := json.Unmarshal([]byte(match[1]), v); err == nil {
			return nil
		}
	}

	// 查找最外层的 { ... }
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(content[start:end+1]), v); err == nil {
			return nil
		}
	}

	return fmt.Errorf("无法从 AI 回复中提取有效 JSON")
}

// ModelInfo 已在 executor.go 中定义，此处复用
// 注意：如果 ModelInfo 未导出，需要确认是否在同一个 package 中
var _ = (*ModelInfo)(nil) // 编译期验证 ModelInfo 可访问
