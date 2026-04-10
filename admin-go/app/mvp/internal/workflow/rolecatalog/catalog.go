package rolecatalog

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/configstore"
)

const (
	ConfigKeyRoleDefinitions  = "workflow.role_definitions"
	ConfigPathRoleDefinitions = "workflow.roleDefinitions"
)

type Definition struct {
	RoleType            string   `json:"roleType"`
	DisplayName         string   `json:"displayName"`
	Color               string   `json:"color,omitempty"`
	Description         string   `json:"description,omitempty"`
	PreferredLevels     []string `json:"preferredLevels,omitempty"`
	DefaultSystemPrompt string   `json:"defaultSystemPrompt,omitempty"`
	AcceptanceJudge     bool     `json:"acceptanceJudge,omitempty"`
	Sort                int      `json:"sort,omitempty"`
}

type ConfigReader interface {
	GetValueByKey(ctx context.Context, key string) (string, error)
}

var defaultConfigReader ConfigReader = configstore.NewStore()

func ListDefinitions(ctx context.Context) []Definition {
	definitions := defaultDefinitions()
	configured := loadConfiguredDefinitions(ctx, defaultConfigReader)
	return mergeDefinitions(definitions, configured)
}

func GetDefinition(ctx context.Context, roleType string) (Definition, bool) {
	roleType = strings.TrimSpace(roleType)
	if roleType == "" {
		return Definition{}, false
	}
	for _, definition := range ListDefinitions(ctx) {
		if definition.RoleType == roleType {
			return definition, true
		}
	}
	return Definition{}, false
}

func loadConfiguredDefinitions(ctx context.Context, reader ConfigReader) []Definition {
	raw := strings.TrimSpace(loadRawConfig(ctx, reader))
	if raw == "" {
		return nil
	}

	var definitions []Definition
	if err := json.Unmarshal([]byte(raw), &definitions); err != nil {
		g.Log().Warningf(ctx, "[RoleCatalog] 解析 workflow.role_definitions 失败: %v", err)
		return nil
	}
	return normalizeDefinitions(definitions)
}

func loadRawConfig(ctx context.Context, reader ConfigReader) string {
	if reader != nil {
		if value, err := reader.GetValueByKey(ctx, ConfigKeyRoleDefinitions); err == nil && strings.TrimSpace(value) != "" {
			return value
		}
	}

	cfgVal := g.Cfg().MustGet(ctx, ConfigPathRoleDefinitions)
	if cfgVal != nil && !cfgVal.IsEmpty() {
		return cfgVal.String()
	}
	return ""
}

func mergeDefinitions(base []Definition, overrides []Definition) []Definition {
	result := make([]Definition, 0, len(base)+len(overrides))
	index := make(map[string]int, len(base)+len(overrides))

	for _, definition := range normalizeDefinitions(base) {
		index[definition.RoleType] = len(result)
		result = append(result, definition)
	}
	for _, definition := range normalizeDefinitions(overrides) {
		if idx, ok := index[definition.RoleType]; ok {
			result[idx] = mergeDefinition(result[idx], definition)
			continue
		}
		index[definition.RoleType] = len(result)
		result = append(result, definition)
	}

	sort.SliceStable(result, func(i, j int) bool {
		left := result[i].Sort
		right := result[j].Sort
		if left == right {
			return result[i].RoleType < result[j].RoleType
		}
		return left < right
	})
	return result
}

func mergeDefinition(base Definition, override Definition) Definition {
	if strings.TrimSpace(override.DisplayName) != "" {
		base.DisplayName = strings.TrimSpace(override.DisplayName)
	}
	if strings.TrimSpace(override.Color) != "" {
		base.Color = strings.TrimSpace(override.Color)
	}
	if strings.TrimSpace(override.Description) != "" {
		base.Description = strings.TrimSpace(override.Description)
	}
	if len(override.PreferredLevels) > 0 {
		base.PreferredLevels = uniqueLevels(override.PreferredLevels)
	}
	if strings.TrimSpace(override.DefaultSystemPrompt) != "" {
		base.DefaultSystemPrompt = strings.TrimSpace(override.DefaultSystemPrompt)
	}
	base.AcceptanceJudge = override.AcceptanceJudge
	if override.Sort != 0 {
		base.Sort = override.Sort
	}
	return base
}

func normalizeDefinitions(definitions []Definition) []Definition {
	result := make([]Definition, 0, len(definitions))
	for _, definition := range definitions {
		definition.RoleType = strings.TrimSpace(definition.RoleType)
		if definition.RoleType == "" {
			continue
		}
		definition.DisplayName = strings.TrimSpace(definition.DisplayName)
		if definition.DisplayName == "" {
			definition.DisplayName = definition.RoleType
		}
		definition.Color = strings.TrimSpace(definition.Color)
		definition.Description = strings.TrimSpace(definition.Description)
		definition.DefaultSystemPrompt = strings.TrimSpace(definition.DefaultSystemPrompt)
		definition.PreferredLevels = uniqueLevels(definition.PreferredLevels)
		result = append(result, definition)
	}
	return result
}

func uniqueLevels(levels []string) []string {
	seen := make(map[string]struct{}, len(levels))
	result := make([]string, 0, len(levels))
	for _, level := range levels {
		level = strings.TrimSpace(level)
		if level == "" {
			continue
		}
		if _, ok := seen[level]; ok {
			continue
		}
		seen[level] = struct{}{}
		result = append(result, level)
	}
	return result
}

func defaultDefinitions() []Definition {
	return []Definition{
		{RoleType: "architect", DisplayName: "架构师", Color: "purple", Description: "负责需求澄清、方案拆分与任务编排", PreferredLevels: []string{"max", "pro", "lite"}, Sort: 10},
		{RoleType: "implementer", DisplayName: "实现者", Color: "blue", Description: "负责功能实现、交付落地与验证闭环", PreferredLevels: []string{"pro", "max", "lite"}, Sort: 20},
		{RoleType: "auditor", DisplayName: "审核者", Color: "green", Description: "负责方案质量、约束一致性与风险审查", PreferredLevels: []string{"pro", "max", "lite"}, Sort: 30},
		{RoleType: "coordinator", DisplayName: "协调者", Color: "orange", Description: "负责依赖同步、节奏控制与跨角色协同", PreferredLevels: []string{"lite", "pro", "max"}, Sort: 40},
		{RoleType: "operator", DisplayName: "运维恢复师", Color: "cyan", Description: "负责故障恢复、环境健康与风险处置", PreferredLevels: []string{"pro", "max", "lite"}, Sort: 50},
		{RoleType: "experience_reviewer", DisplayName: "体验评审师", Color: "magenta", Description: "负责验收阶段的产品体验、UI 交互与关键路径评审", PreferredLevels: []string{"max", "pro", "lite"}, AcceptanceJudge: true, Sort: 60},
	}
}

func findBuiltinDefinition(roleType string) (Definition, bool) {
	roleType = strings.TrimSpace(roleType)
	if roleType == "" {
		return Definition{}, false
	}
	for _, definition := range defaultDefinitions() {
		if definition.RoleType == roleType {
			return definition, true
		}
	}
	return Definition{}, false
}
