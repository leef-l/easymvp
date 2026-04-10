package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"easymvp/app/mvp/internal/workflow/resourcepath"
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

	result.AllowedPaths = normalizePathSlice(result.AllowedPaths)
	result.FilePaths = normalizePathSlice(result.FilePaths)
	result.DirectoryPaths = normalizePathSlice(result.DirectoryPaths)
	return result
}

func looksLikeDirectoryResource(value string) bool {
	return resourcepath.LooksLikeDirectoryResource(value)
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

func applyExecutionSubdir(baseDir string, targets resourceTargets) (string, resourceTargets) {
	commonDir := detectCommonExecutionDir(targets)
	if commonDir == "" {
		return baseDir, targets
	}
	if !canUseExecutionSubdir(baseDir, commonDir) {
		return baseDir, targets
	}

	rebased := resourceTargets{
		AllowedPaths:   normalizePathSlice(targets.AllowedPaths),
		FilePaths:      normalizePathSlice(trimResourcePrefix(targets.FilePaths, commonDir)),
		DirectoryPaths: normalizePathSlice(trimResourcePrefix(targets.DirectoryPaths, commonDir)),
		Rejected:       append([]string(nil), targets.Rejected...),
	}
	return filepath.Join(baseDir, filepath.FromSlash(commonDir)), rebased
}

func promptAllowedPathsForExecution(baseDir string, targets resourceTargets) []string {
	commonDir := detectCommonExecutionDir(targets)
	if commonDir == "" || !canUseExecutionSubdir(baseDir, commonDir) {
		return normalizePathSlice(targets.AllowedPaths)
	}
	return normalizePathSlice(trimResourcePrefix(targets.AllowedPaths, commonDir))
}

func canUseExecutionSubdir(baseDir, commonDir string) bool {
	info, err := os.Stat(filepath.Join(baseDir, filepath.FromSlash(commonDir)))
	if err != nil {
		return false
	}
	return info.IsDir()
}

func detectCommonExecutionDir(targets resourceTargets) string {
	if len(targets.AllowedPaths) == 0 {
		return ""
	}

	var common []string
	seeded := false

	for _, directory := range targets.DirectoryPaths {
		segments := splitPathSegments(directory)
		if len(segments) == 0 {
			return ""
		}
		if !seeded {
			common = segments
			seeded = true
			continue
		}
		common = sharedLeadingSegments(common, segments)
		if len(common) == 0 {
			return ""
		}
	}

	for _, file := range targets.FilePaths {
		dir := path.Dir(file)
		if dir == "." || dir == "" {
			return ""
		}
		segments := splitPathSegments(dir)
		if len(segments) == 0 {
			return ""
		}
		if !seeded {
			common = segments
			seeded = true
			continue
		}
		common = sharedLeadingSegments(common, segments)
		if len(common) == 0 {
			return ""
		}
	}

	if !seeded || len(common) == 0 {
		return ""
	}

	return strings.Join(common, "/")
}

func trimResourcePrefix(values []string, prefix string) []string {
	if prefix == "" {
		return append([]string(nil), values...)
	}

	normalizedPrefix := strings.Trim(strings.TrimSpace(prefix), "/")
	if normalizedPrefix == "" {
		return append([]string(nil), values...)
	}

	trimmed := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.Trim(strings.TrimSpace(value), "/")
		if value == "" {
			continue
		}
		if value == normalizedPrefix {
			continue
		}
		if strings.HasPrefix(value, normalizedPrefix+"/") {
			value = strings.TrimPrefix(value, normalizedPrefix+"/")
		}
		if value == "" {
			continue
		}
		trimmed = append(trimmed, value)
	}
	return trimmed
}

func normalizePathSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalizedValue, ok := worktreeguard.NormalizeRelativePath(value)
		if !ok {
			continue
		}
		if _, exists := seen[normalizedValue]; exists {
			continue
		}
		seen[normalizedValue] = struct{}{}
		normalized = append(normalized, normalizedValue)
	}
	return normalized
}

func splitPathSegments(value string) []string {
	value = strings.Trim(strings.TrimSpace(value), "/")
	if value == "" || value == "." {
		return nil
	}
	parts := strings.Split(value, "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "." {
			continue
		}
		segments = append(segments, part)
	}
	return segments
}

func sharedLeadingSegments(a, b []string) []string {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		if a[i] != b[i] {
			break
		}
		out = append(out, a[i])
	}
	return out
}
