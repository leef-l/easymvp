package service

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/do"

	_ "modernc.org/sqlite"
)

type normalizedCreateProjectCommand struct {
	Name            string
	ProjectCategory string
	GoalSummary     string
	WorkspaceRoot   string
	RepoRoot        string
}

type derivedWorkspacePaths struct {
	ProjectRoot     string
	MetaRoot        string
	EvidenceRoot    string
	RunsRoot        string
	ReplayRoot      string
	ExportsRoot     string
	CacheRoot       string
	DiagnosticsRoot string
}

var resourceIDCounter uint64

func normalizeCreateProjectCommand(req CreateProjectCommand) (*normalizedCreateProjectCommand, error) {
	normalized := &normalizedCreateProjectCommand{
		Name:            strings.TrimSpace(req.Name),
		ProjectCategory: strings.TrimSpace(req.ProjectCategory),
		GoalSummary:     strings.TrimSpace(req.GoalSummary),
		WorkspaceRoot:   cleanProjectPath(req.WorkspaceRoot),
		RepoRoot:        cleanProjectPath(req.RepoRoot),
	}

	if normalized.Name == "" {
		return nil, gerror.New("name is required")
	}
	if normalized.ProjectCategory == "" {
		return nil, gerror.New("project_category is required")
	}
	if normalized.GoalSummary == "" {
		return nil, gerror.New("goal_summary is required")
	}
	if normalized.WorkspaceRoot == "" {
		return nil, gerror.New("workspace_root is required")
	}
	if normalized.RepoRoot == "." {
		normalized.RepoRoot = ""
	}

	return normalized, nil
}

func cleanProjectPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return filepath.Clean(value)
}

func deriveProjectWorkspacePaths(ctx context.Context, projectID string, workspaceRoot string) (derivedWorkspacePaths, error) {
	dataRoot := cleanProjectPath(g.Cfg().MustGet(ctx, "easymvp.dataRoot").String())
	if dataRoot == "" || dataRoot == "." {
		return derivedWorkspacePaths{}, gerror.New("easymvp.dataRoot is invalid")
	}
	if strings.TrimSpace(projectID) == "" {
		return derivedWorkspacePaths{}, gerror.New("project id is required")
	}

	projectRoot := filepath.Join(dataRoot, "projects", projectID)
	paths := derivedWorkspacePaths{
		ProjectRoot:     projectRoot,
		MetaRoot:        filepath.Join(projectRoot, "meta"),
		EvidenceRoot:    filepath.Join(projectRoot, "evidence"),
		RunsRoot:        filepath.Join(projectRoot, "runs"),
		ReplayRoot:      filepath.Join(projectRoot, "replay"),
		ExportsRoot:     filepath.Join(projectRoot, "exports"),
		CacheRoot:       filepath.Join(projectRoot, "cache"),
		DiagnosticsRoot: filepath.Join(projectRoot, "diagnostics"),
	}
	if err := validateManagedWorkspacePaths(dataRoot, workspaceRoot, paths); err != nil {
		return derivedWorkspacePaths{}, err
	}
	return paths, nil
}

func defaultCategoryProfileVersion(projectCategory string) string {
	if projectCategory == "" {
		return "default/v1"
	}
	return projectCategory + "/v1"
}

func defaultAcceptanceProfileVersion(projectCategory string) string {
	if projectCategory == "" {
		return "default/v1"
	}
	return projectCategory + "/v1"
}

func defaultRoleProfileVersion(projectCategory string) string {
	if projectCategory == "" {
		return "default/v1"
	}
	return projectCategory + "/v1"
}

func nowText() string {
	return time.Now().Format(time.RFC3339)
}

func newResourceID(prefix string) string {
	sequence := atomic.AddUint64(&resourceIDCounter, 1)
	return prefix + "_" + time.Now().UTC().Format("20060102150405.000000000") + "_" + strconv.FormatUint(sequence, 36)
}

func openProjectDatabase(ctx context.Context) (*sql.DB, func(), error) {
	dbPath := g.Cfg().MustGet(ctx, "easymvp.dbPath").String()
	if strings.TrimSpace(dbPath) == "" {
		return nil, nil, gerror.New("easymvp.dbPath is empty")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, nil, gerror.Wrap(err, "open sqlite failed")
	}

	closeFn := func() {
		_ = db.Close()
	}
	return db, closeFn, nil
}

