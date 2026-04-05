package acceptance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/provider"
)

// JudgeResult LLM 质量评审结果。
type JudgeResult struct {
	QualityScore float64  `json:"quality_score"` // 0-100
	Conclusion   string   `json:"conclusion"`    // passed / failed / uncertain
	Summary      string   `json:"summary"`
	Suggestions  []string `json:"suggestions,omitempty"`
}

// Judge LLM 质量评审员，基于证据和规则命中结果调用 LLM 输出质量判断。
type Judge struct{}

// NewJudge 创建 LLM Judge。
func NewJudge() *Judge {
	return &Judge{}
}

// Evaluate 调用 LLM 评审验收结果。
// 复用项目 architect 角色的 AI 模型配置。
// 容错：LLM 调用失败时返回 uncertain（降级为人工审核）。
func (j *Judge) Evaluate(ctx context.Context, in *AcceptContext, evidence []EvidenceItem, hits []RuleHit) (*JudgeResult, error) {
	// 1. 获取项目 architect 角色的模型配置
	modelInfo, err := resolveProjectModel(ctx, in.ProjectID, "architect")
	if err != nil {
		g.Log().Warningf(ctx, "[Judge] 获取模型配置失败(降级): %v", err)
		return &JudgeResult{Conclusion: "uncertain", Summary: "无法获取 AI 模型配置"}, nil
	}

	// 2. 创建 Provider
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
	})
	if err != nil {
		g.Log().Warningf(ctx, "[Judge] 创建 Provider 失败(降级): %v", err)
		return &JudgeResult{Conclusion: "uncertain", Summary: "AI Provider 初始化失败"}, nil
	}

	// 3. 构建 prompt
	systemPrompt := buildJudgeSystemPrompt()
	userPrompt := buildJudgeUserPrompt(in, evidence, hits)

	// 4. 非流式调用 LLM
	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     []provider.Message{{Role: provider.RoleUser, Content: userPrompt}},
		MaxTokens:    2000,
		Temperature:  0.3,
		SystemPrompt: systemPrompt,
	})
	if err != nil {
		g.Log().Warningf(ctx, "[Judge] LLM 调用失败(降级): %v", err)
		return &JudgeResult{Conclusion: "uncertain", Summary: "LLM 调用失败: " + err.Error()}, nil
	}

	// 5. 解析 JSON 返回
	result, err := parseJudgeResponse(resp.Content)
	if err != nil {
		g.Log().Warningf(ctx, "[Judge] 解析 LLM 返回失败(降级): %v, content=%s", err, resp.Content)
		return &JudgeResult{Conclusion: "uncertain", Summary: "LLM 返回格式异常"}, nil
	}

	g.Log().Infof(ctx, "[Judge] 评审完成: score=%.1f conclusion=%s project=%d",
		result.QualityScore, result.Conclusion, in.ProjectID)
	return result, nil
}

// judgeModelInfo 内部模型信息结构。
type judgeModelInfo struct {
	ModelCode    string
	ProviderType string
	BaseURL      string
	APIKey       string
	APISecret    string
}

// resolveProjectModel 获取项目指定角色的模型配置（不依赖 engine 包，避免循环依赖）。
func resolveProjectModel(ctx context.Context, projectID int64, roleType string) (*judgeModelInfo, error) {
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

	return &judgeModelInfo{
		ModelCode:    model["model_code"].String(),
		ProviderType: model["provider_type"].String(),
		BaseURL:      model["base_url"].String(),
		APIKey:       model["api_key"].String(),
		APISecret:    model["api_secret"].String(),
	}, nil
}

func buildJudgeSystemPrompt() string {
	return `你是软件项目质量评审专家。根据提供的验收证据和规则检查结果，综合评估项目质量。

你必须严格输出以下 JSON 格式，不要输出其他内容：
{"quality_score": 0-100, "conclusion": "passed|failed|uncertain", "summary": "一句话质量评语", "suggestions": ["改进建议1", "改进建议2"]}

评分标准：
- 90-100：优秀，所有核心功能完整，代码质量高
- 70-89：良好，核心功能完整，有少量问题
- 60-69：及格，基本功能完成，存在明显不足
- 0-59：不合格，缺失关键功能或存在严重问题

conclusion 判断标准：
- passed：评分 >= 70 且无关键问题
- failed：评分 < 60 或存在关键未解决问题
- uncertain：信息不足以判断，建议人工审核`
}

func buildJudgeUserPrompt(in *AcceptContext, evidence []EvidenceItem, hits []RuleHit) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## 项目信息\n- 项目类型：%s\n- 工作目录：%s\n\n", in.ProjectType, in.WorkDir))

	// 规则检查结果
	b.WriteString("## 规则检查结果\n")
	if len(hits) == 0 {
		b.WriteString("所有硬规则通过，无命中问题。\n")
	} else {
		for _, h := range hits {
			b.WriteString(fmt.Sprintf("- [%s] %s: %s (期望: %s, 实际: %s)\n",
				h.Severity, h.RuleCode, h.Title, h.ExpectedValue, h.ActualValue))
		}
	}

	// 证据摘要
	b.WriteString("\n## 验收证据\n")
	if len(evidence) == 0 {
		b.WriteString("无收集到证据。\n")
	} else {
		for _, e := range evidence {
			summary := e.Summary
			if len(summary) > 500 {
				summary = summary[:500] + "..."
			}
			b.WriteString(fmt.Sprintf("- [%s] %s: %s\n", e.EvidenceType, e.SourceType, summary))
		}
	}

	b.WriteString("\n请综合以上信息，输出 JSON 格式的质量评审结论。")
	return b.String()
}

func parseJudgeResponse(content string) (*JudgeResult, error) {
	content = strings.TrimSpace(content)

	// 尝试直接解析
	var result JudgeResult
	if err := json.Unmarshal([]byte(content), &result); err == nil {
		return validateJudgeResult(&result), nil
	}

	// 尝试从 ```json 代码块中提取
	if idx := strings.Index(content, "{"); idx >= 0 {
		if end := strings.LastIndex(content, "}"); end > idx {
			if err := json.Unmarshal([]byte(content[idx:end+1]), &result); err == nil {
				return validateJudgeResult(&result), nil
			}
		}
	}

	return nil, fmt.Errorf("无法解析 LLM 返回的 JSON")
}

func validateJudgeResult(r *JudgeResult) *JudgeResult {
	if r.QualityScore < 0 {
		r.QualityScore = 0
	}
	if r.QualityScore > 100 {
		r.QualityScore = 100
	}
	switch r.Conclusion {
	case "passed", "failed", "uncertain":
		// valid
	default:
		r.Conclusion = "uncertain"
	}
	return r
}
