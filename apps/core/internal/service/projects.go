package service

import (
	"context"
	"database/sql"

	"github.com/gogf/gf/v2/errors/gerror"

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

type IProjects interface {
	CreateProject(ctx context.Context, req CreateProjectCommand) (res *projectsv1.CreateProjectRes, err error)
	GetProjectWorkspaceView(ctx context.Context, projectID string) (res *projectsv1.ProjectWorkspaceViewRes, err error)
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

	return &projectsv1.CreateProjectRes{
		CommandID:  commandID,
		Accepted:   true,
		ResourceID: projectID,
		NextAction: "open_project_workspace",
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
	}
	_ = persistProjectSnapshot(ctx, projectID, res)
	return res, nil
}

func createProjectRows(
	ctx context.Context,
	projectID string,
	projectRow *do.Projects,
	profileRow *do.ProjectProfiles,
	workspaceRow *do.ProjectWorkspaces,
) error {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	tx, err := db.BeginTx(ctx, nil)
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
	if err = tx.Commit(); err != nil {
		return gerror.Wrap(err, "commit project transaction failed")
	}
	return nil
}
