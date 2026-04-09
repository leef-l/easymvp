package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"easymvp/app/mvp/internal/regression"
	workspacepkg "easymvp/app/mvp/internal/workspace"
)

func TestRunWritesManifestAndRiskSummary(t *testing.T) {
	prevValidate := validateManifest
	prevInspect := inspectRiskDeliveryPolicies
	t.Cleanup(func() {
		validateManifest = prevValidate
		inspectRiskDeliveryPolicies = prevInspect
	})

	validateManifest = func(manifestPath string) (*regression.ValidationReport, error) {
		return &regression.ValidationReport{
			ManifestPath:  manifestPath,
			ScenarioCount: 4,
			ReadyCount:    4,
			PlannedCount:  0,
		}, nil
	}
	inspectRiskDeliveryPolicies = func() workspacepkg.RiskDeliveryPolicyReport {
		return workspacepkg.RiskDeliveryPolicyReport{
			Policies: map[string]workspacepkg.RiskDeliveryPolicy{
				workspacepkg.RiskLevelLow: {
					RiskLevel:    workspacepkg.RiskLevelLow,
					DeliveryMode: workspacepkg.DeliveryModePatch,
					SyncStrategy: workspacepkg.SyncStrategyAutoApply,
				},
				workspacepkg.RiskLevelMedium: {
					RiskLevel:    workspacepkg.RiskLevelMedium,
					DeliveryMode: workspacepkg.DeliveryModePatch,
					SyncStrategy: workspacepkg.SyncStrategyManual,
				},
				workspacepkg.RiskLevelHigh: {
					RiskLevel:    workspacepkg.RiskLevelHigh,
					DeliveryMode: workspacepkg.DeliveryModeManual,
					SyncStrategy: workspacepkg.SyncStrategyManual,
				},
			},
		}
	}

	var stdout bytes.Buffer
	if err := run([]string{"test-workspaces/regression-manifest.json"}, &stdout); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "regression manifest OK: path=test-workspaces/regression-manifest.json scenarios=4 ready=4 planned=0") {
		t.Fatalf("unexpected manifest output: %s", output)
	}
	if !strings.Contains(output, "workspace delivery policies OK: 风险交付矩阵: low=patch/auto_apply, medium=patch/manual, high=manual/manual") {
		t.Fatalf("unexpected risk output: %s", output)
	}
}

