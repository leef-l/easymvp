package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	acceptancev1 "github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/events"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

type acceptanceProfileRecord struct {
	ID                   string
	ProjectID            string
	ProjectCategory      string
	ProfileVersion       string
	RequiredSurfacesJSON string
	RequiredJourneysJSON string
	RequiredEvidenceJSON string
	Status               string
	CreatedAt            string
	UpdatedAt            string
}

type productionAcceptanceProfileRecord struct {
	ID                      string
	ProjectID               string
	AcceptanceProfileID     string
	ProfileVersion          string
	RequiredSurfacesJSON    string
	RequiredJourneysJSON    string
	RequiredEvidenceJSON    string
	ReleaseRequirementsJSON string
	Status                  string
	CreatedAt               string
	UpdatedAt               string
}

type startAcceptanceResolved struct {
	Profile    *entity.ProjectProfiles
	Acceptance *acceptanceProfileRecord
	Production *productionAcceptanceProfileRecord
}

func mapAcceptanceProfiles(ctx context.Context, projectID string) (*braincontracts.AcceptanceMappingResult, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	profile, err := getProjectProfileByProjectID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	evidenceItems, err := listProjectEvidenceItems(ctx, projectID, 200)
	if err != nil {
		return nil, err
	}
	runs, err := listProjectAcceptanceRuns(ctx, db, projectID, 1)
	if err != nil {
		return nil, err
	}

	var (
		surfaceCoverage []entity.AcceptanceSurfaceCoverage
		journeyCoverage []entity.AcceptanceJourneyCoverage
	)
	if len(runs) > 0 {
		surfaceCoverage, err = listAcceptanceSurfaceCoverageByRun(ctx, runs[0].Id)
		if err != nil {
			return nil, err
		}
		journeyCoverage, err = listAcceptanceJourneyCoverageByRun(ctx, runs[0].Id)
		if err != nil {
			return nil, err
		}
	}

	_, result, err := EasyMVPBrain().CallAcceptanceMapping(ctx, braincontracts.AcceptanceMappingInput{
		ProjectCategory:     project.ProjectCategory,
		CategoryProfileJSON: mustMarshalRawJSON(buildCategoryProfilePayload(project, profile)),
		ArtifactSummaryJSON: mustMarshalRawJSON(buildAcceptanceArtifactSummary(evidenceItems)),
		CoverageSummaryJSON: mustMarshalRawJSON(buildAcceptanceCoverageSummary(runs, surfaceCoverage, journeyCoverage)),
	})
	if err != nil {
		return nil, err
	}

	now := nowText()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, gerror.Wrap(err, "begin acceptance profile transaction failed")
	}

	if err = insertAcceptanceProfileRecord(ctx, tx, acceptanceProfileRecord{
		ID:                   result.AcceptanceProfileID,
		ProjectID:            projectID,
		ProjectCategory:      project.ProjectCategory,
		ProfileVersion:       profile.AcceptanceProfileVersion,
		RequiredSurfacesJSON: mustMarshalJSONString(result.RequiredSurfaces, "[]"),
		RequiredJourneysJSON: mustMarshalJSONString(result.RequiredJourneys, "[]"),
		RequiredEvidenceJSON: mustMarshalJSONString(result.RequiredEvidence, "[]"),
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
	}); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err = insertProductionAcceptanceProfileRecord(ctx, tx, productionAcceptanceProfileRecord{
		ID:                      result.ProductionAcceptanceProfileID,
		ProjectID:               projectID,
		AcceptanceProfileID:     result.AcceptanceProfileID,
		ProfileVersion:          profile.AcceptanceProfileVersion,
		RequiredSurfacesJSON:    mustMarshalJSONString(result.RequiredSurfaces, "[]"),
		RequiredJourneysJSON:    mustMarshalJSONString(result.RequiredJourneys, "[]"),
		RequiredEvidenceJSON:    mustMarshalJSONString(result.RequiredEvidence, "[]"),
		ReleaseRequirementsJSON: mustMarshalJSONString(result.ReleaseRequirements, "[]"),
		Status:                  "active",
		CreatedAt:               now,
		UpdatedAt:               now,
	}); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, gerror.Wrap(err, "commit acceptance profile transaction failed")
	}
	return result, nil
}

