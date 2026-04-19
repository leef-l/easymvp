package service

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	projectsv1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

const (
	projectWorkspaceActivityLimit = 12
	projectWorkspaceInboxLimit    = 10
)

type projectWorkspaceData struct {
	ProjectSnapshot      projectsv1.ProjectSnapshot
	StageProgress        []projectsv1.StageProgressItem
	LiveActivity         []projectsv1.LiveActivityItem
	ActionInbox          []projectsv1.ActionInboxItem
	AcceptanceCoverage   projectsv1.AcceptanceCoverage
	WorkspaceExplanation projectsv1.WorkspaceExplanation
}

type projectWorkspaceAggregate struct {
	Project             entity.Projects
	AcceptanceProfile   *acceptanceProfileRecord
	ProductionProfile   *productionAcceptanceProfileRecord
	RepairDraft         *repairPlanDraftRecord
	Tasks               []entity.DomainTasks
	RunBindings         []entity.BrainRunBindings
	AcceptanceRuns      []entity.AcceptanceRuns
	AcceptanceIssues    []entity.AcceptanceIssues
	AuditLogs           []entity.AuditLogs
	LatestAcceptanceRun *entity.AcceptanceRuns
}

func loadProjectWorkspaceData(ctx context.Context, projectID string) (*projectWorkspaceData, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	aggregate, err := loadProjectWorkspaceAggregate(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	projectSnapshot := buildProjectSnapshot(aggregate)
	stageProgress := buildProjectStageProgress(aggregate)
	liveActivity := buildProjectLiveActivity(aggregate)
	actionInbox := buildProjectActionInbox(aggregate)
	acceptanceCoverage := buildProjectAcceptanceCoverage(aggregate)
	workspaceExplanation := buildProjectWorkspaceExplanation(ctx, aggregate, projectSnapshot, stageProgress, liveActivity, actionInbox, acceptanceCoverage)

	return &projectWorkspaceData{
		ProjectSnapshot:      projectSnapshot,
		StageProgress:        stageProgress,
		LiveActivity:         liveActivity,
		ActionInbox:          actionInbox,
		AcceptanceCoverage:   acceptanceCoverage,
		WorkspaceExplanation: workspaceExplanation,
	}, nil
}

func buildProjectWorkspaceExplanation(
	ctx context.Context,
	data *projectWorkspaceAggregate,
	projectSnapshot projectsv1.ProjectSnapshot,
	stageProgress []projectsv1.StageProgressItem,
	liveActivity []projectsv1.LiveActivityItem,
	actionInbox []projectsv1.ActionInboxItem,
	acceptanceCoverage projectsv1.AcceptanceCoverage,
) projectsv1.WorkspaceExplanation {
	input := braincontracts.WorkspaceExplanationInput{
		WorkspaceContextJSON: mustMarshalRawJSON(map[string]any{
			"project_snapshot":     projectSnapshot,
			"stage_progress":       stageProgress,
			"live_activity":        liveActivity,
			"action_inbox":         actionInbox,
			"acceptance_coverage":  acceptanceCoverage,
			"task_count":           len(data.Tasks),
			"run_binding_count":    len(data.RunBindings),
			"acceptance_run_count": len(data.AcceptanceRuns),
		}),
		RiskSummaryJSON: mustMarshalRawJSON(map[string]any{
			"risk_level":           deriveProjectRiskLevel(data),
			"blocking_issue_count": countBlockingIssues(data.AcceptanceIssues),
			"production_status":    deriveProjectProductionStatus(data),
		}),
		LatestDecisionSummaryJSON: mustMarshalRawJSON(map[string]any{
			"current_stage":       projectSnapshot.CurrentStage,
			"production_status":   projectSnapshot.ProductionStatus,
			"manual_release_need": projectSnapshot.ManualReleaseNeed,
			"top_action_count":    len(actionInbox),
			"active_item_title":   deriveCurrentActiveItemTitle(data),
		}),
	}

	_, result, err := EasyMVPBrain().CallWorkspaceExplanation(ctx, input)
	if err != nil || result == nil {
		return buildProjectWorkspaceExplanationFallback(data, actionInbox, err)
	}

	view := projectsv1.WorkspaceExplanation{
		Headline:     result.Headline,
		Summary:      result.Summary,
		TopBlockers:  result.TopBlockers,
		ExplainLinks: result.ExplainLinks,
	}
	for _, item := range result.RecommendedActions {
		view.RecommendedActions = append(view.RecommendedActions, projectsv1.WorkspaceRecommendedAction{
			ActionKey: item.ActionKey,
			Label:     item.Label,
			Reason:    item.Reason,
			DeepLink:  item.DeepLink,
		})
	}
	if strings.TrimSpace(view.Headline) == "" || strings.TrimSpace(view.Summary) == "" {
		return buildProjectWorkspaceExplanationFallback(data, actionInbox, nil)
	}
	return view
}

func buildProjectWorkspaceExplanationFallback(data *projectWorkspaceAggregate, actionInbox []projectsv1.ActionInboxItem, err error) projectsv1.WorkspaceExplanation {
	switch deriveWorkspaceExplanationFallbackMode(data, err) {
	case "denied":
		return buildRuntimeLimitedWorkspaceExplanation(data, "denied")
	case "unsupported":
		return buildRuntimeLimitedWorkspaceExplanation(data, "unsupported")
	default:
		return buildFallbackWorkspaceExplanation(data, actionInbox)
	}
}

func buildFallbackWorkspaceExplanation(data *projectWorkspaceAggregate, actionInbox []projectsv1.ActionInboxItem) projectsv1.WorkspaceExplanation {
	view := projectsv1.WorkspaceExplanation{
		Headline: "Project status overview",
		Summary:  "Workspace explanation is temporarily using local fallback summary.",
	}
	for _, item := range actionInbox {
		view.TopBlockers = append(view.TopBlockers, item.Title)
		view.RecommendedActions = append(view.RecommendedActions, projectsv1.WorkspaceRecommendedAction{
			ActionKey: item.RecommendedAction,
			Label:     item.Title,
			Reason:    item.Severity,
			DeepLink:  item.TargetID,
		})

		if len(view.TopBlockers) >= 3 || len(view.RecommendedActions) >= 3 {
			break
		}
	}
	if len(view.TopBlockers) == 0 {
		view.TopBlockers = []string{"No blocking issue is currently detected."}
	}
	if len(view.RecommendedActions) == 0 {
		view.RecommendedActions = []projectsv1.WorkspaceRecommendedAction{
			{
				ActionKey: "open_project_plan",
				Label:     "Open project plan",
				Reason:    "Review the current project stage and next planned action.",
				DeepLink:  data.Project.Id,
			},
		}
	}
	return view
}

func buildRuntimeLimitedWorkspaceExplanation(data *projectWorkspaceAggregate, mode string) projectsv1.WorkspaceExplanation {
	view := projectsv1.WorkspaceExplanation{
		ExplainLinks: []string{"runtime", "task_review"},
	}
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "denied":
		view.Headline = "Workspace explanation blocked by runtime policy"
		view.Summary = "The latest workspace explanation could not be generated because the runtime denied a required capability. Review the affected task and continue with manual follow-up."
	case "unsupported":
		view.Headline = "Workspace explanation limited by runtime capability"
		view.Summary = "The latest workspace explanation could not be generated because the runtime reported an unsupported capability. Review the affected task and continue with fallback handling."
	default:
		return buildFallbackWorkspaceExplanation(data, nil)
	}

	targetStatus := "run_" + strings.ToLower(strings.TrimSpace(mode))
	for _, binding := range data.RunBindings {
		if normalizeBindingStatus(binding.RunStatus) != targetStatus {
			continue
		}
		view.TopBlockers = append(view.TopBlockers, deriveBindingInboxTitle(binding, data.Tasks))
		view.RecommendedActions = append(view.RecommendedActions, projectsv1.WorkspaceRecommendedAction{
			ActionKey: "open_task_review",
			Label:     deriveBindingTitle(binding, data.Tasks),
			Reason:    deriveBindingInboxTitle(binding, data.Tasks),
			DeepLink:  firstNonEmpty(strings.TrimSpace(binding.TaskId), binding.Id),
		})
		if len(view.TopBlockers) >= 3 || len(view.RecommendedActions) >= 3 {
			break
		}
	}
	if len(view.TopBlockers) == 0 {
		switch targetStatus {
		case "run_denied":
			view.TopBlockers = []string{"Runtime policy denied the latest workspace explanation request."}
		case "run_unsupported":
			view.TopBlockers = []string{"Runtime reported an unsupported capability for the latest workspace explanation request."}
		}
	}
	if len(view.RecommendedActions) == 0 {
		view.RecommendedActions = []projectsv1.WorkspaceRecommendedAction{
			{
				ActionKey: "open_project_plan",
				Label:     "Open project plan",
				Reason:    "Use the project plan and action inbox while workspace explanation is unavailable.",
				DeepLink:  data.Project.Id,
			},
		}
	}
	return view
}

