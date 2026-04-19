package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	workspacev1 "github.com/leef-l/easymvp/apps/core/api/workspace/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

const (
	workspaceHomeProjectLimit   = 12
	workspaceHomeAttentionLimit = 10
	workspaceHomeActivityLimit  = 12
	workspaceHomeReleaseLimit   = 8
)

type workspaceHomeData struct {
	TotalProjects    int
	ActiveProjects   int
	BlockedProjects  int
	PendingActions   int
	ProjectCards     []workspacev1.ProjectCard
	NeedAttention    []workspacev1.NeedAttentionItem
	RecentActivity   []workspacev1.LiveActivityItem
	ReleaseReadiness []workspacev1.ReleaseReadiness
}

type projectHomeAggregate struct {
	Project       entity.Projects
	BlockingCount int
	LatestRun     *entity.AcceptanceRuns
}

func loadWorkspaceHomeData(ctx context.Context) (*workspaceHomeData, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	projects, err := listWorkspaceProjects(ctx, db, workspaceHomeProjectLimit)
	if err != nil {
		return nil, err
	}

	needAttention, pendingActions, blockedProjects, err := listWorkspaceNeedAttention(ctx, db, projects, workspaceHomeAttentionLimit)
	if err != nil {
		return nil, err
	}

	recentActivity, err := listWorkspaceRecentActivity(ctx, db, projects, workspaceHomeActivityLimit)
	if err != nil {
		return nil, err
	}

	releaseReadiness := buildWorkspaceReleaseReadiness(projects, workspaceHomeReleaseLimit)

	activeProjects := 0
	for _, item := range projects {
		if !isFinishedProjectStatus(item.Project.Status) {
			activeProjects++
		}
	}

	return &workspaceHomeData{
		TotalProjects:    len(projects),
		ActiveProjects:   activeProjects,
		BlockedProjects:  blockedProjects,
		PendingActions:   pendingActions,
		ProjectCards:     buildWorkspaceProjectCards(projects),
		NeedAttention:    needAttention,
		RecentActivity:   recentActivity,
		ReleaseReadiness: releaseReadiness,
	}, nil
}

func listWorkspaceProjects(ctx context.Context, db *sql.DB, limit int) ([]projectHomeAggregate, error) {
	query := `
SELECT
  p.id,
  p.name,
  p.project_category,
  p.goal_summary,
  p.status,
  p.production_status,
  p.workspace_root,
  COALESCE(p.repo_root, ''),
  COALESCE(p.current_plan_draft_id, ''),
  COALESCE(p.current_compiled_plan_id, ''),
  p.created_at,
  p.updated_at,
  COALESCE((
    SELECT COUNT(1)
    FROM ` + dao.AcceptanceIssues.Table() + ` ai
    WHERE ai.project_id = p.id
      AND ai.blocking = 1
  ), 0) AS blocking_count
FROM ` + dao.Projects.Table() + ` p
ORDER BY p.updated_at DESC, p.created_at DESC
LIMIT ?`

	rows, err := db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, gerror.Wrap(err, "query workspace projects failed")
	}
	defer rows.Close()

	items := make([]projectHomeAggregate, 0, limit)
	for rows.Next() {
		var item projectHomeAggregate
		if err = rows.Scan(
			&item.Project.Id,
			&item.Project.Name,
			&item.Project.ProjectCategory,
			&item.Project.GoalSummary,
			&item.Project.Status,
			&item.Project.ProductionStatus,
			&item.Project.WorkspaceRoot,
			&item.Project.RepoRoot,
			&item.Project.CurrentPlanDraftId,
			&item.Project.CurrentCompiledPlanId,
			&item.Project.CreatedAt,
			&item.Project.UpdatedAt,
			&item.BlockingCount,
		); err != nil {
			return nil, gerror.Wrap(err, "scan workspace project failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate workspace projects failed")
	}

	for i := range items {
		run, runErr := getLatestAcceptanceRunByProjectID(ctx, db, items[i].Project.Id)
		if runErr != nil {
			return nil, runErr
		}
		items[i].LatestRun = run
	}

	return items, nil
}