func startAcceptanceRun(ctx context.Context, req StartAcceptanceCommand) (string, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return "", err
	}
	defer closeFn()

	project, err := getProjectByID(ctx, db, req.ProjectID)
	if err != nil {
		return "", err
	}
	if err = validateProjectStatusTransition(project.Status, []string{"executing", "running", "acceptance", "completed"}); err != nil {
		return "", err
	}
	resolved, err := resolveAcceptanceProfilesForRun(ctx, db, req)
	if err != nil {
		return "", err
	}

	// P1-04: channel_unavailable automatic downgrade.
	channel := newHighSpecVerificationChannelFromConfig(ctx)
	channelAvailable := channel.Available(ctx)
	fallbackChannel := "github_actions"
	if strings.ToLower(strings.TrimSpace(req.Mode)) == "manual" {
		fallbackChannel = "manual_review"
	}

	now := nowText()
	runID := newResourceID("acceptance_run")
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", gerror.Wrap(err, "begin start acceptance transaction failed")
	}

	// If preferred channel is unavailable, record a runtime escalation and proceed with fallback.
	if !channelAvailable {
		escalation := acceptancev1.RuntimeEscalationView{
			Status:           "escalated",
			ReasonClass:      "channel_unavailable",
			Severity:         "blocking",
			Action:           "switch_verification_channel",
			Summary:          "Preferred verification channel (high_spec_remote) is unavailable. Switched to fallback channel: " + fallbackChannel + ".",
			ResolverKind:     "system",
			ResolutionStatus: "downgraded",
		}
		if err = insertRuntimeEscalationRecord(ctx, tx, entity.RuntimeEscalations{
			Id:               newResourceID("escal"),
			ProjectId:        project.Id,
			AcceptanceRunId:  runID,
			Status:           escalation.Status,
			ReasonClass:      escalation.ReasonClass,
			Severity:         escalation.Severity,
			Action:           escalation.Action,
			Summary:          escalation.Summary,
			ResolverKind:     escalation.ResolverKind,
			ResolutionStatus: escalation.ResolutionStatus,
			CreatedAt:        now,
			UpdatedAt:        now,
		}); err != nil {
			_ = tx.Rollback()
			return "", gerror.Wrap(err, "persist channel_unavailable escalation failed")
		}
	}

	if err = insertAcceptanceRunRecord(ctx, tx, entity.AcceptanceRuns{
		Id:                    runID,
		ProjectId:             project.Id,
		TaskId:                req.TaskID,
		ProfileVersion:        resolved.Profile.AcceptanceProfileVersion,
		Status:                normalizeAcceptanceRunMode(req.Mode),
		FunctionalStatus:      "pending",
		ProductionStatus:      "pending",
		ManualReleaseRequired: 0,
		CreatedAt:             now,
	}); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	for _, surface := range parseStringArrayJSON(resolved.Production.RequiredSurfacesJSON) {
		if err = insertAcceptanceSurfaceCoverageRecord(ctx, tx, entity.AcceptanceSurfaceCoverage{
			Id:              newResourceID("acc_surface"),
			ProjectId:       project.Id,
			AcceptanceRunId: runID,
			Surface:         surface,
			CoverageStatus:  "missing",
			EvidenceCount:   0,
			CreatedAt:       now,
			UpdatedAt:       now,
		}); err != nil {
			_ = tx.Rollback()
			return "", err
		}
	}
	for _, journey := range parseStringArrayJSON(resolved.Production.RequiredJourneysJSON) {
		if err = insertAcceptanceJourneyCoverageRecord(ctx, tx, entity.AcceptanceJourneyCoverage{
			Id:              newResourceID("acc_journey"),
			ProjectId:       project.Id,
			AcceptanceRunId: runID,
			Journey:         journey,
			CoverageStatus:  "missing",
			EvidenceCount:   0,
			CreatedAt:       now,
			UpdatedAt:       now,
		}); err != nil {
			_ = tx.Rollback()
			return "", err
		}
	}
	if err = updateProjectStatusForAcceptanceRun(ctx, tx, project.Id); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if auditErr := insertAuditLogSqlTx(ctx, tx, project.Id, "acceptance.run.started", "user:local_operator", "Acceptance run started", map[string]any{"acceptance_run_id": runID}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	if err = tx.Commit(); err != nil {
		return "", gerror.Wrap(err, "commit start acceptance transaction failed")
	}

	// P1+P2: launch browser/verifier runs in parallel using ExecuteMultiBrain,
	// then save their run IDs back to the acceptance run record.
	go launchValidationBrainsAndSaveRunIDs(ctx, project.Id, req.TaskID, runID)

	return runID, nil
}

