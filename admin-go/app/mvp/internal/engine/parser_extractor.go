package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/provider"
)

// extractJSONFromCodeBlocks 从 markdown 代码块中提取 JSON 字符串。
// 使用字符串分割而非正则，避免嵌套 JSON 的 } 导致非贪婪匹配提前截断。
func extractJSONFromCodeBlocks(text string) []string {
	var results []string
	remaining := text
	for {
		// 找 ```json 或 ``` 开头
		startIdx := strings.Index(remaining, "```json")
		tagLen := 7
		if startIdx == -1 {
			startIdx = strings.Index(remaining, "```")
			tagLen = 3
		}
		if startIdx == -1 {
			break
		}
		// 跳过开始标记，找内容起点
		contentStart := startIdx + tagLen
		// 跳过开始标记后的空白和换行
		for contentStart < len(remaining) && (remaining[contentStart] == '\n' || remaining[contentStart] == '\r' || remaining[contentStart] == ' ') {
			contentStart++
		}
		// 找结束的 ```
		endIdx := strings.Index(remaining[contentStart:], "```")
		if endIdx == -1 {
			break
		}
		block := strings.TrimSpace(remaining[contentStart : contentStart+endIdx])
		remaining = remaining[contentStart+endIdx+3:]

		// 只提取以 { 开头的 JSON 对象
		if len(block) > 0 && block[0] == '{' {
			results = append(results, block)
		}
	}
	return results
}

// extractTaskPlan 从 AI 回复文本中提取 JSON 任务清单
// 支持多种格式，并自动合并分布在多个 JSON 代码块中的任务：
//  1. 标准 ```json ... ``` 代码块（支持多个，自动合并）
//  2. { "tasks": [...] } 直接 JSON（支持多个）
//  3. 混合文本中的 JSON 片段
//  4. 独立的 JSON 数组 [{ ... }]
func (p *TaskParser) extractTaskPlan(text string) (*ArchitectTaskPlan, error) {
	var allTasks []ArchitectTask

	// 策略1：从所有 ```json 代码块中提取并合并
	// 先按 ``` 分割找代码块，再从中提取 JSON，避免正则嵌套匹配问题
	matches := extractJSONFromCodeBlocks(text)
	for _, jsonStr := range matches {
		if plan, err := p.tryParseJSON(jsonStr); err == nil && len(plan.Tasks) > 0 {
			allTasks = append(allTasks, plan.Tasks...)
		}
	}

	// 如果从代码块中已收集到任务，直接返回（最常见的分段输出场景）
	if len(allTasks) > 0 {
		return &ArchitectTaskPlan{Tasks: allTasks}, nil
	}

	// 策略2：查找所有 { "tasks": [...] } 格式的 JSON 块
	tasksRe := regexp.MustCompile(`(?s)\{\s*"tasks"\s*:\s*\[.*?\]\s*\}`)
	tasksMatches := tasksRe.FindAllString(text, -1)
	for _, m := range tasksMatches {
		if plan, err := p.tryParseJSON(m); err == nil && len(plan.Tasks) > 0 {
			allTasks = append(allTasks, plan.Tasks...)
		}
	}
	if len(allTasks) > 0 {
		return &ArchitectTaskPlan{Tasks: allTasks}, nil
	}

	// 策略3：贪婪匹配单个大 JSON（兼容旧的单块输出）
	tasksGreedyRe := regexp.MustCompile(`(?s)\{\s*"tasks"\s*:\s*\[[\s\S]*\]\s*\}`)
	if m := tasksGreedyRe.FindString(text); m != "" {
		if plan, err := p.tryParseJSON(m); err == nil && len(plan.Tasks) > 0 {
			return plan, nil
		}
	}

	// 策略4：查找独立的 JSON 数组 [{ ... }]（直接就是 tasks 数组）
	arrayRe := regexp.MustCompile(`(?s)\[\s*\{[\s\S]*\}\s*\]`)
	if m := arrayRe.FindString(text); m != "" {
		cleaned := p.cleanJSON(m)
		var tasks []ArchitectTask
		if err := json.Unmarshal([]byte(cleaned), &tasks); err == nil && len(tasks) > 0 {
			return &ArchitectTaskPlan{Tasks: tasks}, nil
		}
	}

	// 没找到有效的任务清单
	return &ArchitectTaskPlan{}, nil
}

