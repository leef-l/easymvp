package qualitygate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectSignalsDetectsInteractiveFrontend(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	frontendDir := filepath.Join(root, "frontend")
	if err := os.MkdirAll(filepath.Join(frontendDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir frontend: %v", err)
	}
	if err := os.WriteFile(filepath.Join(frontendDir, "package.json"), []byte(`{
  "dependencies": {"react": "^19.0.0", "react-dom": "^19.0.0"},
  "scripts": {"build": "vite build", "e2e": "playwright test"}
}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(frontendDir, "src", "App.tsx"), []byte("export default function App(){return null}"), 0o644); err != nil {
		t.Fatalf("write App.tsx: %v", err)
	}

	signals := DetectProjectSignals(root, "coding", "software_dev")
	if !signals.HasFrontendApp {
		t.Fatalf("expected frontend app signal, got %+v", signals)
	}
	if !signals.HasBrowserAutomation {
		t.Fatalf("expected browser automation signal, got %+v", signals)
	}
}

func TestDetectProjectSignalsSkipsBackendOnlyProject(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module demo\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	signals := DetectProjectSignals(root, "coding", "software_dev")
	if !signals.HasGoModules {
		t.Fatalf("expected go module signal, got %+v", signals)
	}
	if signals.HasFrontendApp {
		t.Fatalf("did not expect frontend app signal, got %+v", signals)
	}
}

func TestResolveVerificationStandard(t *testing.T) {
	t.Parallel()

	backend := ResolveVerificationStandard(ProjectSignals{FamilyCode: "coding", HasGoModules: true})
	if backend.Code != "coding.backend" {
		t.Fatalf("backend standard = %+v", backend)
	}
	if len(backend.RequiredCheckKinds) != 1 || backend.RequiredCheckKinds[0] != CheckKindTest {
		t.Fatalf("backend required kinds = %+v", backend.RequiredCheckKinds)
	}

	interactive := ResolveVerificationStandard(ProjectSignals{
		FamilyCode:           "coding",
		ProjectTypeCode:      "software_dev",
		HasFrontendApp:       true,
		HasGoModules:         true,
		HasBrowserAutomation: true,
	})
	if interactive.Code != "coding.interactive_delivery" {
		t.Fatalf("interactive standard = %+v", interactive)
	}
	if !interactive.RequireBrowserPlan || !interactive.RequireBrowserEvidence {
		t.Fatalf("interactive standard missing browser requirements: %+v", interactive)
	}
	if len(interactive.RequiredProjectRoles) != 1 || interactive.RequiredProjectRoles[0].RoleType != RoleTypeExperienceReviewer {
		t.Fatalf("interactive standard missing experience reviewer role: %+v", interactive.RequiredProjectRoles)
	}
}

func TestResolveVerificationStandardSkipsBrowserRequirementWithoutAutomation(t *testing.T) {
	t.Parallel()

	interactive := ResolveVerificationStandard(ProjectSignals{
		FamilyCode:      "coding",
		ProjectTypeCode: "software_dev",
		HasFrontendApp:  true,
		HasGoModules:    true,
	})
	if interactive.Code != "coding.interactive_delivery" {
		t.Fatalf("interactive standard = %+v", interactive)
	}
	if interactive.RequireBrowserPlan || interactive.RequireBrowserEvidence {
		t.Fatalf("interactive standard should not require browser verification without automation: %+v", interactive)
	}
	if len(interactive.RequiredCheckKinds) != 2 {
		t.Fatalf("interactive required kinds = %+v", interactive.RequiredCheckKinds)
	}
	if interactive.RequiredCheckKinds[0] != CheckKindBuild || interactive.RequiredCheckKinds[1] != CheckKindTest {
		t.Fatalf("interactive required kinds = %+v, want [build test]", interactive.RequiredCheckKinds)
	}
}

func TestResolveVerificationStandardForGameAndNativeApps(t *testing.T) {
	t.Parallel()

	game := ResolveVerificationStandard(ProjectSignals{
		FamilyCode:      "coding",
		ProjectTypeCode: "game_dev",
		HasFrontendApp:  true,
	})
	if game.Code != "coding.game_client_runtime" {
		t.Fatalf("game standard = %+v", game)
	}
	if got := game.RequiredProjectRoles[0].ReviewProfile; got != ExperienceProfileGameClient {
		t.Fatalf("game review profile = %q, want %q", got, ExperienceProfileGameClient)
	}

	android := ResolveVerificationStandard(ProjectSignals{
		FamilyCode:     "coding",
		HasAndroidApp:  true,
		HasNodePackage: true,
	})
	if android.Code != "coding.android_native_app" {
		t.Fatalf("android standard = %+v", android)
	}
	if android.RequireBrowserPlan || android.RequireBrowserEvidence {
		t.Fatalf("android standard should not require browser evidence yet: %+v", android)
	}
	if got := android.RequiredProjectRoles[0].ReviewProfile; got != ExperienceProfileAndroidNative {
		t.Fatalf("android review profile = %q, want %q", got, ExperienceProfileAndroidNative)
	}
	if len(android.RequiredCheckKinds) != 1 || android.RequiredCheckKinds[0] != CheckKindBuild {
		t.Fatalf("android required kinds = %+v", android.RequiredCheckKinds)
	}
}

func TestBlueprintSignals(t *testing.T) {
	t.Parallel()

	if !BlueprintNeedsBrowserVerification("frontend-game-ui", "实现游戏界面与交互", []string{"frontend/src/App.tsx"}) {
		t.Fatal("expected UI blueprint to require browser verification")
	}
	if !BlueprintProvidesBrowserVerification("frontend-e2e", "补 Playwright 关键路径", []string{"frontend/e2e/smoke.spec.ts"}) {
		t.Fatal("expected e2e blueprint to provide browser verification")
	}
}

func TestInferCheckKindPrefersBrowser(t *testing.T) {
	t.Parallel()

	if got := InferCheckKind("test:e2e", []string{"pnpm", "run", "test:e2e"}, "local"); got != CheckKindBrowser {
		t.Fatalf("InferCheckKind() = %q, want %q", got, CheckKindBrowser)
	}
}
