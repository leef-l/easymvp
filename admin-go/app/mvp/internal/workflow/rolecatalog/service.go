package rolecatalog

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"easymvp/app/mvp/internal/workflow/configstore"
)

type ConfigWriter interface {
	ConfigReader
	DeleteByKey(ctx context.Context, key string) error
	UpsertByKey(ctx context.Context, in configstore.UpsertInput) error
}

type Service struct {
	store ConfigWriter
}

var roleTypePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)

func NewService(store ConfigWriter) *Service {
	return &Service{store: store}
}

func DefaultService() *Service {
	return NewService(configstore.NewStore())
}

func (s *Service) List(ctx context.Context) []Definition {
	if s == nil {
		return ListDefinitions(ctx)
	}
	return mergeDefinitions(defaultDefinitions(), loadConfiguredDefinitions(ctx, s.store))
}

func (s *Service) Save(ctx context.Context, definitions []Definition, createdBy, deptID int64) error {
	if s == nil || s.store == nil {
		return fmt.Errorf("role definition service store is not configured")
	}
	normalized, err := normalizeForSave(definitions)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("序列化角色定义失败: %w", err)
	}
	return s.store.UpsertByKey(ctx, configstore.UpsertInput{
		Key:         ConfigKeyRoleDefinitions,
		Value:       string(payload),
		ConfigType:  "json",
		Category:    "engine",
		Description: "项目角色定义注册表(JSON)，供项目角色、角色预设和验收角色选择使用",
		CreatedBy:   createdBy,
		DeptID:      deptID,
	})
}

func normalizeForSave(definitions []Definition) ([]Definition, error) {
	normalized := make([]Definition, 0, len(definitions))
	seen := make(map[string]struct{}, len(definitions))
	for index, definition := range normalizeDefinitions(definitions) {
		if definition.RoleType == "" {
			return nil, fmt.Errorf("第 %d 条角色定义缺少 roleType", index+1)
		}
		if definition.DisplayName == "" {
			return nil, fmt.Errorf("角色 %s 缺少展示名", definition.RoleType)
		}
		if _, ok := seen[definition.RoleType]; ok {
			return nil, fmt.Errorf("角色类型 %s 重复，请保持唯一", definition.RoleType)
		}
		if !roleTypePattern.MatchString(definition.RoleType) {
			return nil, fmt.Errorf("角色类型 %s 不合法，只允许小写字母、数字、下划线，且必须以字母开头", definition.RoleType)
		}
		for _, level := range definition.PreferredLevels {
			switch level {
			case "lite", "pro", "max":
			default:
				return nil, fmt.Errorf("角色 %s 的推荐等级 %s 不合法，仅支持 lite/pro/max", definition.RoleType, level)
			}
		}
		if builtin, ok := findBuiltinDefinition(definition.RoleType); ok {
			if definition.Color == "" {
				definition.Color = builtin.Color
			}
			if definition.Description == "" {
				definition.Description = builtin.Description
			}
			if len(definition.PreferredLevels) == 0 {
				definition.PreferredLevels = builtin.PreferredLevels
			}
			if definition.Sort == 0 {
				definition.Sort = builtin.Sort
			}
		}
		if definition.Color == "" {
			definition.Color = "default"
		}
		if definition.Sort == 0 {
			definition.Sort = (index + 1) * 10
		}
		seen[definition.RoleType] = struct{}{}
		normalized = append(normalized, definition)
	}

	sort.SliceStable(normalized, func(i, j int) bool {
		left := normalized[i].Sort
		right := normalized[j].Sort
		if left == right {
			return normalized[i].RoleType < normalized[j].RoleType
		}
		return left < right
	})
	return normalized, nil
}

func MergeConfigPrompt(basePrompt string, definition Definition) string {
	if strings.TrimSpace(definition.DefaultSystemPrompt) == "" {
		return basePrompt
	}
	return strings.TrimSpace(definition.DefaultSystemPrompt)
}

func (s *Service) Reset(ctx context.Context) error {
	if s == nil || s.store == nil {
		return fmt.Errorf("role definition service store is not configured")
	}
	return s.store.DeleteByKey(ctx, ConfigKeyRoleDefinitions)
}
