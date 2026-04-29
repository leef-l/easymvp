package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	acceptancev1 "github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
	projectsv1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
	"github.com/leef-l/easymvp/apps/core/internal/repo"
)

const (
	projectWorkspaceActivityLimit = 12
	projectWorkspaceInboxLimit    = 10
)

type projectWorkspaceData struct {
	Overview             projectsv1.WorkspaceOverview
	ProjectSnapshot      projectsv1.ProjectSnapshot
	StageProgress        []projectsv1.StageProgressItem
	LiveActivity         []projectsv1.LiveActivityItem
	ActionInbox          []projectsv1.ActionInboxItem
	AcceptanceCoverage   projectsv1.AcceptanceCoverage
	WorkspaceExplanation projectsv1.WorkspaceExplanation
	VerificationResult   acceptancev1.VerificationResultView
	CompletionVerdict    acceptancev1.CompletionVerdictView
	RuntimeEscalation    acceptancev1.RuntimeEscalationView
	FaultSummary         acceptancev1.FaultSummaryView
	RepairPlanDraft      acceptancev1.RepairPlanDraftSummary
	HealthMetrics        projectsv1.ProjectHealthMetrics
}

type projectWorkspaceAggregate struct {
	Project                  entity.Projects
	AcceptanceProfile        *acceptanceProfileRecord
	ProductionProfile        *productionAcceptanceProfileRecord
	RepairDraft              *repairPlanDraftRecord
	Tasks                    []entity.DomainTasks
	RunBindings              []entity.BrainRunBindings
	AcceptanceRuns           []entity.AcceptanceRuns
	AcceptanceIssues         []entity.AcceptanceIssues
	AuditLogs                []entity.AuditLogs
	LatestAcceptanceRun      *entity.AcceptanceRuns
	PersistedVerdict         *acceptancev1.CompletionVerdictView
	ChannelAvailable         bool
	EnvironmentAvailable     bool
	ChannelUnavailableReason string
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
		Overview:             buildProjectWorkspaceOverview(aggregate, projectSnapshot, actionInbox),
		ProjectSnapshot:      projectSnapshot,
		StageProgress:        stageProgress,
		LiveActivity:         liveActivity,
		ActionInbox:          actionInbox,
		AcceptanceCoverage:   acceptanceCoverage,
		WorkspaceExplanation: workspaceExplanation,
		VerificationResult:   buildWorkspaceVerificationResult(ctx, aggregate, acceptanceCoverage),
		CompletionVerdict:    buildWorkspaceCompletionVerdict(ctx, aggregate, acceptanceCoverage),
		RuntimeEscalation:    buildWorkspaceRuntimeEscalation(aggregate),
		FaultSummary:         buildWorkspaceFaultSummary(aggregate),
		RepairPlanDraft:      buildWorkspaceRepairPlanDraft(aggregate),
		HealthMetrics:        buildProjectHealthMetrics(aggregate),
	}, nil
}

