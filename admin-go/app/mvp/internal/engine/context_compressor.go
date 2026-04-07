package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/provider"
)

// ContextCompressor 上下文压缩器
//
// 三层优化策略：
// 1. 规则压缩优先：< 500字原文存、500-3000字规则截取、> 3000字才调 AI
// 2. 批次合并压缩：同批次任务完成后一次性压缩成批次摘要，不是逐个压缩
// 3. 渐进式全局摘要：新批次摘要合并进全局摘要（恒定 ~3000字），不累加
type ContextCompressor struct {
	projectLocks sync.Map // projectID(int64) → *sync.Mutex，项目级并发锁
}

var defaultCompressor = &ContextCompressor{}

func GetCompressor() *ContextCompressor {
	return defaultCompressor
}

// ----------------------------------------------------------------
// 优化1: 单任务智能压缩（规则优先，减少 AI 调用）
// ----------------------------------------------------------------

// compressionParams 分类族压缩参数
type compressionParams struct {
	ShortThreshold  int    // 短文本阈值（原文保留）
	MediumThreshold int    // 中等文本阈值（规则截取）
	GlobalLimit     int    // 全局上下文上限
	TaskPrompt      string // AI 压缩任务的系统提示词
	ProjectPrompt   string // AI 压缩项目的系统提示词
}

// getCompressionParams 根据项目分类族获取压缩参数
func getCompressionParams(projectCategory string) compressionParams {
	family := GetCategoryFamily(projectCategory)
	switch family {
	case CategoryFamilyCreative:
		return compressionParams{
			ShortThreshold:  800, // 创意内容阈值更高（保留更多原文细节）
			MediumThreshold: 5000,
			GlobalLimit:     5000, // 创意项目需要更多上下文（世界观/人物/风格一致性）
			TaskPrompt:      `将以下创意任务内容压缩为300字以内的摘要。保留：创意产出（章节/场景/人物）、风格约束、剧情走向、世界观/设定变更。纯文本输出。`,
			ProjectPrompt:   `将以下创意项目架构师对话压缩为5000字以内的全局上下文。保留：世界观设定、人物关系图谱、风格指南、剧情主线和支线、创作约束。这将作为所有后续创作任务的背景知识。`,
		}
	case CategoryFamilyAnalysis:
		return compressionParams{
			ShortThreshold:  500,
			MediumThreshold: 3000,
			GlobalLimit:     4000, // 分析项目需要保留数据定义和指标口径
			TaskPrompt:      `将以下分析任务内容压缩为200字以内的摘要。保留：分析方法、数据源、关键结论、可视化说明、指标口径。纯文本输出。`,
			ProjectPrompt:   `将以下分析项目架构师对话压缩为4000字以内的全局上下文。保留：分析目标、数据源定义、指标口径、分析维度、交付物要求。这将作为所有后续分析任务的背景知识。`,
		}
	default:
		return compressionParams{
			ShortThreshold:  500,
			MediumThreshold: 3000,
			GlobalLimit:     3000,
			TaskPrompt:      `将以下任务内容压缩为200字以内的摘要。保留：目标、结果、关键决策、产出物。纯文本输出。`,
			ProjectPrompt:   `将以下项目架构师对话压缩为3000字以内的全局上下文。保留：需求、方案、架构、模块划分、约束、依赖关系。这将作为所有后续任务的背景知识。`,
		}
	}
}

