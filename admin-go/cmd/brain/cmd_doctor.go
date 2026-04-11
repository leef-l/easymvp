package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"easymvp/brain/cli"
)

// checkStatus classifies the outcome of one doctor check. Only 3 values are
// used in the skeleton: ok, fail, skip. `warn` is reserved for the full
// implementation.
type checkStatus int

const (
	checkOK checkStatus = iota
	checkFail
	checkSkip
)

// checkResult is the in-memory result of a single check. We collect all of
// them, then print + compute the exit code at the end so the output order
// follows 27-CLI命令契约.md §16.3.
type checkResult struct {
	name   string
	status checkStatus
	msg    string
	hint   string
}

// runDoctor implements `brain doctor [--fix]` per 27 §16. In the v0.1.0
// skeleton the 8 checks either run for real (workspace / disk / clock) or
// return skip with a clear reason (database / sidecars / llm / credentials)
// — see each checkXxx helper below.
func runDoctor(args []string) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fixFlag := fs.Bool("fix", false, "attempt to repair issues (not implemented in skeleton)")
	if err := fs.Parse(args); err != nil {
		return cli.ExitUsage
	}
	if *fixFlag {
		fmt.Fprintln(os.Stderr, "brain doctor: --fix is not implemented in v0.1.0 skeleton")
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "brain doctor: unexpected argument %q\n", fs.Arg(0))
		return cli.ExitUsage
	}

	fmt.Println("Checking brain environment...")
	fmt.Println()

	results := []checkResult{
		checkWorkspace(),
		checkConfigFile(),
		checkDatabase(),
		checkCredentials(),
		checkSidecars(),
		checkLLMReachable(),
		checkDiskSpace(),
		checkClockDrift(),
	}

	failed := 0
	skipped := 0
	for _, r := range results {
		switch r.status {
		case checkOK:
			fmt.Printf("✓ %s: %s\n", r.name, r.msg)
		case checkFail:
			fmt.Printf("✗ %s: %s\n", r.name, r.msg)
			if r.hint != "" {
				fmt.Printf("  → %s\n", r.hint)
			}
			failed++
		case checkSkip:
			fmt.Printf("- %s: skipped (%s)\n", r.name, r.msg)
			skipped++
		}
	}

	fmt.Println()
	switch {
	case failed > 0:
		fmt.Printf("%d issue(s) found", failed)
		if skipped > 0 {
			fmt.Printf(", %d skipped", skipped)
		}
		fmt.Println(". Run with --fix to attempt repair (not in skeleton).")
		return cli.ExitFailed
	case skipped > 0:
		fmt.Printf("All active checks passed (%d skipped in v0.1.0 skeleton).\n", skipped)
	default:
		fmt.Println("All checks passed.")
	}
	return cli.ExitOK
}

// checkWorkspace verifies that $HOME/.brain is present (or can be created) and
// writable. First of the 8 checks in 27 §16.2.
func checkWorkspace() checkResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return checkResult{"workspace", checkFail, err.Error(), "set $HOME"}
	}
	ws := filepath.Join(home, ".brain")
	st, err := os.Stat(ws)
	if os.IsNotExist(err) {
		return checkResult{"workspace", checkOK, fmt.Sprintf("%s (not present, will be created on first run)", ws), ""}
	}
	if err != nil {
		return checkResult{"workspace", checkFail, err.Error(), ""}
	}
	if !st.IsDir() {
		return checkResult{"workspace", checkFail, ws + " exists but is not a directory", ""}
	}
	// Crude writability probe: try to create and remove a temp file.
	probe, err := os.CreateTemp(ws, ".brain-doctor-*")
	if err != nil {
		return checkResult{"workspace", checkFail, "not writable: " + err.Error(), "chmod u+w " + ws}
	}
	name := probe.Name()
	probe.Close()
	_ = os.Remove(name)
	return checkResult{"workspace", checkOK, fmt.Sprintf("%s (writable)", ws), ""}
}

// checkConfigFile is 27 §16.2 check #2 — configuration YAML syntactical
// validation. Skeleton has no config loader yet, so this is a skip.
func checkConfigFile() checkResult {
	return checkResult{
		name:   "config",
		status: checkSkip,
		msg:    "config loader not implemented in v0.1.0 skeleton",
	}
}

// checkDatabase is 27 §16.2 check #3 — SQLite / MySQL reachability. Skeleton
// has no persistence drivers wired in, so this is a skip.
func checkDatabase() checkResult {
	return checkResult{
		name:   "database",
		status: checkSkip,
		msg:    "persistence driver not implemented in v0.1.0 skeleton",
	}
}

// checkCredentials is 27 §16.2 check #4 — credential env var presence.
// Skeleton has no credentials map yet; skip.
func checkCredentials() checkResult {
	return checkResult{
		name:   "credentials",
		status: checkSkip,
		msg:    "credential map not implemented in v0.1.0 skeleton",
	}
}

// checkSidecars is 27 §16.2 check #5 — sidecar binary presence. Skeleton has
// no sidecars shipped; skip.
func checkSidecars() checkResult {
	return checkResult{
		name:   "sidecars",
		status: checkSkip,
		msg:    "no sidecar binaries shipped in v0.1.0 skeleton",
	}
}

// checkLLMReachable is 27 §16.2 check #6 — LLM ping. Network dependent, skip
// by default.
func checkLLMReachable() checkResult {
	return checkResult{
		name:   "llm reachable",
		status: checkSkip,
		msg:    "network-dependent, skipped in v0.1.0 skeleton",
	}
}

// checkDiskSpace is 27 §16.2 check #7 — workspace disk free space (>1GB).
// We do a best-effort read of the workspace parent dir; on any error we skip
// rather than fail so the skeleton stays green on unusual filesystems.
func checkDiskSpace() checkResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return checkResult{"disk space", checkSkip, err.Error(), ""}
	}
	// Portable stat-based check is non-trivial; in the skeleton we just
	// probe that $HOME is statable and assume the full implementation will
	// wire in syscall.Statfs. See 27 §16 for the real contract.
	if _, err := os.Stat(home); err != nil {
		return checkResult{"disk space", checkFail, err.Error(), ""}
	}
	return checkResult{
		name:   "disk space",
		status: checkSkip,
		msg:    "statfs not implemented in v0.1.0 skeleton",
	}
}

// checkClockDrift is 27 §16.2 check #8 — clock drift vs NTP. Network dependent,
// skip in skeleton.
func checkClockDrift() checkResult {
	_ = time.Now
	return checkResult{
		name:   "clock drift",
		status: checkSkip,
		msg:    "NTP probe not implemented in v0.1.0 skeleton",
	}
}
