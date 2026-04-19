package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func createRepairDraftFromAcceptanceFailure(
	ctx context.Context,
	aggregate *acceptanceAggregate,
	result *braincontracts.CompletionAdjudicationResult,
) (*CreateRepairDraftResult, error) {
	if aggregate == nil || aggregate.LatestAcceptanceRun == nil {
		return nil, gerror.New("latest acceptance run is required")
	}
	if result == nil {
		return nil, gerror.New("completion adjudication result is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	domainTask, err := getDomainTaskByID(ctx, db, aggregate.LatestAcceptanceRun.TaskId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	var compiledTask *entity.WorkflowCompiledTasks
	compiledTaskID := strings.TrimSpace(aggregate.LatestAcceptanceRun.TaskId)
	if domainTask != nil {
		compiledTaskID = firstNonEmpty(domainTask.SourceCompiledTaskId, compiledTaskID)
	}
	if compiledTaskID != "" {
		compiledTask, err = getCompiledTaskByID(ctx, db, compiledTaskID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
	}

	return Plan().CreateRepairDraft(ctx, CreateRepairDraftCommand{
		ProjectID:             aggregate.Project.Id,
		FailedTaskContextJSON: mustMarshalJSONString(buildRepairFailedTaskContext(aggregate, domainTask, compiledTask), "{}"),
		FailureReasonJSON:     mustMarshalJSONString(buildRepairFailureReason(aggregate, result), "{}"),
		OriginalContractsJSON: mustMarshalJSONString(buildRepairOriginalContracts(aggregate, domainTask, compiledTask), "{}"),
		RuntimeSummaryJSON:    mustMarshalJSONString(buildRepairRuntimeSummary(aggregate, result), "{}"),
		ArtifactRefs:          buildRepairArtifactRefs(aggregate.EvidenceItems),
		CreatedBy:             "acceptance.adjudication",
	})
}

func buildRepairFailedTaskContext(
	aggregate *acceptanceAggregate,
	domainTask *entity.DomainTasks,
	compiledTask *entity.WorkflowCompiledTasks,
) map[string]any {
	context := map[string]any{
		"project_id":         aggregate.Project.Id,
		"project_name":       aggregate.Project.Name,
		"project_category":   aggregate.Project.ProjectCategory,
		"acceptance_run_id":  aggregate.LatestAcceptanceRun.Id,
		"acceptance_task_id": aggregate.LatestAcceptanceRun.TaskId,
		"current_stage":      normalizeProjectStage(aggregate.Project.Status),
	}
	if domainTask != nil {
		context["domain_task"] = map[string]any{
			"id":                      domainTask.Id,
			"name":                    domainTask.Name,
			"phase":                   domainTask.Phase,
			"task_kind":               domainTask.TaskKind,
			"role_type":               domainTask.RoleType,
			"brain_kind":              domainTask.BrainKind,
			"risk_level":              domainTask.RiskLevel,
			"status":                  domainTask.Status,
			"source_compiled_plan_id": domainTask.SourceCompiledPlanId,
			"source_compiled_task_id": domainTask.SourceCompiledTaskId,
			"source_task_key":         domainTask.SourceTaskKey,
		}
	}
	if compiledTask != nil {
		context["compiled_task"] = map[string]any{
			"id":         compiledTask.Id,
			"task_key":   compiledTask.TaskKey,
			"name":       compiledTask.Name,
			"phase":      compiledTask.Phase,
			"task_kind":  compiledTask.TaskKind,
			"role_type":  compiledTask.RoleType,
			"brain_kind": compiledTask.BrainKind,
			"risk_level": compiledTask.RiskLevel,
			"status":     compiledTask.Status,
		}
	}
	return context
}

func buildRepairFailureReason(
	aggregate *acceptanceAggregate,
	result *braincontracts.CompletionAdjudicationResult,
) map[string]any {
	return map[string]any{
		"acceptance_run_id":       aggregate.LatestAcceptanceRun.Id,
		"final_status":            result.FinalStatus,
		"decision_reason":         result.DecisionReason,
		"functional_passed":       result.FunctionalPassed,
		"production_passed":       result.ProductionPassed,
		"manual_release_required": result.ManualReleaseRequired,
		"blocking_issue_count":    countBlockingIssues(aggregate.Issues),
		"issues":                  buildRepairIssueRows(aggregate.Issues),
	}
}

func buildRepairOriginalContracts(
	aggregate *acceptanceAggregate,
	domainTask *entity.DomainTasks,
	compiledTask *entity.WorkflowCompiledTasks,
) map[string]any {
	payload := map[string]any{
		"project_context": map[string]any{
			"project_id":       aggregate.Project.Id,
			"project_name":     aggregate.Project.Name,
			"goal_summary":     aggregate.Project.GoalSummary,
			"project_category": aggregate.Project.ProjectCategory,
		},
	}
	if domainTask != nil {
		payload["domain_task"] = map[string]any{
			"id":         domainTask.Id,
			"name":       domainTask.Name,
			"phase":      domainTask.Phase,
			"task_kind":  domainTask.TaskKind,
			"role_type":  domainTask.RoleType,
			"brain_kind": domainTask.BrainKind,
			"risk_level": domainTask.RiskLevel,
		}
	}
	if compiledTask != nil {
		payload["compiled_task"] = map[string]any{
			"id":                         compiledTask.Id,
			"compiled_plan_id":           compiledTask.CompiledPlanId,
			"task_key":                   compiledTask.TaskKey,
			"name":                       compiledTask.Name,
			"delivery_contract_json":     mustUnmarshalJSONObject(compiledTask.DeliveryContractJson),
			"verification_contract_json": mustUnmarshalJSONObject(compiledTask.VerificationContractJson),
			"affected_resources":         mustUnmarshalJSONArray(compiledTask.AffectedResourcesJson),
			"depends_on_task_keys":       mustUnmarshalJSONArray(compiledTask.DependsOnTaskKeysJson),
		}
	}
	return payload
}

func buildRepairRuntimeSummary(
	aggregate *acceptanceAggregate,
	result *braincontracts.CompletionAdjudicationResult,
) map[string]any {
	return map[string]any{
		"acceptance_run": map[string]any{
			"id":                      aggregate.LatestAcceptanceRun.Id,
			"status":                  aggregate.LatestAcceptanceRun.Status,
			"functional_status":       aggregate.LatestAcceptanceRun.FunctionalStatus,
			"production_status":       aggregate.LatestAcceptanceRun.ProductionStatus,
			"manual_release_required": aggregate.LatestAcceptanceRun.ManualReleaseRequired == 1,
			"created_at":              aggregate.LatestAcceptanceRun.CreatedAt,
			"finished_at":             aggregate.LatestAcceptanceRun.FinishedAt,
		},
		"adjudication": map[string]any{
			"final_status":             result.FinalStatus,
			"decision_reason":          result.DecisionReason,
			"functional_passed":        result.FunctionalPassed,
			"production_passed":        result.ProductionPassed,
			"manual_release_required":  result.ManualReleaseRequired,
			"manual_release_completed": result.ManualReleaseCompleted,
		},
		"coverage_summary": map[string]any{
			"surface_coverage": buildCoverageSummaryRows(aggregate.SurfaceCoverage),
			"journey_coverage": buildJourneySummaryRows(aggregate.JourneyCoverage),
		},
		"judgements": buildRepairJudgementRows(aggregate.Judgements),
		"issues":     buildRepairIssueRows(aggregate.Issues),
	}
}

func buildRepairArtifactRefs(items []entity.EvidenceItems) []braincontracts.ArtifactRef {
	refs := make([]braincontracts.ArtifactRef, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.FilePath) == "" && strings.TrimSpace(item.Id) == "" {
			continue
		}
		refs = append(refs, braincontracts.ArtifactRef{
			Kind: firstNonEmpty(item.EvidenceType, "evidence"),
			ID:   item.Id,
			Path: item.FilePath,
		})
		if len(refs) >= acceptanceEvidenceLimit {
			break
		}
	}
	return refs
}

func buildRepairIssueRows(items []entity.AcceptanceIssues) []map[string]any {
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		rows = append(rows, map[string]any{
			"id":                item.Id,
			"severity":          item.Severity,
			"issue_kind":        item.IssueKind,
			"blocking":          item.Blocking == 1,
			"summary":           item.Summary,
			"detail_json":       mustUnmarshalJSONObject(item.DetailJson),
			"acceptance_run_id": item.AcceptanceRunId,
			"created_at":        item.CreatedAt,
		})
	}
	return rows
}

func buildRepairJudgementRows(items []entity.AcceptanceJudgements) []map[string]any {
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		rows = append(rows, map[string]any{
			"id":               item.Id,
			"judgement_kind":   item.JudgementKind,
			"judgement_result": item.JudgementResult,
			"summary":          item.Summary,
			"detail_json":      mustUnmarshalJSONObject(item.DetailJson),
			"created_at":       item.CreatedAt,
		})
	}
	return rows
}

