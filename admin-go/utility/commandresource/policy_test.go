package commandresource

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMergeNodeMaxOldSpaceAddsLimitWhenMissing(t *testing.T) {
	t.Parallel()

	got := mergeNodeMaxOldSpace("--trace-warnings", 768)
	want := "--trace-warnings --max-old-space-size=768"
	if got != want {
		t.Fatalf("mergeNodeMaxOldSpace() = %q, want %q", got, want)
	}
}

func TestMergeNodeMaxOldSpaceReplacesExistingValue(t *testing.T) {
	t.Parallel()

	got := mergeNodeMaxOldSpace("--max-old-space-size=8192 --trace-warnings", 1024)
	want := "--max-old-space-size=1024 --trace-warnings"
	if got != want {
		t.Fatalf("mergeNodeMaxOldSpace() = %q, want %q", got, want)
	}
}

func TestPolicyMergeEnvOverridesUnsafeValues(t *testing.T) {
	t.Parallel()

	policy := Policy{
		Enabled:                  true,
		NodeMaxOldSpaceMB:        1024,
		NPMJobs:                  1,
		NPMNetworkConcurrency:    4,
		PNPMChildConcurrency:     1,
		PNPMWorkspaceConcurrency: 1,
		GoMaxProcs:               1,
		GoMemLimitMB:             768,
		ComposeParallelLimit:     1,
		TurboConcurrency:         1,
		MakeJobs:                 1,
		CargoBuildJobs:           1,
		ProcessAddressSpaceMB:    1536,
	}

	env := policy.MergeEnv(map[string]string{
		"NODE_OPTIONS": "--max-old-space-size=8192 --trace-warnings",
		"GOMAXPROCS":   "8",
		"KEEP_ME":      "yes",
	})

	if got := env["NODE_OPTIONS"]; got != "--max-old-space-size=1024 --trace-warnings" {
		t.Fatalf("NODE_OPTIONS = %q", got)
	}
	if got := env["GOMAXPROCS"]; got != "1" {
		t.Fatalf("GOMAXPROCS = %q", got)
	}
	if got := env["GOMEMLIMIT"]; got != "768MiB" {
		t.Fatalf("GOMEMLIMIT = %q", got)
	}
	if _, ok := env["npm_config_jobs"]; ok {
		t.Fatalf("npm_config_jobs should not be injected without command context")
	}
	if got := env["KEEP_ME"]; got != "yes" {
		t.Fatalf("KEEP_ME = %q", got)
	}
}

func TestPolicyEnvSliceDeduplicatesAndOverrides(t *testing.T) {
	t.Parallel()

	policy := Policy{
		Enabled:                  true,
		NodeMaxOldSpaceMB:        512,
		NPMJobs:                  1,
		NPMNetworkConcurrency:    4,
		PNPMChildConcurrency:     1,
		PNPMWorkspaceConcurrency: 1,
		GoMaxProcs:               1,
		GoMemLimitMB:             256,
		ComposeParallelLimit:     1,
		TurboConcurrency:         1,
		MakeJobs:                 1,
		CargoBuildJobs:           1,
		ProcessAddressSpaceMB:    1536,
	}

	env := policy.EnvSlice([]string{
		"NODE_OPTIONS=--max-old-space-size=8192",
		"GOMAXPROCS=16",
		"KEEP_ME=yes",
		"NODE_OPTIONS=--trace-warnings",
	})
	got := envSliceToMap(env)

	if got["NODE_OPTIONS"] != "--trace-warnings --max-old-space-size=512" {
		t.Fatalf("NODE_OPTIONS = %q", got["NODE_OPTIONS"])
	}
	if got["GOMAXPROCS"] != "1" {
		t.Fatalf("GOMAXPROCS = %q", got["GOMAXPROCS"])
	}
	if got["KEEP_ME"] != "yes" {
		t.Fatalf("KEEP_ME = %q", got["KEEP_ME"])
	}
	if _, ok := got["npm_config_child_concurrency"]; ok {
		t.Fatalf("npm_config_child_concurrency should not be injected without command context")
	}
}

