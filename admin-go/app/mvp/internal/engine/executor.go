package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/activity"
	mvpmodel "easymvp/app/mvp/internal/model"
	"easymvp/utility/provider"
	"easymvp/utility/snowflake"
)

// Executor 任务执行器，为每个任务启动 goroutine 调用 AI
type Executor struct {
	scheduler *Scheduler
}

// NewExecutor 创建执行器
func NewExecutor(scheduler *Scheduler) *Executor {
	return &Executor{scheduler: scheduler}
}

// Execute 执行单个任务
// implementer 角色使用 Aider 进行真实代码编辑，其他角色使用 ChatStream 对话
func (e *Executor) Execute(ctx context.Context, projectID int64, taskID int64) {
	// 1. 查询任务信息
	task, err := g.DB().Model("mvp_task").Where("id", taskID).One()
	if err != nil || task.IsEmpty() {
		e.scheduler.OnTaskFailed(projectID, taskID, "任务不存在")
		return
	}

	roleType := task["role_type"].String()
	modelID := task["model_id"].Int64()

	// 2. 获取模型信息
	modelInfo, err := e.resolveTaskModel(ctx, projectID, taskID, roleType, modelID)
	if err != nil {
		e.handleTaskFailure(ctx, projectID, taskID, roleType, classifyTaskConfigError(err), err.Error())
		return
	}

	// 3. implementer 角色 → Aider 代码编辑模式
	if roleType == "implementer" {
		e.executeWithAider(ctx, projectID, taskID, task, modelInfo)
		return
	}

	// 4. 其他角色 → ChatStream 对话模式
	// 查找或创建任务对话
	conversationID, err := e.ensureConversation(ctx, projectID, taskID, roleType)
	if err != nil {
		e.handleTaskFailure(ctx, projectID, taskID, roleType, taskFailureExecution, err.Error())
		return
	}

	// 回写 conversation_id 到 task，方便前端和 watchdog 检测
	if _, err := g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
		"conversation_id": conversationID,
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 回写 conversation_id 失败: task=%d, err=%v", taskID, err)
	}

	// 4. 构建任务指令消息
	taskPrompt := e.buildTaskPrompt(task)

	// 5. 保存指令消息
	userMsgID := int64(snowflake.Generate())
	if _, err := g.DB().Model("mvp_message").Insert(g.Map{
		"id":              userMsgID,
		"conversation_id": conversationID,
		"role":            "user",
		"message_type":    mvpmodel.MessageTypeTaskPrompt,
		"content":         taskPrompt,
		"status":          "completed",
		"created_by":      0,
		"dept_id":         0,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 保存用户消息失败: task=%d, err=%v", taskID, err)
	}

	// 6. 创建 AI 回复消息
	replyID := int64(snowflake.Generate())
	if _, err := g.DB().Model("mvp_message").Insert(g.Map{
		"id":              replyID,
		"conversation_id": conversationID,
		"role":            "assistant",
		"message_type":    mvpmodel.MessageTypeTaskReply,
		"content":         "",
		"model_id":        modelInfo.ModelID,
		"status":          "streaming",
		"created_by":      0,
		"dept_id":         0,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 创建AI回复消息失败: task=%d, err=%v", taskID, err)
	}

	// 7. 调用 AI
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
	})
	if err != nil {
		e.handleTaskFailure(ctx, projectID, taskID, roleType, taskFailureExecution, err.Error())
		return
	}

	// 加载对话历史
	history, _ := e.loadConversationHistory(ctx, conversationID, replyID)

	// 构建包含上下文摘要的 system prompt
	enrichedPrompt := BuildTaskSystemPrompt(ctx, projectID, taskID, roleType, task["role_level"].String(), modelInfo.SystemPrompt)

	req := &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     history,
		MaxTokens:    modelInfo.MaxTokens,
		Temperature:  0.7,
		Stream:       true,
		SystemPrompt: enrichedPrompt,
	}

	var fullContent strings.Builder
	chunkIndex := 0
	hub := GetHub()

	err = p.ChatStream(ctx, req, func(chunk *provider.StreamChunk) error {
		if chunk.Content != "" {
			fullContent.WriteString(chunk.Content)
			chunkIndex++

			if _, insertErr := g.DB().Model("mvp_message_chunk").Insert(g.Map{
				"message_id":  replyID,
				"chunk_index": chunkIndex,
				"content":     chunk.Content,
				"created_at":  gtime.Now(),
			}); insertErr != nil {
				g.Log().Errorf(ctx, "[Executor] 写入 chunk 失败: msg=%d, err=%v", replyID, insertErr)
			}
			activity.TouchMessageActivity(ctx, replyID)
			activity.TouchConversationActivity(ctx, conversationID)
			activity.TouchTaskActivity(ctx, taskID)

			chunkJSON, _ := json.Marshal(map[string]interface{}{
				"content": chunk.Content,
				"index":   chunkIndex,
			})
			hub.Publish(replyID, string(chunkJSON))
		}

		if chunk.FinishReason != "" && chunk.Usage != nil {
			usageJSON, _ := json.Marshal(chunk.Usage)
			if _, err := g.DB().Model("mvp_message").Where("id", replyID).Update(g.Map{
				"token_usage": string(usageJSON),
			}); err != nil {
				g.Log().Errorf(ctx, "[Executor] 更新 token_usage 失败: msg=%d, err=%v", replyID, err)
			}
		}

		return nil
	})

	if err != nil {
		// AI 调用失败
		if _, dbErr := g.DB().Model("mvp_message").Where("id", replyID).Update(g.Map{
			"content":      "AI调用失败: " + err.Error(),
			"message_type": mvpmodel.MessageTypePoison,
			"status":       "failed",
			"updated_at":   gtime.Now(),
		}); dbErr != nil {
			g.Log().Errorf(ctx, "[Executor] 更新失败消息状态失败: msg=%d, err=%v", replyID, dbErr)
		}
		e.handleTaskFailure(ctx, projectID, taskID, roleType, taskFailureExecution, err.Error())
		return
	}

	// 8. 完成消息
	if _, err := g.DB().Model("mvp_message").Where("id", replyID).Update(g.Map{
		"content":    fullContent.String(),
		"status":     "completed",
		"updated_at": gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 更新完成消息失败: msg=%d, err=%v", replyID, err)
	}

	hub.Publish(replyID, `{"done":true}`)
	hub.Done(replyID)

	// 9. 保存任务结果
	result := fullContent.String()
	if _, err := g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
		"result":       result,
		"status":       "completed",
		"completed_at": gtime.Now(),
		"updated_at":   gtime.Now(),
	}); err != nil {
		g.Log().Errorf(ctx, "[Executor] 保存任务结果失败: task=%d, err=%v", taskID, err)
	}

	// 10. 压缩任务上下文为摘要
	go GetCompressor().CompressTaskContext(context.Background(), projectID, taskID)

	// 11. 如果是实施员任务，完成后需要创建对应的审计任务（如果还没有）
	if roleType == "implementer" {
		go e.createAuditTask(ctx, projectID, taskID, task)
	} else if roleType == "architect" {
		name := task["name"].String()
		switch {
		case strings.HasPrefix(name, "Bug分析:"):
			go e.scheduler.AutoDispatchBugFix(context.Background(), projectID, taskID)
		case strings.HasPrefix(name, "失败分析:"):
			go e.scheduler.AutoDispatchFailureFix(context.Background(), projectID, taskID)
		}
	}

	// 12. 通知调度器
	e.scheduler.OnTaskCompleted(projectID, taskID)
}

