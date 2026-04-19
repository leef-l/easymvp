package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	acceptancev1 "github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

const (
	acceptanceIssuesLimit    = 20
	acceptanceEvidenceLimit  = 12
	acceptanceRunsLoadLimit  = 6
	acceptanceAuditLogsLimit = 20
)

type acceptanceViewData struct {
	AcceptanceRun  acceptancev1.AcceptanceRunView
	CoverageMatrix []acceptancev1.CoverageItem
	Issues         []acceptancev1.AcceptanceIssue
	EvidenceCards  []acceptancev1.EvidenceCard
	ReleaseGate    acceptancev1.ReleaseGateView
}

type acceptanceAggregate struct {
	Project             entity.Projects
	AcceptanceProfile   *acceptanceProfileRecord
	ProductionProfile   *productionAcceptanceProfileRecord
	AcceptanceRuns      []entity.AcceptanceRuns
	LatestAcceptanceRun *entity.AcceptanceRuns
	SurfaceCoverage     []entity.AcceptanceSurfaceCoverage
	JourneyCoverage     []entity.AcceptanceJourneyCoverage
	Issues              []entity.AcceptanceIssues
	EvidenceItems       []entity.EvidenceItems
	AuditLogs           []entity.AuditLogs
	Judgements          []entity.AcceptanceJudgements
}