func launchValidationBrainsAndSaveRunIDs(ctx context.Context, projectID, taskID, acceptanceRunID string) {
	browserURL := strings.TrimSpace(g.Cfg().MustGet(ctx, "easymvp.acceptance.browserValidationURL", "").String())
	enabled := g.Cfg().MustGet(ctx, "easymvp.acceptance.verifierEnabled", false).Bool()

	var brains []string
	var prompts []string
	if browserURL != "" {
		brains = append(brains, "browser")
		prompts = append(prompts, fmt.Sprintf("Check URL %s for anomalies. Report DOM issues, HTTP errors, missing elements, and form problems.", browserURL))
	}
	if enabled {
		brains = append(brains, "verifier")
		prompts = append(prompts, "Run project verification checks: unit tests, output assertions, and file validations. Report pass/fail for each check.")
	}
	if len(brains) == 0 {
		return
	}

	// Use a composite prompt when multiple brains are launched.
	prompt := strings.Join(prompts, "; ")
	baseURL, err := runtimeBaseURL(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "resolve brain serve base url for multi-brain validation failed: %v", err)
		return
	}

	multiReq := &MultiBrainRunRequest{
		ProjectID: projectID,
		TaskID:    taskID,
		Prompt:    prompt,
		MaxTurns:  6,
		Brains:    brains,
	}
	multiRes := ExecuteMultiBrain(ctx, &http.Client{Timeout: 15 * time.Second}, baseURL, multiReq)

	// Save run IDs back to acceptance_runs.
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "open db for saving validation run ids failed: %v", err)
		return
	}
	defer closeFn()

	browserRunID := ""
	verifierRunID := ""
	validationResults := map[string]interface{}{}
	for kind, res := range multiRes.Results {
		if res != nil && res.RunID != "" {
			validationResults[kind+"_run_id"] = res.RunID
			validationResults[kind+"_status"] = res.Status
			if kind == "browser" {
				browserRunID = res.RunID
			}
			if kind == "verifier" {
				verifierRunID = res.RunID
			}
		}
	}
	for kind, err := range multiRes.Errors {
		if err != nil {
			validationResults[kind+"_error"] = err.Error()
		}
	}
	validationJSON, _ := json.Marshal(validationResults)

	_, dbErr := db.ExecContext(ctx,
		`UPDATE acceptance_runs SET browser_run_id = ?, verifier_run_id = ?, validation_results_json = ? WHERE id = ?`,
		nullIfEmpty(browserRunID), nullIfEmpty(verifierRunID), string(validationJSON), acceptanceRunID)
	if dbErr != nil {
		g.Log().Errorf(ctx, "save validation run ids failed: %v", dbErr)
	}

	// Publish validation completion events for workflow downstream processing.
	if browserRunID != "" {
		events.Publish(ctx, &events.WorkflowEvent{
			ProjectID: projectID,
			EventType: events.BrowserCheckCompleted,
			Payload: map[string]interface{}{
				"acceptance_run_id": acceptanceRunID,
				"browser_run_id":    browserRunID,
				"status":            "completed",
			},
		})
	}
	if verifierRunID != "" {
		events.Publish(ctx, &events.WorkflowEvent{
			ProjectID: projectID,
			EventType: events.VerifierCheckCompleted,
			Payload: map[string]interface{}{
				"acceptance_run_id": acceptanceRunID,
				"verifier_run_id":   verifierRunID,
				"status":            "completed",
			},
		})
	}
}

