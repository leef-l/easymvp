package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// normalizeTasks 标准化和校验任务列表
func (p *TaskParser) normalizeTasks(ctx context.Context, tasks []ArchitectTask, projectCategory string) []ArchitectTask {
	normalized, _ := p.normalizeTasksWithReport(ctx, tasks, projectCategory)
	return normalized
}

func (p *TaskParser) normalizeTasksWithReport(ctx context.Context, tasks []ArchitectTask, projectCategory string) ([]ArchitectTask, *TaskNormalizationReport) {
	var (
		normalizedRev []ArchitectTask
		family        = GetCategoryFamily(projectCategory)
		seenNames     = make(map[string]struct{}, len(tasks))
		report        = &TaskNormalizationReport{}
	)

	for i := len(tasks) - 1; i >= 0; i-- {
		task := tasks[i]
		task.Name = strings.TrimSpace(task.Name)
		task.Description = strings.TrimSpace(task.Description)
		task.RoleType = strings.TrimSpace(task.RoleType)
		task.RoleLevel = strings.TrimSpace(task.RoleLevel)
		task.ParentName = strings.TrimSpace(task.ParentName)

		if task.Name == "" {
			report.EmptyNameDropped++
			g.Log().Warning(ctx, "[TaskParser] 跳过空任务名的任务项")
			continue
		}

		// 分类感知校验
		switch family {
		case CategoryFamilyCoding:
			if task.RoleType == "" {
				task.RoleType = "implementer"
			}
		case CategoryFamilyCreative:
			if task.RoleType == "" {
				task.RoleType = "implementer"
			}
			if len(task.AffectedResources) == 0 && task.RoleType == "implementer" {
				task.AffectedResources = []string{fmt.Sprintf("content/%s.md", task.Name)}
			}
		case CategoryFamilyAnalysis:
			if task.RoleType == "" {
				task.RoleType = "implementer"
			}
			if len(task.AffectedResources) == 0 && task.RoleType == "implementer" {
				task.AffectedResources = []string{fmt.Sprintf("output/%s.md", task.Name)}
			}
		}

		validLevels := map[string]bool{"lite": true, "pro": true, "max": true}
		if !validLevels[task.RoleLevel] {
			task.RoleLevel = "pro"
		}
		if isExplorationPlaceholderTask(task) {
			report.PlaceholderDropped = append(report.PlaceholderDropped, task.Name)
			g.Log().Warningf(ctx, "[TaskParser] 跳过无交付占位任务: %s", task.Name)
			continue
		}
		if _, exists := seenNames[task.Name]; exists {
			report.DuplicateDropped = append(report.DuplicateDropped, task.Name)
			g.Log().Warningf(ctx, "[TaskParser] 跳过重复任务名（已由后发块覆盖）: %s", task.Name)
			continue
		}

		seenNames[task.Name] = struct{}{}
		normalizedRev = append(normalizedRev, task)
	}

	normalized := make([]ArchitectTask, 0, len(normalizedRev))
	for i := len(normalizedRev) - 1; i >= 0; i-- {
		normalized = append(normalized, normalizedRev[i])
	}
	return normalized, report
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