// resolveTaskModel 解析任务使用的 AI 模型
func (e *Executor) resolveTaskModel(ctx context.Context, projectID int64, taskID int64, roleType string, modelID int64) (*ModelInfo, error) {
	// 如果任务自身指定了 model_id，优先使用
	if modelID > 0 {
		return e.getModelInfo(ctx, modelID, "")
	}

	// 否则从项目角色配置中查找
	task, err := g.DB().Model("mvp_task").Where("id", taskID).Fields("role_level").One()
	if err != nil {
		g.Log().Warningf(ctx, "[Executor] 查询 task role_level 失败: task=%d, err=%v", taskID, err)
	}
	roleLevel := ""
	if !task.IsEmpty() {
		roleLevel = task["role_level"].String()
	}

	// 查找角色配置（匹配 role_type + role_level）
	roleQuery := g.DB().Model("mvp_project_role").
		Where("project_id", projectID).
		Where("role_type", roleType).
		Where("status", 1).
		Where("deleted_at IS NULL")
	if roleLevel != "" {
		roleQuery = roleQuery.Where("role_level", roleLevel)
	}

	role, err := roleQuery.One()
	if err != nil || role.IsEmpty() {
		// 没有匹配等级的，用该角色类型的任意一个
		role, err = g.DB().Model("mvp_project_role").
			Where("project_id", projectID).
			Where("role_type", roleType).
			Where("status", 1).
			Where("deleted_at IS NULL").
			One()
		if err != nil || role.IsEmpty() {
			return nil, fmt.Errorf("项目未配置 %s 角色模型", roleType)
		}
	}

	return e.getModelInfo(ctx, role["model_id"].Int64(), role["system_prompt"].String())
}

