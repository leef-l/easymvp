package regression

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "easymvp/app/mvp/api/mvp/v1"
)

func TestValidateManifestReadyScenarios(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createSpecScenario(t, root, "specs/readme-refresh")
	createRepoScenario(t, root, "workflow-v2-backend")
	manifestPath := writeManifest(t, root, Manifest{
		Version:   2,
		UpdatedAt: "2026-04-09",
		Scenarios: []v1.RegressionScenarioItem{
			{
				ScenarioCode: "readme_refresh",
				Name:         "README 修订",
				WorkspaceDir: "specs/readme-refresh",
				Status:       "ready",
				Goal:         "覆盖低风险 patch + auto_apply 路径",
				Checkpoints:  []string{"任务默认输出 patch"},
			},
			{
				ScenarioCode: "workflow_v2_backend",
				Name:         "最小后端主链",
				WorkspaceDir: "workflow-v2-backend",
				Status:       "ready",
				Goal:         "覆盖 create-project -> complete 主链",
				Checkpoints:  []string{"工作区存在 repo"},
			},
		},
	})

	report, err := ValidateManifest(manifestPath)
	if err != nil {
		t.Fatalf("ValidateManifest() error = %v", err)
	}
	if report.ScenarioCount != 2 || report.ReadyCount != 2 || report.PlannedCount != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestValidateManifestRejectsDuplicateScenarioCode(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manifestPath := writeManifest(t, root, Manifest{
		Version:   2,
		UpdatedAt: "2026-04-09",
		Scenarios: []v1.RegressionScenarioItem{
			{
				ScenarioCode: "duplicate_case",
				Name:         "样例一",
				WorkspaceDir: "specs/one",
				Status:       "planned",
				Goal:         "first",
			},
			{
				ScenarioCode: "duplicate_case",
				Name:         "样例二",
				WorkspaceDir: "specs/two",
				Status:       "planned",
				Goal:         "second",
			},
		},
	})

	_, err := ValidateManifest(manifestPath)
	if err == nil || !strings.Contains(err.Error(), "重复 scenarioCode") {
		t.Fatalf("expected duplicate scenarioCode error, got %v", err)
	}
}

func TestValidateManifestRejectsMissingReadySpecFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	specDir := filepath.Join(root, "specs", "manual-takeover")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "README.md"), []byte("# manual"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "input.md"), []byte("input"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	manifestPath := writeManifest(t, root, Manifest{
		Version:   2,
		UpdatedAt: "2026-04-09",
		Scenarios: []v1.RegressionScenarioItem{
			{
				ScenarioCode: "manual_takeover",
				Name:         "人工接管回归",
				WorkspaceDir: "specs/manual-takeover",
				Status:       "ready",
				Goal:         "覆盖人工接管",
				Checkpoints:  []string{"可见人工节点"},
			},
		},
	})

	_, err := ValidateManifest(manifestPath)
	if err == nil || !strings.Contains(err.Error(), "expected.md") {
		t.Fatalf("expected missing expected.md error, got %v", err)
	}
}

func TestLoadManifestHandlesInvalidJSONAndNilScenarios(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	invalidPath := filepath.Join(root, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte(`{broken`), 0o644); err != nil {
		t.Fatalf("write invalid manifest: %v", err)
	}
	if _, err := LoadManifest(invalidPath); err == nil || !strings.Contains(err.Error(), "解析回归样例清单失败") {
		t.Fatalf("expected parse error, got %v", err)
	}

	validPath := filepath.Join(root, "valid.json")
	if err := os.WriteFile(validPath, []byte(`{"version":2,"updatedAt":"2026-04-09"}`), 0o644); err != nil {
		t.Fatalf("write valid manifest: %v", err)
	}
	manifest, err := LoadManifest(validPath)
	if err != nil {
		t.Fatalf("LoadManifest(valid) error = %v", err)
	}
	if manifest.Scenarios == nil || len(manifest.Scenarios) != 0 {
		t.Fatalf("expected normalized empty scenarios, got %+v", manifest.Scenarios)
	}
}

