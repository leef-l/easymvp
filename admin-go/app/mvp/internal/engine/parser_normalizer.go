package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// normalizeTasks 标准化和校验任务列表
func (p *TaskParser) normalizeTasks(ctx context.Context, tasks []ArchitectTask, projectCategory string) []ArchitectTask {
	normalized := make([]ArchitectTask, 0, len(tasks))
	seenNames := make(map[string]struct{}, len(tasks))
	family := GetCategoryFamily(projectCategory)

	for _, task := range tasks {
		task.Name = strings.TrimSpace(task.Name)
		task.Description = strings.TrimSpace(task.Description)
		task.RoleType = strings.TrimSpace(task.RoleType)
		task.RoleLevel = strings.TrimSpace(task.RoleLevel)
		task.ParentName = strings.TrimSpace(task.ParentName)

		if task.Name == "" {
			g.Log().Warning(ctx, "[TaskParser] 跳过空任务名的任务项")
			continue
		}
		if _, exists := seenNames[task.Name]; exists {
			g.Log().Warningf(ctx, "[TaskParser] 跳过重复任务名: %s", task.Name)
			continue
		}

		// affected_resources 格式验证
		task.AffectedResources = validateAffectedResources(ctx, task.Name, task.AffectedResources)

		// 分类感知校验
		switch family {
		case CategoryFamilyCoding:
			// 编码类：默认角色 implementer，affected_resources 应为代码路径
			if task.RoleType == "" {
				task.RoleType = "implementer"
			}
		case CategoryFamilyCreative:
			// 创意类：默认角色 implementer，affected_resources 为内容文件路径
			if task.RoleType == "" {
				task.RoleType = "implementer"
			}
			// 创意类任务如果没有 affected_resources，自动生成基于任务名的路径
			if len(task.AffectedResources) == 0 && task.RoleType == "implementer" {
				task.AffectedResources = []string{fmt.Sprintf("content/%s.md", task.Name)}
			}
		case CategoryFamilyAnalysis:
			// 分析类：默认角色 implementer
			if task.RoleType == "" {
				task.RoleType = "implementer"
			}
			if len(task.AffectedResources) == 0 && task.RoleType == "implementer" {
				task.AffectedResources = []string{fmt.Sprintf("output/%s.md", task.Name)}
			}
		}

		// RoleLevel 校验
		validLevels := map[string]bool{"lite": true, "pro": true, "max": true}
		if !validLevels[task.RoleLevel] {
			task.RoleLevel = "pro" // 默认 pro
		}
		if isExplorationPlaceholderTask(task) {
			g.Log().Warningf(ctx, "[TaskParser] 跳过无交付占位任务: %s", task.Name)
			continue
		}

		seenNames[task.Name] = struct{}{}
		normalized = append(normalized, task)
	}

	return normalized
}

// validateAffectedResources 验证并过滤无效的 affected_resources
func validateAffectedResources(ctx context.Context, taskName string, resources []string) []string {
	if len(resources) == 0 {
		return resources
	}

	valid := make([]string, 0, len(resources))
	for _, r := range resources {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		// 禁止根路径
		if r == "/" || r == "\\" {
			g.Log().Warningf(ctx, "[TaskParser] 任务 [%s] affected_resources 包含根路径，已跳过: %s", taskName, r)
			continue
		}

		// 禁止通配符
		if strings.Contains(r, "*") {
			g.Log().Warningf(ctx, "[TaskParser] 任务 [%s] affected_resources 包含通配符，已跳过: %s", taskName, r)
			continue
		}

		// 禁止目录路径（以 / 结尾）
		if strings.HasSuffix(r, "/") || strings.HasSuffix(r, "\\") {
			g.Log().Warningf(ctx, "[TaskParser] 任务 [%s] affected_resources 是目录路径，已跳过: %s", taskName, r)
			continue
		}

		// 禁止绝对路径（以 / 或 \ 或盘符开头）
		if strings.HasPrefix(r, "/") && !strings.Contains(r[1:], "/") {
			// 单独的 /path 形式，可能是根路径的变体
			if !strings.Contains(r[1:], ".") && !strings.Contains(r[1:], "/") {
				g.Log().Warningf(ctx, "[TaskParser] 任务 [%s] affected_resources 疑似根路径，已跳过: %s", taskName, r)
				continue
			}
		}

		valid = append(valid, r)
	}

	return valid
}

func isExplorationPlaceholderTask(task ArchitectTask) bool {
	combined := strings.ToLower(strings.TrimSpace(task.Name + "\n" + task.Description))
	if combined == "" {
		return false
	}

	toolMarkers := []string{
		"<minimax:tool_call>",
		"<invoke",
		"cli-mcp-server_run_command",
		"ls -la",
		"pwd",
		"find ",
		"grep ",
		"rg ",
		"cat ",
	}
	if containsAny(combined, toolMarkers) {
		return true
	}

	if len(task.AffectedResources) > 0 || len(task.DependsOn) > 0 {
		return false
	}

	explorationKeywords := []string{
		"查看项目目录",
		"查看目录",
		"目录结构",
		"读取文件",
		"了解现有代码",
		"了解代码结构",
		"熟悉代码",
		"环境分析",
		"项目勘察",
		"查看仓库",
		"检查目录",
		"scan the repo",
		"inspect the repo",
		"repo structure",
		"list files",
	}
	if !containsAny(combined, explorationKeywords) {
		return false
	}

	deliverableKeywords := []string{
		"方案",
		"计划",
		"文档",
		"清单",
		"报告",
		"设计",
		"修复",
		"实现",
		"接口",
		"测试",
		"验收",
		"脚本",
		"迁移",
		"优化",
		"deliverable",
		"plan",
		"spec",
		"report",
	}
	return !containsAny(combined, deliverableKeywords)
}

func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