func getLatestAcceptanceProfile(ctx context.Context, db *sql.DB, projectID string) (*acceptanceProfileRecord, error) {
	query := `
SELECT
  id,
  project_id,
  project_category,
  profile_version,
  required_surfaces_json,
  required_journeys_json,
  required_evidence_json,
  status,
  created_at,
  updated_at
FROM acceptance_profiles
WHERE project_id = ?
ORDER BY updated_at DESC, created_at DESC
LIMIT 1`

	row := db.QueryRowContext(ctx, query, projectID)
	var item acceptanceProfileRecord
	if err := row.Scan(
		&item.ID,
		&item.ProjectID,
		&item.ProjectCategory,
		&item.ProfileVersion,
		&item.RequiredSurfacesJSON,
		&item.RequiredJourneysJSON,
		&item.RequiredEvidenceJSON,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows || isSchemaMissingError(err) {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query latest acceptance profile failed")
	}
	return &item, nil
}

func getLatestProductionAcceptanceProfile(ctx context.Context, db *sql.DB, projectID string) (*productionAcceptanceProfileRecord, error) {
	query := `
SELECT
  id,
  project_id,
  acceptance_profile_id,
  profile_version,
  required_surfaces_json,
  required_journeys_json,
  required_evidence_json,
  release_requirements_json,
  status,
  created_at,
  updated_at
FROM production_acceptance_profiles
WHERE project_id = ?
ORDER BY updated_at DESC, created_at DESC
LIMIT 1`

	row := db.QueryRowContext(ctx, query, projectID)
	var item productionAcceptanceProfileRecord
	if err := row.Scan(
		&item.ID,
		&item.ProjectID,
		&item.AcceptanceProfileID,
		&item.ProfileVersion,
		&item.RequiredSurfacesJSON,
		&item.RequiredJourneysJSON,
		&item.RequiredEvidenceJSON,
		&item.ReleaseRequirementsJSON,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows || isSchemaMissingError(err) {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query latest production acceptance profile failed")
	}
	return &item, nil
}

func getLatestAcceptanceRunForTask(ctx context.Context, db *sql.DB, projectID string, taskID string) (*entity.AcceptanceRuns, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, nil
	}

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
WHERE project_id = ? AND task_id = ?
ORDER BY created_at DESC
LIMIT 1`, projectID, taskID)

	var item entity.AcceptanceRuns
	if err := row.Scan(
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
		if err == sql.ErrNoRows || isSchemaMissingError(err) {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query latest acceptance run for task failed")
	}
	return &item, nil
}

func resolveAcceptanceProfilesForRun(ctx context.Context, db *sql.DB, req StartAcceptanceCommand) (*startAcceptanceResolved, error) {
	profile, err := getProjectProfileByProjectID(ctx, db, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if err = applyRequestedAcceptanceProfileVersion(ctx, db, profile, req.ProfileVersion); err != nil {
		return nil, err
	}

	acceptanceProfile, productionProfile, err := ensureAcceptanceProfilesCurrent(ctx, db, req.ProjectID, profile.AcceptanceProfileVersion)
	if err != nil {
		return nil, err
	}
	if acceptanceProfile == nil || productionProfile == nil {
		return nil, gerror.New("acceptance profiles are required")
	}
	return &startAcceptanceResolved{
		Profile:    profile,
		Acceptance: acceptanceProfile,
		Production: productionProfile,
	}, nil
}

func ensureAcceptanceProfilesCurrent(
	ctx context.Context,
	db *sql.DB,
	projectID string,
	profileVersion string,
) (*acceptanceProfileRecord, *productionAcceptanceProfileRecord, error) {
	acceptanceProfile, err := getLatestAcceptanceProfile(ctx, db, projectID)
	if err != nil {
		return nil, nil, err
	}
	productionProfile, err := getLatestProductionAcceptanceProfile(ctx, db, projectID)
	if err != nil {
		return nil, nil, err
	}
	if acceptanceProfilesNeedRefresh(profileVersion, acceptanceProfile, productionProfile) {
		if _, err = mapAcceptanceProfiles(ctx, projectID); err != nil {
			return nil, nil, err
		}
		acceptanceProfile, err = getLatestAcceptanceProfile(ctx, db, projectID)
		if err != nil {
			return nil, nil, err
		}
		productionProfile, err = getLatestProductionAcceptanceProfile(ctx, db, projectID)
		if err != nil {
			return nil, nil, err
		}
	}
	return acceptanceProfile, productionProfile, nil
}

func acceptanceProfilesNeedRefresh(
	profileVersion string,
	acceptanceProfile *acceptanceProfileRecord,
	productionProfile *productionAcceptanceProfileRecord,
) bool {
	profileVersion = strings.TrimSpace(profileVersion)
	if acceptanceProfile == nil || productionProfile == nil {
		return true
	}
	if profileVersion == "" {
		return false
	}
	return acceptanceProfile.ProfileVersion != profileVersion || productionProfile.ProfileVersion != profileVersion
}

func applyRequestedAcceptanceProfileVersion(ctx context.Context, db *sql.DB, profile *entity.ProjectProfiles, requestedVersion string) error {
	if profile == nil {
		return gerror.New("project profile is required")
	}
	requestedVersion = strings.TrimSpace(requestedVersion)
	if requestedVersion == "" || requestedVersion == profile.AcceptanceProfileVersion {
		return nil
	}
	if _, err := db.ExecContext(
		ctx,
		`UPDATE `+dao.ProjectProfiles.Table()+` SET acceptance_profile_version = ?, updated_at = ? WHERE id = ?`,
		requestedVersion,
		nowText(),
		profile.Id,
	); err != nil {
		return gerror.Wrap(err, "update project acceptance profile version failed")
	}
	profile.AcceptanceProfileVersion = requestedVersion
	return nil
}

func insertAcceptanceProfileRecord(ctx context.Context, tx *sql.Tx, row acceptanceProfileRecord) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO acceptance_profiles (
id, project_id, project_category, profile_version, required_surfaces_json, required_journeys_json, required_evidence_json, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID,
		row.ProjectID,
		row.ProjectCategory,
		row.ProfileVersion,
		row.RequiredSurfacesJSON,
		row.RequiredJourneysJSON,
		row.RequiredEvidenceJSON,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert acceptance profile failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert acceptance profile affected unexpected rows")
	}
	return nil
}

func insertProductionAcceptanceProfileRecord(ctx context.Context, tx *sql.Tx, row productionAcceptanceProfileRecord) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO production_acceptance_profiles (
id, project_id, acceptance_profile_id, profile_version, required_surfaces_json, required_journeys_json, required_evidence_json, release_requirements_json, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID,
		row.ProjectID,
		row.AcceptanceProfileID,
		row.ProfileVersion,
		row.RequiredSurfacesJSON,
		row.RequiredJourneysJSON,
		row.RequiredEvidenceJSON,
		row.ReleaseRequirementsJSON,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert production acceptance profile failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert production acceptance profile affected unexpected rows")
	}
	return nil
}

func insertAcceptanceRunRecord(ctx context.Context, tx *sql.Tx, row entity.AcceptanceRuns) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.AcceptanceRuns.Table()+` (
id, project_id, task_id, profile_version, status, functional_status, production_status, manual_release_required, created_at, finished_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		nullIfEmpty(row.TaskId),
		row.ProfileVersion,
		row.Status,
		row.FunctionalStatus,
		row.ProductionStatus,
		row.ManualReleaseRequired,
		nullIfEmpty(row.BrowserRunID),
		nullIfEmpty(row.VerifierRunID),
		nullIfEmpty(row.ValidationResultsJSON),
		row.CreatedAt,
		nullIfEmpty(row.FinishedAt),
	)
	if err != nil {
		return gerror.Wrap(err, "insert acceptance run failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert acceptance run affected unexpected rows")
	}
	return nil
}

func insertRuntimeEscalationRecord(ctx context.Context, tx *sql.Tx, row entity.RuntimeEscalations) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO runtime_escalations (
	id, project_id, acceptance_run_id, status, reason_class, source_brain, source_task_id,
	run_binding_id, run_status, severity, action, task_id, run_id, summary, policy_denied,
	evidence_refs_json, resolved_at, resolution_status, resolver_kind, linked_fault_id,
	created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.AcceptanceRunId,
		row.Status,
		row.ReasonClass,
		nullIfEmpty(row.SourceBrain),
		nullIfEmpty(row.SourceTaskID),
		nullIfEmpty(row.RunBindingID),
		nullIfEmpty(row.RunStatus),
		nullIfEmpty(row.Severity),
		nullIfEmpty(row.Action),
		nullIfEmpty(row.TaskID),
		nullIfEmpty(row.RunID),
		nullIfEmpty(row.Summary),
		row.PolicyDenied,
		nullIfEmpty(row.EvidenceRefsJSON),
		nullIfEmpty(row.ResolvedAt),
		nullIfEmpty(row.ResolutionStatus),
		nullIfEmpty(row.ResolverKind),
		nullIfEmpty(row.LinkedFaultID),
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert runtime escalation failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert runtime escalation affected unexpected rows")
	}
	return nil
}

