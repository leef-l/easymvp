package regression

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "easymvp/app/mvp/api/mvp/v1"
)

var allowedScenarioStatuses = map[string]struct{}{
	"planned": {},
	"ready":   {},
}

// Manifest 回归样例清单。
type Manifest struct {
	Version   int                         `json:"version"`
	UpdatedAt string                      `json:"updatedAt"`
	Scenarios []v1.RegressionScenarioItem `json:"scenarios"`
}

// ValidationReport 校验结果摘要。
type ValidationReport struct {
	ManifestPath  string
	ScenarioCount int
	ReadyCount    int
	PlannedCount  int
}

// ResolveManifestPath 自动定位回归样例清单。
func ResolveManifestPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return ResolveManifestPathFromCWD(cwd)
}

// ResolveManifestPathFromCWD 基于当前工作目录定位回归样例清单。
func ResolveManifestPathFromCWD(cwd string) (string, error) {
	candidates := []string{
		filepath.Join(cwd, "test-workspaces", "regression-manifest.json"),
		filepath.Join(cwd, "..", "test-workspaces", "regression-manifest.json"),
		filepath.Join(cwd, "..", "..", "test-workspaces", "regression-manifest.json"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Clean(candidate), nil
		}
	}
	return "", fmt.Errorf("regression manifest not found")
}

// LoadManifest 读取回归样例清单。
func LoadManifest(manifestPath string) (*Manifest, error) {
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("读取回归样例清单失败: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("解析回归样例清单失败: %w", err)
	}
	if manifest.Scenarios == nil {
		manifest.Scenarios = []v1.RegressionScenarioItem{}
	}
	return &manifest, nil
}

// ValidateManifest 校验样例清单和 ready 场景结构。
func ValidateManifest(manifestPath string) (*ValidationReport, error) {
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	if manifest.Version <= 0 {
		return nil, fmt.Errorf("manifest version 非法: %d", manifest.Version)
	}
	if strings.TrimSpace(manifest.UpdatedAt) == "" {
		return nil, fmt.Errorf("manifest updatedAt 不能为空")
	}
	if _, err := time.Parse("2006-01-02", manifest.UpdatedAt); err != nil {
		return nil, fmt.Errorf("manifest updatedAt 格式非法: %w", err)
	}
	if len(manifest.Scenarios) == 0 {
		return nil, fmt.Errorf("manifest scenarios 不能为空")
	}

	report := &ValidationReport{
		ManifestPath:  filepath.Clean(manifestPath),
		ScenarioCount: len(manifest.Scenarios),
	}
	seenCodes := make(map[string]struct{}, len(manifest.Scenarios))
	for _, scenario := range manifest.Scenarios {
		if err := validateScenario(manifestPath, scenario, seenCodes); err != nil {
			return nil, err
		}
		switch strings.TrimSpace(scenario.Status) {
		case "ready":
			report.ReadyCount++
		case "planned":
			report.PlannedCount++
		}
	}
	return report, nil
}

func validateScenario(manifestPath string, scenario v1.RegressionScenarioItem, seenCodes map[string]struct{}) error {
	code := strings.TrimSpace(scenario.ScenarioCode)
	if code == "" {
		return fmt.Errorf("存在空 scenarioCode")
	}
	if _, exists := seenCodes[code]; exists {
		return fmt.Errorf("存在重复 scenarioCode: %s", code)
	}
	seenCodes[code] = struct{}{}

	if strings.TrimSpace(scenario.Name) == "" {
		return fmt.Errorf("scenario %s name 不能为空", code)
	}
	if strings.TrimSpace(scenario.WorkspaceDir) == "" {
		return fmt.Errorf("scenario %s workspaceDir 不能为空", code)
	}
	status := strings.TrimSpace(scenario.Status)
	if _, ok := allowedScenarioStatuses[status]; !ok {
		return fmt.Errorf("scenario %s status 非法: %s", code, status)
	}
	if strings.TrimSpace(scenario.Goal) == "" {
		return fmt.Errorf("scenario %s goal 不能为空", code)
	}

	manifestDir := filepath.Dir(manifestPath)
	workspacePath, err := resolveWorkspacePath(manifestDir, scenario.WorkspaceDir)
	if err != nil {
		return fmt.Errorf("scenario %s workspaceDir 非法: %w", code, err)
	}
	if status != "ready" {
		return nil
	}

	if len(scenario.Checkpoints) == 0 {
		return fmt.Errorf("scenario %s ready 场景必须包含 checkpoints", code)
	}
	for _, checkpoint := range scenario.Checkpoints {
		if strings.TrimSpace(checkpoint) == "" {
			return fmt.Errorf("scenario %s checkpoints 不能包含空项", code)
		}
	}
	if err := requireDir(workspacePath); err != nil {
		return fmt.Errorf("scenario %s workspace 缺失: %w", code, err)
	}

	workspaceDir := filepath.ToSlash(filepath.Clean(strings.TrimSpace(scenario.WorkspaceDir)))
	if strings.HasPrefix(workspaceDir, "specs/") {
		for _, name := range []string{"README.md", "input.md", "expected.md"} {
			if err := requireRegularFile(filepath.Join(workspacePath, name)); err != nil {
				return fmt.Errorf("scenario %s 缺少 %s: %w", code, name, err)
			}
		}
		return nil
	}

	hasReadme := requireRegularFile(filepath.Join(workspacePath, "README.md")) == nil
	hasRepo := requireDir(filepath.Join(workspacePath, "repo")) == nil
	if !hasReadme && !hasRepo {
		return fmt.Errorf("scenario %s ready 工作区至少需要 README.md 或 repo/ 目录", code)
	}
	return nil
}

func resolveWorkspacePath(manifestDir, workspaceDir string) (string, error) {
	if strings.TrimSpace(workspaceDir) == "" {
		return "", fmt.Errorf("workspaceDir 为空")
	}
	if filepath.IsAbs(workspaceDir) {
		return "", fmt.Errorf("workspaceDir 不能是绝对路径")
	}
	cleaned := filepath.Clean(filepath.FromSlash(strings.TrimSpace(workspaceDir)))
	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("workspaceDir 为空")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("workspaceDir 不能越出 test-workspaces 根目录")
	}
	return filepath.Join(manifestDir, cleaned), nil
}

func requireDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是目录", path)
	}
	return nil
}

func requireRegularFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s 是目录", path)
	}
	return nil
}