func TestRunReturnsManifestValidationError(t *testing.T) {
	prevValidate := validateManifest
	prevInspect := inspectRiskDeliveryPolicies
	t.Cleanup(func() {
		validateManifest = prevValidate
		inspectRiskDeliveryPolicies = prevInspect
	})

	validateManifest = func(string) (*regression.ValidationReport, error) {
		return nil, errors.New("broken manifest")
	}
	inspectRiskDeliveryPolicies = func() workspacepkg.RiskDeliveryPolicyReport {
		t.Fatal("risk inspection should not run when manifest validation fails")
		return workspacepkg.RiskDeliveryPolicyReport{}
	}

	err := run([]string{"test-workspaces/regression-manifest.json"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "manifest 校验失败: broken manifest") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunReturnsRiskPolicyError(t *testing.T) {
	prevValidate := validateManifest
	prevInspect := inspectRiskDeliveryPolicies
	t.Cleanup(func() {
		validateManifest = prevValidate
		inspectRiskDeliveryPolicies = prevInspect
	})

	validateManifest = func(manifestPath string) (*regression.ValidationReport, error) {
		return &regression.ValidationReport{
			ManifestPath:  manifestPath,
			ScenarioCount: 1,
			ReadyCount:    1,
			PlannedCount:  0,
		}, nil
	}
	inspectRiskDeliveryPolicies = func() workspacepkg.RiskDeliveryPolicyReport {
		return workspacepkg.RiskDeliveryPolicyReport{
			Policies: map[string]workspacepkg.RiskDeliveryPolicy{
				workspacepkg.RiskLevelHigh: {
					RiskLevel:    workspacepkg.RiskLevelHigh,
					DeliveryMode: workspacepkg.DeliveryModePatch,
					SyncStrategy: workspacepkg.SyncStrategyAutoApply,
				},
			},
			Warnings: []string{"high 风险未进入 PR/人工交付路径"},
		}
	}

	err := run([]string{"test-workspaces/regression-manifest.json"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected risk policy error")
	}
	if !strings.Contains(err.Error(), "风险交付矩阵校验失败") {
		t.Fatalf("missing risk policy prefix: %v", err)
	}
	if !strings.Contains(err.Error(), "high 风险未进入 PR/人工交付路径") {
		t.Fatalf("missing risk policy warning: %v", err)
	}
}

func TestResolveManifestPath(t *testing.T) {
	t.Parallel()

	got, err := resolveManifestPath([]string{"./test-workspaces/../test-workspaces/regression-manifest.json"})
	if err != nil {
		t.Fatalf("resolveManifestPath(args) error = %v", err)
	}
	if got != "test-workspaces/regression-manifest.json" {
		t.Fatalf("resolveManifestPath(args) = %q", got)
	}
}

func TestResolveManifestPathAutoDetect(t *testing.T) {
	root := t.TempDir()
	manifestDir := filepath.Join(root, "test-workspaces")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatalf("mkdir manifest dir: %v", err)
	}
	manifestPath := filepath.Join(manifestDir, "regression-manifest.json")
	if err := os.WriteFile(manifestPath, []byte(`{"version":2,"updatedAt":"2026-04-09","scenarios":[{"scenarioCode":"a","name":"n","workspaceDir":"specs/a","status":"planned","goal":"g"}]}`), 0o644); err != nil {
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

	got, err := resolveManifestPath(nil)
	if err != nil {
		t.Fatalf("resolveManifestPath(nil) error = %v", err)
	}
	if filepath.Clean(got) != filepath.Clean(manifestPath) {
		t.Fatalf("resolveManifestPath(nil) = %q, want %q", got, manifestPath)
	}
}

func TestFailWritesToStderrAndExits(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestRegressionCheckHelperProcess")
	cmd.Env = append(os.Environ(),
		"GO_WANT_REGRESSIONCHECK_HELPER=1",
		"REGRESSIONCHECK_HELPER_MODE=fail",
	)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected subprocess to exit with error")
	}
	if !strings.Contains(string(output), "regression guard failed: boom") {
		t.Fatalf("unexpected fail output: %s", string(output))
	}
}

func TestMainWritesSuccessOutput(t *testing.T) {
	manifestPath := filepath.Join(t.TempDir(), "regression-manifest.json")

	cmd := exec.Command(os.Args[0], "-test.run=TestRegressionCheckHelperProcess")
	cmd.Env = append(os.Environ(),
		"GO_WANT_REGRESSIONCHECK_HELPER=1",
		"REGRESSIONCHECK_HELPER_MODE=main",
		"REGRESSIONCHECK_MANIFEST="+manifestPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper main error = %v, output=%s", err, string(output))
	}
	text := string(output)
	if !strings.Contains(text, "regression manifest OK: path="+manifestPath+" scenarios=1 ready=1 planned=0") {
		t.Fatalf("unexpected main output: %s", text)
	}
	if !strings.Contains(text, "workspace delivery policies OK: 风险交付矩阵: low=patch/auto_apply, medium=patch/manual, high=manual/manual") {
		t.Fatalf("unexpected policy summary: %s", text)
	}
}

func TestRegressionCheckHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_REGRESSIONCHECK_HELPER") != "1" {
		return
	}

	switch os.Getenv("REGRESSIONCHECK_HELPER_MODE") {
	case "fail":
		fail(errors.New("boom"))
	case "main":
		validateManifest = func(manifestPath string) (*regression.ValidationReport, error) {
			return &regression.ValidationReport{
				ManifestPath:  manifestPath,
				ScenarioCount: 1,
				ReadyCount:    1,
				PlannedCount:  0,
			}, nil
		}
		inspectRiskDeliveryPolicies = func() workspacepkg.RiskDeliveryPolicyReport {
			return workspacepkg.RiskDeliveryPolicyReport{
				Policies: map[string]workspacepkg.RiskDeliveryPolicy{
					workspacepkg.RiskLevelLow: {
						RiskLevel:    workspacepkg.RiskLevelLow,
						DeliveryMode: workspacepkg.DeliveryModePatch,
						SyncStrategy: workspacepkg.SyncStrategyAutoApply,
					},
					workspacepkg.RiskLevelMedium: {
						RiskLevel:    workspacepkg.RiskLevelMedium,
						DeliveryMode: workspacepkg.DeliveryModePatch,
						SyncStrategy: workspacepkg.SyncStrategyManual,
					},
					workspacepkg.RiskLevelHigh: {
						RiskLevel:    workspacepkg.RiskLevelHigh,
						DeliveryMode: workspacepkg.DeliveryModeManual,
						SyncStrategy: workspacepkg.SyncStrategyManual,
					},
				},
			}
		}
		os.Args = []string{"regressioncheck", os.Getenv("REGRESSIONCHECK_MANIFEST")}
		main()
	default:
		os.Exit(2)
	}
}