// CompressTaskContext 压缩单个任务的上下文
func (c *ContextCompressor) CompressTaskContext(ctx context.Context, projectID int64, taskID int64) error {
	task, err := g.DB().Model("mvp_task").Where("id", taskID).One()
	if err != nil {
		return fmt.Errorf("查询任务失败: %w", err)
	}
	if task.IsEmpty() {
		return nil
	}

	// 获取项目分类
	params := c.getProjectCompressionParams(ctx, projectID)

	// 获取任务结果文本
	content := task["result"].String()
	if content == "" {
		// 尝试从对话中获取
		content = c.collectTaskDialog(ctx, taskID)
	}
	if content == "" {
		return nil
	}

	// 规则压缩：按长度分级（使用分类感知的阈值）
	var summary string
	switch {
	case len(content) < params.ShortThreshold:
		// 很短，原文保存，不调 AI
		summary = content

	case len(content) < params.MediumThreshold:
		// 中等长度，规则截取
		summary = ruleCompress(content)

	default:
		// 长文本，调 AI 压缩
		modelInfo, err := c.getCompressModel(ctx, projectID)
		if err != nil {
			summary = ruleCompress(content)
		} else {
			aiSummary, err := c.aiCompressTask(ctx, modelInfo, task["name"].String(), content, params.TaskPrompt)
			if err != nil {
				summary = ruleCompress(content)
			} else {
				summary = aiSummary
			}
		}
	}

	c.saveSummary(ctx, taskID, summary)
	return nil
}

// getProjectCompressionParams 获取项目的压缩参数
func (c *ContextCompressor) getProjectCompressionParams(ctx context.Context, projectID int64) compressionParams {
	project, err := g.DB().Model("mvp_project").Where("id", projectID).Fields("project_category").One()
	if err != nil || project.IsEmpty() {
		return getCompressionParams("")
	}
	return getCompressionParams(project["project_category"].String())
}

// ----------------------------------------------------------------
// 优化2: 批次完成后合并压缩
// ----------------------------------------------------------------

// CompressBatchContext 一个批次的任务全部完成后，合并压缩成批次摘要
// 然后用批次摘要更新全局上下文（优化3）
func (c *ContextCompressor) CompressBatchContext(ctx context.Context, projectID int64, batchNo int) {
	// 收集该批次所有已完成任务的摘要
	tasks, err := g.DB().Model("mvp_task").
		Where("project_id", projectID).
		Where("batch_no", batchNo).
		Where("status", "completed").
		Where("deleted_at IS NULL").
		Fields("name, role_type, role_level, context_summary").
		Order("sort ASC").
		All()
	if err != nil || len(tasks) == 0 {
		return
	}

	// 拼接批次内所有任务摘要
	var batchText strings.Builder
	batchText.WriteString(fmt.Sprintf("=== 批次 %d 完成摘要 ===\n", batchNo))
	for _, t := range tasks {
		summary := t["context_summary"].String()
		if summary == "" {
			continue
		}
		batchText.WriteString(fmt.Sprintf("\n[%s-%s] %s:\n%s\n",
			t["role_type"].String(),
			t["role_level"].String(),
			t["name"].String(),
			summary,
		))
	}

	batchSummary := batchText.String()

	// 优化3: 将批次摘要合并进全局上下文
	c.mergeIntoGlobalContext(ctx, projectID, batchSummary)
}

// ----------------------------------------------------------------
// 优化3: 渐进式全局摘要（恒定 ~3000字）
// ----------------------------------------------------------------

