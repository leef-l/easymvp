package verification

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFindComposeFileAndDockerfile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	composePath := filepath.Join(root, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte("services:{}"), 0o644); err != nil {
		t.Fatalf("write compose: %v", err)
	}
	dockerfilePath := filepath.Join(root, "docker", "Dockerfile")
	if err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755); err != nil {
		t.Fatalf("mkdir docker dir: %v", err)
	}
	if err := os.WriteFile(dockerfilePath, []byte("FROM busybox"), 0o644); err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}

	gotCompose, err := findComposeFile(root)
	if err != nil {
		t.Fatalf("findComposeFile() error = %v", err)
	}
	if gotCompose != composePath {
		t.Fatalf("findComposeFile() = %q, want %q", gotCompose, composePath)
	}

	gotDockerfile, err := findDockerfile(root)
	if err != nil {
		t.Fatalf("findDockerfile() error = %v", err)
	}
	if gotDockerfile != dockerfilePath {
		t.Fatalf("findDockerfile() = %q, want %q", gotDockerfile, dockerfilePath)
	}
}

func TestDiscoverLocalStepsFindsGoAndNodeCommands(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module demo\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	frontendDir := filepath.Join(root, "frontend")
	if err := os.MkdirAll(frontendDir, 0o755); err != nil {
		t.Fatalf("mkdir frontend: %v", err)
	}
	packageJSON := `{
  "name": "demo-frontend",
  "scripts": {
    "lint": "eslint .",
    "test": "vitest",
    "build": "vite build"
  }
}`
	if err := os.WriteFile(filepath.Join(frontendDir, "package.json"), []byte(packageJSON), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(frontendDir, "pnpm-lock.yaml"), []byte("lockfileVersion: '9.0'"), 0o644); err != nil {
		t.Fatalf("write pnpm lock: %v", err)
	}

	svc := NewService(nil, nil, nil)
	meta := &runMeta{WorkflowRunID: 0}
	steps := svc.discoverLocalSteps(context.Background(), meta, root)

	if len(steps) < 4 {
		t.Fatalf("discoverLocalSteps() got %d steps, want at least 4", len(steps))
	}

	assertHasStep(t, steps, "go test ./...", []string{"go", "test", "./..."})
	assertHasStep(t, steps, "frontend lint", []string{"pnpm", "run", "lint"})
	assertHasStep(t, steps, "frontend test", []string{"pnpm", "run", "test", "--", "--run"})
	assertHasStep(t, steps, "frontend build", []string{"pnpm", "run", "build"})
}

func TestBuildTestScriptCommand(t *testing.T) {
	t.Parallel()

	gotVitest := buildTestScriptCommand("npm", "vitest")
	wantVitest := []string{"npm", "run", "test", "--", "--run"}
	assertCommandEqual(t, gotVitest, wantVitest)

	gotJest := buildTestScriptCommand("yarn", "react-scripts test")
	wantJest := []string{"yarn", "test", "--watchAll=false"}
	assertCommandEqual(t, gotJest, wantJest)
}

func TestNormalizeVerificationGate(t *testing.T) {
	t.Parallel()

	gate, err := normalizeVerificationGate(verificationGate{
		AllowedDecisions:   []string{"Passed", "passed", "manual_review"},
		MinExecutedSteps:   -2,
		RequiredCheckKinds: []string{"Test", "build"},
		AllowedRunnerTypes: []string{"Compose", "dockerfile"},
	})
	if err != nil {
		t.Fatalf("normalizeVerificationGate() error = %v", err)
	}

	assertCommandEqual(t, gate.AllowedDecisions, []string{"passed", "manual_review"})
	assertCommandEqual(t, gate.RequiredCheckKinds, []string{"test", "build"})
	assertCommandEqual(t, gate.AllowedRunnerTypes, []string{"docker_compose", "dockerfile"})
	if gate.MinExecutedSteps != 0 {
		t.Fatalf("MinExecutedSteps = %d, want 0", gate.MinExecutedSteps)
	}
}

func TestEvaluateVerificationGateFindsMissingRequiredKinds(t *testing.T) {
	t.Parallel()

	plan := &executionPlan{
		RunnerType: modeLocal,
		Gate: &verificationGate{
			AllowedDecisions:   []string{decisionPassed},
			MinExecutedSteps:   2,
			RequiredCheckKinds: []string{"test", "build"},
		},
		GateSource: "category:game_dev",
		VerifySteps: []verificationStep{
			{Name: "unit test", Command: []string{"go", "test", "./..."}},
		},
	}
	stepExecutions := []stepExecution{
		{
			Stage: "verify",
			Step:  verificationStep{Name: "unit test", Command: []string{"go", "test", "./..."}},
			Result: commandResult{
				ExitCode: 0,
			},
		},
	}

	issues := evaluateVerificationGate(plan, 1, stepExecutions, nil)
	if len(issues) != 2 {
		t.Fatalf("evaluateVerificationGate() issues = %d, want 2; got=%+v", len(issues), issues)
	}
}

func TestEvaluateVerificationGateRespectsRunnerAndDecision(t *testing.T) {
	t.Parallel()

	plan := &executionPlan{
		RunnerType: modeDockerCompose,
		Gate: &verificationGate{
			AllowedDecisions:   []string{decisionPassed},
			AllowedRunnerTypes: []string{modeLocal},
		},
		GateSource: "category:software_dev",
	}

	existingIssues := []issueDraft{
		{Severity: "error", Title: "step failed"},
	}
	issues := evaluateVerificationGate(plan, 1, nil, existingIssues)
	if len(issues) != 2 {
		t.Fatalf("evaluateVerificationGate() issues = %d, want 2; got=%+v", len(issues), issues)
	}
}