func buildProjectWorkspaceOverview(
	data *projectWorkspaceAggregate,
	projectSnapshot projectsv1.ProjectSnapshot,
	actionInbox []projectsv1.ActionInboxItem,
) projectsv1.WorkspaceOverview {
	completionVerdict := buildWorkspaceCompletionVerdict(context.Background(), data, buildProjectAcceptanceCoverage(data))
	nextAction := "open_project_plan"
	if strings.TrimSpace(completionVerdict.NextAction) != "" {
		nextAction = strings.TrimSpace(completionVerdict.NextAction)
	} else if len(actionInbox) > 0 {
		nextAction = firstNonEmpty(actionInbox[0].RecommendedAction, nextAction)
	}

	acceptanceRunStatus := ""
	if data.LatestAcceptanceRun != nil {
		acceptanceRunStatus = normalizeAcceptanceRunStatus(data.LatestAcceptanceRun.Status)
	}

	return projectsv1.WorkspaceOverview{
		ProjectID:            data.Project.Id,
		CurrentStage:         projectSnapshot.CurrentStage,
		StageStatus:          deriveProjectStageStatus(data.Project.Status, countBlockingIssues(data.AcceptanceIssues)),
		RiskLevel:            projectSnapshot.RiskLevel,
		ProductionStatus:     projectSnapshot.ProductionStatus,
		NextAction:           nextAction,
		ActionRequiredCount:  len(actionInbox),
		BlockingIssueCount:   countBlockingIssues(data.AcceptanceIssues),
		ManualReleaseNeeded:  projectSnapshot.ManualReleaseNeed,
		AcceptanceRunStatus:  acceptanceRunStatus,
		ManualReviewRequired: projectSnapshot.ManualReviewRequired,
		VerificationConflict: projectSnapshot.VerificationConflict,
		FaultLoopDetected:    projectSnapshot.FaultLoopDetected,
		PolicyDenied:         projectSnapshot.PolicyDenied,
	}
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
			"completion_verdict":   buildWorkspaceCompletionVerdict(context.Background(), data, acceptanceCoverage),
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

	baseView := buildStructuredWorkspaceExplanation(data, actionInbox)

	_, result, err := EasyMVPBrain().CallWorkspaceExplanation(ctx, input)
	if err != nil || result == nil {
		return baseView
	}

	view := projectsv1.WorkspaceExplanation{
		Headline:     firstNonEmpty(strings.TrimSpace(result.Headline), baseView.Headline),
		Summary:      firstNonEmpty(strings.TrimSpace(result.Summary), baseView.Summary),
		TopBlockers:  append([]string(nil), result.TopBlockers...),
		ExplainLinks: append([]string(nil), result.ExplainLinks...),
	}
	for _, item := range result.RecommendedActions {
		view.RecommendedActions = append(view.RecommendedActions, projectsv1.WorkspaceRecommendedAction{
			ActionKey: item.ActionKey,
			Label:     item.Label,
			Reason:    item.Reason,
			DeepLink:  item.DeepLink,
		})
	}
	if len(view.TopBlockers) == 0 {
		view.TopBlockers = baseView.TopBlockers
	}
	if len(view.RecommendedActions) == 0 {
		view.RecommendedActions = baseView.RecommendedActions
	}
	if len(view.ExplainLinks) == 0 {
		view.ExplainLinks = baseView.ExplainLinks
	}
	return view
}

func buildStructuredWorkspaceExplanation(data *projectWorkspaceAggregate, actionInbox []projectsv1.ActionInboxItem) projectsv1.WorkspaceExplanation {
	switch deriveWorkspaceExplanationFallbackMode(data, nil) {
	case "denied":
		return buildRuntimeLimitedWorkspaceExplanation(data, "denied")
	case "unsupported":
		return buildRuntimeLimitedWorkspaceExplanation(data, "unsupported")
	default:
		return buildDeterministicWorkspaceExplanation(data, actionInbox)
	}
}

func buildDeterministicWorkspaceExplanation(data *projectWorkspaceAggregate, actionInbox []projectsv1.ActionInboxItem) projectsv1.WorkspaceExplanation {
	view := projectsv1.WorkspaceExplanation{
		Headline: "Project status overview",
		Summary:  deriveDeterministicWorkspaceSummary(data, actionInbox),
		ExplainLinks: []string{
			"workspace",
			"acceptance",
			"runtime",
		},
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
		view.Summary = "The latest workspace explanation could not be generated because the runtime reported an unsupported capability. Review the affected task and continue with manual handling."
	default:
		return buildDeterministicWorkspaceExplanation(data, nil)
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
				Reason:    "Use the project plan and action inbox to continue delivery while runtime capability is limited.",
				DeepLink:  data.Project.Id,
			},
		}
	}
	return view
}