func TestPolicyApplyKeepsNPMCommandUnwrappedAndInjectsNPMEnv(t *testing.T) {
	t.Parallel()

	policy := Policy{
		Enabled:                  true,
		NodeMaxOldSpaceMB:        1024,
		NPMJobs:                  1,
		NPMNetworkConcurrency:    4,
		PNPMChildConcurrency:     1,
		PNPMWorkspaceConcurrency: 1,
		GoMaxProcs:               1,
		GoMemLimitMB:             768,
		ComposeParallelLimit:     1,
		TurboConcurrency:         1,
		MakeJobs:                 1,
		CargoBuildJobs:           1,
		ProcessAddressSpaceMB:    1536,
	}

	cmd := exec.Command("npm", "run", "build")
	cmd.Env = []string{"NODE_OPTIONS=--trace-warnings"}
	policy.Apply(cmd)

	if base := filepath.Base(cmd.Path); base == "prlimit" {
		t.Fatalf("npm command should not be wrapped with prlimit")
	}

	env := envSliceToMap(cmd.Env)
	if got := env["NODE_OPTIONS"]; got != "--trace-warnings --max-old-space-size=1024" {
		t.Fatalf("NODE_OPTIONS = %q", got)
	}
	if got := env["npm_config_maxsockets"]; got != "4" {
		t.Fatalf("npm_config_maxsockets = %q", got)
	}
	if _, ok := env["npm_config_child_concurrency"]; ok {
		t.Fatalf("npm_config_child_concurrency should not be injected for npm")
	}
}

func TestPolicyApplyWrapsGoCommandWithPrlimit(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "linux" {
		t.Skip("prlimit is only used on linux")
	}
	if _, err := exec.LookPath("prlimit"); err != nil {
		t.Skip("prlimit not installed")
	}

	policy := Policy{
		Enabled:                  true,
		NodeMaxOldSpaceMB:        1024,
		NPMJobs:                  1,
		NPMNetworkConcurrency:    4,
		PNPMChildConcurrency:     1,
		PNPMWorkspaceConcurrency: 1,
		GoMaxProcs:               1,
		GoMemLimitMB:             768,
		ComposeParallelLimit:     1,
		TurboConcurrency:         1,
		MakeJobs:                 1,
		CargoBuildJobs:           1,
		ProcessAddressSpaceMB:    1536,
	}

	cmd := exec.Command("go", "test", "./...")
	policy.Apply(cmd)

	if base := filepath.Base(cmd.Path); base != "prlimit" {
		t.Fatalf("cmd.Path = %q, want prlimit", cmd.Path)
	}
	if got := strings.Join(cmd.Args, " "); !strings.Contains(got, "--as=1610612736") {
		t.Fatalf("cmd.Args = %q", got)
	}
}

func TestPolicyApplyDetectsPNPMInsideShellScript(t *testing.T) {
	t.Parallel()

	policy := Policy{
		Enabled:                  true,
		NodeMaxOldSpaceMB:        1024,
		NPMJobs:                  1,
		NPMNetworkConcurrency:    4,
		PNPMChildConcurrency:     1,
		PNPMWorkspaceConcurrency: 1,
		GoMaxProcs:               1,
		GoMemLimitMB:             768,
		ComposeParallelLimit:     1,
		TurboConcurrency:         1,
		MakeJobs:                 1,
		CargoBuildJobs:           1,
		ProcessAddressSpaceMB:    1536,
	}

	cmd := exec.Command("sh", "-lc", "CI=1 pnpm build")
	policy.Apply(cmd)

	env := envSliceToMap(cmd.Env)
	if got := env["npm_config_network_concurrency"]; got != "4" {
		t.Fatalf("npm_config_network_concurrency = %q", got)
	}
	if got := env["npm_config_child_concurrency"]; got != "1" {
		t.Fatalf("npm_config_child_concurrency = %q", got)
	}
	if got := env["npm_config_workspace_concurrency"]; got != "1" {
		t.Fatalf("npm_config_workspace_concurrency = %q", got)
	}
}
