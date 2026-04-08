package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/provider"
)

// AuditorReviewResult 审计员 AI 审核输出
type AuditorReviewResult struct {
	Approved    bool          `json:"approved"`
	Issues      []ReviewIssue `json:"issues"`
	Suggestions string        `json:"suggestions"`
}

// auditorReview 审计员 AI 审核（legacy 入口）
func auditorReview(ctx context.Context, projectID int64, tasks gdb.Result) (*AuditorReviewResult, error) {
	modelInfo, err := getReviewRoleModel(ctx, projectID, "auditor")
	if err != nil {
		return nil, fmt.Errorf("获取审计员模型失败: %w", err)
	}
	project, projErr := g.DB().Model("mvp_project").Ctx(ctx).
		Fields("project_category, name, description").
		Where("id", projectID).WhereNull("deleted_at").One()
	if projErr != nil {
		g.Log().Warningf(ctx, "[AuditorReview] 查询项目失败: projectID=%d err=%v", projectID, projErr)
	}
	projectCategory := "软件开发"
	projectName := ""
	projectDesc := ""
	if !project.IsEmpty() {
		projectCategory = project["project_category"].String()
		projectName = project["name"].String()
		projectDesc = project["description"].String()
	}
	return doAuditorReview(ctx, modelInfo, tasks, projectCategory, projectName, projectDesc)
}

// doAuditorReview 审计员 AI 审核核心逻辑。
// projectCategory 用于选择分类感知的审核标准。
// doAuditorReview 审计员 AI 审核核心逻辑。
// projectCategory 用于选择分类感知的审核标准。
func doAuditorReview(ctx context.Context, modelInfo *ModelInfo, tasks gdb.Result, projectCategory, projectName, projectDesc string) (*AuditorReviewResult, error) {
	taskSummaries := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		taskSummaries = append(taskSummaries, map[string]interface{}{
			"name":               t["name"].String(),
			"description":        truncate(t["description"].String(), 500),
			"role_type":          t["role_type"].String(),
			"role_level":         t["role_level"].String(),
			"batch_no":           t["batch_no"].Int(),
			"affected_resources": t["affected_resources"].String(),
			"depends_on":         t["depends_on"].String(),
		})
	}
	summaryJSON, _ := json.MarshalIndent(taskSummaries, "", "  ")

	prompt := buildAuditorPrompt(projectCategory, projectName, projectDesc, len(tasks), string(summaryJSON))

	// 调用 AI
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
	})
	if err != nil {
		return nil, err
	}

	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens:    modelInfo.MaxTokens,
		Temperature:  0.3,
		SystemPrompt: modelInfo.SystemPrompt,
	})
	if err != nil {
		return nil, err
	}

	// 解析审核结果
	var auditorResult AuditorReviewResult
	if err := parseJSONFromAI(resp.Content, &auditorResult); err != nil {
		return nil, fmt.Errorf("解析审计员审核结果失败: %w, 原文: %s", err, truncate(resp.Content, 500))
	}

	return &auditorResult, nil
}

// buildAuditorPrompt 根据项目分类族构建审核 prompt。
func buildAuditorPrompt(projectCategory, projectName, projectDesc string, taskCount int, summaryJSON string) string {
	family := GetCategoryFamily(projectCategory)

	// 项目上下文（所有分类通用）
	projectContext := fmt.Sprintf("项目名称：%s\n项目分类：%s", projectName, projectCategory)
	if projectDesc != "" {
		desc := projectDesc
		if len([]rune(desc)) > 500 {
			desc = string([]rune(desc)[:500]) + "..."
		}
		projectContext += fmt.Sprintf("\n项目简介：%s", desc)
	}

	var dimensions string
	var preamble string

	switch family {
	case CategoryFamilyCreative:
		preamble = "你是一位创意项目审核员。请以创意项目的标准，从整体方案的角度审核任务清单。"
		dimensions = `1. 任务划分是否覆盖了项目的主要内容模块（世界观、人物、主线剧情等）
2. 任务之间的先后顺序和依赖关系是否合理（设定先行，创作在后）
3. 任务描述是否足够清晰，执行者能理解要产出什么
4. 批次安排是否合理（前置准备 → 主体创作 → 收尾完善）

注意：这是创意类项目，不要用软件工程标准（数据库、API、测试框架等）来审核。
affected_resources 为空是正常的。`

	case CategoryFamilyAnalysis:
		preamble = "你是一位分析项目审核员。请以分析项目的标准，从整体方案的角度审核任务清单。"
		dimensions = `1. 分析流程是否完整（数据收集 → 清洗 → 分析 → 结论/报告）
2. 任务之间的数据依赖关系是否合理（上游产出 → 下游消费）
3. 任务描述是否明确了输入数据来源和预期产出
4. 批次安排是否合理（数据准备 → 并行分析 → 汇总报告）

注意：这是分析类项目，不要用软件工程标准来审核。`

	default: // coding
		preamble = "你是一位软件项目审核员。请以软件工程的标准，从整体方案的角度审核任务清单。"
		dimensions = `1. 整体方案是否覆盖了项目需求的核心功能模块
2. 任务之间的依赖关系和执行顺序是否合理（基础设施 → 核心功能 → 集成/测试）
3. 批次安排是否合理，前置任务是否包含了项目初始化和基础框架
4. 任务粒度是否合理（不能太大也不能太碎）

注意：后续批次的任务依赖前面批次创建的文件，这是正确的设计模式。`
	}

	return fmt.Sprintf(`%s

===== 项目信息 =====
%s

===== 审核原则 =====
1. 你的目标是确保方案整体可执行，而不是追求完美
2. 先通读所有任务，理解项目的整体目标和任务之间的关联关系
3. 这是一个全新项目，所有文件/内容都需要从零创建，后续任务依赖前面任务的产出是正确的设计
4. 只有会导致项目无法执行的严重问题才标记为 error（如关键模块完全遗漏、依赖死循环）
5. 改进建议标记为 warning，不阻塞审核
6. 如果整体方案覆盖了项目核心需求且执行顺序合理，就应该通过

===== 审核维度 =====
%s

===== 任务清单（共 %d 个）=====
%s

请先整体理解项目目标和所有任务的关联关系，再给出审核结论。
请严格输出 JSON，格式如下：
{"approved": true/false, "issues": [{"task_name": "xxx", "severity": "error/warning", "message": "问题描述"}], "suggestions": "整体建议"}`,
		preamble, projectContext, dimensions, taskCount, summaryJSON)
}