func loadProjectWorkspaceAggregate(ctx context.Context, db *sql.DB, projectID string) (*projectWorkspaceAggregate, error) {
	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	tasks, err := listProjectDomainTasks(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	runBindings, err := listProjectBrainRunBindings(ctx, db, projectID, projectWorkspaceActivityLimit)
	if err != nil {
		return nil, err
	}
	acceptanceRuns, err := listProjectAcceptanceRuns(ctx, db, projectID, projectWorkspaceActivityLimit)
	if err != nil {
		return nil, err
	}
	acceptanceIssues, err := listProjectAcceptanceIssues(ctx, db, projectID, projectWorkspaceInboxLimit)
	if err != nil {
		return nil, err
	}
	auditLogs, err := listProjectAuditLogs(ctx, db, projectID, projectWorkspaceActivityLimit)
	if err != nil {
		return nil, err
	}

	aggregate := &projectWorkspaceAggregate{
		Project:          *project,
		Tasks:            tasks,
		RunBindings:      runBindings,
		AcceptanceRuns:   acceptanceRuns,
		AcceptanceIssues: acceptanceIssues,
		AuditLogs:        auditLogs,
	}
	aggregate.AcceptanceProfile, err = getLatestAcceptanceProfile(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	aggregate.ProductionProfile, err = getLatestProductionAcceptanceProfile(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	aggregate.RepairDraft, err = getLatestRepairDraftForProject(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	if len(acceptanceRuns) > 0 {
		aggregate.LatestAcceptanceRun = &acceptanceRuns[0]
	}
	return aggregate, nil
}

func getProjectByID(ctx context.Context, db *sql.DB, projectID string) (*entity.Projects, error) {
	query := `
SELECT
  id,
  name,
  project_category,
  goal_summary,
  status,
  production_status,
  workspace_root,
  COALESCE(repo_root, ''),
  COALESCE(current_plan_draft_id, ''),
  COALESCE(current_compiled_plan_id, ''),
  created_at,
  updated_at
FROM ` + dao.Projects.Table() + `
WHERE id = ?
LIMIT 1`

	row := db.QueryRowContext(ctx, query, projectID)
	var project entity.Projects
	if err := row.Scan(
		&project.Id,
		&project.Name,
		&project.ProjectCategory,
		&project.GoalSummary,
		&project.Status,
		&project.ProductionStatus,
		&project.WorkspaceRoot,
		&project.RepoRoot,
		&project.CurrentPlanDraftId,
		&project.CurrentCompiledPlanId,
		&project.CreatedAt,
		&project.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.Newf("project not found: %s", projectID)
		}
		return nil, gerror.Wrap(err, "query project by id failed")
	}
	return &project, nil
}

func listProjectDomainTasks(ctx context.Context, db *sql.DB, projectID string) ([]entity.DomainTasks, error) {
	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	var (
		query string
		args  []any
	)
	if strings.TrimSpace(project.CurrentCompiledPlanId) != "" {
		query = `
SELECT
  id,
  project_id,
  COALESCE(source_compiled_plan_id, ''),
  COALESCE(source_compiled_task_id, ''),
  COALESCE(source_task_key, ''),
  compiled_version,
  name,
  phase,
  task_kind,
  role_type,
  brain_kind,
  risk_level,
  status,
  manual_review_required,
  created_at,
  updated_at
FROM ` + dao.DomainTasks.Table() + `
WHERE project_id = ? AND source_compiled_plan_id = ?
ORDER BY updated_at DESC, created_at DESC`
		args = []any{projectID, project.CurrentCompiledPlanId}
	} else {
		query = `
SELECT
  id,
  project_id,
  COALESCE(source_compiled_plan_id, ''),
  COALESCE(source_compiled_task_id, ''),
  COALESCE(source_task_key, ''),
  compiled_version,
  name,
  phase,
  task_kind,
  role_type,
  brain_kind,
  risk_level,
  status,
  manual_review_required,
  created_at,
  updated_at
FROM ` + dao.DomainTasks.Table() + `
WHERE project_id = ?
ORDER BY updated_at DESC, created_at DESC`
		args = []any{projectID}
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, gerror.Wrap(err, "query project domain tasks failed")
	}
	defer rows.Close()

	items := make([]entity.DomainTasks, 0)
	for rows.Next() {
		var item entity.DomainTasks
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.SourceCompiledPlanId,
			&item.SourceCompiledTaskId,
			&item.SourceTaskKey,
			&item.CompiledVersion,
			&item.Name,
			&item.Phase,
			&item.TaskKind,
			&item.RoleType,
			&item.BrainKind,
			&item.RiskLevel,
			&item.Status,
			&item.ManualReviewRequired,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan project domain task failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate project domain tasks failed")
	}
	return items, nil
}

func listProjectBrainRunBindings(ctx context.Context, db *sql.DB, projectID string, limit int) ([]entity.BrainRunBindings, error) {
	query := `
SELECT
  id,
  project_id,
  COALESCE(task_id, ''),
  brain_kind,
  brain_run_id,
  run_status,
  COALESCE(started_at, ''),
  COALESCE(finished_at, ''),
  COALESCE(last_sync_at, ''),
  created_at,
  updated_at
FROM ` + dao.BrainRunBindings.Table() + `
WHERE project_id = ?
ORDER BY COALESCE(last_sync_at, updated_at, started_at, created_at) DESC
LIMIT ?`

	rows, err := db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, gerror.Wrap(err, "query project brain run bindings failed")
	}
	defer rows.Close()

	items := make([]entity.BrainRunBindings, 0, limit)
	for rows.Next() {
		var item entity.BrainRunBindings
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.TaskId,
			&item.BrainKind,
			&item.BrainRunId,
			&item.RunStatus,
			&item.StartedAt,
			&item.FinishedAt,
			&item.LastSyncAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan project brain run binding failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate project brain run bindings failed")
	}
	return items, nil
}

func listProjectAcceptanceRuns(ctx context.Context, db *sql.DB, projectID string, limit int) ([]entity.AcceptanceRuns, error) {
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
LIMIT ?`

	rows, err := db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, gerror.Wrap(err, "query project acceptance runs failed")
	}
	defer rows.Close()

	items := make([]entity.AcceptanceRuns, 0, limit)
	for rows.Next() {
		var item entity.AcceptanceRuns
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.TaskId,
			&item.ProfileVersion,
			&item.Status,
			&item.FunctionalStatus,
			&item.ProductionStatus,
			&item.ManualReleaseRequired,
			&item.CreatedAt,
			&item.FinishedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan project acceptance run failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate project acceptance runs failed")
	}
	return items, nil
}

func listProjectAcceptanceIssues(ctx context.Context, db *sql.DB, projectID string, limit int) ([]entity.AcceptanceIssues, error) {
	query := `
SELECT
  id,
  project_id,
  acceptance_run_id,
  severity,
  issue_kind,
  blocking,
  summary,
  COALESCE(detail_json, ''),
  created_at
FROM ` + dao.AcceptanceIssues.Table() + `
WHERE project_id = ?
ORDER BY created_at DESC
LIMIT ?`

	rows, err := db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, gerror.Wrap(err, "query project acceptance issues failed")
	}
	defer rows.Close()

	items := make([]entity.AcceptanceIssues, 0, limit)
	for rows.Next() {
		var item entity.AcceptanceIssues
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.AcceptanceRunId,
			&item.Severity,
			&item.IssueKind,
			&item.Blocking,
			&item.Summary,
			&item.DetailJson,
			&item.CreatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan project acceptance issue failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate project acceptance issues failed")
	}
	return items, nil
}

func listProjectAuditLogs(ctx context.Context, db *sql.DB, projectID string, limit int) ([]entity.AuditLogs, error) {
	query := `
SELECT
  id,
  project_id,
  event_type,
  actor_kind,
  summary,
  COALESCE(payload_json, ''),
  created_at
FROM ` + dao.AuditLogs.Table() + `
WHERE project_id = ?
ORDER BY created_at DESC
LIMIT ?`

	rows, err := db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, gerror.Wrap(err, "query project audit logs failed")
	}
	defer rows.Close()

	items := make([]entity.AuditLogs, 0, limit)
	for rows.Next() {
		var item entity.AuditLogs
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.EventType,
			&item.ActorKind,
			&item.Summary,
			&item.PayloadJson,
			&item.CreatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan project audit log failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate project audit logs failed")
	}
	return items, nil
}

func buildProjectSnapshot(data *projectWorkspaceAggregate) projectsv1.ProjectSnapshot {
	return projectsv1.ProjectSnapshot{
		ProjectID:         data.Project.Id,
		Name:              data.Project.Name,
		ProjectCategory:   data.Project.ProjectCategory,
		CurrentStage:      normalizeProjectStage(data.Project.Status),
		ProgressPercent:   deriveProjectProgress(data.Project.Status, data.Project.ProductionStatus),
		RiskLevel:         deriveProjectRiskLevel(data),
		ProductionStatus:  deriveProjectProductionStatus(data),
		ManualReleaseNeed: data.LatestAcceptanceRun != nil && data.LatestAcceptanceRun.ManualReleaseRequired == 1,
	}
}

func buildProjectStageProgress(data *projectWorkspaceAggregate) []projectsv1.StageProgressItem {
	stageDefs := []struct {
		Key  string
		Name string
	}{
		{Key: "design", Name: "Design"},
		{Key: "review", Name: "Review"},
		{Key: "compile", Name: "Compile"},
		{Key: "execute", Name: "Execute"},
		{Key: "acceptance", Name: "Acceptance"},
		{Key: "complete", Name: "Complete"},
	}

	currentKey := mapProjectStatusToWorkspaceStage(data.Project.Status)
	currentIndex := 0
	for i, def := range stageDefs {
		if def.Key == currentKey {
			currentIndex = i
			break
		}
	}

	activeTitle := deriveCurrentActiveItemTitle(data)
	blockingCount := countBlockingIssues(data.AcceptanceIssues)
	stageDuration := deriveCurrentStageDurationSeconds(data)

	items := make([]projectsv1.StageProgressItem, 0, len(stageDefs))
	for i, def := range stageDefs {
		status := "pending"
		duration := int64(0)
		active := ""
		stageBlocking := 0

		switch {
		case i < currentIndex:
			status = "completed"
		case i == currentIndex:
			status = "running"
			if blockingCount > 0 {
				status = "blocked"
				stageBlocking = blockingCount
			}
			duration = stageDuration
			active = activeTitle
			if def.Key == "acceptance" {
				stageBlocking = blockingCount
			}
		default:
			if def.Key == "complete" && isFinishedProjectStatus(data.Project.Status) {
				status = "completed"
			}
		}

		if def.Key == "complete" && isFinishedProjectStatus(data.Project.Status) {
			status = "completed"
			duration = stageDuration
		}

		items = append(items, projectsv1.StageProgressItem{
			StageKey:         def.Key,
			StageName:        def.Name,
			Status:           status,
			DurationSeconds:  duration,
			ActiveItemTitle:  active,
			BlockingIssueCnt: stageBlocking,
		})
	}
	return items
}

func buildProjectLiveActivity(data *projectWorkspaceAggregate) []projectsv1.LiveActivityItem {
	items := make([]projectsv1.LiveActivityItem, 0, projectWorkspaceActivityLimit)
	for _, logItem := range data.AuditLogs {
		if len(items) >= projectWorkspaceActivityLimit {
			break
		}
		items = append(items, projectsv1.LiveActivityItem{
			EventID:        logItem.Id,
			EventType:      firstNonEmpty(logItem.EventType, "audit_event"),
			Title:          firstNonEmpty(logItem.Summary, "Project activity updated"),
			SourceBrain:    firstNonEmpty(logItem.ActorKind, "system"),
			SourceTaskID:   "",
			OccurredAt:     firstNonEmpty(logItem.CreatedAt, data.Project.UpdatedAt, data.Project.CreatedAt),
			RequiresAction: false,
		})
	}
	if len(items) > 0 {
		return items
	}

	for _, binding := range data.RunBindings {
		if len(items) >= projectWorkspaceActivityLimit {
			break
		}
		items = append(items, projectsv1.LiveActivityItem{
			EventID:        binding.Id,
			EventType:      bindingActivityEventType(binding.RunStatus),
			Title:          deriveBindingTitle(binding, data.Tasks),
			SourceBrain:    firstNonEmpty(binding.BrainKind, "brain"),
			SourceTaskID:   binding.TaskId,
			OccurredAt:     firstNonEmpty(binding.LastSyncAt, binding.UpdatedAt, binding.StartedAt, binding.CreatedAt),
			RequiresAction: bindingNeedsAttention(binding.RunStatus),
		})
	}
	if len(items) > 0 {
		return items
	}

	if activeTask := findLatestTask(data.Tasks); activeTask != nil {
		items = append(items, projectsv1.LiveActivityItem{
			EventID:        "derived_task_" + activeTask.Id,
			EventType:      "task_" + normalizeTaskStatus(activeTask.Status),
			Title:          firstNonEmpty(activeTask.Name, "Task available"),
			SourceBrain:    firstNonEmpty(activeTask.BrainKind, "system"),
			SourceTaskID:   activeTask.Id,
			OccurredAt:     firstNonEmpty(activeTask.UpdatedAt, activeTask.CreatedAt, data.Project.UpdatedAt),
			RequiresAction: activeTask.ManualReviewRequired == 1,
		})
		return items
	}

	return []projectsv1.LiveActivityItem{
		{
			EventID:        "derived_project_" + data.Project.Id,
			EventType:      "project_" + normalizeProjectStage(data.Project.Status),
			Title:          "Project workspace is ready",
			SourceBrain:    "system",
			SourceTaskID:   "",
			OccurredAt:     firstNonEmpty(data.Project.UpdatedAt, data.Project.CreatedAt),
			RequiresAction: countBlockingIssues(data.AcceptanceIssues) > 0,
		},
	}
}

func buildProjectActionInbox(data *projectWorkspaceAggregate) []projectsv1.ActionInboxItem {
	items := make([]projectsv1.ActionInboxItem, 0, projectWorkspaceInboxLimit)
	if data.RepairDraft != nil {
		items = append(items, projectsv1.ActionInboxItem{
			ItemID:            data.RepairDraft.ID,
			Title:             "Repair plan draft is ready for review",
			Severity:          "error",
			IsBlocking:        true,
			RecommendedAction: "open_repair_draft",
			TargetID:          data.RepairDraft.ID,
		})
	}
	for _, issue := range data.AcceptanceIssues {
		if len(items) >= projectWorkspaceInboxLimit {
			break
		}
		items = append(items, projectsv1.ActionInboxItem{
			ItemID:            issue.Id,
			Title:             firstNonEmpty(issue.Summary, "Acceptance issue requires review"),
			Severity:          normalizeSeverity(issue.Severity),
			IsBlocking:        issue.Blocking == 1,
			RecommendedAction: deriveIssueRecommendedAction(issue),
			TargetID:          firstNonEmpty(issue.AcceptanceRunId, data.Project.Id),
		})
	}

	if len(items) < projectWorkspaceInboxLimit {
		for _, binding := range data.RunBindings {
			if len(items) >= projectWorkspaceInboxLimit {
				break
			}
			if !bindingNeedsAttention(binding.RunStatus) {
				continue
			}
			items = append(items, projectsv1.ActionInboxItem{
				ItemID:            "run_attention_" + binding.Id,
				Title:             deriveBindingInboxTitle(binding, data.Tasks),
				Severity:          deriveBindingSeverity(binding.RunStatus),
				IsBlocking:        isBlockingBindingStatus(binding.RunStatus),
				RecommendedAction: "open_task_review",
				TargetID:          firstNonEmpty(binding.TaskId, binding.Id),
			})
		}
	}

	if len(items) < projectWorkspaceInboxLimit {
		for _, task := range data.Tasks {
			if len(items) >= projectWorkspaceInboxLimit {
				break
			}
			if task.ManualReviewRequired != 1 {
				continue
			}
			items = append(items, projectsv1.ActionInboxItem{
				ItemID:            "task_review_" + task.Id,
				Title:             "Manual review required: " + firstNonEmpty(task.Name, task.TaskKind, task.Id),
				Severity:          deriveTaskSeverity(task),
				IsBlocking:        isBlockingTaskStatus(task.Status),
				RecommendedAction: "open_task_review",
				TargetID:          task.Id,
			})
		}
	}

	if len(items) < projectWorkspaceInboxLimit && data.LatestAcceptanceRun != nil && data.LatestAcceptanceRun.ManualReleaseRequired == 1 && !containsRecommendedAction(items, "open_acceptance_center") {
		items = append(items, projectsv1.ActionInboxItem{
			ItemID:            "manual_release_" + data.LatestAcceptanceRun.Id,
			Title:             "Manual release confirmation is required",
			Severity:          "warning",
			IsBlocking:        false,
			RecommendedAction: "open_acceptance_center",
			TargetID:          data.LatestAcceptanceRun.Id,
		})
	}

	if len(items) == 0 && !isProductionReady(deriveProjectProductionStatus(data)) {
		items = append(items, projectsv1.ActionInboxItem{
			ItemID:            "project_progress_" + data.Project.Id,
			Title:             "Project is not production-ready yet",
			Severity:          deriveFallbackProjectSeverity(data),
			IsBlocking:        countBlockingIssues(data.AcceptanceIssues) > 0,
			RecommendedAction: "open_project_plan",
			TargetID:          data.Project.Id,
		})
	}

	return items
}

func buildProjectAcceptanceCoverage(data *projectWorkspaceAggregate) projectsv1.AcceptanceCoverage {
	requiredSurfaces, requiredJourneys, requiredEvidence := acceptanceCoverageRequirements(data.Project.ProjectCategory)
	if data.ProductionProfile != nil {
		if items := parseStringArrayJSON(data.ProductionProfile.RequiredSurfacesJSON); len(items) > 0 {
			requiredSurfaces = len(items)
		}
		if items := parseStringArrayJSON(data.ProductionProfile.RequiredJourneysJSON); len(items) > 0 {
			requiredJourneys = len(items)
		}
		if items := parseStringArrayJSON(data.ProductionProfile.RequiredEvidenceJSON); len(items) > 0 {
			requiredEvidence = len(items)
		}
	}
	coveredSurfaces := 0
	coveredJourneys := 0
	evidenceReady := 0
	productionPassed := false

	if data.LatestAcceptanceRun != nil {
		coveredSurfaces = 1
		coveredJourneys = 1
		evidenceReady = 1

		if isFunctionalPassed(data.LatestAcceptanceRun.FunctionalStatus) {
			coveredSurfaces = minInt(requiredSurfaces, maxInt(coveredSurfaces, 2))
			coveredJourneys = minInt(requiredJourneys, maxInt(coveredJourneys, 3))
			evidenceReady = minInt(requiredEvidence, maxInt(evidenceReady, 2))
		}
		if isProductionReady(data.LatestAcceptanceRun.ProductionStatus) {
			coveredSurfaces = requiredSurfaces
			coveredJourneys = requiredJourneys
			evidenceReady = requiredEvidence
			productionPassed = true
		}
	}

	if len(data.AcceptanceIssues) > 0 && !productionPassed {
		evidenceReady = maxInt(0, evidenceReady-minInt(len(data.AcceptanceIssues), evidenceReady))
	}

	return projectsv1.AcceptanceCoverage{
		Category:         firstNonEmpty(data.Project.ProjectCategory, "default"),
		CoveredSurfaces:  coveredSurfaces,
		RequiredSurfaces: requiredSurfaces,
		CoveredJourneys:  coveredJourneys,
		RequiredJourneys: requiredJourneys,
		EvidenceReady:    evidenceReady,
		EvidenceRequired: requiredEvidence,
		ProductionPassed: productionPassed || isProductionReady(deriveProjectProductionStatus(data)),
	}
}

func deriveProjectProductionStatus(data *projectWorkspaceAggregate) string {
	if data.LatestAcceptanceRun != nil && strings.TrimSpace(data.LatestAcceptanceRun.ProductionStatus) != "" {
		return normalizeProductionStatus(data.LatestAcceptanceRun.ProductionStatus)
	}
	return normalizeProductionStatus(data.Project.ProductionStatus)
}

func deriveProjectRiskLevel(data *projectWorkspaceAggregate) string {
	if hasCriticalIssue(data.AcceptanceIssues) || hasHighRiskTask(data.Tasks) {
		return "high"
	}
	if len(data.AcceptanceIssues) > 0 || hasMediumRiskTask(data.Tasks) {
		return "medium"
	}
	return "low"
}

func mapProjectStatusToWorkspaceStage(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "created", "planning", "plan_draft":
		return "design"
	case "plan_review", "review":
		return "review"
	case "compiled", "execution_ready":
		return "compile"
	case "executing", "running":
		return "execute"
	case "acceptance":
		return "acceptance"
	case "completed":
		return "complete"
	default:
		return "design"
	}
}

func containsRecommendedAction(items []projectsv1.ActionInboxItem, action string) bool {
	action = strings.TrimSpace(action)
	if action == "" {
		return false
	}
	for _, item := range items {
		if item.RecommendedAction == action {
			return true
		}
	}
	return false
}

func deriveCurrentActiveItemTitle(data *projectWorkspaceAggregate) string {
	if task := findActiveTask(data.Tasks); task != nil {
		return firstNonEmpty(task.Name, task.TaskKind, task.Id)
	}
	if data.LatestAcceptanceRun != nil {
		return "Acceptance run " + normalizeBindingStatus(data.LatestAcceptanceRun.Status)
	}
	if len(data.AcceptanceIssues) > 0 {
		return firstNonEmpty(data.AcceptanceIssues[0].Summary, "Acceptance issue under review")
	}
	return "Project is progressing"
}

func deriveCurrentStageDurationSeconds(data *projectWorkspaceAggregate) int64 {
	if task := findActiveTask(data.Tasks); task != nil {
		return maxInt64(1, timestampDeltaSeconds(task.CreatedAt, task.UpdatedAt))
	}
	if data.LatestAcceptanceRun != nil {
		return maxInt64(1, timestampDeltaSeconds(data.LatestAcceptanceRun.CreatedAt, firstNonEmpty(data.LatestAcceptanceRun.FinishedAt, nowText())))
	}
	return maxInt64(1, timestampDeltaSeconds(data.Project.CreatedAt, data.Project.UpdatedAt))
}

func countBlockingIssues(items []entity.AcceptanceIssues) int {
	count := 0
	for _, item := range items {
		if item.Blocking == 1 {
			count++
		}
	}
	return count
}

func hasCriticalIssue(items []entity.AcceptanceIssues) bool {
	for _, item := range items {
		if normalizeSeverity(item.Severity) == "critical" || normalizeSeverity(item.Severity) == "error" {
			return true
		}
	}
	return false
}

func hasHighRiskTask(items []entity.DomainTasks) bool {
	for _, item := range items {
		switch strings.ToLower(strings.TrimSpace(item.RiskLevel)) {
		case "high", "critical":
			return true
		}
	}
	return false
}

func hasMediumRiskTask(items []entity.DomainTasks) bool {
	for _, item := range items {
		switch strings.ToLower(strings.TrimSpace(item.RiskLevel)) {
		case "medium":
			return true
		}
	}
	return false
}

func findActiveTask(items []entity.DomainTasks) *entity.DomainTasks {
	for i := range items {
		if isActiveTaskStatus(items[i].Status) {
			return &items[i]
		}
	}
	if len(items) == 0 {
		return nil
	}
	return &items[0]
}

func findLatestTask(items []entity.DomainTasks) *entity.DomainTasks {
	if len(items) == 0 {
		return nil
	}
	return &items[0]
}

func isActiveTaskStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "running", "executing", "in_progress", "review", "pending_review", "acceptance":
		return true
	default:
		return false
	}
}

func isBlockingTaskStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "failed", "blocked":
		return true
	default:
		return false
	}
}

func normalizeTaskStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "pending"
	}
	return strings.ReplaceAll(status, " ", "_")
}

func normalizeBindingStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "running"
	}
	if !strings.HasPrefix(status, "run_") {
		status = mapBrainRunStatus(status)
	}
	return strings.ReplaceAll(status, " ", "_")
}

func deriveWorkspaceExplanationFallbackMode(data *projectWorkspaceAggregate, err error) string {
	if mode := deriveWorkspaceExplanationFallbackModeFromError(err); mode != "" {
		return mode
	}
	hasUnsupported := false
	for _, binding := range data.RunBindings {
		switch normalizeBindingStatus(binding.RunStatus) {
		case "run_denied":
			return "denied"
		case "run_unsupported":
			hasUnsupported = true
		}
	}
	if hasUnsupported {
		return "unsupported"
	}
	return ""
}

func deriveWorkspaceExplanationFallbackModeFromError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case strings.Contains(message, "run_denied"),
		strings.Contains(message, "permission_denied"),
		strings.Contains(message, "tool_denied"),
		strings.Contains(message, "forbidden"),
		strings.Contains(message, "denied by runtime policy"):
		return "denied"
	case strings.Contains(message, "run_unsupported"),
		strings.Contains(message, "tool_unsupported"),
		strings.Contains(message, "not_supported"),
		strings.Contains(message, "unsupported capability"):
		return "unsupported"
	default:
		return ""
	}
}

func bindingActivityEventType(status string) string {
	normalized := normalizeBindingStatus(status)
	return "brain_run_" + strings.TrimPrefix(normalized, "run_")
}

func bindingNeedsAttention(status string) bool {
	switch normalizeBindingStatus(status) {
	case "run_failed", "run_unsupported", "run_denied":
		return true
	default:
		return false
	}
}

func isBlockingBindingStatus(status string) bool {
	switch normalizeBindingStatus(status) {
	case "run_failed", "run_denied":
		return true
	default:
		return false
	}
}

func deriveBindingTitle(binding entity.BrainRunBindings, tasks []entity.DomainTasks) string {
	title := "Brain run " + bindingDisplayStatus(binding.RunStatus)
	for _, task := range tasks {
		if task.Id == binding.TaskId {
			return firstNonEmpty(task.Name, title)
		}
	}
	return title
}

func bindingDisplayStatus(status string) string {
	raw := strings.ToLower(strings.TrimSpace(status))
	if raw == "" {
		return "running"
	}
	if strings.HasPrefix(raw, "run_") {
		return strings.TrimPrefix(raw, "run_")
	}
	return strings.ReplaceAll(raw, " ", "_")
}

func deriveBindingInboxTitle(binding entity.BrainRunBindings, tasks []entity.DomainTasks) string {
	base := deriveBindingTitle(binding, tasks)
	switch normalizeBindingStatus(binding.RunStatus) {
	case "run_unsupported":
		return base + " requires fallback because runtime reported unsupported capability"
	case "run_denied":
		return base + " was denied by runtime policy and needs manual follow-up"
	case "run_failed":
		return base + " failed and needs review"
	default:
		return base
	}
}

func deriveBindingSeverity(status string) string {
	switch normalizeBindingStatus(status) {
	case "run_denied":
		return "error"
	case "run_unsupported":
		return "warning"
	case "run_failed":
		return "error"
	default:
		return "warning"
	}
}

func deriveIssueRecommendedAction(issue entity.AcceptanceIssues) string {
	if issue.Blocking == 1 {
		return "open_acceptance_issue"
	}
	if strings.Contains(strings.ToLower(issue.IssueKind), "release") {
		return "open_acceptance_center"
	}
	return "review_issue"
}

func deriveTaskSeverity(task entity.DomainTasks) string {
	switch strings.ToLower(strings.TrimSpace(task.RiskLevel)) {
	case "critical":
		return "critical"
	case "high":
		return "error"
	default:
		return "warning"
	}
}

func deriveFallbackProjectSeverity(data *projectWorkspaceAggregate) string {
	if countBlockingIssues(data.AcceptanceIssues) > 0 {
		return "error"
	}
	if deriveProjectRiskLevel(data) == "high" {
		return "error"
	}
	return "warning"
}

func acceptanceCoverageRequirements(category string) (requiredSurfaces int, requiredJourneys int, requiredEvidence int) {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "web":
		return 4, 6, 5
	case "game":
		return 5, 7, 6
	case "video_editing":
		return 4, 6, 5
	default:
		return 4, 6, 5
	}
}

func timestampDeltaSeconds(start string, end string) int64 {
	startTime, ok := parseTimestamp(start)
	if !ok {
		return 0
	}
	endTime, ok := parseTimestamp(end)
	if !ok {
		return 0
	}
	if endTime.Before(startTime) {
		return 0
	}
	return int64(endTime.Sub(startTime).Seconds())
}

func parseTimestamp(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
