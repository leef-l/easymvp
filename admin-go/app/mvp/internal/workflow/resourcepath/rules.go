package resourcepath

import (
	"path"
	"sort"
	"strings"
)

var codingBareFileNames = map[string]struct{}{
	"authors":      {},
	"brewfile":     {},
	"changelog":    {},
	"contributing": {},
	"dockerfile":   {},
	"gemfile":      {},
	"jenkinsfile":  {},
	"license":      {},
	"makefile":     {},
	"notice":       {},
	"procfile":     {},
	"rakefile":     {},
	"readme":       {},
	"taskfile":     {},
	"tiltfile":     {},
	"vagrantfile":  {},
}

var governedRootFiles = map[string]struct{}{
	".gitignore": {},
}

func LooksLikeDirectoryResource(value string) bool {
	value = strings.TrimSpace(strings.Trim(value, "`'\""))
	if value == "" {
		return false
	}
	value = strings.TrimSpace(strings.TrimPrefix(value, "-"))
	if strings.HasSuffix(value, "/") || strings.HasSuffix(value, "\\") {
		return true
	}

	value = strings.Trim(value, "/\\")
	if value == "" {
		return false
	}
	return !LooksLikeExplicitFile(value)
}

func LooksLikeExplicitFile(value string) bool {
	value = strings.TrimSpace(strings.Trim(value, "`'\""))
	if value == "" {
		return false
	}
	value = strings.Trim(value, "/\\")
	base := path.Base(strings.ReplaceAll(value, "\\", "/"))
	if base == "" || base == "." {
		return false
	}
	if strings.HasPrefix(base, ".") && len(base) > 1 {
		return true
	}
	if strings.Contains(base, ".") {
		return true
	}
	_, ok := codingBareFileNames[strings.ToLower(base)]
	return ok
}

func FindCodingDirectoryPlaceholders(resources []string) []string {
	if len(resources) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(resources))
	placeholders := make([]string, 0, len(resources))
	for _, resource := range resources {
		if !LooksLikeDirectoryResource(resource) {
			continue
		}
		normalized := strings.TrimSpace(strings.Trim(resource, "`'\""))
		normalized = strings.TrimSpace(strings.TrimPrefix(normalized, "-"))
		normalized = strings.ReplaceAll(normalized, "\\", "/")
		normalized = strings.Trim(normalized, "/")
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		placeholders = append(placeholders, normalized)
	}
	sort.Strings(placeholders)
	return placeholders
}

func FindNewlyIntroducedGovernedRootFiles(existingResources []string, nextResources []string) []string {
	if len(nextResources) == 0 {
		return nil
	}

	existing := make(map[string]struct{}, len(existingResources))
	for _, resource := range existingResources {
		if normalized, ok := normalizeGovernedRootFile(resource); ok {
			existing[normalized] = struct{}{}
		}
	}

	introduced := make([]string, 0, len(nextResources))
	seen := make(map[string]struct{}, len(nextResources))
	for _, resource := range nextResources {
		normalized, ok := normalizeGovernedRootFile(resource)
		if !ok {
			continue
		}
		if _, owned := existing[normalized]; owned {
			continue
		}
		if _, duplicated := seen[normalized]; duplicated {
			continue
		}
		seen[normalized] = struct{}{}
		introduced = append(introduced, normalized)
	}

	sort.Strings(introduced)
	return introduced
}

func normalizeGovernedRootFile(value string) (string, bool) {
	value = strings.TrimSpace(strings.Trim(value, "`'\""))
	if value == "" {
		return "", false
	}

	value = strings.ReplaceAll(value, "\\", "/")
	value = path.Clean(value)
	if value == "." || value == "/" {
		return "", false
	}

	value = strings.TrimPrefix(value, "./")
	value = strings.Trim(value, "/")
	if value == "" || strings.Contains(value, "/") {
		return "", false
	}

	value = strings.ToLower(value)
	_, ok := governedRootFiles[value]
	return value, ok
}
