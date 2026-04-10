package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/provider"
)

var (
	planMetaPrefixRe     = `(?:"plan_meta"\s*:\s*\{[\s\S]*?\}\s*,\s*)?`
	tasksRe              = regexp.MustCompile(`(?s)\{\s*` + planMetaPrefixRe + `"tasks"\s*:\s*\[.*?\]\s*\}`)
	tasksGreedyRe        = regexp.MustCompile(`(?s)\{\s*` + planMetaPrefixRe + `"tasks"\s*:\s*\[[\s\S]*\]\s*\}`)
	arrayRe              = regexp.MustCompile(`(?s)\[\s*\{[\s\S]*\}\s*\]`)
	multiCommentRe       = regexp.MustCompile(`(?s)/\*.*?\*/`)
	trailingCommaRe      = regexp.MustCompile(`,\s*([\]\}])`)
	codeBlockReParserExt = regexp.MustCompile("(?s)```(?:json)?\\s*\\n?([\\{\\[][\\s\\S]*?[\\}\\]])\\s*```")
)

type architectPlanBlock struct {
	Plan *ArchitectTaskPlan
}

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
	plan, _, err := p.extractTaskPlanWithReport(text)
	return plan, err
}

func (p *TaskParser) extractTaskPlanWithReport(text string) (*ArchitectTaskPlan, *ArchitectPlanParseReport, error) {
	var allTasks []ArchitectTask
	var (
		report           = &ArchitectPlanParseReport{}
		chunkedBlocks    = make(map[int]architectPlanBlock)
		standaloneBlocks []architectPlanBlock
	)

	// 策略1：从所有 ```json 代码块中提取并合并
	// 先按 ``` 分割找代码块，再从中提取 JSON，避免正则嵌套匹配问题
	matches := extractJSONFromCodeBlocks(text)
	report.CodeBlockCount = len(matches)
	for _, jsonStr := range matches {
		plan, err := p.tryParseJSON(jsonStr)
		if err != nil || len(plan.Tasks) == 0 {
			report.InvalidCodeBlockCount++
			continue
		}
		report.ParsedBlockCount++
		report.RawTaskCount += len(plan.Tasks)
		report.mergePlanMeta(plan.PlanMeta)

		if plan.PlanMeta != nil && plan.PlanMeta.ChunkIndex != nil && *plan.PlanMeta.ChunkIndex > 0 {
			chunkIndex := *plan.PlanMeta.ChunkIndex
			if len(standaloneBlocks) > 0 {
				report.MixedChunkEncoding = true
			}
			if _, exists := chunkedBlocks[chunkIndex]; exists {
				report.ReplacedChunkIndexes = append(report.ReplacedChunkIndexes, chunkIndex)
			}
			chunkedBlocks[chunkIndex] = architectPlanBlock{
				Plan: plan,
			}
			continue
		}
		if len(chunkedBlocks) > 0 {
			report.MixedChunkEncoding = true
		}
		standaloneBlocks = append(standaloneBlocks, architectPlanBlock{
			Plan: plan,
		})
	}

	// 如果从代码块中已收集到任务，直接返回（最常见的分段输出场景）
	if len(chunkedBlocks) > 0 || len(standaloneBlocks) > 0 {
		if len(chunkedBlocks) > 0 {
			indexes := make([]int, 0, len(chunkedBlocks))
			for chunkIndex := range chunkedBlocks {
				indexes = append(indexes, chunkIndex)
			}
			sort.Ints(indexes)
			for _, chunkIndex := range indexes {
				report.ChunkIndexes = append(report.ChunkIndexes, chunkIndex)
				allTasks = append(allTasks, chunkedBlocks[chunkIndex].Plan.Tasks...)
			}
		}
		for _, block := range standaloneBlocks {
			allTasks = append(allTasks, block.Plan.Tasks...)
		}
		report.RawTaskCount = len(allTasks)
		report.finalize()
		return &ArchitectTaskPlan{
			PlanMeta: &ArchitectPlanMeta{
				PlanID:        report.PlanID,
				DeclaredTotal: report.DeclaredTotal,
				ChunkTotal:    report.ChunkTotal,
			},
			Tasks: allTasks,
		}, report, nil
	}

	// 策略2：查找所有 { "tasks": [...] } 格式的 JSON 块
	tasksMatches := tasksRe.FindAllString(text, -1)
	for _, m := range tasksMatches {
		if plan, err := p.tryParseJSON(m); err == nil && len(plan.Tasks) > 0 {
			report.ParsedBlockCount++
			allTasks = append(allTasks, plan.Tasks...)
			report.mergePlanMeta(plan.PlanMeta)
			report.RawTaskCount += len(plan.Tasks)
		}
	}
	if len(allTasks) > 0 {
		report.finalize()
		return &ArchitectTaskPlan{Tasks: allTasks}, report, nil
	}

	// 策略3：贪婪匹配单个大 JSON（兼容旧的单块输出）
	if m := tasksGreedyRe.FindString(text); m != "" {
		if plan, err := p.tryParseJSON(m); err == nil && len(plan.Tasks) > 0 {
			report.ParsedBlockCount = 1
			report.mergePlanMeta(plan.PlanMeta)
			report.RawTaskCount = len(plan.Tasks)
			report.finalize()
			return plan, report, nil
		}
	}

	// 策略4：查找独立的 JSON 数组 [{ ... }]（直接就是 tasks 数组）
	if m := arrayRe.FindString(text); m != "" {
		cleaned := p.cleanJSON(m)
		var tasks []ArchitectTask
		if err := json.Unmarshal([]byte(cleaned), &tasks); err == nil && len(tasks) > 0 {
			report.ParsedBlockCount = 1
			report.RawTaskCount = len(tasks)
			report.finalize()
			return &ArchitectTaskPlan{Tasks: tasks}, report, nil
		}
	}

	// 没找到有效的任务清单
	report.finalize()
	return &ArchitectTaskPlan{}, report, nil
}

