package service

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

// ReadinessCheckCategory classifies production readiness checks.
type ReadinessCheckCategory string

const (
	ReadinessCategoryCode           ReadinessCheckCategory = "code"
	ReadinessCategoryTest           ReadinessCheckCategory = "test"
	ReadinessCategorySecurity       ReadinessCheckCategory = "security"
	ReadinessCategoryPerformance    ReadinessCheckCategory = "performance"
	ReadinessCategoryDocumentation  ReadinessCheckCategory = "documentation"
	ReadinessCategoryInfrastructure ReadinessCheckCategory = "infrastructure"
)

// ReadinessCheckStatus is the outcome of a single readiness check.
type ReadinessCheckStatus string

const (
	ReadinessStatusPass    ReadinessCheckStatus = "pass"
	ReadinessStatusFail    ReadinessCheckStatus = "fail"
	ReadinessStatusWarning ReadinessCheckStatus = "warning"
	ReadinessStatusSkipped ReadinessCheckStatus = "skipped"
)

// ---------------------------------------------------------------------------
// Core structs
// ---------------------------------------------------------------------------

// ReadinessCheck represents a single production readiness check item.
type ReadinessCheck struct {
	Name     string                 `json:"name"`
	Category ReadinessCheckCategory `json:"category"`
	Status   ReadinessCheckStatus   `json:"status"`
	Message  string                 `json:"message"`
	Required bool                   `json:"required"`
}

// ReadinessReport is the aggregated result of a full production readiness scan.
type ReadinessReport struct {
	ProjectID    string           `json:"projectId"`
	Checks       []ReadinessCheck `json:"checks"`
	OverallReady bool             `json:"overallReady"`
	Score        int              `json:"score"`
	MaxScore     int              `json:"maxScore"`
	GeneratedAt  time.Time        `json:"generatedAt"`
}

// ReadinessVerdict is the final judgment derived from a ReadinessReport.
type ReadinessVerdict struct {
	Ready          bool     `json:"ready"`
	BlockingIssues []string `json:"blockingIssues"`
	Warnings       []string `json:"warnings"`
	Recommendation string   `json:"recommendation"`
}

// ReadinessCheckFunc is the signature for a custom check function.
type ReadinessCheckFunc func(ctx context.Context, projectID string) ReadinessCheck

// ---------------------------------------------------------------------------
// Interface & singleton
// ---------------------------------------------------------------------------

// IProductionReadiness defines the production readiness checking service.
type IProductionReadiness interface {
	RunFullCheck(ctx context.Context, projectID string) *ReadinessReport
	RunCategoryCheck(ctx context.Context, projectID string, category ReadinessCheckCategory) []ReadinessCheck
	GetDefaultChecklist(projectType string) []ReadinessCheck
	EvaluateReport(report *ReadinessReport) ReadinessVerdict
	RegisterCustomCheck(name string, category ReadinessCheckCategory, required bool, fn ReadinessCheckFunc)
}

var (
	localProductionReadiness     IProductionReadiness
	localProductionReadinessOnce sync.Once
)

type sProductionReadiness struct {
	mu           sync.RWMutex
	customChecks []customCheckEntry
}

type customCheckEntry struct {
	name     string
	category ReadinessCheckCategory
	required bool
	fn       ReadinessCheckFunc
}

// ProductionReadiness returns the singleton IProductionReadiness implementation.
func ProductionReadiness() IProductionReadiness {
	localProductionReadinessOnce.Do(func() {
		localProductionReadiness = &sProductionReadiness{}
	})
	return localProductionReadiness
}

// ---------------------------------------------------------------------------
// Interface implementation
// ---------------------------------------------------------------------------

