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

		seenNames[task.Name] = struct{}{}
		normalized = append(normalized, task)
	}

	return normalized
}