// tryParseJSON 尝试解析 JSON 为 ArchitectTaskPlan
// 如果首次解析失败，会逐级尝试修复常见的 AI 输出格式问题后重试。
func (p *TaskParser) tryParseJSON(jsonStr string) (*ArchitectTaskPlan, error) {
	jsonStr = p.cleanJSON(jsonStr)

	var plan ArchitectTaskPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err == nil {
		return &plan, nil
	}

	// 第一级修复：修复多重转义
	fixed1 := fixBrokenJSON(jsonStr)
	if fixed1 != jsonStr {
		if err := json.Unmarshal([]byte(fixed1), &plan); err == nil {
			return &plan, nil
		}
	}

	// 第二级修复：更激进——将所有 \\" 替换为空（去除字符串值中的错误转义引号）
	fixed2 := strings.ReplaceAll(jsonStr, `\\"`, ``)
	if fixed2 != jsonStr {
		if err := json.Unmarshal([]byte(fixed2), &plan); err == nil {
			return &plan, nil
		}
	}

	// 第三级修复：将 \\" 替换为 '（用单引号替代转义引号）
	fixed3 := strings.ReplaceAll(jsonStr, `\\"`, `'`)
	if fixed3 != jsonStr {
		if err := json.Unmarshal([]byte(fixed3), &plan); err == nil {
			return &plan, nil
		}
	}

	return nil, fmt.Errorf("JSON 解析失败，已尝试多级修复")
}

// fixBrokenJSON 修复 AI 输出中常见的 JSON 格式问题
func fixBrokenJSON(s string) string {
	// 1. 修复多重转义：\\\" → \"（三重转义降为一重）
	//    AI 常见问题：在 JSON 字符串值中对引号过度转义
	s = strings.ReplaceAll(s, `\\\"`, `\"`)

	// 2. 修复 \\\\ → \\（四重反斜杠降为两重）
	for strings.Contains(s, `\\\\`) {
		s = strings.ReplaceAll(s, `\\\\`, `\\`)
	}

	// 3. 修复半角/全角引号混用
	s = strings.ReplaceAll(s, "\u201c", `"`) // 左双引号 "
	s = strings.ReplaceAll(s, "\u201d", `"`) // 右双引号 "
	s = strings.ReplaceAll(s, "\u2018", `'`) // 左单引号 '
	s = strings.ReplaceAll(s, "\u2019", `'`) // 右单引号 '

	return s
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
	s = multiCommentRe.ReplaceAllString(s, "")

	// 移除尾随逗号（数组或对象最后一个元素后的逗号）
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

	prompt := fmt.Sprintf(`你是一个任务提取助手。请基于“项目目标 / 最近用户要求 / 架构师回复”三个上下文，提取能够直接交付目标的最终任务，并转换为标准 JSON 任务列表。

项目分类：%s

规则：
1. 每个任务必须有 name 和 description
2. 有批次/阶段信息映射到 batch_no（从1开始）
3. 小说/创意类项目：每个"卷"或"阶段"作为一个任务，章节范围和概要放在 description 中
4. 任务数量控制在 2-30 个，太细的合并
5. 忽略 tool_call / invoke / cli-mcp-server_run_command / XML 标签 / 命令探测痕迹，这些不是任务
6. 不要保留“查看目录 / 读取文件 / 环境分析”这类无交付占位任务，除非它本身就是明确交付物
7. 如果架构师回复只包含勘察动作，也要回到项目目标和最近用户要求，推导出最小可交付任务清单
8. 只输出 JSON，不要其他内容

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
	matches := codeBlockReParserExt.FindAllStringSubmatch(text, -1)
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
		Fields("m.model_code, m.max_tokens, pv.provider_type, pv.supported_protocols, pv.base_url, p.api_key, p.api_secret, m.role_prompt").
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
		ModelCode:          model["model_code"].String(),
		ProviderType:       model["provider_type"].String(),
		SupportedProtocols: decodeProviderProtocols(model["supported_protocols"].String(), model["provider_type"].String()),
		BaseURL:            model["base_url"].String(),
		APIKey:             model["api_key"].String(),
		APISecret:          model["api_secret"].String(),
		SystemPrompt:       sp,
		MaxTokens:          model["max_tokens"].Int(),
	}
}

// callAIForJSON 调用 AI 获取 JSON 结果（非流式）。
func callAIForJSON(ctx context.Context, modelInfo *ModelInfo, prompt string) (string, error) {
	p, err := provider.GetProvider(provider.Config{
		ProviderType:       modelInfo.ProviderType,
		SupportedProtocols: modelInfo.SupportedProtocols,
		BaseURL:            modelInfo.BaseURL,
		APIKey:             modelInfo.APIKey,
		APISecret:          modelInfo.APISecret,
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