func loadAcceptanceViewData(ctx context.Context, projectID string) (*acceptanceViewData, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	aggregate, err := loadAcceptanceAggregate(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	return &acceptanceViewData{
		AcceptanceRun:  buildAcceptanceRunView(aggregate),
		CoverageMatrix: buildAcceptanceCoverageMatrix(aggregate),
		Issues:         buildAcceptanceIssues(aggregate),
		EvidenceCards:  buildAcceptanceEvidenceCards(aggregate),
		ReleaseGate:    buildAcceptanceReleaseGate(aggregate),
	}, nil
}

func loadAcceptanceAggregate(ctx context.Context, db *sql.DB, projectID string) (*acceptanceAggregate, error) {
	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	runs, err := listProjectAcceptanceRuns(ctx, db, projectID, acceptanceRunsLoadLimit)
	if err != nil {
		return nil, err
	}
	auditLogs, err := listProjectAuditLogs(ctx, db, projectID, acceptanceAuditLogsLimit)
	if err != nil {
		return nil, err
	}
	evidenceItems, err := listProjectEvidenceItems(ctx, db, projectID, acceptanceEvidenceLimit)
	if err != nil {
		return nil, err
	}

	aggregate := &acceptanceAggregate{
		Project:        *project,
		AcceptanceRuns: runs,
		EvidenceItems:  evidenceItems,
		AuditLogs:      auditLogs,
	}
	aggregate.AcceptanceProfile, err = getLatestAcceptanceProfile(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	aggregate.ProductionProfile, err = getLatestProductionAcceptanceProfile(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	if len(runs) > 0 {
		aggregate.LatestAcceptanceRun = &runs[0]
		aggregate.SurfaceCoverage, err = listAcceptanceSurfaceCoverageByRun(ctx, db, runs[0].Id)
		if err != nil {
			return nil, err
		}
		aggregate.JourneyCoverage, err = listAcceptanceJourneyCoverageByRun(ctx, db, runs[0].Id)
		if err != nil {
			return nil, err
		}
		aggregate.Issues, err = listAcceptanceIssuesByRun(ctx, db, runs[0].Id, acceptanceIssuesLimit)
		if err != nil {
			return nil, err
		}
		aggregate.Judgements, err = listAcceptanceJudgementsByRun(ctx, db, runs[0].Id)
		if err != nil {
			return nil, err
		}
	}
	if aggregate.LatestAcceptanceRun == nil {
		aggregate.Issues, err = listProjectAcceptanceIssues(ctx, db, projectID, acceptanceIssuesLimit)
		if err != nil {
			return nil, err
		}
	}

	return aggregate, nil
}

func loadAcceptanceAggregateByRunID(ctx context.Context, db *sql.DB, projectID string, acceptanceRunID string) (*acceptanceAggregate, error) {
	aggregate, err := loadAcceptanceAggregate(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	acceptanceRunID = strings.TrimSpace(acceptanceRunID)
	if acceptanceRunID == "" {
		return nil, gerror.New("acceptance run id is required")
	}
	if aggregate.LatestAcceptanceRun != nil && aggregate.LatestAcceptanceRun.Id == acceptanceRunID {
		return aggregate, nil
	}

	run, err := getAcceptanceRunByID(ctx, db, acceptanceRunID)
	if err != nil {
		return nil, err
	}
	if run == nil || strings.TrimSpace(run.ProjectId) != strings.TrimSpace(projectID) {
		return nil, gerror.New("acceptance run does not belong to project")
	}

	aggregate.LatestAcceptanceRun = run
	aggregate.SurfaceCoverage, err = listAcceptanceSurfaceCoverageByRun(ctx, db, run.Id)
	if err != nil {
		return nil, err
	}
	aggregate.JourneyCoverage, err = listAcceptanceJourneyCoverageByRun(ctx, db, run.Id)
	if err != nil {
		return nil, err
	}
	aggregate.Issues, err = listAcceptanceIssuesByRun(ctx, db, run.Id, acceptanceIssuesLimit)
	if err != nil {
		return nil, err
	}
	aggregate.Judgements, err = listAcceptanceJudgementsByRun(ctx, db, run.Id)
	if err != nil {
		return nil, err
	}
	return aggregate, nil
}

func getAcceptanceRunByID(ctx context.Context, db *sql.DB, acceptanceRunID string) (*entity.AcceptanceRuns, error) {
	row := db.QueryRowContext(ctx, `
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
FROM `+dao.AcceptanceRuns.Table()+`
WHERE id = ?
LIMIT 1`, acceptanceRunID)

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
		return nil, gerror.Wrap(err, "query acceptance run by id failed")
	}
	return &run, nil
}

func listAcceptanceIssuesByRun(ctx context.Context, db *sql.DB, acceptanceRunID string, limit int) ([]entity.AcceptanceIssues, error) {
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
WHERE acceptance_run_id = ?
ORDER BY created_at DESC
LIMIT ?`

	rows, err := db.QueryContext(ctx, query, acceptanceRunID, limit)
	if err != nil {
		return nil, gerror.Wrap(err, "query acceptance issues by run failed")
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
			return nil, gerror.Wrap(err, "scan acceptance issue by run failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate acceptance issues by run failed")
	}
	return items, nil
}

func listAcceptanceSurfaceCoverageByRun(ctx context.Context, db *sql.DB, acceptanceRunID string) ([]entity.AcceptanceSurfaceCoverage, error) {
	query := `
SELECT
  id,
  project_id,
  acceptance_run_id,
  surface,
  coverage_status,
  evidence_count,
  created_at,
  updated_at
FROM ` + dao.AcceptanceSurfaceCoverage.Table() + `
WHERE acceptance_run_id = ?
ORDER BY surface ASC`

	rows, err := db.QueryContext(ctx, query, acceptanceRunID)
	if err != nil {
		return nil, gerror.Wrap(err, "query acceptance surface coverage failed")
	}
	defer rows.Close()

	items := make([]entity.AcceptanceSurfaceCoverage, 0)
	for rows.Next() {
		var item entity.AcceptanceSurfaceCoverage
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.AcceptanceRunId,
			&item.Surface,
			&item.CoverageStatus,
			&item.EvidenceCount,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan acceptance surface coverage failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate acceptance surface coverage failed")
	}
	return items, nil
}

func listAcceptanceJourneyCoverageByRun(ctx context.Context, db *sql.DB, acceptanceRunID string) ([]entity.AcceptanceJourneyCoverage, error) {
	query := `
SELECT
  id,
  project_id,
  acceptance_run_id,
  journey,
  coverage_status,
  evidence_count,
  created_at,
  updated_at
FROM ` + dao.AcceptanceJourneyCoverage.Table() + `
WHERE acceptance_run_id = ?
ORDER BY journey ASC`

	rows, err := db.QueryContext(ctx, query, acceptanceRunID)
	if err != nil {
		return nil, gerror.Wrap(err, "query acceptance journey coverage failed")
	}
	defer rows.Close()

	items := make([]entity.AcceptanceJourneyCoverage, 0)
	for rows.Next() {
		var item entity.AcceptanceJourneyCoverage
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.AcceptanceRunId,
			&item.Journey,
			&item.CoverageStatus,
			&item.EvidenceCount,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan acceptance journey coverage failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate acceptance journey coverage failed")
	}
	return items, nil
}

func listAcceptanceJudgementsByRun(ctx context.Context, db *sql.DB, acceptanceRunID string) ([]entity.AcceptanceJudgements, error) {
	query := `
SELECT
  id,
  project_id,
  acceptance_run_id,
  judgement_kind,
  judgement_result,
  summary,
  COALESCE(detail_json, ''),
  created_at
FROM ` + dao.AcceptanceJudgements.Table() + `
WHERE acceptance_run_id = ?
ORDER BY created_at DESC`

	rows, err := db.QueryContext(ctx, query, acceptanceRunID)
	if err != nil {
		return nil, gerror.Wrap(err, "query acceptance judgements failed")
	}
	defer rows.Close()

	items := make([]entity.AcceptanceJudgements, 0)
	for rows.Next() {
		var item entity.AcceptanceJudgements
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.AcceptanceRunId,
			&item.JudgementKind,
			&item.JudgementResult,
			&item.Summary,
			&item.DetailJson,
			&item.CreatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan acceptance judgement failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate acceptance judgements failed")
	}
	return items, nil
}

func listProjectEvidenceItems(ctx context.Context, db *sql.DB, projectID string, limit int) ([]entity.EvidenceItems, error) {
	query := `
SELECT
  id,
  project_id,
  COALESCE(run_id, ''),
  COALESCE(surface, ''),
  COALESCE(journey, ''),
  evidence_type,
  file_path,
  content_hash,
  file_size,
  COALESCE(captured_at, ''),
  created_at
FROM ` + dao.EvidenceItems.Table() + `
WHERE project_id = ?
ORDER BY COALESCE(captured_at, created_at) DESC, created_at DESC
LIMIT ?`

	rows, err := db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, gerror.Wrap(err, "query project evidence items failed")
	}
	defer rows.Close()

	items := make([]entity.EvidenceItems, 0, limit)
	for rows.Next() {
		var item entity.EvidenceItems
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.RunId,
			&item.Surface,
			&item.Journey,
			&item.EvidenceType,
			&item.FilePath,
			&item.ContentHash,
			&item.FileSize,
			&item.CapturedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan project evidence item failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate project evidence items failed")
	}
	return items, nil
}

func buildAcceptanceRunView(data *acceptanceAggregate) acceptancev1.AcceptanceRunView {
	if data.LatestAcceptanceRun == nil {
		return acceptancev1.AcceptanceRunView{
			ID:                     "",
			Status:                 deriveAcceptanceRunStatusFromProject(data.Project.Status),
			FunctionalStatus:       "pending",
			ProductionStatus:       normalizeProductionStatus(data.Project.ProductionStatus),
			ManualReleaseRequired:  false,
			LatestJudgementSummary: "",
		}
	}

	view := acceptancev1.AcceptanceRunView{
		ID:                    data.LatestAcceptanceRun.Id,
		TaskID:                strings.TrimSpace(data.LatestAcceptanceRun.TaskId),
		ProfileVersion:        strings.TrimSpace(data.LatestAcceptanceRun.ProfileVersion),
		Status:                normalizeAcceptanceRunStatus(data.LatestAcceptanceRun.Status),
		FunctionalStatus:      normalizeFunctionalStatus(data.LatestAcceptanceRun.FunctionalStatus),
		ProductionStatus:      normalizeProductionStatus(data.LatestAcceptanceRun.ProductionStatus),
		ManualReleaseRequired: data.LatestAcceptanceRun.ManualReleaseRequired == 1,
		FinishedAt:            strings.TrimSpace(data.LatestAcceptanceRun.FinishedAt),
	}
	if len(data.Judgements) > 0 {
		view.LatestJudgementKind = strings.TrimSpace(data.Judgements[0].JudgementKind)
		view.LatestJudgementResult = strings.TrimSpace(data.Judgements[0].JudgementResult)
		view.LatestJudgementSummary = strings.TrimSpace(data.Judgements[0].Summary)
		view.LatestJudgementAt = strings.TrimSpace(data.Judgements[0].CreatedAt)
	}
	return view
}

func buildAcceptanceCoverageMatrix(data *acceptanceAggregate) []acceptancev1.CoverageItem {
	requiredSurfaces := requiredAcceptanceSurfaces(data.Project.ProjectCategory)
	requiredJourneys := requiredAcceptanceJourneys(data.Project.ProjectCategory)
	if data.ProductionProfile != nil {
		if items := parseStringArrayJSON(data.ProductionProfile.RequiredSurfacesJSON); len(items) > 0 {
			requiredSurfaces = items
		}
		if items := parseStringArrayJSON(data.ProductionProfile.RequiredJourneysJSON); len(items) > 0 {
			requiredJourneys = items
		}
	}
	evidenceBySurface, evidenceByJourney := groupEvidenceCounts(data.EvidenceItems)

	items := make([]acceptancev1.CoverageItem, 0, len(requiredSurfaces)+len(requiredJourneys))
	surfaceMap := make(map[string]entity.AcceptanceSurfaceCoverage, len(data.SurfaceCoverage))
	for _, item := range data.SurfaceCoverage {
		surfaceMap[strings.ToLower(strings.TrimSpace(item.Surface))] = item
	}
	journeyMap := make(map[string]entity.AcceptanceJourneyCoverage, len(data.JourneyCoverage))
	for _, item := range data.JourneyCoverage {
		journeyMap[strings.ToLower(strings.TrimSpace(item.Journey))] = item
	}

	for _, key := range requiredSurfaces {
		lowerKey := strings.ToLower(key)
		if item, ok := surfaceMap[lowerKey]; ok {
			items = append(items, acceptancev1.CoverageItem{
				Key:            key,
				Kind:           "surface",
				Name:           humanizeAcceptanceKey(key),
				CoverageStatus: normalizeCoverageStatus(item.CoverageStatus),
				EvidenceCount:  maxInt(item.EvidenceCount, evidenceBySurface[lowerKey]),
			})
			continue
		}
		items = append(items, acceptancev1.CoverageItem{
			Key:            key,
			Kind:           "surface",
			Name:           humanizeAcceptanceKey(key),
			CoverageStatus: deriveCoverageStatusFromEvidence(evidenceBySurface[lowerKey], data),
			EvidenceCount:  evidenceBySurface[lowerKey],
		})
	}

	for _, key := range requiredJourneys {
		lowerKey := strings.ToLower(key)
		if item, ok := journeyMap[lowerKey]; ok {
			items = append(items, acceptancev1.CoverageItem{
				Key:            key,
				Kind:           "journey",
				Name:           humanizeAcceptanceKey(key),
				CoverageStatus: normalizeCoverageStatus(item.CoverageStatus),
				EvidenceCount:  maxInt(item.EvidenceCount, evidenceByJourney[lowerKey]),
			})
			continue
		}
		items = append(items, acceptancev1.CoverageItem{
			Key:            key,
			Kind:           "journey",
			Name:           humanizeAcceptanceKey(key),
			CoverageStatus: deriveCoverageStatusFromEvidence(evidenceByJourney[lowerKey], data),
			EvidenceCount:  evidenceByJourney[lowerKey],
		})
	}

	return items
}

func buildAcceptanceIssues(data *acceptanceAggregate) []acceptancev1.AcceptanceIssue {
	items := make([]acceptancev1.AcceptanceIssue, 0, len(data.Issues))
	for _, issue := range data.Issues {
		items = append(items, acceptancev1.AcceptanceIssue{
			ID:       issue.Id,
			Severity: normalizeSeverity(issue.Severity),
			Blocking: issue.Blocking == 1,
			Summary:  firstNonEmpty(issue.Summary, fmt.Sprintf("Acceptance issue: %s", issue.IssueKind)),
		})
	}
	return items
}

func buildAcceptanceEvidenceCards(data *acceptanceAggregate) []acceptancev1.EvidenceCard {
	items := make([]acceptancev1.EvidenceCard, 0, len(data.EvidenceItems))
	for _, item := range data.EvidenceItems {
		items = append(items, acceptancev1.EvidenceCard{
			ID:           item.Id,
			Surface:      item.Surface,
			Journey:      item.Journey,
			EvidenceType: item.EvidenceType,
			FilePath:     item.FilePath,
			CapturedAt:   firstNonEmpty(item.CapturedAt, item.CreatedAt),
		})
	}
	return items
}

func buildAcceptanceReleaseGate(data *acceptanceAggregate) acceptancev1.ReleaseGateView {
	blockingCount := countBlockingIssues(data.Issues)
	releasedByHuman := hasHumanReleaseApproval(data)

	switch {
	case data.LatestAcceptanceRun == nil:
		return acceptancev1.ReleaseGateView{
			Status:     "pending_acceptance",
			NextAction: "start_acceptance_run",
			Summary:    "Acceptance has not started yet.",
		}
	case blockingCount > 0:
		return acceptancev1.ReleaseGateView{
			Status:     "blocked",
			NextAction: "resolve_blocking_issues",
			Summary:    fmt.Sprintf("%d blocking acceptance issue(s) must be resolved before release.", blockingCount),
		}
	case isProductionReady(data.LatestAcceptanceRun.ProductionStatus) && data.LatestAcceptanceRun.ManualReleaseRequired == 1 && releasedByHuman:
		return acceptancev1.ReleaseGateView{
			Status:     "released",
			NextAction: "open_release_record",
			Summary:    "Production acceptance passed and manual release has been approved.",
		}
	case isProductionReady(data.LatestAcceptanceRun.ProductionStatus) && data.LatestAcceptanceRun.ManualReleaseRequired == 1:
		return acceptancev1.ReleaseGateView{
			Status:     "awaiting_manual_release",
			NextAction: "apply_manual_release",
			Summary:    "Production acceptance passed, waiting for manual release approval.",
		}
	case isProductionReady(data.LatestAcceptanceRun.ProductionStatus):
		return acceptancev1.ReleaseGateView{
			Status:     "ready",
			NextAction: "complete_project",
			Summary:    "Production acceptance passed and the project can be delivered.",
		}
	case isFunctionalPassed(data.LatestAcceptanceRun.FunctionalStatus):
		return acceptancev1.ReleaseGateView{
			Status:     "functional_passed",
			NextAction: "continue_production_acceptance",
			Summary:    "Functional acceptance passed, but production readiness is still incomplete.",
		}
	default:
		return acceptancev1.ReleaseGateView{
			Status:     normalizeAcceptanceRunStatus(data.LatestAcceptanceRun.Status),
			NextAction: "watch_acceptance_progress",
			Summary:    "Acceptance is still gathering coverage and evidence.",
		}
	}
}

func requiredAcceptanceSurfaces(category string) []string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "game":
		return []string{"game_runtime", "admin_backend", "api_backend", "build_output", "ops_console"}
	case "video_editing":
		return []string{"editor_runtime", "admin_backend", "api_backend", "export_output"}
	case "web_app", "web":
		return []string{"user_frontend", "admin_backend", "api_backend", "build_output"}
	default:
		return []string{"user_frontend", "admin_backend", "api_backend", "build_output"}
	}
}

func requiredAcceptanceJourneys(category string) []string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "game":
		return []string{"launch_game", "core_game_loop", "pause_resume", "finish_settlement", "restart_game", "admin_sync", "build_verification"}
	case "video_editing":
		return []string{"import_media", "timeline_edit", "preview_playback", "export_video", "admin_sync", "build_verification"}
	case "web_app", "web":
		return []string{"user_sign_in", "primary_user_flow", "admin_operate", "state_write_read", "api_verification", "build_verification"}
	default:
		return []string{"user_sign_in", "primary_user_flow", "admin_operate", "state_write_read", "api_verification", "build_verification"}
	}
}

func groupEvidenceCounts(items []entity.EvidenceItems) (map[string]int, map[string]int) {
	bySurface := make(map[string]int)
	byJourney := make(map[string]int)
	for _, item := range items {
		if key := strings.ToLower(strings.TrimSpace(item.Surface)); key != "" {
			bySurface[key]++
		}
		if key := strings.ToLower(strings.TrimSpace(item.Journey)); key != "" {
			byJourney[key]++
		}
	}
	return bySurface, byJourney
}

func deriveCoverageStatusFromEvidence(evidenceCount int, data *acceptanceAggregate) string {
	switch {
	case evidenceCount <= 0:
		return "missing"
	case data.LatestAcceptanceRun != nil && isProductionReady(data.LatestAcceptanceRun.ProductionStatus):
		return "pass"
	default:
		return "partial"
	}
}

func normalizeCoverageStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "pass", "passed", "ready", "covered":
		return "pass"
	case "partial", "in_progress":
		return "partial"
	default:
		return "missing"
	}
}

func normalizeAcceptanceRunStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "pending"
	}
	return strings.ReplaceAll(status, " ", "_")
}

func normalizeFunctionalStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "passed", "pass", "functional_passed":
		return "passed"
	case "failed", "error":
		return "failed"
	case "partial":
		return "partial"
	default:
		return "pending"
	}
}

func deriveAcceptanceRunStatusFromProject(projectStatus string) string {
	switch strings.ToLower(strings.TrimSpace(projectStatus)) {
	case "acceptance":
		return "running"
	case "completed":
		return "completed"
	default:
		return "pending"
	}
}

func humanizeAcceptanceKey(key string) string {
	parts := strings.Split(strings.TrimSpace(key), "_")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}

func hasHumanReleaseApproval(data *acceptanceAggregate) bool {
	for _, item := range data.Judgements {
		kind := strings.ToLower(strings.TrimSpace(item.JudgementKind))
		result := strings.ToLower(strings.TrimSpace(item.JudgementResult))
		if strings.Contains(kind, "release") && (result == "approved" || result == "released" || result == "passed") {
			return true
		}
	}
	for _, item := range data.AuditLogs {
		eventType := strings.ToLower(strings.TrimSpace(item.EventType))
		summary := strings.ToLower(strings.TrimSpace(item.Summary))
		if strings.Contains(eventType, "release") && (strings.Contains(summary, "approve") || strings.Contains(summary, "released")) {
			return true
		}
	}
	return false
}
