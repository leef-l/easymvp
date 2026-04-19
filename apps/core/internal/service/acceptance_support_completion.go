package service

import (
	"context"
	"database/sql"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
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

	_, result, err := EasyMVPBrain().CallCompletionAdjudication(ctx, braincontracts.CompletionAdjudicationInput{
		ExecutionSummaryJSON:    mustMarshalRawJSON(buildCompletionExecutionSummary(aggregate)),
		DeliverySummaryJSON:     mustMarshalRawJSON(buildCompletionDeliverySummary(aggregate)),
		VerificationSummaryJSON: mustMarshalRawJSON(buildCompletionVerificationSummary(aggregate)),
		AcceptanceSummaryJSON:   mustMarshalRawJSON(buildCompletionAcceptanceSummary(aggregate)),
		ManualReleaseStateJSON:  mustMarshalRawJSON(buildCompletionManualReleaseState(aggregate)),
	})
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, gerror.Wrap(err, "begin completion adjudication transaction failed")
	}

	now := nowText()
	if err = updateAcceptanceRunFromAdjudication(ctx, tx, aggregate.LatestAcceptanceRun.Id, result, now); err != nil {
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
	if err = tx.Commit(); err != nil {
		return nil, gerror.Wrap(err, "commit completion adjudication transaction failed")
	}
	if shouldCreateRepairDraftAfterAdjudication(result) {
		if _, repairErr := createRepairDraftFromAcceptanceFailure(ctx, aggregate, result); repairErr != nil {
			g.Log().Warningf(ctx, "create repair draft after acceptance adjudication failed: project=%s run=%s err=%v", aggregate.Project.Id, aggregate.LatestAcceptanceRun.Id, repairErr)
		}
	}
	return result, nil
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

func updateProjectFromAdjudication(ctx context.Context, tx *sql.Tx, projectID string, result *braincontracts.CompletionAdjudicationResult, now string) error {
	status := "acceptance"
	if result.FinalStatus == "released_by_human" || (result.FinalStatus == "production_passed" && !result.ManualReleaseRequired) {
		status = "completed"
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
	case result.ManualReleaseRequired && !result.ManualReleaseCompleted:
		return "awaiting_manual_release"
	case result.FinalStatus == "released_by_human" || result.FinalStatus == "production_passed":
		return "completed"
	default:
		return "completed"
	}
}

func deriveReleaseJudgementResult(result *braincontracts.CompletionAdjudicationResult) string {
	if result.ManualReleaseCompleted || result.FinalStatus == "released_by_human" {
		return "approved"
	}
	return "awaiting_manual_release"
}

func shouldCreateRepairDraftAfterAdjudication(result *braincontracts.CompletionAdjudicationResult) bool {
	if result == nil {
		return false
	}
	return result.FinalStatus == "failed"
}