// tryParseJSON 尝试解析 JSON 为 ArchitectTaskPlan
func (p *TaskParser) tryParseJSON(jsonStr string) (*ArchitectTaskPlan, error) {
	jsonStr = p.cleanJSON(jsonStr)

	var plan ArchitectTaskPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

// cleanJSON 清理 AI 输出的非标准 JSON（注释、尾随逗号等）
func (p *TaskParser) cleanJSON(s string) string {
	s = strings.TrimSpace(s)

	// 移除单行注释 // ... （但不破坏字符串内的 URL 如 https://）
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = removeLineComment(line)
	}
	s = strings.Join(lines, "\n")

	// 移除多行注释 /* ... */
	multiCommentRe := regexp.MustCompile(`(?s)/\*.*?\*/`)
	s = multiCommentRe.ReplaceAllString(s, "")

	// 移除尾随逗号（数组或对象最后一个元素后的逗号）
	trailingCommaRe := regexp.MustCompile(`,\s*([\]\}])`)
	s = trailingCommaRe.ReplaceAllString(s, "$1")

	return s
}

// removeLineComment 移除一行中不在引号内的 // 注释
func removeLineComment(line string) string {
	inString := false
	escape := false
	for i, ch := range line {
		if escape {
			escape = false
			continue
		}
		if ch == '\\' && inString {
			escape = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if !inString && ch == '/' && i+1 < len(line) && line[i+1] == '/' {
			return line[:i]
		}
	}
	return line
}

// aiExtractTasks 使用 AI 从非标准格式的回复中提取任务列表。
// 适用于：跨消息 JSON 碎片、非 tasks 格式（如小说章节 chapters）、纯文本描述等。
func (p *TaskParser) aiExtractTasks(ctx context.Context, aiReply, projectCategory string) ([]ArchitectTask, error) {
	// 先提取 JSON 代码块的结构摘要（而非全文），大幅缩减 AI 输入
	content := p.extractJSONSummary(aiReply)
	if content == "" {
		// 没有 JSON 代码块，截取纯文本
		content = aiReply
		if len([]rune(content)) > 5000 {
			runes := []rune(content)
			content = string(runes[:3000]) + "\n\n...(省略)...\n\n" + string(runes[len(runes)-1500:])
		}
	}

	prompt := fmt.Sprintf(`你是一个任务提取助手。请从以下内容中提取所有任务/章节/模块，转换为标准 JSON 任务列表。

项目分类：%s

规则：
1. 每个任务必须有 name 和 description
2. 有批次/阶段信息映射到 batch_no（从1开始）
3. 小说/创意类项目：每个"卷"或"阶段"作为一个任务，章节范围和概要放在 description 中
4. 任务数量控制在 5-30 个，太细的合并
5. 只输出 JSON，不要其他内容

输出格式：
{"tasks": [{"name": "任务名", "description": "描述", "batch_no": 1, "affected_resources": [], "depends_on": []}]}

--- 内容 ---
%s`, projectCategory, content)

	modelInfo := getDefaultReviewModel(ctx)
	if modelInfo == nil {
		return nil, fmt.Errorf("无可用 AI 模型进行二次提取")
	}

	// 用独立 context，脱离 HTTP 请求生命周期（避免前端超时导致 cancel）
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	result, err := callAIForJSON(timeoutCtx, modelInfo, prompt)
	if err != nil {
		return nil, err
	}

	var plan ArchitectTaskPlan
	if err := parseJSONFromAI(result, &plan); err != nil {
		return nil, fmt.Errorf("AI 二次提取结果解析失败: %w", err)
	}
	return plan.Tasks, nil
}

// extractJSONSummary 从 AI 回复中提取 JSON 代码块的结构摘要。
// 只保留 key 结构和前几个元素，大幅缩减内容。
func (p *TaskParser) extractJSONSummary(text string) string {
	codeBlockRe := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?([\\{\\[][\\s\\S]*?[\\}\\]])\\s*```")
	matches := codeBlockRe.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return ""
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("共 %d 个 JSON 代码块：\n\n", len(matches)))

	for i, m := range matches {
		jsonStr := m[1]
		var raw map[string]json.RawMessage
		if json.Unmarshal([]byte(jsonStr), &raw) == nil {
			summary.WriteString(fmt.Sprintf("=== 代码块 %d ===\n", i+1))
			for key, val := range raw {
				valStr := string(val)
				if len(valStr) > 500 {
					if strings.HasPrefix(strings.TrimSpace(valStr), "[") {
						var arr []json.RawMessage
						if json.Unmarshal(val, &arr) == nil {
							summary.WriteString(fmt.Sprintf("  %s: 数组，共 %d 个元素\n", key, len(arr)))
							for j := 0; j < 3 && j < len(arr); j++ {
								elemStr := string(arr[j])
								if len(elemStr) > 200 {
									elemStr = elemStr[:200] + "..."
								}
								summary.WriteString(fmt.Sprintf("    [%d] %s\n", j, elemStr))
							}
							if len(arr) > 3 {
								lastStr := string(arr[len(arr)-1])
								if len(lastStr) > 200 {
									lastStr = lastStr[:200] + "..."
								}
								summary.WriteString(fmt.Sprintf("    [%d] %s\n", len(arr)-1, lastStr))
							}
							continue
						}
					}
					valStr = valStr[:500] + "..."
				}
				summary.WriteString(fmt.Sprintf("  %s: %s\n", key, valStr))
			}
		} else {
			if len(jsonStr) > 300 {
				jsonStr = jsonStr[:300] + "..."
			}
			summary.WriteString(fmt.Sprintf("=== 代码块 %d（解析失败）===\n%s\n", i+1, jsonStr))
		}
		summary.WriteString("\n")
	}

	return summary.String()
}

// getDefaultReviewModel 获取一个可用的 AI 模型用于辅助提取。
func getDefaultReviewModel(ctx context.Context) *ModelInfo {
	// 优先使用默认 auditor 预设，避免依赖项目级角色配置。
	preset, err := repo.GetRolePreset(ctx, repo.RolePresetQuery{
		RoleType:    "auditor",
		DefaultOnly: true,
	})
	if err != nil || preset == nil {
		return nil
	}

	model, err := g.DB().Model("ai_model m").Ctx(ctx).
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.model_code, m.max_tokens, pv.provider_type, pv.base_url, p.api_key, p.api_secret, m.role_prompt").
		Where("m.deleted_at IS NULL").
		Where("m.id", preset["model_id"].Int64()).
		One()
	if err != nil || model.IsEmpty() {
		return nil
	}

	sp := preset["system_prompt"].String()
	if sp == "" {
		sp = model["role_prompt"].String()
	}

	return &ModelInfo{
		ModelCode:    model["model_code"].String(),
		ProviderType: model["provider_type"].String(),
		BaseURL:      model["base_url"].String(),
		APIKey:       model["api_key"].String(),
		APISecret:    model["api_secret"].String(),
		SystemPrompt: sp,
		MaxTokens:    model["max_tokens"].Int(),
	}
}

// callAIForJSON 调用 AI 获取 JSON 结果（非流式）。
func callAIForJSON(ctx context.Context, modelInfo *ModelInfo, prompt string) (string, error) {
	p, err := provider.GetProvider(provider.Config{
		ProviderType: modelInfo.ProviderType,
		BaseURL:      modelInfo.BaseURL,
		APIKey:       modelInfo.APIKey,
		APISecret:    modelInfo.APISecret,
	})
	if err != nil {
		return "", err
	}

	resp, err := p.Chat(ctx, &provider.ChatRequest{
		Model:        modelInfo.ModelCode,
		Messages:     []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens:    modelInfo.MaxTokens,
		Temperature:  0.2,
		SystemPrompt: "你是一个精确的 JSON 提取助手，只输出 JSON，不要输出其他内容。",
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