// getProjectLock 获取项目级互斥锁（延迟创建）
func (c *ContextCompressor) getProjectLock(projectID int64) *sync.Mutex {
	actual, _ := c.projectLocks.LoadOrStore(projectID, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

// mergeIntoGlobalContext 将新内容合并进全局上下文
// 如果合并后超过 3000 字，调 AI 重新压缩为 3000 字
// 使用项目级锁防止多批次同时完成时互相覆盖
func (c *ContextCompressor) mergeIntoGlobalContext(ctx context.Context, projectID int64, newContent string) {
	mu := c.getProjectLock(projectID)
	mu.Lock()
	defer mu.Unlock()
	project, err := g.DB().Model("mvp_project").Where("id", projectID).
		Fields("global_context, name, project_category").One()
	if err != nil || project.IsEmpty() {
		return
	}

	params := getCompressionParams(project["project_category"].String())
	existing := project["global_context"].String()
	merged := existing + "\n\n" + newContent

	if len(merged) < params.GlobalLimit {
		// 还没超，直接追加
		c.saveProjectContext(ctx, projectID, merged)
		return
	}

	// 超了，调 AI 重新压缩
	modelInfo, err := c.getCompressModel(ctx, projectID)
	if err != nil {
		c.saveProjectContext(ctx, projectID, ruleCompress(merged))
		return
	}

	compressed, err := c.aiMergeGlobal(ctx, modelInfo, project["name"].String(), merged, params.GlobalLimit)
	if err != nil {
		c.saveProjectContext(ctx, projectID, ruleCompress(merged))
		return
	}

	c.saveProjectContext(ctx, projectID, compressed)
}

// CompressProjectContext 压缩架构师对话为初始全局上下文（确认方案时调用）
func (c *ContextCompressor) CompressProjectContext(ctx context.Context, projectID int64) error {
	messages, err := g.DB().Model("mvp_message m").
		LeftJoin("mvp_conversation cv", "cv.id = m.conversation_id").
		Where("cv.project_id", projectID).
		Where("cv.task_id IS NULL").
		Where("cv.role_type", "architect").
		Where("m.deleted_at IS NULL").
		Where("m.status", "completed").
		Fields("m.role, m.content").
		Order("m.created_at ASC").
		All()
	if err != nil {
		return fmt.Errorf("查询架构师对话失败: %w", err)
	}
	if len(messages) == 0 {
		return nil
	}

	var dialogText strings.Builder
	for _, msg := range messages {
		dialogText.WriteString(fmt.Sprintf("[%s]: %s\n\n", msg["role"].String(), msg["content"].String()))
	}
	dialog := dialogText.String()

	params := c.getProjectCompressionParams(ctx, projectID)

	if len(dialog) < params.GlobalLimit {
		c.saveProjectContext(ctx, projectID, dialog)
		return nil
	}

	modelInfo, err := c.getCompressModel(ctx, projectID)
	if err != nil {
		c.saveProjectContext(ctx, projectID, ruleCompress(dialog))
		return nil
	}

	project, _ := g.DB().Model("mvp_project").Where("id", projectID).Fields("name").One()
	projectName := ""
	if !project.IsEmpty() {
		projectName = project["name"].String()
	}

	summary, err := c.aiCompressProject(ctx, modelInfo, projectName, dialog, params.ProjectPrompt)
	if err != nil {
		c.saveProjectContext(ctx, projectID, ruleCompress(dialog))
		return nil
	}
	c.saveProjectContext(ctx, projectID, summary)
	return nil
}

// ----------------------------------------------------------------
// BuildTaskSystemPrompt: 构建任务的 system prompt
// 优化后：架构师/max 只读全局摘要（恒定 ~3000字），不再逐条读任务摘要
// ----------------------------------------------------------------

func BuildTaskSystemPrompt(ctx context.Context, projectID int64, taskID int64, roleType string, roleLevel string, basePrompt string) string {
	var sb strings.Builder
	sb.WriteString(basePrompt)

	if roleType == "architect" || roleLevel == "max" {
		// 架构师/max：读全局摘要（恒定大小，已包含所有历史）
		project, _ := g.DB().Model("mvp_project").
			Where("id", projectID).
			Fields("global_context, name, description").One()
		if !project.IsEmpty() {
			sb.WriteString("\n\n## 项目信息\n")
			sb.WriteString(fmt.Sprintf("项目名称：%s\n项目简介：%s\n", project["name"].String(), project["description"].String()))

			globalCtx := project["global_context"].String()
			if globalCtx != "" {
				sb.WriteString("\n## 项目全局上下文（含已完成工作摘要）\n")
				sb.WriteString(globalCtx)
			}
		}
	} else {
		// pro/lite：只读直接依赖任务的摘要
		deps, _ := g.DB().Model("mvp_task_dependency d").
			LeftJoin("mvp_task t", "t.id = d.depends_on_id").
			Where("d.task_id", taskID).
			Where("t.status", "completed").
			Where("t.context_summary IS NOT NULL").
			Fields("t.name, t.context_summary").
			All()

		if len(deps) > 0 {
			sb.WriteString("\n\n## 前置任务摘要\n")
			for _, d := range deps {
				sb.WriteString(fmt.Sprintf("\n### %s\n%s\n", d["name"].String(), d["context_summary"].String()))
			}
		}
	}

	return sb.String()
}

// ----------------------------------------------------------------
// 规则压缩（不调 AI，零 token 消耗）
// ----------------------------------------------------------------

// ruleCompress 规则截取：前500字 + ... + 末200字
func ruleCompress(content string) string {
	runes := []rune(content)
	if len(runes) <= 700 {
		return content
	}
	head := string(runes[:500])
	tail := string(runes[len(runes)-200:])
	return head + "\n\n...(中间省略)...\n\n" + tail
}

// ----------------------------------------------------------------
// AI 压缩方法
// ----------------------------------------------------------------

func (c *ContextCompressor) aiCompressTask(ctx context.Context, modelInfo *ModelInfo, taskName string, content string, systemPrompt string) (string, error) {
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
	})
	if err != nil {
		return "", err
	}

	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		MaxTokens:    500,
		SystemPrompt: systemPrompt,
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: fmt.Sprintf("任务：%s\n\n内容：\n%s", taskName, truncate(content, 8000))},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (c *ContextCompressor) aiCompressProject(ctx context.Context, modelInfo *ModelInfo, projectName string, dialog string, systemPrompt string) (string, error) {
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
	})
	if err != nil {
		return "", err
	}

	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		MaxTokens:    2000,
		SystemPrompt: systemPrompt,
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: fmt.Sprintf("项目：%s\n\n对话：\n%s", projectName, truncate(dialog, 15000))},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (c *ContextCompressor) aiMergeGlobal(ctx context.Context, modelInfo *ModelInfo, projectName string, merged string, globalLimit int) (string, error) {
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
	})
	if err != nil {
		return "", err
	}

	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		MaxTokens:    2000,
		SystemPrompt: fmt.Sprintf(`将以下项目全局上下文重新压缩为%d字以内。这是项目的完整知识库，包含需求、方案和所有已完成工作。必须保留所有关键信息，不能丢失任何对后续任务有影响的决策和产出。`, globalLimit),
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: fmt.Sprintf("项目：%s\n\n当前全局上下文：\n%s", projectName, truncate(merged, 15000))},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// ----------------------------------------------------------------
// 辅助方法
// ----------------------------------------------------------------