func getLatestAcceptanceRunByProjectID(ctx context.Context, db *sql.DB, projectID string) (*entity.AcceptanceRuns, error) {
	query := `
SELECT
  id,
  project_id,
  COALESCE(task_id, ''),
  profile_version,
  status,
  functional_status,
  production_status,
  manual_release_required,
  created_at,
  COALESCE(finished_at, '')
FROM ` + dao.AcceptanceRuns.Table() + `
WHERE project_id = ?
ORDER BY created_at DESC
LIMIT 1`

	row := db.QueryRowContext(ctx, query, projectID)
	var run entity.AcceptanceRuns
	if err := row.Scan(
		&run.Id,
		&run.ProjectId,
		&run.TaskId,
		&run.ProfileVersion,
		&run.Status,
		&run.FunctionalStatus,
		&run.ProductionStatus,
		&run.ManualReleaseRequired,
		&run.CreatedAt,
		&run.FinishedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query latest acceptance run failed")
	}
	return &run, nil
}

func listWorkspaceNeedAttention(ctx context.Context, db *sql.DB, projects []projectHomeAggregate, limit int) ([]workspacev1.NeedAttentionItem, int, int, error) {
	query := `
SELECT
  ai.id,
  ai.project_id,
  ai.summary,
  ai.severity,
  ai.blocking
FROM ` + dao.AcceptanceIssues.Table() + ` ai
ORDER BY ai.created_at DESC
LIMIT ?`

	rows, err := db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, 0, 0, gerror.Wrap(err, "query workspace need-attention failed")
	}
	defer rows.Close()

	items := make([]workspacev1.NeedAttentionItem, 0, limit)
	blockedProjectSet := make(map[string]struct{})
	for rows.Next() {
		var (
			item     workspacev1.NeedAttentionItem
			summary  string
			severity string
			blocking int
		)
		if err = rows.Scan(&item.ItemID, &item.ProjectID, &summary, &severity, &blocking); err != nil {
			return nil, 0, 0, gerror.Wrap(err, "scan workspace need-attention failed")
		}
		item.Title = summary
		item.Severity = normalizeSeverity(severity)
		item.IsBlocking = blocking == 1
		item.RecommendedAct = "open_project_workspace"
		items = append(items, item)
		if item.IsBlocking {
			blockedProjectSet[item.ProjectID] = struct{}{}
		}
	}
	if err = rows.Err(); err != nil {
		return nil, 0, 0, gerror.Wrap(err, "iterate workspace need-attention failed")
	}

	if len(items) == 0 {
		for _, project := range projects {
			if len(items) >= limit {
				break
			}
			if isProductionReady(project.Project.ProductionStatus) {
				continue
			}
			items = append(items, workspacev1.NeedAttentionItem{
				ItemID:         "derived_attention_" + project.Project.Id,
				ProjectID:      project.Project.Id,
				Title:          "Project is not production-ready yet",
				Severity:       deriveAttentionSeverity(project),
				IsBlocking:     project.BlockingCount > 0,
				RecommendedAct: "open_project_workspace",
			})
			if project.BlockingCount > 0 {
				blockedProjectSet[project.Project.Id] = struct{}{}
			}
		}
	}

	return items, len(items), len(blockedProjectSet), nil
}

func listWorkspaceRecentActivity(ctx context.Context, db *sql.DB, projects []projectHomeAggregate, limit int) ([]workspacev1.LiveActivityItem, error) {
	query := `
SELECT
  id,
  COALESCE(project_id, ''),
  event_type,
  summary,
  actor_kind,
  created_at
FROM ` + dao.AuditLogs.Table() + `
ORDER BY created_at DESC
LIMIT ?`

	rows, err := db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, gerror.Wrap(err, "query workspace recent activity failed")
	}
	defer rows.Close()

	items := make([]workspacev1.LiveActivityItem, 0, limit)
	for rows.Next() {
		var item workspacev1.LiveActivityItem
		if err = rows.Scan(
			&item.EventID,
			&item.ProjectID,
			&item.EventType,
			&item.Title,
			&item.SourceBrain,
			&item.OccurredAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan workspace recent activity failed")
		}
		item.NeedsAttention = false
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate workspace recent activity failed")
	}

	if len(items) > 0 {
		return items, nil
	}

	for _, project := range projects {
		if len(items) >= limit {
			break
		}
		items = append(items, workspacev1.LiveActivityItem{
			EventID:        "derived_activity_" + project.Project.Id,
			ProjectID:      project.Project.Id,
			EventType:      "project_created",
			Title:          "Project available in workspace",
			SourceBrain:    "system",
			OccurredAt:     firstNonEmpty(project.Project.UpdatedAt, project.Project.CreatedAt),
			NeedsAttention: project.BlockingCount > 0,
		})
	}
	return items, nil
}