func getDomainTaskByID(ctx context.Context, db *sql.DB, taskID string) (*entity.DomainTasks, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, nil
	}

	row := db.QueryRowContext(
		ctx,
		`SELECT
id, project_id, COALESCE(source_compiled_plan_id, ''), COALESCE(source_compiled_task_id, ''), COALESCE(source_task_key, ''), compiled_version, name, phase, task_kind, role_type, brain_kind, risk_level, status, manual_review_required, created_at, updated_at
FROM `+dao.DomainTasks.Table()+` WHERE id = ? LIMIT 1`,
		taskID,
	)

	var item entity.DomainTasks
	if err := row.Scan(
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
		if err == sql.ErrNoRows || isSchemaMissingError(err) {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query domain task failed")
	}
	return &item, nil
}

func getCompiledTaskByID(ctx context.Context, db *sql.DB, compiledTaskID string) (*entity.WorkflowCompiledTasks, error) {
	compiledTaskID = strings.TrimSpace(compiledTaskID)
	if compiledTaskID == "" {
		return nil, nil
	}

	row := db.QueryRowContext(
		ctx,
		`SELECT
id, compiled_plan_id, task_key, name, COALESCE(description, ''), phase, task_kind, role_type, brain_kind, risk_level, affected_resources_json, delivery_contract_json, verification_contract_json, manual_review_required, COALESCE(depends_on_task_keys_json, ''), status, created_at, updated_at
FROM `+dao.WorkflowCompiledTasks.Table()+` WHERE id = ? LIMIT 1`,
		compiledTaskID,
	)

	var item entity.WorkflowCompiledTasks
	if err := row.Scan(
		&item.Id,
		&item.CompiledPlanId,
		&item.TaskKey,
		&item.Name,
		&item.Description,
		&item.Phase,
		&item.TaskKind,
		&item.RoleType,
		&item.BrainKind,
		&item.RiskLevel,
		&item.AffectedResourcesJson,
		&item.DeliveryContractJson,
		&item.VerificationContractJson,
		&item.ManualReviewRequired,
		&item.DependsOnTaskKeysJson,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows || isSchemaMissingError(err) {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query compiled task failed")
	}
	return &item, nil
}