func (c *ContextCompressor) collectTaskDialog(ctx context.Context, taskID int64) string {
	messages, err := g.DB().Model("mvp_message m").
		LeftJoin("mvp_conversation cv", "cv.id = m.conversation_id").
		Where("cv.task_id", taskID).
		Where("m.deleted_at IS NULL").
		Where("m.status", "completed").
		Fields("m.role, m.content").
		Order("m.created_at ASC").
		All()
	if err != nil || len(messages) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n\n", msg["role"].String(), msg["content"].String()))
	}
	return sb.String()
}

func (c *ContextCompressor) getCompressModel(ctx context.Context, projectID int64) (*ModelInfo, error) {
	role, err := ResolveProjectRole(ctx, projectID, "architect")
	if err != nil {
		return nil, fmt.Errorf("找不到架构师模型配置: %w", err)
	}

	model, err := g.DB().Model("ai_model m").
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.model_code, m.max_tokens, pv.provider_type, pv.base_url, p.api_key, p.api_secret").
		Where("m.id", role["model_id"].Int64()).
		One()
	if err != nil || model.IsEmpty() {
		return nil, fmt.Errorf("架构师模型不存在")
	}

	return &ModelInfo{
		ModelID:      role["model_id"].Int64(),
		ModelCode:    model["model_code"].String(),
		ProviderType: model["provider_type"].String(),
		BaseURL:      model["base_url"].String(),
		APIKey:       model["api_key"].String(),
		APISecret:    model["api_secret"].String(),
		MaxTokens:    2000,
	}, nil
}

func (c *ContextCompressor) saveSummary(ctx context.Context, taskID int64, summary string) {
	g.DB().Model("mvp_task").Where("id", taskID).Update(g.Map{
		"context_summary": summary,
		"updated_at":      gtime.Now(),
	})
}

func (c *ContextCompressor) saveProjectContext(ctx context.Context, projectID int64, globalCtx string) {
	g.DB().Model("mvp_project").Where("id", projectID).Update(g.Map{
		"global_context": globalCtx,
		"updated_at":     gtime.Now(),
	})
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "...(已截取)"
}
