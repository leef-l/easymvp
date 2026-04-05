package autonomy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/provider"
)

// Replanner 自动重规划器。
// 当 rework 失败、批次全军覆没、验收连续不过时，由 LLM 生成重规划建议。
type Replanner struct {
	decisionRepo *repo.AutonomyDecisionRepo
}

// NewReplanner 创建重规划器。
func NewReplanner(decisionRepo *repo.AutonomyDecisionRepo) *Replanner {
	return &Replanner{decisionRepo: decisionRepo}
}

// Evaluate 评估是否需要重规划，生成建议并写入决策表。
func (r *Replanner) Evaluate(ctx context.Context, input *ReplanInput) (*ReplanRecommendation, error) {
	// 检查重规划次数限制（同一项目最多 2 次全量重规划）
	replanCount, _ := r.decisionRepo.CountByType(ctx, input.WorkflowRunID, DecisionReplan)
	if replanCount >= 2 {
		g.Log().Warningf(ctx, "[Replanner] 已达重规划上限(2次): workflowRun=%d", input.WorkflowRunID)
		return &ReplanRecommendation{
			Action:    ReplanAbort,
			Reasoning: "已达最大重规划次数限制(2次)，建议人工介入",
		}, nil
	}

	// 获取项目 architect 角色的模型配置
	modelInfo, err := resolveProjectModel(ctx, input.ProjectID, "architect")
	if err != nil {
		g.Log().Warningf(ctx, "[Replanner] 获取模型配置失败(降级): %v", err)
		return &ReplanRecommendation{
			Action:    ReplanAbort,
			Reasoning: "无法获取 AI 模型配置: " + err.Error(),
		}, nil
	}

	// 创建 Provider
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
	})
	if err != nil {
		return &ReplanRecommendation{
			Action:    ReplanAbort,
			Reasoning: "AI Provider 初始化失败: " + err.Error(),
		}, nil
	}

	// 构建 prompt
	systemPrompt := buildReplanSystemPrompt()
	userPrompt := buildReplanUserPrompt(input)

	// 调用 LLM
	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     []provider.Message{{Role: provider.RoleUser, Content: userPrompt}},
		MaxTokens:    2000,
		Temperature:  0.3,
		SystemPrompt: systemPrompt,
	})
	if err != nil {
		g.Log().Warningf(ctx, "[Replanner] LLM 调用失败(降级): %v", err)
		return &ReplanRecommendation{
			Action:    ReplanAbort,
			Reasoning: "LLM 调用失败: " + err.Error(),
		}, nil
	}

	// 解析返回
	rec, err := parseReplanResponse(resp.Content)
	if err != nil {
		g.Log().Warningf(ctx, "[Replanner] 解析 LLM 返回失败: %v, content=%s", err, resp.Content)
		return &ReplanRecommendation{
			Action:    ReplanAbort,
			Reasoning: "LLM 返回格式异常",
		}, nil
	}

	// 写入决策记录（模式跟随全局配置）
	mode := GetAutonomyMode(ctx)
	humanAction := ActionPending
	if mode == ModeAuto {
		humanAction = ActionApproved
	}
	triggerCtx, _ := json.Marshal(input)
	recommendation, _ := json.Marshal(rec)
	decisionID, err := r.decisionRepo.Create(ctx, g.Map{
		"workflow_run_id": input.WorkflowRunID,
		"project_id":      input.ProjectID,
		"decision_type":   DecisionReplan,
		"trigger_source":  input.TriggerSource,
		"trigger_context": string(triggerCtx),
		"recommendation":  string(recommendation),
		"decision_mode":   mode,
		"human_action":    humanAction,
	})
	if err != nil {
		g.Log().Warningf(ctx, "[Replanner] 写入决策记录失败: %v", err)
	}

	g.Log().Infof(ctx, "[Replanner] 重规划建议已生成: workflowRun=%d action=%s decisionID=%d",
		input.WorkflowRunID, rec.Action, decisionID)

	return rec, nil
}