func insertProjectRow(ctx context.Context, tx *sql.Tx, row *do.Projects) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.Projects.Table()+` (
id, name, project_category, goal_summary, status, production_status, workspace_root, repo_root, current_plan_draft_id, current_compiled_plan_id, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.Name,
		row.ProjectCategory,
		row.GoalSummary,
		row.Status,
		row.ProductionStatus,
		row.WorkspaceRoot,
		nullIfEmpty(row.RepoRoot),
		nullIfEmpty(row.CurrentPlanDraftId),
		nullIfEmpty(row.CurrentCompiledPlanId),
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert project failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert project affected unexpected rows")
	}
	return nil
}

func insertProjectProfileRow(ctx context.Context, tx *sql.Tx, row *do.ProjectProfiles) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.ProjectProfiles.Table()+` (
id, project_id, category_profile_version, acceptance_profile_version, role_profile_version, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.CategoryProfileVersion,
		row.AcceptanceProfileVersion,
		nullIfEmpty(row.RoleProfileVersion),
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert project profile failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert project profile affected unexpected rows")
	}
	return nil
}

func insertProjectWorkspaceRow(ctx context.Context, tx *sql.Tx, row *do.ProjectWorkspaces) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.ProjectWorkspaces.Table()+` (
id, project_id, workspace_root, evidence_root, runs_root, replay_root, diagnostics_root, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.WorkspaceRoot,
		row.EvidenceRoot,
		row.RunsRoot,
		row.ReplayRoot,
		row.DiagnosticsRoot,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert project workspace failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert project workspace affected unexpected rows")
	}
	return nil
}

func ensureProjectWorkspaceDirs(ctx context.Context, projectID string, row *do.ProjectWorkspaces) error {
	paths, err := deriveProjectWorkspacePaths(ctx, projectID, asString(row.WorkspaceRoot))
	if err != nil {
		return err
	}

	dirs := []string{
		filepath.Clean(asString(row.WorkspaceRoot)),
		filepath.Clean(paths.ProjectRoot),
		filepath.Clean(paths.MetaRoot),
		filepath.Clean(paths.EvidenceRoot),
		filepath.Clean(paths.RunsRoot),
		filepath.Clean(paths.ReplayRoot),
		filepath.Clean(paths.ExportsRoot),
		filepath.Clean(paths.CacheRoot),
		filepath.Clean(paths.DiagnosticsRoot),
	}
	for _, dir := range dirs {
		if dir == "" || dir == "." {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return gerror.Wrapf(err, "create workspace dir failed: %s", dir)
		}
	}
	for _, dir := range []string{
		paths.ProjectRoot,
		paths.MetaRoot,
		paths.EvidenceRoot,
		paths.RunsRoot,
		paths.ReplayRoot,
		paths.ExportsRoot,
		paths.CacheRoot,
		paths.DiagnosticsRoot,
		asString(row.WorkspaceRoot),
	} {
		if err := ensureDirWritable(dir); err != nil {
			return err
		}
	}
	return nil
}

func validateManagedWorkspacePaths(dataRoot string, workspaceRoot string, paths derivedWorkspacePaths) error {
	if strings.TrimSpace(workspaceRoot) == "" {
		return gerror.New("workspace_root is required")
	}

	managedRoots := []string{
		paths.ProjectRoot,
		paths.MetaRoot,
		paths.EvidenceRoot,
		paths.RunsRoot,
		paths.ReplayRoot,
		paths.ExportsRoot,
		paths.CacheRoot,
		paths.DiagnosticsRoot,
	}
	base := filepath.Clean(dataRoot)
	for _, path := range managedRoots {
		cleaned := filepath.Clean(path)
		rel, err := filepath.Rel(base, cleaned)
		if err != nil {
			return gerror.Wrapf(err, "validate managed path failed: %s", cleaned)
		}
		if rel == "." || strings.HasPrefix(rel, "..") {
			return gerror.Newf("managed path escapes data root: %s", cleaned)
		}
	}
	return nil
}

func ensureDirWritable(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return gerror.Wrapf(err, "stat dir failed: %s", dir)
	}
	if !info.IsDir() {
		return gerror.Newf("path is not a directory: %s", dir)
	}

	probePath := filepath.Join(dir, ".writecheck")
	if err = os.WriteFile(probePath, []byte("ok"), 0o644); err != nil {
		return gerror.Wrapf(err, "directory is not writable: %s", dir)
	}
	if err = os.Remove(probePath); err != nil && !os.IsNotExist(err) {
		return gerror.Wrapf(err, "cleanup write check failed: %s", dir)
	}
	return nil
}

func nullIfEmpty(value any) any {
	text := asString(value)
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return text
}

func asString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}
