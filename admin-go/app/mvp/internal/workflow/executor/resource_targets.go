package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"easymvp/utility/worktreeguard"
)

type resourceTargets struct {
	AllowedPaths   []string
	FilePaths      []string
	DirectoryPaths []string
	Rejected       []string
}

func parseResourceTargets(jsonStr string) resourceTargets {
	result := resourceTargets{}
	if strings.TrimSpace(jsonStr) == "" || jsonStr == "null" {
		return result
	}

	var rawValues []string
	if err := json.Unmarshal([]byte(jsonStr), &rawValues); err != nil {
		result.Rejected = append(result.Rejected, strings.TrimSpace(jsonStr))
		return result
	}

	allowedSeen := make(map[string]struct{}, len(rawValues))
	fileSeen := make(map[string]struct{}, len(rawValues))
	dirSeen := make(map[string]struct{}, len(rawValues))

	for _, rawValue := range rawValues {
		normalized, ok := worktreeguard.NormalizeRelativePath(rawValue)
		if !ok {
			if trimmed := strings.TrimSpace(rawValue); trimmed != "" {
				result.Rejected = append(result.Rejected, trimmed)
			}
			continue
		}

		if _, exists := allowedSeen[normalized]; !exists {
			allowedSeen[normalized] = struct{}{}
			result.AllowedPaths = append(result.AllowedPaths, normalized)
		}

		if looksLikeDirectoryResource(rawValue) {
			if _, exists := dirSeen[normalized]; !exists {
				dirSeen[normalized] = struct{}{}
				result.DirectoryPaths = append(result.DirectoryPaths, normalized)
			}
			continue
		}

		if _, exists := fileSeen[normalized]; !exists {
			fileSeen[normalized] = struct{}{}
			result.FilePaths = append(result.FilePaths, normalized)
		}
	}

	return result
}

func looksLikeDirectoryResource(value string) bool {
	value = strings.TrimSpace(strings.Trim(value, "`'\""))
	if value == "" {
		return false
	}
	value = strings.TrimSpace(strings.TrimPrefix(value, "-"))
	return strings.HasSuffix(value, "/") || strings.HasSuffix(value, "\\")
}

func ensureDirectoryTargets(baseDir string, directories []string) error {
	for _, directory := range directories {
		if strings.TrimSpace(directory) == "" {
			continue
		}
		if err := os.MkdirAll(filepath.Join(baseDir, directory), 0755); err != nil {
			return fmt.Errorf("创建目录失败: %s: %w", directory, err)
		}
	}
	return nil
}