// RunFullCheck executes every default and custom check for the given project.
func (s *sProductionReadiness) RunFullCheck(ctx context.Context, projectID string) *ReadinessReport {
	repoRoot, projectCategory := resolveProjectForReadiness(ctx, projectID)
	if projectCategory == "" {
		projectCategory = "general"
	}
	checks := s.GetDefaultChecklist(projectCategory)

	for i := range checks {
		checks[i] = runDefaultCheck(ctx, checks[i], repoRoot)
	}

	s.mu.RLock()
	customs := make([]customCheckEntry, len(s.customChecks))
	copy(customs, s.customChecks)
	s.mu.RUnlock()

	for _, cc := range customs {
		checks = append(checks, safeRunCustomCheck(ctx, cc, projectID))
	}

	return s.buildReport(projectID, checks)
}

// RunCategoryCheck executes only the checks belonging to the given category.
func (s *sProductionReadiness) RunCategoryCheck(ctx context.Context, projectID string, category ReadinessCheckCategory) []ReadinessCheck {
	repoRoot, projectCategory := resolveProjectForReadiness(ctx, projectID)
	if projectCategory == "" {
		projectCategory = "general"
	}
	all := s.GetDefaultChecklist(projectCategory)

	var filtered []ReadinessCheck
	for _, c := range all {
		if c.Category == category {
			filtered = append(filtered, runDefaultCheck(ctx, c, repoRoot))
		}
	}

	s.mu.RLock()
	customs := make([]customCheckEntry, len(s.customChecks))
	copy(customs, s.customChecks)
	s.mu.RUnlock()

	for _, cc := range customs {
		if cc.category == category {
			filtered = append(filtered, safeRunCustomCheck(ctx, cc, projectID))
		}
	}

	return filtered
}

// GetDefaultChecklist returns the standard set of readiness checks for a project type.
func (s *sProductionReadiness) GetDefaultChecklist(projectType string) []ReadinessCheck {
	checks := []ReadinessCheck{
		// Code
		{Name: "compilation", Category: ReadinessCategoryCode, Required: true},
		{Name: "static_analysis", Category: ReadinessCategoryCode, Required: true},
		{Name: "code_coverage", Category: ReadinessCategoryCode, Required: false},
		// Test
		{Name: "unit_tests_pass", Category: ReadinessCategoryTest, Required: true},
		{Name: "integration_tests_pass", Category: ReadinessCategoryTest, Required: true},
		// Security
		{Name: "dependency_vulnerability_scan", Category: ReadinessCategorySecurity, Required: true},
		{Name: "sensitive_info_check", Category: ReadinessCategorySecurity, Required: true},
		// Performance
		{Name: "benchmark_exists", Category: ReadinessCategoryPerformance, Required: false},
		// Documentation
		{Name: "readme_exists", Category: ReadinessCategoryDocumentation, Required: true},
		{Name: "api_docs", Category: ReadinessCategoryDocumentation, Required: false},
		// Infrastructure
		{Name: "docker_config", Category: ReadinessCategoryInfrastructure, Required: true},
		{Name: "env_template", Category: ReadinessCategoryInfrastructure, Required: false},
	}

	switch projectType {
	case "api":
		readinessSetRequired(checks, "api_docs", true)
		readinessSetRequired(checks, "benchmark_exists", true)
	case "web_app":
		readinessSetRequired(checks, "code_coverage", true)
	case "game":
		readinessSetRequired(checks, "benchmark_exists", true)
		readinessSetRequired(checks, "docker_config", false)
	case "data_pipeline":
		readinessSetRequired(checks, "benchmark_exists", true)
		readinessSetRequired(checks, "code_coverage", true)
	}

	return checks
}