// getModelInfo 查询模型详情
func (e *Executor) getModelInfo(ctx context.Context, modelID int64, systemPrompt string) (*ModelInfo, error) {
	model, err := g.DB().Model("ai_model m").
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.model_code, m.max_tokens, pv.provider_type, pv.base_url, p.api_key, p.api_secret, m.role_prompt").
		Where("m.id", modelID).
		Where("m.deleted_at IS NULL").
		One()
	if err != nil || model.IsEmpty() {
		return nil, fmt.Errorf("AI模型 %d 不存在", modelID)
	}

	prompt := systemPrompt
	if prompt == "" {
		prompt = model["role_prompt"].String()
	}

	return &ModelInfo{
		ModelID:      modelID,
		ModelCode:    model["model_code"].String(),
		ProviderType: model["provider_type"].String(),
		BaseURL:      model["base_url"].String(),
		APIKey:       model["api_key"].String(),
		APISecret:    model["api_secret"].String(),
		SystemPrompt: prompt,
		MaxTokens:    model["max_tokens"].Int(),
	}, nil
}

// ensureConversation 确保任务有对应的对话
func (e *Executor) ensureConversation(ctx context.Context, projectID int64, taskID int64, roleType string) (int64, error) {
	// 查找已有的任务对话
	conv, err := g.DB().Model("mvp_conversation").
		Where("project_id", projectID).
		Where("task_id", taskID).
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return 0, err
	}
	if !conv.IsEmpty() {
		return conv["id"].Int64(), nil
	}

	// 创建新对话
	project, err := g.DB().Model("mvp_project").
		Fields("created_by, dept_id").
		Where("id", projectID).
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return 0, err
	}
	if project.IsEmpty() {
		return 0, fmt.Errorf("项目不存在")
	}

	convID := int64(snowflake.Generate())
	_, err = g.DB().Model("mvp_conversation").Insert(g.Map{
		"id":         convID,
		"project_id": projectID,
		"task_id":    taskID,
		"title":      "任务对话",
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
	return convID, nil
}

// loadConversationHistory 加载对话历史
func (e *Executor) loadConversationHistory(ctx context.Context, conversationID int64, excludeID int64) ([]provider.Message, error) {
	records, err := g.DB().Model("mvp_message").
		Where("conversation_id", conversationID).
		Where("deleted_at IS NULL").
		Where("status", "completed").
		Where("(message_type IS NULL OR message_type <> ?)", mvpmodel.MessageTypePoison).
		Where("id != ?", excludeID).
		Order("created_at ASC").
		All()
	if err != nil {
		return nil, err
	}

	messages := make([]provider.Message, 0, len(records))
	for _, r := range records {
		messages = append(messages, provider.Message{
			Role:    provider.Role(r["role"].String()),
			Content: r["content"].String(),
		})
	}
	return messages, nil
}

// buildTaskPrompt 构建任务指令
func (e *Executor) buildTaskPrompt(task gdb.Record) string {
	name := task["name"].String()
	desc := task["description"].String()

	prompt := fmt.Sprintf("任务名称：%s\n任务描述：%s", name, desc)

	// 如果有依赖任务的结果，附加上下文
	taskID := task["id"].Int64()
	deps, _ := g.DB().Model("mvp_task_dependency d").
		LeftJoin("mvp_task t", "t.id = d.depends_on_id").
		Fields("t.name, t.result").
		Where("d.task_id", taskID).
		Where("t.status", "completed").
		All()

	if len(deps) > 0 {
		prompt += "\n\n## 前置任务结果（供参考）"
		for _, dep := range deps {
			depName := dep["name"].String()
			depResult := dep["result"].String()
			if len(depResult) > 2000 {
				depResult = depResult[:2000] + "...(截断)"
			}
			prompt += fmt.Sprintf("\n\n### %s\n%s", depName, depResult)
		}
	}

	return prompt
}

// buildAiderTaskPrompt 为 Aider 构建更紧凑的任务指令，避免上下文过大触发 token limit。
func (e *Executor) buildAiderTaskPrompt(task gdb.Record, resources []string) string {
	name := task["name"].String()
	desc := task["description"].String()

	prompt := fmt.Sprintf("## 任务\n%s\n\n## 任务描述\n%s", name, desc)

	if len(resources) > 0 {
		limitedResources := resources
		if len(limitedResources) > 12 {
			limitedResources = limitedResources[:12]
		}
		prompt += "\n\n允许修改的文件或目录："
		for _, resource := range limitedResources {
			if resource == "" {
				continue
			}
			prompt += "\n- " + resource
		}
		if len(resources) > len(limitedResources) {
			prompt += fmt.Sprintf("\n- 其余 %d 个文件暂不优先处理，必要时再扩展", len(resources)-len(limitedResources))
		}
	}

	taskID := task["id"].Int64()
	deps, _ := g.DB().Model("mvp_task_dependency d").
		LeftJoin("mvp_task t", "t.id = d.depends_on_id").
		Fields("t.name, t.result").
		Where("d.task_id", taskID).
		Where("t.status", "completed").
		Limit(2).
		All()

	if len(deps) > 0 {
		prompt += "\n\n前置结果摘要："
		for _, dep := range deps {
			depName := dep["name"].String()
			depResult := dep["result"].String()
			if len(depResult) > 800 {
				depResult = depResult[:800] + "...(截断)"
			}
			prompt += fmt.Sprintf("\n- %s：%s", depName, depResult)
		}
	}

	prompt += "\n\n执行约束：只允许修改上面列出的文件或目录；不要输出“1. 标题（路径）”或其他 Markdown 章节标题来描述文件；不要把说明标题、编号、括号说明当成文件名；请直接完成最小必要改动。"
	return prompt
}

// createAuditTask 为实施员任务创建对应的审计任务
func (e *Executor) createAuditTask(ctx context.Context, projectID int64, implTaskID int64, implTask gdb.Record) {
	// 检查是否已有审计任务（通过依赖关系）
	count, _ := g.DB().Model("mvp_task_dependency").
		Where("depends_on_id", implTaskID).
		Count()
	if count > 0 {
		return
	}

	auditTaskID := int64(snowflake.Generate())
	g.DB().Model("mvp_task").Insert(g.Map{
		"id":          auditTaskID,
		"project_id":  projectID,
		"parent_id":   implTask["parent_id"].Int64(),
		"name":        fmt.Sprintf("审计: %s", implTask["name"].String()),
		"description": fmt.Sprintf("审计实施员任务「%s」的结果，检查是否正确完成，是否有 bug。", implTask["name"].String()),
		"role_type":   "auditor",
		"role_level":  implTask["role_level"].String(),
		"status":      "pending",
		"batch_no":    implTask["batch_no"].Int() + 1,
		"created_by":  0,
		"dept_id":     0,
		"created_at":  gtime.Now(),
		"updated_at":  gtime.Now(),
	})

	// 添加依赖关系
	g.DB().Model("mvp_task_dependency").Insert(g.Map{
		"task_id":       auditTaskID,
		"depends_on_id": implTaskID,
	})
}

// executeWithAider 使用 Aider 执行实施类任务（真实代码编辑）
func (e *Executor) executeWithAider(ctx context.Context, projectID int64, taskID int64, task gdb.Record, modelInfo *ModelInfo) {
	// 1. 查询项目获取工作目录
	project, err := g.DB().Model("mvp_project").Where("id", projectID).One()
	if err != nil || project.IsEmpty() {
		e.failTask(ctx, projectID, taskID, "项目不存在")
		return
	}

	// 工作目录：优先项目配置的 work_dir，兜底用默认
	workDir := project["work_dir"].String()
	if workDir == "" {
		workDir = "/www/wwwroot/project/easymvp"
	}

	// 2. 解析 affected_resources 作为需要编辑的文件
	resourceResult := parseResourcesDetail(task["affected_resources"].String())
	if len(resourceResult.Rejected) > 0 {
		e.escalateImplementerResourceIssue(ctx, projectID, taskID, task, resourceResult)
		return
	}
	resources := resourceResult.Resources

	// 3. 构建更紧凑的 Aider prompt，避免一次性塞入过多上下文
	taskPrompt := e.buildAiderTaskPrompt(task, resources)

	// 4. 创建对话记录（用于前端展示 Aider 过程）
	conversationID, _ := e.ensureConversation(ctx, projectID, taskID, "implementer")
	g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
		"conversation_id": conversationID,
	})
	activity.TouchTaskActivity(ctx, taskID)
	activity.TouchConversationActivity(ctx, conversationID)

	// 保存指令消息
	userMsgID := int64(snowflake.Generate())
	g.DB().Model("mvp_message").Insert(g.Map{
		"id":              userMsgID,
		"conversation_id": conversationID,
		"role":            "user",
		"message_type":    mvpmodel.MessageTypeTaskPrompt,
		"content":         taskPrompt,
		"status":          "completed",
		"created_by":      0,
		"dept_id":         0,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	})

	// 5. 调用 Aider
	g.Log().Infof(ctx, "[Executor] 任务 %d 使用 Aider 执行: model=%s files=%v", taskID, modelInfo.ModelCode, resources)

	result := GetAiderRunner().RunTask(ctx, projectID, taskID, modelInfo, taskPrompt, workDir, resources, nil)

	// 6. 保存 Aider 输出为 AI 回复消息
	replyID := int64(snowflake.Generate())
	replyStatus := "completed"
	if result.Error != nil {
		replyStatus = "failed"
	}
	g.DB().Model("mvp_message").Insert(g.Map{
		"id":              replyID,
		"conversation_id": conversationID,
		"role":            "assistant",
		"message_type":    map[bool]string{true: mvpmodel.MessageTypePoison, false: mvpmodel.MessageTypeTaskReply}[result.Error != nil],
		"content":         result.Output,
		"model_id":        modelInfo.ModelID,
		"status":          replyStatus,
		"created_by":      0,
		"dept_id":         0,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	})

	// 7. 判断结果
	if result.Error != nil {
		errMsg := result.FailureHint
		if errMsg == "" {
			errMsg = result.Error.Error()
		}
		category := result.Category
		if category == "" {
			category = taskFailureExecution
		}
		e.handleTaskFailure(ctx, projectID, taskID, task["role_type"].String(), category, fmt.Sprintf("Aider执行失败(code=%d): %s", result.ExitCode, errMsg))
		return
	}

	// 8. 更新任务为完成
	g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
		"result":       result.Output,
		"status":       "completed",
		"completed_at": gtime.Now(),
		"updated_at":   gtime.Now(),
	})

	// 9. 压缩上下文
	go GetCompressor().CompressTaskContext(context.Background(), projectID, taskID)

	// 10. 创建审计任务
	go e.createAuditTask(ctx, projectID, taskID, task)

	// 11. 通知调度器
	e.scheduler.OnTaskCompleted(projectID, taskID)
}