func buildWorkspaceProjectCards(projects []projectHomeAggregate) []workspacev1.ProjectCard {
	cards := make([]workspacev1.ProjectCard, 0, len(projects))
	for _, item := range projects {
		cards = append(cards, workspacev1.ProjectCard{
			ProjectID:        item.Project.Id,
			Name:             item.Project.Name,
			ProjectCategory:  item.Project.ProjectCategory,
			CurrentStage:     normalizeProjectStage(item.Project.Status),
			StageStatus:      deriveProjectStageStatus(item.Project.Status, item.BlockingCount),
			ProgressPercent:  deriveProjectProgress(item.Project.Status, item.Project.ProductionStatus),
			ProductionStatus: normalizeProductionStatus(item.Project.ProductionStatus),
		})
	}
	return cards
}

func buildWorkspaceReleaseReadiness(projects []projectHomeAggregate, limit int) []workspacev1.ReleaseReadiness {
	items := make([]workspacev1.ReleaseReadiness, 0, minInt(len(projects), limit))
	for _, item := range projects {
		if len(items) >= limit {
			break
		}
		missingItems := deriveReleaseMissingItems(item)
		if missingItems == 0 && isProductionReady(item.Project.ProductionStatus) {
			continue
		}
		items = append(items, workspacev1.ReleaseReadiness{
			ProjectID:        item.Project.Id,
			Name:             item.Project.Name,
			ProductionStatus: normalizeProductionStatus(item.Project.ProductionStatus),
			MissingItems:     missingItems,
		})
	}
	return items
}

func deriveReleaseMissingItems(item projectHomeAggregate) int {
	if item.BlockingCount > 0 {
		return item.BlockingCount
	}
	if item.LatestRun == nil {
		if isProductionReady(item.Project.ProductionStatus) {
			return 0
		}
		return 1
	}

	missing := 0
	if !isFunctionalPassed(item.LatestRun.FunctionalStatus) {
		missing++
	}
	if !isProductionReady(item.LatestRun.ProductionStatus) {
		missing++
	}
	if item.LatestRun.ManualReleaseRequired == 1 && !isFinishedProjectStatus(item.Project.Status) {
		missing++
	}
	return missing
}

func deriveAttentionSeverity(item projectHomeAggregate) string {
	if item.BlockingCount > 0 {
		return "error"
	}
	if strings.EqualFold(item.Project.ProductionStatus, "pending") {
		return "warning"
	}
	return "warning"
}

func normalizeProjectStage(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "created"
	}
	return status
}

func deriveProjectStageStatus(status string, blockingCount int) string {
	if blockingCount > 0 {
		return "blocked"
	}
	if isFinishedProjectStatus(status) {
		return "completed"
	}
	return "running"
}

func deriveProjectProgress(status string, productionStatus string) int {
	switch strings.TrimSpace(status) {
	case "created":
		return 5
	case "planning", "plan_draft":
		return 18
	case "plan_review":
		return 32
	case "compiled", "execution_ready":
		return 48
	case "executing":
		return 68
	case "acceptance":
		return 84
	case "completed":
		if isProductionReady(productionStatus) {
			return 100
		}
		return 92
	default:
		if isProductionReady(productionStatus) {
			return 100
		}
		return 12
	}
}

func normalizeProductionStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "pending"
	}
	return status
}

func normalizeSeverity(severity string) string {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		return "critical"
	case "error":
		return "error"
	default:
		return "warning"
	}
}

func isProductionReady(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "production_passed") ||
		strings.EqualFold(strings.TrimSpace(status), "production_ready")
}

func isFunctionalPassed(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "functional_passed") ||
		isProductionReady(status)
}

func isFinishedProjectStatus(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "completed")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return nowText()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