// EvaluateReport produces a ReadinessVerdict from a completed report.
func (s *sProductionReadiness) EvaluateReport(report *ReadinessReport) ReadinessVerdict {
	var blocking []string
	var warnings []string

	for _, c := range report.Checks {
		switch c.Status {
		case ReadinessStatusFail:
			if c.Required {
				blocking = append(blocking, c.Name+": "+c.Message)
			} else {
				warnings = append(warnings, c.Name+": "+c.Message)
			}
		case ReadinessStatusWarning:
			warnings = append(warnings, c.Name+": "+c.Message)
		case ReadinessStatusSkipped:
			// Skipped checks degrade to warnings — they signal the framework
			// could not evaluate the item, not that the item failed.
			if c.Required {
				warnings = append(warnings, c.Name+": skipped (no project workspace or runner unavailable)")
			}
		}
	}

	ready := len(blocking) == 0
	recommendation := "All required checks passed. Ready for production."
	if !ready {
		recommendation = "Resolve all blocking issues before deploying to production."
	} else if len(warnings) > 0 {
		recommendation = "Production-ready with warnings. Review non-critical items."
	}

	return ReadinessVerdict{
		Ready:          ready,
		BlockingIssues: blocking,
		Warnings:       warnings,
		Recommendation: recommendation,
	}
}

// RegisterCustomCheck adds a user-defined check to the readiness framework.
func (s *sProductionReadiness) RegisterCustomCheck(name string, category ReadinessCheckCategory, required bool, fn ReadinessCheckFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.customChecks = append(s.customChecks, customCheckEntry{
		name:     name,
		category: category,
		required: required,
		fn:       fn,
	})
}

// ---------------------------------------------------------------------------
// Default check executors
// ---------------------------------------------------------------------------

// runDefaultCheck executes a check based on its name. If the project workspace
// is unavailable or the check is not yet implemented, it returns Skipped.
func runDefaultCheck(ctx context.Context, c ReadinessCheck, repoRoot string) ReadinessCheck {
	if repoRoot == "" {
		c.Status = ReadinessStatusSkipped
		c.Message = "no project workspace resolved"
		return c
	}

	// P1.1 — refuse to run build/vet against EasyMVP itself; that would
	// recursively spawn `go build ./...` inside the very binary that is
	// running, which has hung the dev box more than once.
	if readinessIsSelfRepo(repoRoot) {
		c.Status = ReadinessStatusSkipped
		c.Message = "repo_root points to EasyMVP itself; skipping to avoid self-recursion"
		return c
	}

	switch c.Name {
	case "compilation":
		return checkCompilation(ctx, c, repoRoot)
	case "static_analysis":
		return checkStaticAnalysis(ctx, c, repoRoot)
	case "readme_exists":
		return checkFileExists(c, repoRoot, []string{"README.md", "README", "readme.md"}, "no README found at project root")
	case "api_docs":
		return checkAnyPath(c, repoRoot,
			[]string{"docs/api.md", "docs/api/", "openapi.yaml", "openapi.json", "swagger.yaml", "swagger.json"},
			"no API docs (docs/api.* / openapi.* / swagger.*) found")
	case "docker_config":
		return checkAnyPath(c, repoRoot,
			[]string{"Dockerfile", "docker-compose.yml", "docker-compose.yaml", "docker/", "compose.yaml"},
			"no Dockerfile or docker-compose found")
	case "env_template":
		return checkAnyPath(c, repoRoot,
			[]string{".env.example", ".env.template", ".env.sample"},
			"no .env.example / .env.template found")
	case "sensitive_info_check":
		return checkSensitiveInfo(c, repoRoot)
	default:
		c.Status = ReadinessStatusSkipped
		c.Message = "check not implemented; register a custom check to evaluate"
		return c
	}
}

// checkCompilation runs `go build ./...` in the repo root if it appears to be a Go project.
func checkCompilation(ctx context.Context, c ReadinessCheck, repoRoot string) ReadinessCheck {
	if !readinessFileExists(filepath.Join(repoRoot, "go.mod")) {
		c.Status = ReadinessStatusSkipped
		c.Message = "no go.mod found; compilation check is Go-specific"
		return c
	}
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "go", "build", "./...")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		c.Status = ReadinessStatusFail
		c.Message = "go build failed: " + readinessTruncate(string(out), 200)
		return c
	}
	c.Status = ReadinessStatusPass
	c.Message = "go build ./... succeeded"
	return c
}