// replanModelInfo 内部模型信息结构（复用 judge.go 的模式）。
type replanModelInfo struct {
	ModelCode    string
	ProviderType string
	BaseURL      string
	APIKey       string
	APISecret    string
}

// resolveProjectModel 获取项目指定角色的模型配置。
func resolveProjectModel(ctx context.Context, projectID int64, roleType string) (*replanModelInfo, error) {
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
		return nil, fmt.Errorf("项目 %d 未配置 %s 角色", projectID, roleType)
	}

	modelID := role["model_id"].Int64()
	model, err := g.DB().Model("ai_model m").
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.model_code, pv.provider_type, pv.base_url, p.api_key, p.api_secret").
		Where("m.id", modelID).
		Where("m.deleted_at IS NULL").
		One()
	if err != nil {
		return nil, fmt.Errorf("查询模型信息失败: %w", err)
	}
	if model.IsEmpty() {
		return nil, fmt.Errorf("AI 模型 %d 不存在", modelID)
	}

	return &replanModelInfo{
		ModelCode:    model["model_code"].String(),
		ProviderType: model["provider_type"].String(),
		BaseURL:      model["base_url"].String(),
		APIKey:       model["api_key"].String(),
		APISecret:    model["api_secret"].String(),
	}, nil
}

func buildReplanSystemPrompt() string {
	return `你是软件项目管理专家，负责在项目执行遇到系统性困难时给出重规划建议。

你必须严格输出以下 JSON 格式，不要输出其他内容：
{"action": "replan_partial|replan_full|abort", "affected_task_ids": [任务ID列表（仅 replan_partial 时需要）], "new_plan_summary": "新方案摘要", "reasoning": "判断理由"}

action 判断标准：
- replan_partial：仅部分任务需要调整（如依赖错误、接口不匹配），其他任务不受影响
- replan_full：整体方案方向有误或大范围架构需要调整，需要全部重做
- abort：项目目标不可行或缺少关键前置条件，建议终止`
}

func buildReplanUserPrompt(input *ReplanInput) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## 重规划触发\n- 触发源：%s\n- 工作流ID：%d\n", input.TriggerSource, input.WorkflowRunID))
	if input.BreakReason != "" {
		b.WriteString(fmt.Sprintf("- 熔断原因：%s\n", input.BreakReason))
	}
	b.WriteString("\n")

	b.WriteString("## 失败任务列表\n")
	if len(input.FailedTasks) == 0 {
		b.WriteString("无失败任务信息。\n")
	} else {
		for _, t := range input.FailedTasks {
			b.WriteString(fmt.Sprintf("- 任务 %d [%s]: %s (重试 %d 次)\n",
				t.TaskID, t.TaskName, t.ErrorMessage, t.RetryCount))
		}
	}

	if len(input.AcceptIssues) > 0 {
		b.WriteString("\n## 验收问题\n")
		for _, issue := range input.AcceptIssues {
			b.WriteString("- " + issue + "\n")
		}
	}

	b.WriteString("\n请综合分析以上信息，输出 JSON 格式的重规划建议。")
	return b.String()
}

func parseReplanResponse(content string) (*ReplanRecommendation, error) {
	content = strings.TrimSpace(content)

	var rec ReplanRecommendation
	if err := json.Unmarshal([]byte(content), &rec); err == nil {
		return validateReplanResult(&rec), nil
	}

	// 从代码块提取
	if idx := strings.Index(content, "{"); idx >= 0 {
		if end := strings.LastIndex(content, "}"); end > idx {
			if err := json.Unmarshal([]byte(content[idx:end+1]), &rec); err == nil {
				return validateReplanResult(&rec), nil
			}
		}
	}

	return nil, fmt.Errorf("无法解析 LLM 返回的 JSON")
}

func validateReplanResult(r *ReplanRecommendation) *ReplanRecommendation {
	switch r.Action {
	case ReplanPartial, ReplanFull, ReplanAbort:
		// valid
	default:
		r.Action = ReplanAbort
	}
	return r
}