func TestResolveIssueTaskIDFromCandidatesChoosesBackendRootTaskForCrossTaskCompileFailure(t *testing.T) {
	t.Parallel()

	tasks := []taskBindingCandidate{
		{
			ID:                100,
			Name:              "backend-goframe-init",
			AffectedResources: []string{"backend/go.mod", "backend/main.go", "backend/internal/cmd/cmd.go"},
			Depth:             0,
		},
		{
			ID:                110,
			Name:              "backend-config-websocket",
			ParentTaskID:      100,
			AffectedResources: []string{"backend/internal/controller/websocket/client.go"},
			Depth:             1,
		},
		{
			ID:                120,
			Name:              "backend-game-engine-core",
			ParentTaskID:      100,
			AffectedResources: []string{"backend/internal/logic/game/engine.go"},
			Depth:             1,
		},
		{
			ID:                130,
			Name:              "backend-food-generator",
			ParentTaskID:      100,
			AffectedResources: []string{"backend/internal/logic/game/food_generator.go"},
			Depth:             1,
		},
		{
			ID:                140,
			Name:              "backend-engine-unit-tests",
			ParentTaskID:      120,
			AffectedResources: []string{"backend/internal/logic/game/engine_test.go", "backend/internal/logic/game/collision_test.go"},
			Depth:             2,
		},
		{
			ID:                150,
			Name:              "backend-integration-tests",
			ParentTaskID:      110,
			AffectedResources: []string{"backend/internal/controller/websocket/handler_test.go", "backend/internal/logic/session/manager_test.go"},
			Depth:             2,
		},
	}

	got := resolveIssueTaskIDFromCandidates(issueTaskResolutionInput{
		CurrentTaskID: 150,
		Title:         "验证步骤失败: backend go test ./...",
		ResourceRef:   "backend",
		Detail: `main.go:6:2: missing go.sum entry
internal/cmd/cmd.go:1:1: expected 'package', found 'EOF'
internal/controller/websocket/client.go:9:2: missing go.sum entry
internal/controller/websocket/handler_test.go:1:1: expected 'package', found 'EOF'
internal/logic/session/manager_test.go:1:1: expected 'package', found 'EOF'
internal/logic/game/food_generator.go:25:6: Position redeclared
internal/logic/game/engine.go:322:6: cannot use int64
internal/logic/game/engine_test.go:85:16: undefined: moveSnake
internal/logic/game/collision_test.go:76:14: undefined: checkCollision`,
	}, tasks)

	if got != 100 {
		t.Fatalf("resolveIssueTaskIDFromCandidates() = %d, want 100", got)
	}
}

func TestResolveIssueTaskIDFromCandidatesChoosesFrontendInitForFrontendBuildFailure(t *testing.T) {
	t.Parallel()

	tasks := []taskBindingCandidate{
		{
			ID:                200,
			Name:              "frontend-vite-react-init",
			AffectedResources: []string{"frontend/package.json", "frontend/vite.config.ts", "frontend/tsconfig.json"},
			Depth:             0,
		},
		{
			ID:                210,
			Name:              "frontend-main-app",
			ParentTaskID:      200,
			AffectedResources: []string{"frontend/src/App.tsx", "frontend/src/App.css"},
			Depth:             1,
		},
		{
			ID:                220,
			Name:              "frontend-e2e-test",
			ParentTaskID:      210,
			AffectedResources: []string{"frontend/e2e/snake.spec.ts", "frontend/playwright.config.ts"},
			Depth:             2,
		},
	}

	got := resolveIssueTaskIDFromCandidates(issueTaskResolutionInput{
		CurrentTaskID: 220,
		Title:         "验证步骤失败: frontend build",
		ResourceRef:   "frontend",
		Detail: `命令执行失败：'npm' 'run' 'build'
sh: line 1: tsc: command not found`,
	}, tasks)

	if got != 200 {
		t.Fatalf("resolveIssueTaskIDFromCandidates() = %d, want 200", got)
	}
}

func TestResolveIssueTaskIDFromCandidatesChoosesProductionBuildForRootBuildFailure(t *testing.T) {
	t.Parallel()

	tasks := []taskBindingCandidate{
		{
			ID:                300,
			Name:              "cli-root-init",
			AffectedResources: []string{"package.json", "scripts/dev.js", "scripts/build.js", ".gitignore"},
			Depth:             0,
		},
		{
			ID:                310,
			Name:              "cli-production-build",
			ParentTaskID:      300,
			AffectedResources: []string{"scripts/build.js", "Makefile", "docker/Dockerfile"},
			Depth:             1,
		},
	}

	got := resolveIssueTaskIDFromCandidates(issueTaskResolutionInput{
		Title:       "验证步骤失败: build",
		ResourceRef: ".",
		Detail: `命令执行失败：'npm' 'run' 'build'
> node scripts/build.js
> npm ci
npm error The npm ci command can only install with an existing package-lock.json`,
	}, tasks)

	if got != 310 {
		t.Fatalf("resolveIssueTaskIDFromCandidates() = %d, want 310", got)
	}
}

func assertHasStep(t *testing.T, steps []verificationStep, name string, command []string) {
	t.Helper()
	for _, step := range steps {
		if step.Name == name {
			assertCommandEqual(t, step.Command, command)
			return
		}
	}
	t.Fatalf("step %q not found in %+v", name, steps)
}

func assertCommandEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("command len = %d, want %d; got=%v want=%v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("command[%d] = %q, want %q; got=%v want=%v", i, got[i], want[i], got, want)
		}
	}
}