// checkStaticAnalysis runs `go vet ./...` in the repo root if it appears to be a Go project.
func checkStaticAnalysis(ctx context.Context, c ReadinessCheck, repoRoot string) ReadinessCheck {
	if !readinessFileExists(filepath.Join(repoRoot, "go.mod")) {
		c.Status = ReadinessStatusSkipped
		c.Message = "no go.mod found; static analysis check is Go-specific"
		return c
	}
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "go", "vet", "./...")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		c.Status = ReadinessStatusFail
		c.Message = "go vet failed: " + readinessTruncate(string(out), 200)
		return c
	}
	c.Status = ReadinessStatusPass
	c.Message = "go vet ./... passed"
	return c
}

// checkFileExists passes if any of the candidate filenames exists at repoRoot.
func checkFileExists(c ReadinessCheck, repoRoot string, candidates []string, failMsg string) ReadinessCheck {
	for _, name := range candidates {
		if readinessFileExists(filepath.Join(repoRoot, name)) {
			c.Status = ReadinessStatusPass
			c.Message = "found: " + name
			return c
		}
	}
	c.Status = ReadinessStatusFail
	c.Message = failMsg
	return c
}

// checkAnyPath passes if any candidate path (file or directory) exists.
func checkAnyPath(c ReadinessCheck, repoRoot string, candidates []string, failMsg string) ReadinessCheck {
	for _, p := range candidates {
		if readinessPathExists(filepath.Join(repoRoot, p)) {
			c.Status = ReadinessStatusPass
			c.Message = "found: " + p
			return c
		}
	}
	c.Status = ReadinessStatusFail
	c.Message = failMsg
	return c
}

// checkSensitiveInfo flags common .env files committed at repo root.
// It uses readinessPathExists so that an `.env` directory is also caught
// (fileExists would silently skip directories).
func checkSensitiveInfo(c ReadinessCheck, repoRoot string) ReadinessCheck {
	suspects := []string{".env", ".env.local", ".env.production", "secrets.json", "credentials.json"}
	var found []string
	for _, name := range suspects {
		if readinessPathExists(filepath.Join(repoRoot, name)) {
			found = append(found, name)
		}
	}
	if len(found) > 0 {
		c.Status = ReadinessStatusFail
		c.Message = "potential secrets at repo root: " + strings.Join(found, ", ")
		return c
	}
	c.Status = ReadinessStatusPass
	c.Message = "no plaintext secrets detected at repo root"
	return c
}

// safeRunCustomCheck executes a custom check guarded against panics.
func safeRunCustomCheck(ctx context.Context, cc customCheckEntry, projectID string) (result ReadinessCheck) {
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "readiness custom check %s panicked: %v\n%s", cc.name, r, debug.Stack())
			result = ReadinessCheck{
				Name:     cc.name,
				Category: cc.category,
				Required: cc.required,
				Status:   ReadinessStatusFail,
				Message:  "custom check panicked",
			}
		}
	}()
	return cc.fn(ctx, projectID)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// buildReport assembles a ReadinessReport from a slice of completed checks.
func (s *sProductionReadiness) buildReport(projectID string, checks []ReadinessCheck) *ReadinessReport {
	score := 0
	maxScore := 0

	for _, c := range checks {
		weight := 1
		if c.Required {
			weight = 2
		}
		maxScore += weight

		switch c.Status {
		case ReadinessStatusPass:
			score += weight
		case ReadinessStatusWarning:
			score += weight / 2
		}
	}

	overallReady := true
	for _, c := range checks {
		if !c.Required {
			continue
		}
		if c.Status == ReadinessStatusFail {
			overallReady = false
			break
		}
	}

	return &ReadinessReport{
		ProjectID:    projectID,
		Checks:       checks,
		OverallReady: overallReady,
		Score:        score,
		MaxScore:     maxScore,
		GeneratedAt:  time.Now(),
	}
}

