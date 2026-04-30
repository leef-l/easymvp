package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	projectsv1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/do"
)

type CreateProjectCommand struct {
	Name            string
	ProjectCategory string
	GoalSummary     string
	WorkspaceRoot   string
	RepoRoot        string
}

type UpdateProjectCommand struct {
	ProjectID     string
	Name          string
	GoalSummary   string
	WorkspaceRoot string
	RepoRoot      string
}

type IProjects interface {
	CreateProject(ctx context.Context, req CreateProjectCommand) (res *projectsv1.CreateProjectRes, err error)
	UpdateProject(ctx context.Context, req UpdateProjectCommand) (res *projectsv1.UpdateProjectRes, err error)
	DeleteProject(ctx context.Context, projectID string) (res *projectsv1.DeleteProjectRes, err error)
	GetProjectWorkspaceView(ctx context.Context, projectID string) (res *projectsv1.ProjectWorkspaceViewRes, err error)
	ProjectProgressStream(ctx context.Context, req *projectsv1.ProjectProgressStreamReq) error
}

var localProjects IProjects = (*sProjects)(nil)

type sProjects struct{}

func Projects() IProjects {
	if localProjects == nil {
		localProjects = &sProjects{}
	}
	return localProjects
}

func (s *sProjects) CreateProject(ctx context.Context, req CreateProjectCommand) (res *projectsv1.CreateProjectRes, err error) {
	normalized, err := normalizeCreateProjectCommand(req)
	if err != nil {
		return nil, err
	}

	commandID := newResourceID("cmd")
	projectID := newResourceID("project")
	profileID := newResourceID("profile")
	workspaceID := newResourceID("workspace")
	now := nowText()
	workspacePaths, err := deriveProjectWorkspacePaths(ctx, projectID, normalized.WorkspaceRoot)
	if err != nil {
		return nil, err
	}

	projectRow := &do.Projects{
		Id:               projectID,
		Name:             normalized.Name,
		ProjectCategory:  normalized.ProjectCategory,
		GoalSummary:      normalized.GoalSummary,
		Status:           "created",
		ProductionStatus: "pending",
		WorkspaceRoot:    normalized.WorkspaceRoot,
		RepoRoot:         normalized.RepoRoot,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	profileRow := &do.ProjectProfiles{
		Id:                       profileID,
		ProjectId:                projectID,
		CategoryProfileVersion:   defaultCategoryProfileVersion(normalized.ProjectCategory),
		AcceptanceProfileVersion: defaultAcceptanceProfileVersion(normalized.ProjectCategory),
		RoleProfileVersion:       defaultRoleProfileVersion(normalized.ProjectCategory),
		CreatedAt:                now,
		UpdatedAt:                now,
	}
	workspaceRow := &do.ProjectWorkspaces{
		Id:              workspaceID,
		ProjectId:       projectID,
		WorkspaceRoot:   normalized.WorkspaceRoot,
		EvidenceRoot:    workspacePaths.EvidenceRoot,
		RunsRoot:        workspacePaths.RunsRoot,
		ReplayRoot:      workspacePaths.ReplayRoot,
		DiagnosticsRoot: workspacePaths.DiagnosticsRoot,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err = createProjectRows(ctx, projectID, projectRow, profileRow, workspaceRow); err != nil {
		return nil, err
	}
	if err = Plan().CreateInitialDraft(ctx, projectID); err != nil {
		return nil, err
	}
	if _, err = ArchitectChat().CreateConversation(ctx, projectID); err != nil {
		return nil, err
	}

	// Auto-compile plan in background for end-to-end auto-execution pipeline.
	// Use CompilePlan directly to preserve the initial draft tasks (avoiding
	// architect_chat fallback that overwrites them with a generic single task).
	go func(pid string) {
		bgCtx := context.Background()
		g.Log().Infof(bgCtx, "auto-compiling plan for project %s", pid)
		_, compileErr := Plan().CompilePlan(bgCtx, CompilePlanCommand{
			ProjectID:    pid,
			AutoRedesign: true,
			MaxRedesignAttempts: 2,
		})
		if compileErr != nil {
			g.Log().Warningf(bgCtx, "auto-compile plan failed for project %s: %v", pid, compileErr)
		} else {
			g.Log().Infof(bgCtx, "auto-compile plan succeeded for project %s", pid)
		}
	}(projectID)

	return &projectsv1.CreateProjectRes{
		CommandID:  commandID,
		Accepted:   true,
		ResourceID: projectID,
		NextAction: "open_project_workspace",
	}, nil
}

func (s *sProjects) UpdateProject(ctx context.Context, req UpdateProjectCommand) (res *projectsv1.UpdateProjectRes, err error) {
	if req.ProjectID == "" {
		return nil, gerror.New("project id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	project, err := getProjectByID(ctx, db, req.ProjectID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]any)
	if strings.TrimSpace(req.Name) != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.GoalSummary) != "" {
		updates["goal_summary"] = strings.TrimSpace(req.GoalSummary)
	}
	if strings.TrimSpace(req.WorkspaceRoot) != "" {
		updates["workspace_root"] = cleanProjectPath(req.WorkspaceRoot)
	}
	if req.RepoRoot != "" {
		updates["repo_root"] = nullIfEmpty(cleanProjectPath(req.RepoRoot))
	}
	if len(updates) == 0 {
		return nil, gerror.New("no fields to update")
	}
	updates["updated_at"] = nowText()

	if err = updateProjectRow(ctx, req.ProjectID, updates); err != nil {
		return nil, err
	}
	if auditErr := insertAuditLog(ctx, req.ProjectID, "project.updated", "user:local_operator", "Project updated", updates); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	recordSecurityAuditNoFail(ctx, SecurityAuditEvent{
		ProjectID:   req.ProjectID,
		EventType:   AuditEventProjectStateChange,
		Severity:    SeverityInfo,
		Operator:    "user:local_operator",
		Resource:    req.ProjectID,
		Description: "Project updated",
	})

	return &projectsv1.UpdateProjectRes{
		CommandID:  newResourceID("cmd"),
		Accepted:   true,
		ResourceID: project.Id,
		NextAction: "open_project_workspace",
	}, nil
}

func (s *sProjects) DeleteProject(ctx context.Context, projectID string) (res *projectsv1.DeleteProjectRes, err error) {
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	if err = deleteProjectByID(ctx, projectID); err != nil {
		return nil, err
	}
	if auditErr := insertAuditLog(ctx, projectID, "project.deleted", "user:local_operator", "Project deleted", nil); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	recordSecurityAuditNoFail(ctx, SecurityAuditEvent{
		ProjectID:   projectID,
		EventType:   AuditEventProjectStateChange,
		Severity:    SeverityWarning,
		Operator:    "user:local_operator",
		Resource:    projectID,
		Description: "Project deleted",
	})

	return &projectsv1.DeleteProjectRes{
		CommandID:  newResourceID("cmd"),
		Accepted:   true,
		ResourceID: projectID,
		NextAction: "return_to_projects",
	}, nil
}

func (s *sProjects) GetProjectWorkspaceView(ctx context.Context, projectID string) (res *projectsv1.ProjectWorkspaceViewRes, err error) {
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	data, err := loadProjectWorkspaceData(ctx, projectID)
	if err != nil {
		snapshot, snapshotErr := loadProjectWorkspaceSnapshot(ctx, projectID)
		if snapshotErr == nil {
			return snapshot, nil
		}
		if snapshotErr != nil && snapshotErr != sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	res = &projectsv1.ProjectWorkspaceViewRes{
		Overview:             data.Overview,
		ProjectSnapshot:      data.ProjectSnapshot,
		StageProgress:        data.StageProgress,
		LiveActivity:         data.LiveActivity,
		ActionInbox:          data.ActionInbox,
		AcceptanceCoverage:   data.AcceptanceCoverage,
		WorkspaceExplanation: data.WorkspaceExplanation,
		VerificationResult:   data.VerificationResult,
		CompletionVerdict:    data.CompletionVerdict,
		RuntimeEscalation:    data.RuntimeEscalation,
		FaultSummary:         data.FaultSummary,
		RepairPlanDraft:      data.RepairPlanDraft,
		HealthMetrics:        data.HealthMetrics,
	}
	persistProjectSnapshotAsync(projectID, res)
	return res, nil
}

func (s *sProjects) ProjectProgressStream(ctx context.Context, req *projectsv1.ProjectProgressStreamReq) error {
	return StreamProjectProgress(ctx, req)
}

func createProjectRows(
	ctx context.Context,
	projectID string,
	projectRow *do.Projects,
	profileRow *do.ProjectProfiles,
	workspaceRow *do.ProjectWorkspaces,
) error {
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return gerror.Wrap(err, "begin project transaction failed")
	}

	if err = insertProjectRow(ctx, tx, projectRow); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = insertProjectProfileRow(ctx, tx, profileRow); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = insertProjectWorkspaceRow(ctx, tx, workspaceRow); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = ensureProjectWorkspaceDirs(ctx, projectID, workspaceRow); err != nil {
		_ = tx.Rollback()
		return err
	}
	if auditErr := insertAuditLogTx(ctx, tx, projectID, "project.created", "user:local_operator", "Project created", nil); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	if err = tx.Commit(); err != nil {
		return gerror.Wrap(err, "commit project transaction failed")
	}
	recordSecurityAuditNoFail(ctx, SecurityAuditEvent{
		ProjectID:   projectID,
		EventType:   AuditEventProjectStateChange,
		Severity:    SeverityInfo,
		Operator:    "user:local_operator",
		Resource:    projectID,
		Description: "Project created",
	})
	return nil
}