func deriveDeterministicWorkspaceSummary(data *projectWorkspaceAggregate, actionInbox []projectsv1.ActionInboxItem) string {
	projectName := firstNonEmpty(strings.TrimSpace(data.Project.Name), strings.TrimSpace(data.Project.Id), "project")
	productionStatus := deriveProjectProductionStatus(data)
	blockingCount := countBlockingIssues(data.AcceptanceIssues)
	nextAction := "review workspace signals"
	verdict := buildWorkspaceCompletionVerdict(context.Background(), data, buildProjectAcceptanceCoverage(data))
	if strings.TrimSpace(verdict.NextAction) != "" {
		nextAction = strings.TrimSpace(verdict.NextAction)
	} else if len(actionInbox) > 0 {
		nextAction = firstNonEmpty(strings.TrimSpace(actionInbox[0].Title), nextAction)
	}
	if strings.TrimSpace(verdict.Summary) != "" {
		return fmt.Sprintf("%s Status: %s Next action: %s", projectName, verdict.Summary, nextAction)
	}
	return fmt.Sprintf("%s is currently %s with %d blocking issue(s). Next focus: %s.", projectName, productionStatus, blockingCount, nextAction)
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
		Project:              *project,
		Tasks:                tasks,
		RunBindings:          runBindings,
		AcceptanceRuns:       acceptanceRuns,
		AcceptanceIssues:     acceptanceIssues,
		AuditLogs:            auditLogs,
		ChannelAvailable:     true,
		EnvironmentAvailable: true,
	}
	aggregate.AcceptanceProfile, err = getLatestAcceptanceProfile(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	aggregate.ProductionProfile, err = getLatestProductionAcceptanceProfile(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	aggregate.RepairDraft, err = getLatestRepairDraftForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(acceptanceRuns) > 0 {
		aggregate.LatestAcceptanceRun = &acceptanceRuns[0]
	}
	aggregate.PersistedVerdict, err = loadPersistedCompletionVerdict(ctx, projectID, latestWorkspaceRunID(aggregate))
	if err != nil {
		return nil, err
	}
	return aggregate, nil
}

func getProjectByID(ctx context.Context, db *sql.DB, projectID string) (*entity.Projects, error) {
	return repo.GetProjectByID(ctx, projectID)
}

func listProjectDomainTasks(ctx context.Context, db *sql.DB, projectID string) ([]entity.DomainTasks, error) {
	return repo.ListProjectDomainTasks(ctx, projectID)
}

func listProjectBrainRunBindings(ctx context.Context, db *sql.DB, projectID string, limit int) ([]entity.BrainRunBindings, error) {
	return repo.ListProjectBrainRunBindings(ctx, projectID, limit)
}

func listProjectAcceptanceRuns(ctx context.Context, db *sql.DB, projectID string, limit int) ([]entity.AcceptanceRuns, error) {
	return repo.ListProjectAcceptanceRuns(ctx, projectID, limit)
}

func listProjectAcceptanceIssues(ctx context.Context, db *sql.DB, projectID string, limit int) ([]entity.AcceptanceIssues, error) {
	return repo.ListProjectAcceptanceIssues(ctx, projectID, limit)
}

func listProjectAuditLogs(ctx context.Context, db *sql.DB, projectID string, limit int) ([]entity.AuditLogs, error) {
	return repo.ListProjectAuditLogs(ctx, projectID, limit)
}

func buildProjectSnapshot(data *projectWorkspaceAggregate) projectsv1.ProjectSnapshot {
	return projectsv1.ProjectSnapshot{
		ProjectID:            data.Project.Id,
		Name:                 data.Project.Name,
		ProjectCategory:      data.Project.ProjectCategory,
		CurrentStage:         normalizeProjectStage(data.Project.Status),
		ProgressPercent:      deriveProjectProgress(data.Project.Status, data.Project.ProductionStatus),
		RiskLevel:            deriveProjectRiskLevel(data),
		ProductionStatus:     deriveProjectProductionStatus(data),
		ManualReleaseNeed:    data.LatestAcceptanceRun != nil && data.LatestAcceptanceRun.ManualReleaseRequired == 1,
		ManualReviewRequired: projectHasManualReviewRequirement(data),
		VerificationConflict: projectHasVerificationConflict(data),
		FaultLoopDetected:    projectHasFaultLoop(data),
		PolicyDenied:         projectHasPolicyDeniedRun(data),
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

func buildWorkspaceVerificationResult(
	ctx context.Context,
	data *projectWorkspaceAggregate,
	coverage projectsv1.AcceptanceCoverage,
) acceptancev1.VerificationResultView {
	requiredChecks := []string{"acceptance_profile", "coverage_matrix", "evidence_artifacts"}
	requiredEvidence := []string(nil)
	if data.ProductionProfile != nil {
		if items := parseStringArrayJSON(data.ProductionProfile.ReleaseRequirementsJSON); len(items) > 0 {
			requiredChecks = items
		}
		requiredEvidence = parseStringArrayJSON(data.ProductionProfile.RequiredEvidenceJSON)
	}

	failedChecks := make([]string, 0, len(data.AcceptanceIssues))
	for _, issue := range data.AcceptanceIssues {
		label := firstNonEmpty(strings.TrimSpace(issue.IssueKind), strings.TrimSpace(issue.Summary), issue.Id)
		if label != "" {
			failedChecks = append(failedChecks, label)
		}
	}

	missingEvidence := make([]string, 0, len(requiredEvidence))
	if len(requiredEvidence) > coverage.EvidenceReady {
		missingEvidence = append(missingEvidence, requiredEvidence[coverage.EvidenceReady:]...)
	}

	status := "pending"
	decision := "continue_verification"
	completed := false
	summary := "Workspace is waiting for verification evidence."
	manualReviewRequired := projectHasManualReviewRequirement(data)
	if len(failedChecks) > 0 {
		status = "failed"
		decision = "repair_required"
		summary = "Blocking acceptance issues require repair."
	} else if len(missingEvidence) > 0 {
		status = "incomplete"
		decision = "collect_evidence"
		summary = "Verification evidence is still incomplete."
	} else if coverage.ProductionPassed {
		status = "passed"
		decision = "ready_for_completion"
		completed = true
		summary = "Verification requirements are currently satisfied."
	}

	// P1+P2: look up latest browser/verifier run results from RunBindings.
	browserResult := findLatestBrowserCheckResult(data.RunBindings)
	verifierResult := findLatestVerifierCheckResult(data.RunBindings)

	return acceptancev1.VerificationResultView{
		Status:           status,
		Decision:         decision,
		Completed:        completed,
		Summary:          summary,
		PreferredChannel: deriveVerificationCurrentChannel(manualReviewRequired),
		RequiredChecks:   requiredChecks,
		RequiredEvidence: requiredEvidence,
		MissingEvidence:  missingEvidence,
		FailedChecks:     failedChecks,
		VerificationContractJSON: buildVerificationContractJSON(ctx, verificationContractParams{
			ProjectCategory:          strings.TrimSpace(data.Project.ProjectCategory),
			ProfileVersion:           firstNonEmpty(profileVersionForVerification(data.AcceptanceProfile, data.ProductionProfile), strings.TrimSpace(data.Project.Id)),
			RequiredChecks:           requiredChecks,
			RequiredEvidence:         requiredEvidence,
			ManualReviewRequired:     manualReviewRequired,
			ManualReviewSummary:      summary,
			ChannelUnavailable:       !data.ChannelAvailable,
			EnvironmentUnavailable:   !data.EnvironmentAvailable,
			ChannelUnavailableReason: data.ChannelUnavailableReason,
		}),
		SourceRunID:          firstNonEmpty(latestWorkspaceRunID(data), data.Project.Id),
		UpdatedAt:            firstNonEmpty(latestWorkspaceUpdateAt(data), data.Project.UpdatedAt),
		ChannelAvailable:     data.ChannelAvailable,
		EnvironmentAvailable: data.EnvironmentAvailable,
		BrowserCheckResult:   browserResult,
		VerifierCheckResult:  verifierResult,
	}
}

func buildWorkspaceCompletionVerdict(
	ctx context.Context,
	data *projectWorkspaceAggregate,
	coverage projectsv1.AcceptanceCoverage,
) acceptancev1.CompletionVerdictView {
	if data != nil && data.PersistedVerdict != nil {
		return *data.PersistedVerdict
	}
	verification := buildWorkspaceVerificationResult(context.Background(), data, coverage)
	runtimeEscalation := buildWorkspaceRuntimeEscalation(data)
	faultSummary := buildWorkspaceFaultSummary(data)
	verificationConflict := projectHasVerificationConflict(data)
	manualReleaseRequired := data.LatestAcceptanceRun != nil && data.LatestAcceptanceRun.ManualReleaseRequired == 1
	manualReviewRequired := projectHasManualReviewRequirement(data)
	return deriveCompletionVerdictView(
		verification,
		runtimeEscalation,
		faultSummary,
		verificationConflict,
		manualReleaseRequired,
		manualReviewRequired,
		countBlockingIssues(data.AcceptanceIssues),
		firstNonEmpty(latestWorkspaceUpdateAt(data), data.Project.UpdatedAt),
		firstNonEmpty(latestWorkspaceRunID(data), data.Project.Id),
	)
}

func buildWorkspaceRuntimeEscalation(data *projectWorkspaceAggregate) acceptancev1.RuntimeEscalationView {
	if !data.ChannelAvailable {
		reason := data.ChannelUnavailableReason
		if reason == "" {
			reason = "Preferred verification channel (high_spec_remote) is unavailable."
		}
		return acceptancev1.RuntimeEscalationView{
			Status:      "escalated",
			ReasonClass: "channel_unavailable",
			Severity:    "blocking",
			Action:      "switch_verification_channel",
			Summary:     reason,
		}
	}
	if !data.EnvironmentAvailable {
		return acceptancev1.RuntimeEscalationView{
			Status:      "escalated",
			ReasonClass: "environment_unavailable",
			Severity:    "blocking",
			Action:      "restore_environment",
			Summary:     "Execution environment is unavailable. Cannot proceed with verification.",
		}
	}

	for _, binding := range data.RunBindings {
		status := normalizeBindingStatus(binding.RunStatus)
		if !bindingNeedsAttention(status) {
			continue
		}
		reasonClass := "runtime_attention"
		policyDenied := false
		switch status {
		case "run_denied":
			reasonClass = "policy_denied"
			policyDenied = true
		case "run_unsupported":
			reasonClass = "capability_gap"
		case "run_failed":
			reasonClass = "execution_failed"
		}
		return acceptancev1.RuntimeEscalationView{
			Status:       "escalated",
			ReasonClass:  reasonClass,
			SourceBrain:  strings.TrimSpace(binding.BrainKind),
			SourceTaskID: strings.TrimSpace(binding.TaskId),
			RunBindingID: strings.TrimSpace(binding.Id),
			RunStatus:    status,
			Severity:     deriveBindingSeverity(status),
			Action:       "open_task_review",
			TaskID:       strings.TrimSpace(binding.TaskId),
			RunID:        firstNonEmpty(strings.TrimSpace(binding.BrainRunId), strings.TrimSpace(binding.Id)),
			UpdatedAt:    firstNonEmpty(strings.TrimSpace(binding.LastSyncAt), strings.TrimSpace(binding.UpdatedAt)),
			Summary:      deriveBindingInboxTitle(binding, data.Tasks),
			PolicyDenied: policyDenied,
		}
	}
	return acceptancev1.RuntimeEscalationView{Status: "none"}
}

func buildWorkspaceFaultSummary(data *projectWorkspaceAggregate) acceptancev1.FaultSummaryView {
	blocking := 0
	advisory := 0
	topIssue := ""
	failedChecks := make([]string, 0, len(data.AcceptanceIssues))
	affectedTasks := make([]string, 0, len(data.Tasks))
	for _, issue := range data.AcceptanceIssues {
		label := firstNonEmpty(strings.TrimSpace(issue.Summary), strings.TrimSpace(issue.IssueKind), issue.Id)
		if strings.TrimSpace(topIssue) == "" {
			topIssue = label
		}
		if label != "" {
			failedChecks = append(failedChecks, label)
		}
		if issue.Blocking == 1 {
			blocking++
		} else {
			advisory++
		}
	}
	for _, task := range data.Tasks {
		if task.ManualReviewRequired == 1 || isBlockingTaskStatus(task.Status) {
			affectedTasks = append(affectedTasks, firstNonEmpty(strings.TrimSpace(task.Name), task.Id))
		}
		if len(affectedTasks) >= 5 {
			break
		}
	}
	status := "healthy"
	if blocking > 0 {
		status = "blocking"
	} else if advisory > 0 || projectHasFaultLoop(data) {
		status = "attention"
	}
	return acceptancev1.FaultSummaryView{
		Status:             status,
		BlockingIssueCount: blocking,
		AdvisoryIssueCount: advisory,
		TopIssue:           topIssue,
		FaultLoopDetected:  projectHasFaultLoop(data),
		FaultKind:          firstNonEmpty(topIssue, status),
		Severity:           deriveFaultSeverity(blocking, advisory),
		Summary:            deriveFaultSummaryText(status, topIssue, blocking, advisory, projectHasFaultLoop(data)),
		FailedChecks:       failedChecks,
		AffectedTasks:      affectedTasks,
		UpdatedAt:          firstNonEmpty(latestWorkspaceUpdateAt(data), data.Project.UpdatedAt),
	}
}

func buildWorkspaceRepairPlanDraft(data *projectWorkspaceAggregate) acceptancev1.RepairPlanDraftSummary {
	if data.RepairDraft == nil {
		return acceptancev1.RepairPlanDraftSummary{Status: "idle"}
	}
	updatedTasks := make([]string, 0, 4)
	for _, task := range data.Tasks {
		if task.ManualReviewRequired == 1 || isBlockingTaskStatus(task.Status) {
			updatedTasks = append(updatedTasks, firstNonEmpty(strings.TrimSpace(task.Name), task.Id))
		}
		if len(updatedTasks) >= 4 {
			break
		}
	}
	return acceptancev1.RepairPlanDraftSummary{
		ID:                   strings.TrimSpace(data.RepairDraft.ID),
		Status:               normalizePlanState(data.RepairDraft.Status, "ready"),
		ReasonClass:          "acceptance_failure",
		RepairStrategy:       "repair_plan_draft",
		ReasoningSummary:     strings.TrimSpace(data.RepairDraft.RepairReasoningSummary),
		Summary:              strings.TrimSpace(data.RepairDraft.RepairReasoningSummary),
		UpdatedAt:            strings.TrimSpace(data.RepairDraft.UpdatedAt),
		UpdatedTasks:         updatedTasks,
		ManualReviewRequired: projectHasManualReviewRequirement(data),
	}
}

func latestWorkspaceRunID(data *projectWorkspaceAggregate) string {
	if data.LatestAcceptanceRun != nil {
		return strings.TrimSpace(data.LatestAcceptanceRun.Id)
	}
	if len(data.RunBindings) > 0 {
		return firstNonEmpty(strings.TrimSpace(data.RunBindings[0].BrainRunId), strings.TrimSpace(data.RunBindings[0].Id))
	}
	return ""
}

func latestWorkspaceUpdateAt(data *projectWorkspaceAggregate) string {
	if data.LatestAcceptanceRun != nil {
		return firstNonEmpty(strings.TrimSpace(data.LatestAcceptanceRun.FinishedAt), strings.TrimSpace(data.LatestAcceptanceRun.CreatedAt))
	}
	if len(data.RunBindings) > 0 {
		return firstNonEmpty(strings.TrimSpace(data.RunBindings[0].LastSyncAt), strings.TrimSpace(data.RunBindings[0].UpdatedAt))
	}
	return ""
}

func deriveProjectProductionStatus(data *projectWorkspaceAggregate) string {
	if data.LatestAcceptanceRun != nil && strings.TrimSpace(data.LatestAcceptanceRun.ProductionStatus) != "" {
		return normalizeProductionStatus(data.LatestAcceptanceRun.ProductionStatus)
	}
	return normalizeProductionStatus(data.Project.ProductionStatus)
}

func projectHasManualReviewRequirement(data *projectWorkspaceAggregate) bool {
	if data.LatestAcceptanceRun != nil && data.LatestAcceptanceRun.ManualReleaseRequired == 1 {
		return true
	}
	for _, task := range data.Tasks {
		if task.ManualReviewRequired == 1 {
			return true
		}
	}
	return false
}

func projectHasVerificationConflict(data *projectWorkspaceAggregate) bool {
	if countBlockingIssues(data.AcceptanceIssues) == 0 {
		return false
	}
	if data.LatestAcceptanceRun == nil {
		return false
	}
	if isFunctionalPassed(data.LatestAcceptanceRun.FunctionalStatus) || isProductionReady(data.LatestAcceptanceRun.ProductionStatus) {
		return true
	}
	return data.LatestAcceptanceRun.ManualReleaseRequired == 1
}

func projectHasFaultLoop(data *projectWorkspaceAggregate) bool {
	failedCount := 0
	for _, run := range data.AcceptanceRuns {
		switch normalizeAcceptanceRunStatus(run.Status) {
		case "failed", "blocked":
			failedCount++
		}
	}
	return failedCount >= 2 && data.RepairDraft != nil
}

func projectHasPolicyDeniedRun(data *projectWorkspaceAggregate) bool {
	for _, binding := range data.RunBindings {
		if normalizeBindingStatus(binding.RunStatus) == "run_denied" {
			return true
		}
	}
	return false
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

// findLatestBrowserCheckResult scans RunBindings for the most recent browser run.
func findLatestBrowserCheckResult(bindings []entity.BrainRunBindings) *acceptancev1.BrowserCheckResultView {
	for _, b := range bindings {
		if strings.ToLower(strings.TrimSpace(b.BrainKind)) != "browser" {
			continue
		}
		status := normalizeBindingStatus(b.RunStatus)
		resultStatus := "skipped"
		switch status {
		case "run_success", "run_completed":
			resultStatus = "passed"
		case "run_failed", "run_error":
			resultStatus = "failed"
		case "run_denied", "run_unsupported":
			resultStatus = "error"
		default:
			continue
		}
		return &acceptancev1.BrowserCheckResultView{
			URL:     "",
			Status:  resultStatus,
			Summary: "Browser check run status: " + status,
		}
	}
	return nil
}

// findLatestVerifierCheckResult scans RunBindings for the most recent verifier run.
func findLatestVerifierCheckResult(bindings []entity.BrainRunBindings) *acceptancev1.VerifierCheckResultView {
	for _, b := range bindings {
		if strings.ToLower(strings.TrimSpace(b.BrainKind)) != "verifier" {
			continue
		}
		status := normalizeBindingStatus(b.RunStatus)
		resultStatus := "skipped"
		switch status {
		case "run_success", "run_completed":
			resultStatus = "passed"
		case "run_failed", "run_error":
			resultStatus = "failed"
		case "run_denied", "run_unsupported":
			resultStatus = "error"
		default:
			continue
		}
		return &acceptancev1.VerifierCheckResultView{
			Status:  resultStatus,
			Summary: "Verifier check run status: " + status,
		}
	}
	return nil
}

// buildProjectHealthMetrics computes near-term run statistics from brain_run_bindings.
func buildProjectHealthMetrics(data *projectWorkspaceAggregate) projectsv1.ProjectHealthMetrics {
	bindings := data.RunBindings
	if len(bindings) == 0 {
		return projectsv1.ProjectHealthMetrics{}
	}

	total := 0
	success := 0
	var totalLatencyMs int64
	failureModes := make(map[string]int)
	var latestUpdate string

	for _, b := range bindings {
		status := normalizeBindingStatus(b.RunStatus)
		if status == "" {
			continue
		}
		total++

		if status == "run_success" || status == "run_completed" || status == "run_succeeded" {
			success++
		} else if status == "run_failed" || status == "run_error" || status == "run_denied" || status == "run_unsupported" {
			failureModes[status]++
		}

		latency := timestampDeltaSeconds(b.StartedAt, b.FinishedAt) * 1000
		if latency > 0 {
			totalLatencyMs += latency
		}

		cand := firstNonEmpty(b.LastSyncAt, b.UpdatedAt, b.FinishedAt, b.StartedAt)
		if cand > latestUpdate {
			latestUpdate = cand
		}
	}

	topFailureMode := ""
	topFailureCount := 0
	for mode, count := range failureModes {
		if count > topFailureCount {
			topFailureCount = count
			topFailureMode = mode
		}
	}

	avgLatency := 0
	if total > 0 {
		avgLatency = int(totalLatencyMs / int64(total))
	}

	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total)
	}

	return projectsv1.ProjectHealthMetrics{
		RecentRunCount:  total,
		SuccessRate:     successRate,
		AvgLatencyMs:    avgLatency,
		TopFailureMode:  topFailureMode,
		TopFailureCount: topFailureCount,
		LastUpdated:     latestUpdate,
	}
}