// readinessSetRequired adjusts the Required flag of a named check in the slice.
func readinessSetRequired(checks []ReadinessCheck, name string, required bool) {
	for i := range checks {
		if checks[i].Name == name {
			checks[i].Required = required
			return
		}
	}
}

// resolveProjectForReadiness reads the project's repo_root (falling back to
// workspace_root) and project_category from the database. Returns empty
// strings if the project cannot be resolved — callers must handle gracefully.
func resolveProjectForReadiness(ctx context.Context, projectID string) (repoRoot, projectCategory string) {
	if projectID == "" {
		return "", ""
	}
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return "", ""
	}
	defer closeFn()
	project, err := getProjectByID(ctx, db, projectID)
	if err != nil || project == nil {
		return "", ""
	}
	root := strings.TrimSpace(project.RepoRoot)
	if root == "" {
		root = strings.TrimSpace(project.WorkspaceRoot)
	}
	return root, strings.TrimSpace(project.ProjectCategory)
}

func readinessFileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func readinessPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readinessTruncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// ---------------------------------------------------------------------------
// Self-recursion guard
// ---------------------------------------------------------------------------

var (
	readinessSelfRepoOnce sync.Once
	readinessSelfRepoPath string
)

// readinessIsSelfRepo reports whether repoRoot resolves to EasyMVP's own
// source tree. We check three independent signals because any single one may
// be wrong on a developer machine:
//  1. The hard-coded production install path (/www/wwwroot/project/easymvp).
//  2. Whether the path contains apps/core (the readiness service's own pkg).
//  3. The repo root discovered from os.Executable() at process start.
func readinessIsSelfRepo(repoRoot string) bool {
	if strings.TrimSpace(repoRoot) == "" {
		return false
	}
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		// If we can't even resolve it, treat as self to be safe — better to
		// skip a check than to recurse into ourselves.
		return true
	}
	abs = filepath.Clean(abs)

	// Signal 1: known production install path.
	const easymvpInstall = "/www/wwwroot/project/easymvp"
	if abs == easymvpInstall || strings.HasPrefix(abs+string(os.PathSeparator), easymvpInstall+string(os.PathSeparator)) {
		return true
	}

	// Signal 2: path explicitly references the readiness service's own
	// containing module (apps/core). This catches workspace clones too.
	if strings.Contains(abs, string(os.PathSeparator)+filepath.Join("apps", "core")) {
		return true
	}

	// Signal 3: derive selfRepo from the running binary's location. We walk
	// upward from the executable's directory looking for a go.mod and cache
	// the result for the life of the process.
	selfRepo := readinessSelfRepoRoot()
	if selfRepo == "" {
		return false
	}
	if abs == selfRepo {
		return true
	}
	// abs sits underneath selfRepo, e.g. selfRepo/apps/core.
	if strings.HasPrefix(abs+string(os.PathSeparator), selfRepo+string(os.PathSeparator)) {
		return true
	}
	// selfRepo sits underneath abs, e.g. abs is the parent monorepo and
	// selfRepo is .../apps/core within it.
	if strings.HasPrefix(selfRepo+string(os.PathSeparator), abs+string(os.PathSeparator)) {
		return true
	}
	return false
}

// readinessSelfRepoRoot returns the directory containing the go.mod that
// produced the current process binary. Cached after first call.
func readinessSelfRepoRoot() string {
	readinessSelfRepoOnce.Do(func() {
		exe, err := os.Executable()
		if err != nil {
			return
		}
		exe, err = filepath.EvalSymlinks(exe)
		if err != nil {
			// Fall back to the unevaluated path; better than giving up.
			exe, _ = os.Executable()
		}
		dir := filepath.Dir(exe)
		for i := 0; i < 12; i++ {
			if readinessFileExists(filepath.Join(dir, "go.mod")) {
				readinessSelfRepoPath = filepath.Clean(dir)
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				return
			}
			dir = parent
		}
	})
	return readinessSelfRepoPath
}