func insertAcceptanceSurfaceCoverageRecord(ctx context.Context, tx *sql.Tx, row entity.AcceptanceSurfaceCoverage) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.AcceptanceSurfaceCoverage.Table()+` (
id, project_id, acceptance_run_id, surface, coverage_status, evidence_count, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.AcceptanceRunId,
		row.Surface,
		row.CoverageStatus,
		row.EvidenceCount,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert acceptance surface coverage failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert acceptance surface coverage affected unexpected rows")
	}
	return nil
}

func insertAcceptanceJourneyCoverageRecord(ctx context.Context, tx *sql.Tx, row entity.AcceptanceJourneyCoverage) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.AcceptanceJourneyCoverage.Table()+` (
id, project_id, acceptance_run_id, journey, coverage_status, evidence_count, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.AcceptanceRunId,
		row.Journey,
		row.CoverageStatus,
		row.EvidenceCount,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert acceptance journey coverage failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert acceptance journey coverage affected unexpected rows")
	}
	return nil
}

func updateProjectStatusForAcceptanceRun(ctx context.Context, tx *sql.Tx, projectID string) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE `+dao.Projects.Table()+` SET status = ?, updated_at = ? WHERE id = ?`,
		"acceptance",
		nowText(),
		projectID,
	); err != nil {
		return gerror.Wrap(err, "update project status for acceptance run failed")
	}
	return nil
}

func buildAcceptanceArtifactSummary(items []entity.EvidenceItems) map[string]any {
	byType := make(map[string]int)
	bySurface := make(map[string]int)
	byJourney := make(map[string]int)
	for _, item := range items {
		if key := strings.TrimSpace(item.EvidenceType); key != "" {
			byType[key]++
		}
		if key := strings.TrimSpace(item.Surface); key != "" {
			bySurface[key]++
		}
		if key := strings.TrimSpace(item.Journey); key != "" {
			byJourney[key]++
		}
	}
	return map[string]any{
		"evidence_count": len(items),
		"by_type":        byType,
		"by_surface":     bySurface,
		"by_journey":     byJourney,
	}
}

func buildAcceptanceCoverageSummary(
	runs []entity.AcceptanceRuns,
	surfaces []entity.AcceptanceSurfaceCoverage,
	journeys []entity.AcceptanceJourneyCoverage,
) map[string]any {
	surfaceItems := make([]map[string]any, 0, len(surfaces))
	for _, item := range surfaces {
		surfaceItems = append(surfaceItems, map[string]any{
			"surface":         item.Surface,
			"coverage_status": item.CoverageStatus,
			"evidence_count":  item.EvidenceCount,
		})
	}
	journeyItems := make([]map[string]any, 0, len(journeys))
	for _, item := range journeys {
		journeyItems = append(journeyItems, map[string]any{
			"journey":         item.Journey,
			"coverage_status": item.CoverageStatus,
			"evidence_count":  item.EvidenceCount,
		})
	}

	summary := map[string]any{
		"surface_coverage": surfaceItems,
		"journey_coverage": journeyItems,
	}
	if len(runs) > 0 {
		summary["latest_acceptance_run"] = map[string]any{
			"id":                      runs[0].Id,
			"status":                  runs[0].Status,
			"functional_status":       runs[0].FunctionalStatus,
			"production_status":       runs[0].ProductionStatus,
			"manual_release_required": runs[0].ManualReleaseRequired == 1,
		}
	}
	return summary
}

func parseStringArrayJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func normalizeAcceptanceRunMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		return "queued"
	}
	if isAllowedValue(mode, "queued", "running", "manual", "ad_hoc", "functional", "production") {
		return mode
	}
	return "queued"
}

func isTerminalAcceptanceRunStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "failed", "awaiting_manual_release", "released":
		return true
	default:
		return false
	}
}
