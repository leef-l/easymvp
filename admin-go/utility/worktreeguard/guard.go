package worktreeguard

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

type Snapshot struct {
	enabled bool
	paths   map[string]string
}

type ValidationResult struct {
	Enabled    bool
	DeltaPaths []string
	Invalid    []string
	Suspicious []string
}

var (
	bulletPrefixPattern    = regexp.MustCompile(`^\s*(?:[-*]\s+|\d+\.\s+)?`)
	titleWrappedPathRegexp = regexp.MustCompile(`^[^()\r\n（）]*[（(]([^()\r\n（）]+)[)）]\s*$`)
	numberedTitlePattern   = regexp.MustCompile(`^\d+\.\s+`)
	simpleNamePattern      = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
)

func Capture(ctx context.Context, workDir string) (*Snapshot, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return &Snapshot{}, nil
	}

	cmd := exec.CommandContext(ctx, "git", "-C", workDir, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) == "" {
		return &Snapshot{}, nil
	}

	paths, err := readGitStatus(ctx, workDir)
	if err != nil {
		return nil, err
	}
	return &Snapshot{
		enabled: true,
		paths:   paths,
	}, nil
}

func (s *Snapshot) Validate(ctx context.Context, workDir string, allowPaths []string) (*ValidationResult, error) {
	result := &ValidationResult{
		Enabled: s != nil && s.enabled,
	}
	if s == nil || !s.enabled {
		return result, nil
	}

	currentPaths, err := readGitStatus(ctx, workDir)
	if err != nil {
		return nil, err
	}

	allowList, _ := NormalizeRelativePaths(allowPaths)
	for currentPath := range currentPaths {
		if _, existed := s.paths[currentPath]; existed {
			continue
		}
		result.DeltaPaths = append(result.DeltaPaths, currentPath)
		if IsSuspiciousPath(currentPath) {
			result.Suspicious = append(result.Suspicious, currentPath)
			continue
		}
		if len(allowList) > 0 && !isAllowedPath(currentPath, allowList) {
			result.Invalid = append(result.Invalid, currentPath)
		}
	}

	return result, nil
}

func (r *ValidationResult) HasIssues() bool {
	if r == nil {
		return false
	}
	return len(r.Invalid) > 0 || len(r.Suspicious) > 0
}

func (r *ValidationResult) Summary() string {
	if r == nil || !r.HasIssues() {
		return ""
	}

	var issues []string
	if len(r.Suspicious) > 0 {
		issues = append(issues, "检测到可疑文件: "+strings.Join(r.Suspicious, ", "))
	}
	if len(r.Invalid) > 0 {
		issues = append(issues, "检测到越界修改: "+strings.Join(r.Invalid, ", "))
	}
	return strings.Join(issues, "；")
}

func NormalizeRelativePaths(values []string) ([]string, []string) {
	var (
		normalized []string
		dropped    []string
		seen       = make(map[string]struct{})
	)

	for _, value := range values {
		normalizedValue, ok := NormalizeRelativePath(value)
		if !ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				dropped = append(dropped, trimmed)
			}
			continue
		}
		if _, exists := seen[normalizedValue]; exists {
			continue
		}
		seen[normalizedValue] = struct{}{}
		normalized = append(normalized, normalizedValue)
	}
	return normalized, dropped
}

func NormalizeRelativePath(value string) (string, bool) {
	value = strings.TrimSpace(strings.Trim(value, "`'\""))
	if value == "" {
		return "", false
	}

	value = bulletPrefixPattern.ReplaceAllString(value, "")
	value = strings.TrimSpace(value)

	if matches := titleWrappedPathRegexp.FindStringSubmatch(value); len(matches) == 2 && looksLikePath(matches[1]) {
		value = matches[1]
	}

	value = strings.ReplaceAll(value, "\\", "/")
	value = strings.TrimPrefix(value, "./")
	value = path.Clean(value)
	if value == "" || value == "." {
		return "", false
	}
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "../") || value == ".." {
		return "", false
	}
	if strings.Contains(value, "\n") || strings.Contains(value, "\r") || strings.Contains(value, ":") {
		return "", false
	}
	if !looksLikePath(value) {
		return "", false
	}
	return value, true
}

func IsSuspiciousPath(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	if strings.ContainsAny(value, "（）`") {
		return true
	}

	firstSegment := value
	if idx := strings.Index(firstSegment, "/"); idx >= 0 {
		firstSegment = firstSegment[:idx]
	}
	return numberedTitlePattern.MatchString(firstSegment)
}

func looksLikePath(value string) bool {
	if strings.Contains(value, "/") {
		return true
	}
	return simpleNamePattern.MatchString(value)
}

func isAllowedPath(value string, allowPaths []string) bool {
	for _, allowPath := range allowPaths {
		if value == allowPath || strings.HasPrefix(value, allowPath+"/") {
			return true
		}
	}
	return false
}

func readGitStatus(ctx context.Context, workDir string) (map[string]string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", workDir, "status", "--porcelain=v1", "--untracked-files=all")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("读取 git 变更失败: %w", err)
	}

	result := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" || len(line) < 4 {
			continue
		}

		status := strings.TrimSpace(line[:2])
		filePath := strings.TrimSpace(line[3:])
		if strings.Contains(filePath, " -> ") {
			parts := strings.Split(filePath, " -> ")
			filePath = strings.TrimSpace(parts[len(parts)-1])
		}
		filePath = strings.ReplaceAll(filePath, "\\", "/")
		filePath = path.Clean(filePath)
		if filePath == "." || filePath == "" {
			continue
		}
		result[filePath] = status
	}
	return result, nil
}