func (e *Executor) escalateImplementerResourceIssue(ctx context.Context, projectID int64, taskID int64, task gdb.Record, resourceResult resourceParseResult) {
	errMsg := fmt.Sprintf(
		"任务涉及的 affected_resources 存在歧义，无法安全继续执行。可修复资源=%v；无法解析资源=%v；原始值=%s。请架构师重新检查任务拆分和 affected_resources。",
		resourceResult.Resources,
		resourceResult.Rejected,
		task["affected_resources"].String(),
	)

	conversationID, err := e.ensureConversation(ctx, projectID, taskID, "implementer")
	if err == nil {
		replyID := int64(snowflake.Generate())
		_, _ = g.DB().Model("mvp_message").Insert(g.Map{
			"id":              replyID,
			"conversation_id": conversationID,
			"role":            "assistant",
			"message_type":    mvpmodel.MessageTypePoison,
			"content":         errMsg,
			"status":          "failed",
			"created_by":      0,
			"dept_id":         0,
			"created_at":      gtime.Now(),
			"updated_at":      gtime.Now(),
		})
	}

	_, _ = g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
		"status":        "escalated",
		"error_message": errMsg,
		"updated_at":    gtime.Now(),
	})

	e.scheduler.OnTaskEscalated(projectID, taskID, errMsg)
	go e.scheduler.EscalateFailedTask(context.Background(), projectID, taskID, task["role_type"].String(), errMsg)
}

// failTask 标记任务失败
func (e *Executor) failTask(ctx context.Context, projectID int64, taskID int64, errMsg string) {
	g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
		"status":        "failed",
		"error_message": errMsg,
		"updated_at":    gtime.Now(),
	})
	e.scheduler.OnTaskFailed(projectID, taskID, errMsg)
}

func (e *Executor) handleTaskFailure(ctx context.Context, projectID int64, taskID int64, roleType string, category taskFailureCategory, errMsg string) {
	switch category {
	case taskFailurePlanning, taskFailurePolicyGuard:
		g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
			"status":        "escalated",
			"error_message": errMsg,
			"updated_at":    gtime.Now(),
		})
		e.scheduler.OnTaskEscalated(projectID, taskID, errMsg)
		go e.scheduler.EscalateFailedTask(context.Background(), projectID, taskID, roleType, errMsg)
	default:
		e.failTask(ctx, projectID, taskID, errMsg)
	}
}
