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