func TestValidateManifestRejectsInvalidMeta(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		manifest Manifest
		wantErr  string
	}{
		{
			name: "invalid version",
			manifest: Manifest{
				Version:   0,
				UpdatedAt: "2026-04-09",
				Scenarios: []v1.RegressionScenarioItem{{ScenarioCode: "a", Name: "n", WorkspaceDir: "specs/a", Status: "planned", Goal: "g"}},
			},
			wantErr: "manifest version 非法",
		},
		{
			name: "empty updatedAt",
			manifest: Manifest{
				Version:   2,
				UpdatedAt: "",
				Scenarios: []v1.RegressionScenarioItem{{ScenarioCode: "a", Name: "n", WorkspaceDir: "specs/a", Status: "planned", Goal: "g"}},
			},
			wantErr: "manifest updatedAt 不能为空",
		},
		{
			name: "invalid updatedAt",
			manifest: Manifest{
				Version:   2,
				UpdatedAt: "2026/04/09",
				Scenarios: []v1.RegressionScenarioItem{{ScenarioCode: "a", Name: "n", WorkspaceDir: "specs/a", Status: "planned", Goal: "g"}},
			},
			wantErr: "manifest updatedAt 格式非法",
		},
		{
			name: "empty scenarios",
			manifest: Manifest{
				Version:   2,
				UpdatedAt: "2026-04-09",
				Scenarios: []v1.RegressionScenarioItem{},
			},
			wantErr: "manifest scenarios 不能为空",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			manifestPath := writeManifest(t, root, tc.manifest)
			if _, err := ValidateManifest(manifestPath); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidateManifestRejectsInvalidScenarioFields(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		scenario v1.RegressionScenarioItem
		wantErr  string
	}{
		{name: "empty name", scenario: v1.RegressionScenarioItem{ScenarioCode: "a", Name: "", WorkspaceDir: "specs/a", Status: "planned", Goal: "g"}, wantErr: "scenario a name 不能为空"},
		{name: "empty workspace", scenario: v1.RegressionScenarioItem{ScenarioCode: "a", Name: "n", WorkspaceDir: "", Status: "planned", Goal: "g"}, wantErr: "scenario a workspaceDir 不能为空"},
		{name: "invalid status", scenario: v1.RegressionScenarioItem{ScenarioCode: "a", Name: "n", WorkspaceDir: "specs/a", Status: "draft", Goal: "g"}, wantErr: "scenario a status 非法"},
		{name: "empty goal", scenario: v1.RegressionScenarioItem{ScenarioCode: "a", Name: "n", WorkspaceDir: "specs/a", Status: "planned", Goal: ""}, wantErr: "scenario a goal 不能为空"},
		{name: "ready no checkpoints", scenario: v1.RegressionScenarioItem{ScenarioCode: "a", Name: "n", WorkspaceDir: "specs/a", Status: "ready", Goal: "g"}, wantErr: "ready 场景必须包含 checkpoints"},
		{name: "ready empty checkpoint", scenario: v1.RegressionScenarioItem{ScenarioCode: "a", Name: "n", WorkspaceDir: "specs/a", Status: "ready", Goal: "g", Checkpoints: []string{" "}}, wantErr: "checkpoints 不能包含空项"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			manifestPath := writeManifest(t, root, Manifest{
				Version:   2,
				UpdatedAt: "2026-04-09",
				Scenarios: []v1.RegressionScenarioItem{tc.scenario},
			})
			if _, err := ValidateManifest(manifestPath); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidateManifestReadyRepoNeedsReadmeOrRepo(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workspaceDir := filepath.Join(root, "workflow-empty")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}

	manifestPath := writeManifest(t, root, Manifest{
		Version:   2,
		UpdatedAt: "2026-04-09",
		Scenarios: []v1.RegressionScenarioItem{
			{
				ScenarioCode: "workflow_empty",
				Name:         "empty repo",
				WorkspaceDir: "workflow-empty",
				Status:       "ready",
				Goal:         "g",
				Checkpoints:  []string{"c1"},
			},
		},
	})

	if _, err := ValidateManifest(manifestPath); err == nil || !strings.Contains(err.Error(), "至少需要 README.md 或 repo/ 目录") {
		t.Fatalf("expected missing README/repo error, got %v", err)
	}
}

func TestResolveManifestPathFromCWD(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manifestDir := filepath.Join(root, "test-workspaces")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatalf("mkdir manifest dir: %v", err)
	}
	manifestPath := filepath.Join(manifestDir, "regression-manifest.json")
	if err := os.WriteFile(manifestPath, []byte(`{"version":1,"updatedAt":"2026-04-09","scenarios":[]}`), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	got, err := ResolveManifestPathFromCWD(filepath.Join(root, "admin-go"))
	if err != nil {
		t.Fatalf("ResolveManifestPathFromCWD() error = %v", err)
	}
	if filepath.Clean(got) != filepath.Clean(manifestPath) {
		t.Fatalf("manifest path = %s, want %s", got, manifestPath)
	}
}

func TestResolveManifestPathFromCWDNotFound(t *testing.T) {
	t.Parallel()

	if _, err := ResolveManifestPathFromCWD(t.TempDir()); err == nil {
		t.Fatal("expected error when manifest is missing")
	}
}

func TestResolveManifestPath(t *testing.T) {
	root := t.TempDir()
	manifestDir := filepath.Join(root, "test-workspaces")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatalf("mkdir manifest dir: %v", err)
	}
	manifestPath := filepath.Join(manifestDir, "regression-manifest.json")
	if err := os.WriteFile(manifestPath, []byte(`{"version":1,"updatedAt":"2026-04-09","scenarios":[{"scenarioCode":"a","name":"n","workspaceDir":"specs/a","status":"planned","goal":"g"}]}`), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	got, err := ResolveManifestPath()
	if err != nil {
		t.Fatalf("ResolveManifestPath() error = %v", err)
	}
	if filepath.Clean(got) != filepath.Clean(manifestPath) {
		t.Fatalf("manifest path = %s, want %s", got, manifestPath)
	}
}

func TestResolveWorkspacePath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	got, err := resolveWorkspacePath(root, "specs/demo")
	if err != nil {
		t.Fatalf("resolveWorkspacePath() error = %v", err)
	}
	if filepath.Clean(got) != filepath.Join(root, "specs", "demo") {
		t.Fatalf("workspace path = %s", got)
	}

	cases := []string{"", "/tmp/abs", "..", "../escape"}
	for _, workspaceDir := range cases {
		if _, err := resolveWorkspacePath(root, workspaceDir); err == nil {
			t.Fatalf("expected error for workspaceDir=%q", workspaceDir)
		}
	}
}

func TestRequireDirAndRequireRegularFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := filepath.Join(root, "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir dir: %v", err)
	}
	file := filepath.Join(root, "README.md")
	if err := os.WriteFile(file, []byte("# demo\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := requireDir(dir); err != nil {
		t.Fatalf("requireDir(dir) error = %v", err)
	}
	if err := requireRegularFile(file); err != nil {
		t.Fatalf("requireRegularFile(file) error = %v", err)
	}
	if err := requireDir(file); err == nil {
		t.Fatal("requireDir(file) expected error")
	}
	if err := requireRegularFile(dir); err == nil {
		t.Fatal("requireRegularFile(dir) expected error")
	}
}

func createSpecScenario(t *testing.T, root string, relDir string) {
	t.Helper()

	dir := filepath.Join(root, filepath.FromSlash(relDir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir spec scenario: %v", err)
	}
	files := map[string]string{
		"README.md":   "# spec\n",
		"input.md":    "input\n",
		"expected.md": "expected\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

func createRepoScenario(t *testing.T, root string, relDir string) {
	t.Helper()

	dir := filepath.Join(root, filepath.FromSlash(relDir))
	if err := os.MkdirAll(filepath.Join(dir, "repo"), 0o755); err != nil {
		t.Fatalf("mkdir repo scenario: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# repo\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
}

func writeManifest(t *testing.T, root string, manifest Manifest) string {
	t.Helper()

	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	path := filepath.Join(root, "regression-manifest.json")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}
