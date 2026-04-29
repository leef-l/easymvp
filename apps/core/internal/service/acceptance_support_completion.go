package service

import (
	"context"
	"database/sql"
	"encoding/json"
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

func adjudicateLatestAcceptanceRun(ctx context.Context, projectID string) (*braincontracts.CompletionAdjudicationResult, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	aggregate, err := loadAcceptanceAggregate(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	if aggregate.LatestAcceptanceRun == nil {
		return nil, gerror.New("latest acceptance run is required")
	}
	return adjudicateAcceptanceAggregate(ctx, db, aggregate)
}

// collectValidationResults queries the browser/verifier brain runs associated
// with an acceptance run and returns their status as JSON payloads for the
// completion adjudication LLM context.
func collectValidationResults(ctx context.Context, run *entity.AcceptanceRuns) (browserJSON, verifierJSON json.RawMessage) {
	if run == nil {
		return nil, nil
	}

	baseURL, err := runtimeBaseURL(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "[collectValidationResults] no brain serve base url: %v", err)
		return nil, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	results := map[string]interface{}{
		"acceptance_run_id": run.Id,
		"project_id":        run.ProjectId,
		"task_id":           run.TaskId,
	}

	// Include previously saved launch results (run IDs and initial statuses).
	if run.ValidationResultsJSON != "" {
		var saved map[string]interface{}
		_ = json.Unmarshal([]byte(run.ValidationResultsJSON), &saved)
		for k, v := range saved {
			results[k] = v
		}
	}

	// Query live run status for browser.
	if run.BrowserRunID != "" {
		state, err := runtimeGetRun(ctx, client, baseURL, run.BrowserRunID)
		if err != nil {
			results["browser_error"] = err.Error()
		} else if state != nil {
			results["browser_status"] = state.Status
			results["brain"] = state.Brain
			if len(state.Result) > 0 {
				results["browser_result"] = string(state.Result)
			}
		}
	}

	// Query live run status for verifier.
	if run.VerifierRunID != "" {
		state, err := runtimeGetRun(ctx, client, baseURL, run.VerifierRunID)
		if err != nil {
			results["verifier_error"] = err.Error()
		} else if state != nil {
			results["verifier_status"] = state.Status
			results["brain"] = state.Brain
			if len(state.Result) > 0 {
				results["verifier_result"] = string(state.Result)
			}
		}
	}

	// Split into browser-specific and verifier-specific payloads.
	browserRes := map[string]interface{}{
		"acceptance_run_id": run.Id,
	}
	verifierRes := map[string]interface{}{
		"acceptance_run_id": run.Id,
	}
	for k, v := range results {
		if strings.HasPrefix(k, "browser_") {
			browserRes[k] = v
		} else if strings.HasPrefix(k, "verifier_") {
			verifierRes[k] = v
		}
	}

	browserJSON, _ = json.Marshal(browserRes)
	verifierJSON, _ = json.Marshal(verifierRes)
	return browserJSON, verifierJSON
}

func adjudicateAcceptanceRunByID(ctx context.Context, projectID string, acceptanceRunID string) (*braincontracts.CompletionAdjudicationResult, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	aggregate, err := loadAcceptanceAggregateByRunID(ctx, db, projectID, acceptanceRunID)
	if err != nil {
		return nil, err
	}
	if aggregate.LatestAcceptanceRun == nil {
		return nil, gerror.New("acceptance run is required")
	}
	return adjudicateAcceptanceAggregate(ctx, db, aggregate)
}

func adjudicateAcceptanceAggregate(ctx context.Context, db *sql.DB, aggregate *acceptanceAggregate) (*braincontracts.CompletionAdjudicationResult, error) {
	if aggregate == nil || aggregate.LatestAcceptanceRun == nil {
		return nil, gerror.New("latest acceptance run is required")
	}

	// P1+P2: collect browser/verifier validation results before adjudication.
	browserValidationJSON, verifierCheckJSON := collectValidationResults(ctx, aggregate.LatestAcceptanceRun)

	_, result, err := EasyMVPBrain().CallCompletionAdjudication(ctx, braincontracts.CompletionAdjudicationInput{
		ExecutionSummaryJSON:        mustMarshalRawJSON(buildCompletionExecutionSummary(aggregate)),
		DeliverySummaryJSON:         mustMarshalRawJSON(buildCompletionDeliverySummary(aggregate)),
		VerificationSummaryJSON:     mustMarshalRawJSON(buildCompletionVerificationSummary(aggregate)),
		AcceptanceSummaryJSON:       mustMarshalRawJSON(buildCompletionAcceptanceSummary(aggregate)),
		ManualReleaseStateJSON:      mustMarshalRawJSON(buildCompletionManualReleaseState(aggregate)),
		BrowserValidationResultJSON: browserValidationJSON,
		VerifierCheckResultJSON:     verifierCheckJSON,
	})
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, gerror.Wrap(err, "begin completion adjudication transaction failed")
	}

	now := nowText()
	verdict := buildPersistedCompletionVerdict(aggregate, result, now)
	if err = updateAcceptanceRunFromAdjudication(ctx, tx, aggregate.LatestAcceptanceRun.Id, result, now); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err = persistCompletionVerdict(ctx, tx, aggregate.Project.Id, aggregate.LatestAcceptanceRun.Id, verdict, now); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	verification := buildVerificationResultView(ctx, aggregate)
	if err = persistVerificationResult(ctx, tx, aggregate.Project.Id, aggregate.LatestAcceptanceRun.Id, verification, now); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	runtimeEsc := buildRuntimeEscalationView(aggregate)
	if err = persistRuntimeEscalation(ctx, tx, aggregate.Project.Id, aggregate.LatestAcceptanceRun.Id, runtimeEsc, now); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	faultSum := buildFaultSummaryView(aggregate)
	if err = persistFaultSummary(ctx, tx, aggregate.Project.Id, aggregate.LatestAcceptanceRun.Id, faultSum, now); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	for _, row := range buildCompletionJudgementRows(aggregate.Project.Id, aggregate.LatestAcceptanceRun.Id, result, now) {
		if err = insertAcceptanceJudgementRecord(ctx, tx, row); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	if err = updateProjectFromAdjudication(ctx, tx, aggregate.Project.Id, result, now); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if auditErr := insertAuditLogSqlTx(ctx, tx, aggregate.Project.Id, "acceptance.adjudicated", "user:local_operator", "Acceptance run adjudicated", map[string]any{"acceptance_run_id": aggregate.LatestAcceptanceRun.Id}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	if err = tx.Commit(); err != nil {
		return nil, gerror.Wrap(err, "commit completion adjudication transaction failed")
	}
	// Detect fault loop from acceptance runs (Engineering Cybernetics ch.11 nonlinearity / limit cycle)
	faultLoopDetected := acceptanceHasFaultLoop(aggregate)
	needsRepair := shouldCreateRepairDraftAfterAdjudication(result) || faultLoopDetected
	if needsRepair {
		if _, repairErr := createRepairDraftFromAcceptanceFailure(ctx, aggregate, result); repairErr != nil {
			g.Log().Warningf(ctx, "create repair draft after acceptance adjudication failed: project=%s run=%s err=%v", aggregate.Project.Id, aggregate.LatestAcceptanceRun.Id, repairErr)
		}
	}

	// Publish acceptance outcome event for workflow downstream processing.
	if result.Decision == "pass" || result.Decision == "release" {
		events.Publish(ctx, &events.WorkflowEvent{
			ProjectID: aggregate.Project.Id,
			EventType: events.AcceptancePassed,
			Payload: map[string]interface{}{
				"acceptance_run_id": aggregate.LatestAcceptanceRun.Id,
				"decision":          result.Decision,
				"manual_release":    result.ManualReleaseRequired,
			},
		})
	} else {
		events.Publish(ctx, &events.WorkflowEvent{
			ProjectID: aggregate.Project.Id,
			EventType: events.AcceptanceFailed,
			Payload: map[string]interface{}{
				"acceptance_run_id": aggregate.LatestAcceptanceRun.Id,
				"decision":          result.Decision,
				"reason":            result.DecisionReason,
				"repair_needed":     needsRepair,
			},
		})
	}

	return result, nil
}

func buildPersistedCompletionVerdict(
	data *acceptanceAggregate,
	result *braincontracts.CompletionAdjudicationResult,
	now string,
) acceptancev1.CompletionVerdictView {
	verdict := deriveCompletionVerdictView(
		buildVerificationResultView(context.Background(), data),
		buildRuntimeEscalationView(data),
		buildFaultSummaryView(data),
		acceptanceHasVerificationConflict(data),
		result.ManualReleaseRequired,
		projectHasBlockingManualReview(data.Issues),
		countBlockingIssues(data.Issues),
		now,
		firstNonEmpty(strings.TrimSpace(acceptanceRunID(data)), strings.TrimSpace(data.Project.Id)),
	)
	if summary := strings.TrimSpace(result.DecisionReason); summary != "" {
		verdict.Reason = summary
		verdict.Summary = summary
	}
	// Override with brain-provided four-layer state if available (backward-compatible)
	if result.Decision != "" {
		verdict.Decision = result.Decision
	}
	if result.ExecutorSucceeded || result.DeliveryVerified || result.AcceptancePassed || result.Completed {
		verdict.ExecutorSucceeded = result.ExecutorSucceeded
		verdict.DeliveryVerified = result.DeliveryVerified
		verdict.AcceptancePassed = result.AcceptancePassed
		verdict.Completed = result.Completed
	}
	return verdict
}

func persistCompletionVerdict(
	ctx context.Context,
	tx *sql.Tx,
	projectID string,
	acceptanceRunID string,
	verdict acceptancev1.CompletionVerdictView,
	now string,
) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO completion_verdicts (
id, project_id, acceptance_run_id, decision, final_status, reason, summary, next_action, blocker_count,
release_ready, completed, manual_review_required, manual_release_required, manual_release_completed,
executor_succeeded, delivery_verified, acceptance_passed,
source_run_id, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(acceptance_run_id) DO UPDATE SET
decision = excluded.decision,
final_status = excluded.final_status,
reason = excluded.reason,
summary = excluded.summary,
next_action = excluded.next_action,
blocker_count = excluded.blocker_count,
release_ready = excluded.release_ready,
completed = excluded.completed,
manual_review_required = excluded.manual_review_required,
manual_release_required = excluded.manual_release_required,
manual_release_completed = excluded.manual_release_completed,
executor_succeeded = excluded.executor_succeeded,
delivery_verified = excluded.delivery_verified,
acceptance_passed = excluded.acceptance_passed,
source_run_id = excluded.source_run_id,
updated_at = excluded.updated_at`,
		newResourceID("verdict"),
		projectID,
		acceptanceRunID,
		verdict.Decision,
		verdict.FinalStatus,
		verdict.Reason,
		verdict.Summary,
		verdict.NextAction,
		verdict.BlockerCount,
		boolToInt(verdict.ReleaseReady),
		boolToInt(verdict.Completed),
		boolToInt(verdict.ManualReviewRequired),
		boolToInt(verdict.ManualReleaseRequired),
		boolToInt(verdict.ManualReleaseCompleted),
		boolToInt(verdict.ExecutorSucceeded),
		boolToInt(verdict.DeliveryVerified),
		boolToInt(verdict.AcceptancePassed),
		firstNonEmpty(strings.TrimSpace(verdict.SourceRunID), acceptanceRunID, projectID),
		now,
		now,
	)
	if err != nil {
		return gerror.Wrap(err, "persist completion verdict failed")
	}
	if affected, _ := result.RowsAffected(); affected < 1 {
		return gerror.New("persist completion verdict affected unexpected rows")
	}
	return nil
}

func buildCompletionExecutionSummary(data *acceptanceAggregate) map[string]any {
	return map[string]any{
		"acceptance_run_id":       data.LatestAcceptanceRun.Id,
		"run_status":              data.LatestAcceptanceRun.Status,
		"functional_status":       data.LatestAcceptanceRun.FunctionalStatus,
		"production_status":       data.LatestAcceptanceRun.ProductionStatus,
		"manual_release_required": data.LatestAcceptanceRun.ManualReleaseRequired == 1,
	}
}

func buildCompletionDeliverySummary(data *acceptanceAggregate) map[string]any {
	summary := buildAcceptanceArtifactSummary(data.EvidenceItems)
	if data.ProductionProfile != nil {
		summary["required_evidence"] = parseStringArrayJSON(data.ProductionProfile.RequiredEvidenceJSON)
		summary["release_requirements"] = parseStringArrayJSON(data.ProductionProfile.ReleaseRequirementsJSON)
	}
	return summary
}

func buildCompletionVerificationSummary(data *acceptanceAggregate) map[string]any {
	return map[string]any{
		"surface_coverage_count": len(data.SurfaceCoverage),
		"journey_coverage_count": len(data.JourneyCoverage),
		"blocking_issue_count":   countBlockingIssues(data.Issues),
		"warning_count":          len(data.Issues) - countBlockingIssues(data.Issues),
	}
}

func buildCompletionAcceptanceSummary(data *acceptanceAggregate) map[string]any {
	return map[string]any{
		"required_surfaces":  profileRequiredSurfaces(data),
		"required_journeys":  profileRequiredJourneys(data),
		"required_evidence":  profileRequiredEvidence(data),
		"surface_coverage":   buildCoverageSummaryRows(data.SurfaceCoverage),
		"journey_coverage":   buildJourneySummaryRows(data.JourneyCoverage),
		"issue_count":        len(data.Issues),
		"judgement_count":    len(data.Judgements),
		"release_gate_state": buildAcceptanceReleaseGate(data).Status,
	}
}

func buildCompletionManualReleaseState(data *acceptanceAggregate) map[string]any {
	return map[string]any{
		"manual_release_required":  data.LatestAcceptanceRun.ManualReleaseRequired == 1,
		"manual_release_completed": hasHumanReleaseApproval(data),
	}
}

func buildCoverageSummaryRows(items []entity.AcceptanceSurfaceCoverage) []map[string]any {
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		rows = append(rows, map[string]any{
			"surface":         item.Surface,
			"coverage_status": item.CoverageStatus,
			"evidence_count":  item.EvidenceCount,
		})
	}
	return rows
}

func buildJourneySummaryRows(items []entity.AcceptanceJourneyCoverage) []map[string]any {
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		rows = append(rows, map[string]any{
			"journey":         item.Journey,
			"coverage_status": item.CoverageStatus,
			"evidence_count":  item.EvidenceCount,
		})
	}
	return rows
}

func profileRequiredSurfaces(data *acceptanceAggregate) []string {
	if data.ProductionProfile != nil {
		if items := parseStringArrayJSON(data.ProductionProfile.RequiredSurfacesJSON); len(items) > 0 {
			return items
		}
	}
	return requiredAcceptanceSurfaces(data.Project.ProjectCategory)
}

func profileRequiredJourneys(data *acceptanceAggregate) []string {
	if data.ProductionProfile != nil {
		if items := parseStringArrayJSON(data.ProductionProfile.RequiredJourneysJSON); len(items) > 0 {
			return items
		}
	}
	return requiredAcceptanceJourneys(data.Project.ProjectCategory)
}

func profileRequiredEvidence(data *acceptanceAggregate) []string {
	if data.ProductionProfile != nil {
		return parseStringArrayJSON(data.ProductionProfile.RequiredEvidenceJSON)
	}
	return nil
}

func updateAcceptanceRunFromAdjudication(ctx context.Context, tx *sql.Tx, acceptanceRunID string, result *braincontracts.CompletionAdjudicationResult, now string) error {
	if result == nil {
		return gerror.New("completion adjudication result is required")
	}
	status := deriveAcceptanceRunStatusFromAdjudication(result)
	finishedAt := ""
	if status == "completed" || status == "failed" {
		finishedAt = now
	}
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE `+dao.AcceptanceRuns.Table()+` SET status = ?, functional_status = ?, production_status = ?, manual_release_required = ?, finished_at = ? WHERE id = ?`,
		status,
		deriveFunctionalStatusFromAdjudication(result),
		deriveProductionStatusFromAdjudication(result),
		boolToInt(result.ManualReleaseRequired),
		nullIfEmpty(finishedAt),
		acceptanceRunID,
	); err != nil {
		return gerror.Wrap(err, "update acceptance run from adjudication failed")
	}
	return nil
}

// updateProjectFromAdjudication is the SOLE authorized path for a project to
// transition into "completed" or "reworking". Any code that sets project.status
// directly to "completed" without going through completion_adjudication is a bug.
// P1-09 enforcement: completion_adjudication is the mandatory gate before completed.
func updateProjectFromAdjudication(ctx context.Context, tx *sql.Tx, projectID string, result *braincontracts.CompletionAdjudicationResult, now string) error {
	status := "acceptance"
	switch {
	case result.FinalStatus == "released_by_human" || (result.FinalStatus == "production_passed" && !result.ManualReleaseRequired):
		status = "completed"
	case result.Decision == "blocked":
		status = "paused"
	case result.Decision == "manual_checkpoint":
		status = "paused"
	case result.Decision == "rework" || result.FinalStatus == "failed":
		status = "reworking"
	}
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE `+dao.Projects.Table()+` SET status = ?, production_status = ?, updated_at = ? WHERE id = ?`,
		status,
		deriveProductionStatusFromAdjudication(result),
		now,
		projectID,
	); err != nil {
		return gerror.Wrap(err, "update project from adjudication failed")
	}
	return nil
}

func buildCompletionJudgementRows(projectID string, acceptanceRunID string, result *braincontracts.CompletionAdjudicationResult, now string) []entity.AcceptanceJudgements {
	rows := []entity.AcceptanceJudgements{
		{
			Id:              newResourceID("judge"),
			ProjectId:       projectID,
			AcceptanceRunId: acceptanceRunID,
			JudgementKind:   "functional_gate",
			JudgementResult: deriveFunctionalStatusFromAdjudication(result),
			Summary:         "Functional gate adjudicated",
			DetailJson:      mustMarshalJSONString(map[string]any{"functional_passed": result.FunctionalPassed}, "{}"),
			CreatedAt:       now,
		},
		{
			Id:              newResourceID("judge"),
			ProjectId:       projectID,
			AcceptanceRunId: acceptanceRunID,
			JudgementKind:   "production_gate",
			JudgementResult: deriveProductionStatusFromAdjudication(result),
			Summary:         "Production gate adjudicated",
			DetailJson:      mustMarshalJSONString(map[string]any{"production_passed": result.ProductionPassed}, "{}"),
			CreatedAt:       now,
		},
		{
			Id:              newResourceID("judge"),
			ProjectId:       projectID,
			AcceptanceRunId: acceptanceRunID,
			JudgementKind:   "final_decision",
			JudgementResult: result.FinalStatus,
			Summary:         firstNonEmpty(result.DecisionReason, "Completion decision generated"),
			DetailJson: mustMarshalJSONString(map[string]any{
				"functional_passed":        result.FunctionalPassed,
				"production_passed":        result.ProductionPassed,
				"manual_release_required":  result.ManualReleaseRequired,
				"manual_release_completed": result.ManualReleaseCompleted,
				"final_status":             result.FinalStatus,
			}, "{}"),
			CreatedAt: now,
		},
	}
	if result.ManualReleaseRequired {
		rows = append(rows, entity.AcceptanceJudgements{
			Id:              newResourceID("judge"),
			ProjectId:       projectID,
			AcceptanceRunId: acceptanceRunID,
			JudgementKind:   "release_gate",
			JudgementResult: deriveReleaseJudgementResult(result),
			Summary:         "Manual release gate evaluated",
			DetailJson: mustMarshalJSONString(map[string]any{
				"manual_release_required":  result.ManualReleaseRequired,
				"manual_release_completed": result.ManualReleaseCompleted,
			}, "{}"),
			CreatedAt: now,
		})
	}
	return rows
}

func insertAcceptanceJudgementRecord(ctx context.Context, tx *sql.Tx, row entity.AcceptanceJudgements) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.AcceptanceJudgements.Table()+` (
id, project_id, acceptance_run_id, judgement_kind, judgement_result, summary, detail_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.AcceptanceRunId,
		row.JudgementKind,
		row.JudgementResult,
		row.Summary,
		nullIfEmpty(row.DetailJson),
		row.CreatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert acceptance judgement failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert acceptance judgement affected unexpected rows")
	}
	return nil
}

func deriveFunctionalStatusFromAdjudication(result *braincontracts.CompletionAdjudicationResult) string {
	if result.ProductionPassed || result.FunctionalPassed {
		return "functional_passed"
	}
	return "failed"
}

func deriveProductionStatusFromAdjudication(result *braincontracts.CompletionAdjudicationResult) string {
	switch result.FinalStatus {
	case "released_by_human", "production_passed":
		return "production_passed"
	case "functional_passed":
		return "functional_passed"
	default:
		if result.ProductionPassed {
			return "production_passed"
		}
		if result.FunctionalPassed {
			return "functional_passed"
		}
		return "failed"
	}
}

func deriveAcceptanceRunStatusFromAdjudication(result *braincontracts.CompletionAdjudicationResult) string {
	switch {
	case result.FinalStatus == "failed":
		return "failed"
	case result.Decision == "blocked":
		return "blocked"
	case result.Decision == "manual_checkpoint":
		return "manual_checkpoint"
	case result.ManualReleaseRequired && !result.ManualReleaseCompleted:
		return "awaiting_manual_release"
	case result.FinalStatus == "released_by_human" || result.FinalStatus == "production_passed":
		return "completed"
	default:
		// P0-01: conservative default — unknown status must NOT auto-complete.
		return "acceptance"
	}
}

func deriveReleaseJudgementResult(result *braincontracts.CompletionAdjudicationResult) string {
	if result.ManualReleaseCompleted || result.FinalStatus == "released_by_human" {
		return "approved"
	}
	return "awaiting_manual_release"
}

// shouldCreateRepairDraftAfterAdjudication enforces the hard rule:
// fault → easymvp-brain.repair_design → reworking is the ONLY allowed path
// when a task fails or needs rework. Direct retry without repair design is
// explicitly prohibited.
func shouldCreateRepairDraftAfterAdjudication(result *braincontracts.CompletionAdjudicationResult) bool {
	if result == nil {
		return false
	}
	// P1-08: any failure or rework decision MUST trigger repair design.
	return result.FinalStatus == "failed" || result.Decision == "rework"
}

func persistVerificationResult(
	ctx context.Context,
	tx *sql.Tx,
	projectID string,
	acceptanceRunID string,
	view acceptancev1.VerificationResultView,
	now string,
) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO verification_results (
	id, project_id, acceptance_run_id, status, decision, completed, summary, preferred_channel,
	required_checks_json, required_evidence_json, missing_evidence_json, failed_checks_json,
	verification_contract_json, source_run_id, channel_available, environment_available,
	created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(acceptance_run_id) DO UPDATE SET
	status = excluded.status,
	decision = excluded.decision,
	completed = excluded.completed,
	summary = excluded.summary,
	preferred_channel = excluded.preferred_channel,
	required_checks_json = excluded.required_checks_json,
	required_evidence_json = excluded.required_evidence_json,
	missing_evidence_json = excluded.missing_evidence_json,
	failed_checks_json = excluded.failed_checks_json,
	verification_contract_json = excluded.verification_contract_json,
	source_run_id = excluded.source_run_id,
	channel_available = excluded.channel_available,
	environment_available = excluded.environment_available,
	updated_at = excluded.updated_at`,
		newResourceID("verif"),
		projectID,
		acceptanceRunID,
		view.Status,
		view.Decision,
		boolToInt(view.Completed),
		view.Summary,
		view.PreferredChannel,
		mustMarshalJSONString(view.RequiredChecks, "[]"),
		mustMarshalJSONString(view.RequiredEvidence, "[]"),
		mustMarshalJSONString(view.MissingEvidence, "[]"),
		mustMarshalJSONString(view.FailedChecks, "[]"),
		view.VerificationContractJSON,
		view.SourceRunID,
		boolToInt(view.ChannelAvailable),
		boolToInt(view.EnvironmentAvailable),
		now,
		now,
	)
	if err != nil {
		return gerror.Wrap(err, "persist verification result failed")
	}
	if affected, _ := result.RowsAffected(); affected < 1 {
		return gerror.New("persist verification result affected unexpected rows")
	}
	return nil
}

func persistRuntimeEscalation(
	ctx context.Context,
	tx *sql.Tx,
	projectID string,
	acceptanceRunID string,
	view acceptancev1.RuntimeEscalationView,
	now string,
) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO runtime_escalations (
	id, project_id, acceptance_run_id, status, reason_class, source_brain, source_task_id,
	run_binding_id, run_status, severity, action, task_id, run_id, summary, policy_denied,
	evidence_refs_json, resolved_at, resolution_status, resolver_kind, linked_fault_id,
	created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(acceptance_run_id) DO UPDATE SET
	status = excluded.status,
	reason_class = excluded.reason_class,
	source_brain = excluded.source_brain,
	source_task_id = excluded.source_task_id,
	run_binding_id = excluded.run_binding_id,
	run_status = excluded.run_status,
	severity = excluded.severity,
	action = excluded.action,
	task_id = excluded.task_id,
	run_id = excluded.run_id,
	summary = excluded.summary,
	policy_denied = excluded.policy_denied,
	evidence_refs_json = excluded.evidence_refs_json,
	resolved_at = excluded.resolved_at,
	resolution_status = excluded.resolution_status,
	resolver_kind = excluded.resolver_kind,
	linked_fault_id = excluded.linked_fault_id,
	updated_at = excluded.updated_at`,
		newResourceID("escal"),
		projectID,
		acceptanceRunID,
		view.Status,
		view.ReasonClass,
		view.SourceBrain,
		view.SourceTaskID,
		view.RunBindingID,
		view.RunStatus,
		view.Severity,
		view.Action,
		view.TaskID,
		view.RunID,
		view.Summary,
		boolToInt(view.PolicyDenied),
		view.EvidenceRefsJSON,
		view.ResolvedAt,
		view.ResolutionStatus,
		view.ResolverKind,
		view.LinkedFaultID,
		now,
		now,
	)
	if err != nil {
		return gerror.Wrap(err, "persist runtime escalation failed")
	}
	if affected, _ := result.RowsAffected(); affected < 1 {
		return gerror.New("persist runtime escalation affected unexpected rows")
	}
	return nil
}

func persistFaultSummary(
	ctx context.Context,
	tx *sql.Tx,
	projectID string,
	acceptanceRunID string,
	view acceptancev1.FaultSummaryView,
	now string,
) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO fault_summaries (
	id, project_id, acceptance_run_id, status, blocking_issue_count, advisory_issue_count,
	top_issue, fault_loop_detected, fault_kind, severity, summary,
	failed_checks_json, affected_tasks_json,
	created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(acceptance_run_id) DO UPDATE SET
	status = excluded.status,
	blocking_issue_count = excluded.blocking_issue_count,
	advisory_issue_count = excluded.advisory_issue_count,
	top_issue = excluded.top_issue,
	fault_loop_detected = excluded.fault_loop_detected,
	fault_kind = excluded.fault_kind,
	severity = excluded.severity,
	summary = excluded.summary,
	failed_checks_json = excluded.failed_checks_json,
	affected_tasks_json = excluded.affected_tasks_json,
	updated_at = excluded.updated_at`,
		newResourceID("fault"),
		projectID,
		acceptanceRunID,
		view.Status,
		view.BlockingIssueCount,
		view.AdvisoryIssueCount,
		view.TopIssue,
		boolToInt(view.FaultLoopDetected),
		view.FaultKind,
		view.Severity,
		view.Summary,
		mustMarshalJSONString(view.FailedChecks, "[]"),
		mustMarshalJSONString(view.AffectedTasks, "[]"),
		now,
		now,
	)
	if err != nil {
		return gerror.Wrap(err, "persist fault summary failed")
	}
	if affected, _ := result.RowsAffected(); affected < 1 {
		return gerror.New("persist fault summary affected unexpected rows")
	}
	return nil
}
