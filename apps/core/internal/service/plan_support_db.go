package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	planv1 "github.com/leef-l/easymvp/apps/core/api/plan/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

const planTaskProjectionLimit = 64

type planViewData struct {
	Overview       planv1.PlanOverview
	Draft          planv1.PlanDraftView
	Review         planv1.PlanReviewView
	Compiled       planv1.CompiledPlanView
	RepairDraft    planv1.RepairDraftView
	TaskProjection []planv1.CompiledTaskView
	DiffSummary    planv1.DiffSummary
}

type planAggregate struct {
	Project       entity.Projects
	Draft         *entity.WorkflowPlanDrafts
	Review        *entity.WorkflowPlanReviewResults
	Compiled      *entity.WorkflowCompiledPlans
	RepairDraft   *repairPlanDraftRecord
	CompiledTasks []entity.WorkflowCompiledTasks
	DomainTasks   []entity.DomainTasks
}

type repairPlanDraftRecord struct {
	ID                      string
	ProjectID               string
	FailedTaskContextJSON   string
	FailureReasonJSON       string
	OriginalContractsJSON   string
	RuntimeSummaryJSON      string
	RepairPlanJSON          string
	RepairReasoningSummary  string
	ReplacedConstraintsJSON string
	Status                  string
	CreatedBy               string
	CreatedAt               string
	UpdatedAt               string
}

func loadPlanViewData(ctx context.Context, projectID string) (*planViewData, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	aggregate, err := loadPlanAggregate(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	return &planViewData{
		Overview:       buildPlanOverview(aggregate),
		Draft:          buildPlanDraftView(aggregate),
		Review:         buildPlanReviewView(aggregate),
		Compiled:       buildCompiledPlanView(aggregate),
		RepairDraft:    buildRepairDraftView(aggregate),
		TaskProjection: buildPlanTaskProjection(aggregate),
		DiffSummary:    buildPlanDiffSummary(aggregate),
	}, nil
}

func loadPlanAggregate(ctx context.Context, db *sql.DB, projectID string) (*planAggregate, error) {
	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	aggregate := &planAggregate{
		Project: *project,
	}

	aggregate.Draft, err = getPlanDraftForProject(ctx, db, *project)
	if err != nil {
		return nil, err
	}
	aggregate.Review, err = getPlanReviewForProject(ctx, db, *project, aggregate.Draft)
	if err != nil {
		return nil, err
	}
	aggregate.Compiled, err = getCompiledPlanForProject(ctx, db, *project, aggregate.Draft, aggregate.Review)
	if err != nil {
		return nil, err
	}
	aggregate.RepairDraft, err = getLatestRepairDraftForProject(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	aggregate.CompiledTasks, err = listCompiledTasksForProject(ctx, db, *project, aggregate.Compiled)
	if err != nil {
		return nil, err
	}
	aggregate.DomainTasks, err = listDomainTasksForPlan(ctx, db, *project, aggregate.Compiled)
	if err != nil {
		return nil, err
	}

	return aggregate, nil
}

func getPlanDraftForProject(ctx context.Context, db *sql.DB, project entity.Projects) (*entity.WorkflowPlanDrafts, error) {
	var (
		query string
		args  []any
	)

	if project.CurrentPlanDraftId != "" {
		query = `
SELECT
  id,
  project_id,
  version,
  source_kind,
  COALESCE(source_run_id, ''),
  project_category,
  goal_summary,
  input_requirements_json,
  draft_tasks_json,
  status,
  created_by,
  created_at,
  updated_at
FROM ` + dao.WorkflowPlanDrafts.Table() + `
WHERE id = ?
LIMIT 1`
		args = []any{project.CurrentPlanDraftId}
	} else {
		query = `
SELECT
  id,
  project_id,
  version,
  source_kind,
  COALESCE(source_run_id, ''),
  project_category,
  goal_summary,
  input_requirements_json,
  draft_tasks_json,
  status,
  created_by,
  created_at,
  updated_at
FROM ` + dao.WorkflowPlanDrafts.Table() + `
WHERE project_id = ?
ORDER BY version DESC
LIMIT 1`
		args = []any{project.Id}
	}

	row := db.QueryRowContext(ctx, query, args...)
	item, err := scanPlanDraftRow(row)
	if err == nil {
		return item, nil
	}
	if isSchemaMissingError(err) || err == sql.ErrNoRows {
		return nil, nil
	}
	return nil, err
}

func getPlanReviewForProject(ctx context.Context, db *sql.DB, project entity.Projects, draft *entity.WorkflowPlanDrafts) (*entity.WorkflowPlanReviewResults, error) {
	if draft == nil && project.CurrentPlanDraftId == "" {
		return nil, nil
	}

	var (
		query string
		args  []any
	)

	if draft != nil {
		query = `
SELECT
  id,
  project_id,
  plan_draft_id,
  review_version,
  COALESCE(review_run_id, ''),
  decision,
  blocking_issue_count,
  advisory_issue_count,
  issues_json,
  COALESCE(split_suggestions_json, ''),
  COALESCE(override_suggestions_json, ''),
  status,
  reviewed_at
FROM ` + dao.WorkflowPlanReviewResults.Table() + `
WHERE plan_draft_id = ?
ORDER BY review_version DESC
LIMIT 1`
		args = []any{draft.Id}
	} else {
		query = `
SELECT
  id,
  project_id,
  plan_draft_id,
  review_version,
  COALESCE(review_run_id, ''),
  decision,
  blocking_issue_count,
  advisory_issue_count,
  issues_json,
  COALESCE(split_suggestions_json, ''),
  COALESCE(override_suggestions_json, ''),
  status,
  reviewed_at
FROM ` + dao.WorkflowPlanReviewResults.Table() + `
WHERE project_id = ?
ORDER BY reviewed_at DESC, review_version DESC
LIMIT 1`
		args = []any{project.Id}
	}

	row := db.QueryRowContext(ctx, query, args...)
	item, err := scanPlanReviewRow(row)
	if err == nil {
		return item, nil
	}
	if isSchemaMissingError(err) || err == sql.ErrNoRows {
		return nil, nil
	}
	return nil, err
}

func getCompiledPlanForProject(
	ctx context.Context,
	db *sql.DB,
	project entity.Projects,
	draft *entity.WorkflowPlanDrafts,
	review *entity.WorkflowPlanReviewResults,
) (*entity.WorkflowCompiledPlans, error) {
	var (
		query string
		args  []any
	)

	switch {
	case project.CurrentCompiledPlanId != "":
		query = `
SELECT
  id,
  project_id,
  plan_draft_id,
  plan_review_result_id,
  compiled_version,
  COALESCE(compile_run_id, ''),
  project_category,
  status,
  COALESCE(risk_summary_json, ''),
  COALESCE(compile_diff_json, ''),
  generated_at,
  COALESCE(activated_at, '')
FROM ` + dao.WorkflowCompiledPlans.Table() + `
WHERE id = ?
LIMIT 1`
		args = []any{project.CurrentCompiledPlanId}
	case review != nil:
		query = `
SELECT
  id,
  project_id,
  plan_draft_id,
  plan_review_result_id,
  compiled_version,
  COALESCE(compile_run_id, ''),
  project_category,
  status,
  COALESCE(risk_summary_json, ''),
  COALESCE(compile_diff_json, ''),
  generated_at,
  COALESCE(activated_at, '')
FROM ` + dao.WorkflowCompiledPlans.Table() + `
WHERE plan_review_result_id = ?
ORDER BY compiled_version DESC
LIMIT 1`
		args = []any{review.Id}
	case draft != nil:
		query = `
SELECT
  id,
  project_id,
  plan_draft_id,
  plan_review_result_id,
  compiled_version,
  COALESCE(compile_run_id, ''),
  project_category,
  status,
  COALESCE(risk_summary_json, ''),
  COALESCE(compile_diff_json, ''),
  generated_at,
  COALESCE(activated_at, '')
FROM ` + dao.WorkflowCompiledPlans.Table() + `
WHERE plan_draft_id = ?
ORDER BY compiled_version DESC
LIMIT 1`
		args = []any{draft.Id}
	default:
		query = `
SELECT
  id,
  project_id,
  plan_draft_id,
  plan_review_result_id,
  compiled_version,
  COALESCE(compile_run_id, ''),
  project_category,
  status,
  COALESCE(risk_summary_json, ''),
  COALESCE(compile_diff_json, ''),
  generated_at,
  COALESCE(activated_at, '')
FROM ` + dao.WorkflowCompiledPlans.Table() + `
WHERE project_id = ?
ORDER BY compiled_version DESC
LIMIT 1`
		args = []any{project.Id}
	}

	row := db.QueryRowContext(ctx, query, args...)
	item, err := scanCompiledPlanRow(row)
	if err == nil {
		return item, nil
	}
	if isSchemaMissingError(err) || err == sql.ErrNoRows {
		return nil, nil
	}
	return nil, err
}

func listCompiledTasksForProject(ctx context.Context, db *sql.DB, project entity.Projects, compiled *entity.WorkflowCompiledPlans) ([]entity.WorkflowCompiledTasks, error) {
	if compiled == nil {
		return nil, nil
	}

	query := `
SELECT
  id,
  compiled_plan_id,
  task_key,
  name,
  COALESCE(description, ''),
  phase,
  task_kind,
  role_type,
  brain_kind,
  risk_level,
  affected_resources_json,
  delivery_contract_json,
  verification_contract_json,
  manual_review_required,
  COALESCE(depends_on_task_keys_json, ''),
  status,
  created_at,
  updated_at
FROM ` + dao.WorkflowCompiledTasks.Table() + `
WHERE compiled_plan_id = ?
ORDER BY created_at ASC, task_key ASC
LIMIT ?`

	rows, err := db.QueryContext(ctx, query, compiled.Id, planTaskProjectionLimit)
	if err != nil {
		if isSchemaMissingError(err) {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query compiled tasks failed")
	}
	defer rows.Close()

	items := make([]entity.WorkflowCompiledTasks, 0)
	for rows.Next() {
		var item entity.WorkflowCompiledTasks
		if err = rows.Scan(
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
			return nil, gerror.Wrap(err, "scan compiled task failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate compiled tasks failed")
	}
	return items, nil
}

func listDomainTasksForPlan(ctx context.Context, db *sql.DB, project entity.Projects, compiled *entity.WorkflowCompiledPlans) ([]entity.DomainTasks, error) {
	var (
		query string
		args  []any
	)

	if compiled != nil {
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
ORDER BY updated_at DESC, created_at DESC
LIMIT ?`
		args = []any{project.Id, compiled.Id, planTaskProjectionLimit}
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
ORDER BY updated_at DESC, created_at DESC
LIMIT ?`
		args = []any{project.Id, planTaskProjectionLimit}
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		if isSchemaMissingError(err) {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query domain tasks for plan failed")
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
			return nil, gerror.Wrap(err, "scan domain task for plan failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate domain tasks for plan failed")
	}
	return items, nil
}

func buildPlanDraftView(aggregate *planAggregate) planv1.PlanDraftView {
	if aggregate.Draft != nil {
		return planv1.PlanDraftView{
			ID:          aggregate.Draft.Id,
			Version:     aggregate.Draft.Version,
			Status:      normalizePlanState(aggregate.Draft.Status, "ready"),
			GoalSummary: planFirstNonEmpty(aggregate.Draft.GoalSummary, aggregate.Project.GoalSummary),
		}
	}

	return planv1.PlanDraftView{
		ID:          aggregate.Project.CurrentPlanDraftId,
		Version:     0,
		Status:      deriveDraftFallbackStatus(aggregate.Project.Status),
		GoalSummary: aggregate.Project.GoalSummary,
	}
}

func buildPlanReviewView(aggregate *planAggregate) planv1.PlanReviewView {
	if aggregate.Review != nil {
		return planv1.PlanReviewView{
			ID:                 aggregate.Review.Id,
			ReviewVersion:      aggregate.Review.ReviewVersion,
			Decision:           normalizePlanState(aggregate.Review.Decision, "pending"),
			BlockingIssueCount: aggregate.Review.BlockingIssueCount,
			AdvisoryIssueCount: aggregate.Review.AdvisoryIssueCount,
		}
	}

	return planv1.PlanReviewView{
		ID:                 "",
		ReviewVersion:      0,
		Decision:           deriveReviewFallbackDecision(aggregate),
		BlockingIssueCount: 0,
		AdvisoryIssueCount: 0,
	}
}

func buildCompiledPlanView(aggregate *planAggregate) planv1.CompiledPlanView {
	if aggregate.Compiled != nil {
		return planv1.CompiledPlanView{
			ID:              aggregate.Compiled.Id,
			CompiledVersion: aggregate.Compiled.CompiledVersion,
			Status:          normalizePlanState(aggregate.Compiled.Status, "ready"),
			RiskSummary:     summarizeRiskJSON(aggregate.Compiled.RiskSummaryJson, aggregate.CompiledTasks, aggregate.DomainTasks),
		}
	}

	return planv1.CompiledPlanView{
		ID:              aggregate.Project.CurrentCompiledPlanId,
		CompiledVersion: 0,
		Status:          deriveCompiledFallbackStatus(aggregate),
		RiskSummary:     summarizeRiskJSON("", aggregate.CompiledTasks, aggregate.DomainTasks),
	}
}

func buildPlanTaskProjection(aggregate *planAggregate) []planv1.CompiledTaskView {
	if len(aggregate.CompiledTasks) > 0 {
		taskMap := make(map[string]entity.DomainTasks, len(aggregate.DomainTasks))
		keyMap := make(map[string]entity.DomainTasks, len(aggregate.DomainTasks))
		for _, item := range aggregate.DomainTasks {
			if item.SourceCompiledTaskId != "" {
				taskMap[item.SourceCompiledTaskId] = item
			}
			if item.SourceTaskKey != "" {
				keyMap[item.SourceTaskKey] = item
			}
		}

		result := make([]planv1.CompiledTaskView, 0, len(aggregate.CompiledTasks))
		for _, item := range aggregate.CompiledTasks {
			mapped, ok := taskMap[item.Id]
			if !ok && item.TaskKey != "" {
				mapped, ok = keyMap[item.TaskKey]
			}
			view := planv1.CompiledTaskView{
				TaskID:               item.Id,
				TaskKey:              item.TaskKey,
				TaskName:             item.Name,
				Phase:                item.Phase,
				TaskKind:             item.TaskKind,
				RoleType:             item.RoleType,
				BrainKind:            item.BrainKind,
				RiskLevel:            item.RiskLevel,
				Status:               normalizePlanState(item.Status, "planned"),
				DeliverySummary:      summarizeContractJSON(item.DeliveryContractJson),
				VerificationSummary:  summarizeContractJSON(item.VerificationContractJson),
				AffectedResources:    summarizeResourceJSON(item.AffectedResourcesJson),
				ManualReviewRequired: item.ManualReviewRequired > 0,
			}
			if ok {
				view.MappedDomainTaskID = mapped.Id
				view.MappedDomainTaskState = normalizePlanState(mapped.Status, "queued")
				if view.Status == "planned" {
					view.Status = view.MappedDomainTaskState
				}
			}
			result = append(result, view)
		}
		return result
	}

	result := make([]planv1.CompiledTaskView, 0, len(aggregate.DomainTasks))
	for _, item := range aggregate.DomainTasks {
		result = append(result, planv1.CompiledTaskView{
			TaskID:                planFirstNonEmpty(item.SourceCompiledTaskId, item.Id),
			TaskKey:               item.SourceTaskKey,
			TaskName:              item.Name,
			Phase:                 item.Phase,
			TaskKind:              item.TaskKind,
			RoleType:              item.RoleType,
			BrainKind:             item.BrainKind,
			RiskLevel:             item.RiskLevel,
			Status:                normalizePlanState(item.Status, "queued"),
			DeliverySummary:       "derived_from_domain_task",
			VerificationSummary:   "derived_from_domain_task",
			AffectedResources:     nil,
			ManualReviewRequired:  item.ManualReviewRequired > 0,
			MappedDomainTaskID:    item.Id,
			MappedDomainTaskState: normalizePlanState(item.Status, "queued"),
		})
	}
	return result
}

func buildRepairDraftView(aggregate *planAggregate) planv1.RepairDraftView {
	if aggregate.RepairDraft == nil {
		return planv1.RepairDraftView{
			Status: "idle",
		}
	}

	return planv1.RepairDraftView{
		ID:                  aggregate.RepairDraft.ID,
		Status:              normalizePlanState(aggregate.RepairDraft.Status, "ready"),
		ReasoningSummary:    strings.TrimSpace(aggregate.RepairDraft.RepairReasoningSummary),
		ReplacedConstraints: parseStringArrayJSON(aggregate.RepairDraft.ReplacedConstraintsJSON),
		UpdatedAt:           strings.TrimSpace(aggregate.RepairDraft.UpdatedAt),
	}
}

func buildPlanDiffSummary(aggregate *planAggregate) planv1.DiffSummary {
	var (
		items          = summarizeDiffItems(aggregate.Compiled, aggregate.Review)
		splitCount     int
		overrideCount  int
		dropCount      int
		unchangedCount int
	)

	for _, item := range items {
		switch item.DiffKind {
		case "split":
			splitCount++
		case "override":
			overrideCount++
		case "drop":
			dropCount++
		default:
			unchangedCount++
		}
	}

	if len(items) == 0 && len(aggregate.CompiledTasks) > 0 {
		unchangedCount = len(aggregate.CompiledTasks)
	}

	if len(items) == 0 && aggregate.Draft != nil && countJSONArrayItems(aggregate.Draft.DraftTasksJson) > 0 {
		unchangedCount = countJSONArrayItems(aggregate.Draft.DraftTasksJson)
	}

	summary := buildDiffSummaryText(aggregate, splitCount, overrideCount, dropCount, unchangedCount)

	return planv1.DiffSummary{
		TotalChanges:     splitCount + overrideCount + dropCount + unchangedCount,
		SplitCount:       splitCount,
		OverrideCount:    overrideCount,
		DropCount:        dropCount,
		UnchangedCount:   unchangedCount,
		ReviewIssueCount: reviewIssueCount(aggregate.Review),
		Summary:          summary,
		Items:            items,
	}
}

func buildPlanOverview(aggregate *planAggregate) planv1.PlanOverview {
	var (
		draft      = buildPlanDraftView(aggregate)
		review     = buildPlanReviewView(aggregate)
		compiled   = buildCompiledPlanView(aggregate)
		repair     = buildRepairDraftView(aggregate)
		tasks      = buildPlanTaskProjection(aggregate)
		nextAction = "refresh_plan_view"
		manualCnt  int
	)

	for _, item := range tasks {
		if item.ManualReviewRequired {
			manualCnt++
		}
	}

	switch {
	case repair.Status != "" && repair.Status != "idle":
		nextAction = "open_repair_draft"
	case compiled.Status == "ready" || compiled.Status == "active":
		nextAction = "open_task_projection"
	case review.Decision == "blocked":
		nextAction = "resolve_review_issues"
	case draft.Status == "pending":
		nextAction = "complete_plan_draft"
	default:
		nextAction = "refresh_plan_view"
	}

	return planv1.PlanOverview{
		ProjectID:             aggregate.Project.Id,
		DraftStatus:           draft.Status,
		ReviewDecision:        review.Decision,
		CompiledStatus:        compiled.Status,
		RepairDraftStatus:     repair.Status,
		CurrentStage:          normalizeProjectStage(aggregate.Project.Status),
		NextAction:            nextAction,
		TaskCount:             len(tasks),
		ManualReviewTaskCount: manualCnt,
		BlockingIssueCount:    review.BlockingIssueCount,
		AdvisoryIssueCount:    review.AdvisoryIssueCount,
		CompiledVersion:       compiled.CompiledVersion,
	}
}

func scanPlanDraftRow(row *sql.Row) (*entity.WorkflowPlanDrafts, error) {
	var item entity.WorkflowPlanDrafts
	if err := row.Scan(
		&item.Id,
		&item.ProjectId,
		&item.Version,
		&item.SourceKind,
		&item.SourceRunId,
		&item.ProjectCategory,
		&item.GoalSummary,
		&item.InputRequirementsJson,
		&item.DraftTasksJson,
		&item.Status,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if isSchemaMissingError(err) {
			return nil, err
		}
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, gerror.Wrap(err, "query plan draft failed")
	}
	return &item, nil
}

func getLatestRepairDraftForProject(ctx context.Context, db *sql.DB, projectID string) (*repairPlanDraftRecord, error) {
	query := `
SELECT
  id,
  project_id,
  COALESCE(failed_task_context_json, ''),
  COALESCE(failure_reason_json, ''),
  COALESCE(original_contracts_json, ''),
  COALESCE(runtime_summary_json, ''),
  COALESCE(repair_plan_json, ''),
  repair_reasoning_summary,
  COALESCE(replaced_constraints_json, ''),
  status,
  COALESCE(created_by, ''),
  COALESCE(created_at, ''),
  updated_at
FROM ` + repairPlanDraftsTable + `
WHERE project_id = ?
ORDER BY updated_at DESC, created_at DESC
LIMIT 1`

	row := db.QueryRowContext(ctx, query, projectID)
	var item repairPlanDraftRecord
	if err := row.Scan(
		&item.ID,
		&item.ProjectID,
		&item.FailedTaskContextJSON,
		&item.FailureReasonJSON,
		&item.OriginalContractsJSON,
		&item.RuntimeSummaryJSON,
		&item.RepairPlanJSON,
		&item.RepairReasoningSummary,
		&item.ReplacedConstraintsJSON,
		&item.Status,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if isSchemaMissingError(err) || err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query latest repair draft failed")
	}
	return &item, nil
}

func loadRepairDraftView(ctx context.Context, projectID string) (*planv1.RepairDraftDetailView, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	item, err := getLatestRepairDraftForProject(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &planv1.RepairDraftDetailView{
			Status: "idle",
		}, nil
	}

	return &planv1.RepairDraftDetailView{
		ID:                    item.ID,
		Status:                normalizePlanState(item.Status, "ready"),
		ReasoningSummary:      strings.TrimSpace(item.RepairReasoningSummary),
		ReplacedConstraints:   parseStringArrayJSON(item.ReplacedConstraintsJSON),
		FailedTaskContextJSON: strings.TrimSpace(item.FailedTaskContextJSON),
		FailureReasonJSON:     strings.TrimSpace(item.FailureReasonJSON),
		OriginalContractsJSON: strings.TrimSpace(item.OriginalContractsJSON),
		RuntimeSummaryJSON:    strings.TrimSpace(item.RuntimeSummaryJSON),
		RepairPlanJSON:        strings.TrimSpace(item.RepairPlanJSON),
		CreatedBy:             strings.TrimSpace(item.CreatedBy),
		CreatedAt:             strings.TrimSpace(item.CreatedAt),
		UpdatedAt:             strings.TrimSpace(item.UpdatedAt),
	}, nil
}

func scanPlanReviewRow(row *sql.Row) (*entity.WorkflowPlanReviewResults, error) {
	var item entity.WorkflowPlanReviewResults
	if err := row.Scan(
		&item.Id,
		&item.ProjectId,
		&item.PlanDraftId,
		&item.ReviewVersion,
		&item.ReviewRunId,
		&item.Decision,
		&item.BlockingIssueCount,
		&item.AdvisoryIssueCount,
		&item.IssuesJson,
		&item.SplitSuggestionsJson,
		&item.OverrideSuggestionsJson,
		&item.Status,
		&item.ReviewedAt,
	); err != nil {
		if isSchemaMissingError(err) {
			return nil, err
		}
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, gerror.Wrap(err, "query plan review failed")
	}
	return &item, nil
}

func scanCompiledPlanRow(row *sql.Row) (*entity.WorkflowCompiledPlans, error) {
	var item entity.WorkflowCompiledPlans
	if err := row.Scan(
		&item.Id,
		&item.ProjectId,
		&item.PlanDraftId,
		&item.PlanReviewResultId,
		&item.CompiledVersion,
		&item.CompileRunId,
		&item.ProjectCategory,
		&item.Status,
		&item.RiskSummaryJson,
		&item.CompileDiffJson,
		&item.GeneratedAt,
		&item.ActivatedAt,
	); err != nil {
		if isSchemaMissingError(err) {
			return nil, err
		}
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, gerror.Wrap(err, "query compiled plan failed")
	}
	return &item, nil
}

func summarizeRiskJSON(raw string, compiledTasks []entity.WorkflowCompiledTasks, domainTasks []entity.DomainTasks) string {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		var object map[string]any
		if json.Unmarshal([]byte(raw), &object) == nil {
			for _, key := range []string{"summary", "risk_summary", "text"} {
				if value := strings.TrimSpace(anyToString(object[key])); value != "" {
					return value
				}
			}
			if level := strings.TrimSpace(anyToString(object["risk_level"])); level != "" {
				return "risk_level=" + level
			}
		}
		if compact := compactJSONString(raw); compact != "" {
			return compact
		}
	}

	highCount := 0
	mediumCount := 0
	for _, item := range compiledTasks {
		switch strings.ToLower(strings.TrimSpace(item.RiskLevel)) {
		case "high", "critical":
			highCount++
		case "medium":
			mediumCount++
		}
	}
	if highCount == 0 && mediumCount == 0 {
		for _, item := range domainTasks {
			switch strings.ToLower(strings.TrimSpace(item.RiskLevel)) {
			case "high", "critical":
				highCount++
			case "medium":
				mediumCount++
			}
		}
	}
	if highCount == 0 && mediumCount == 0 {
		return ""
	}
	return fmt.Sprintf("high=%d, medium=%d", highCount, mediumCount)
}

func summarizeContractJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var object map[string]any
	if json.Unmarshal([]byte(raw), &object) == nil {
		for _, key := range []string{"summary", "title", "description", "kind"} {
			if value := strings.TrimSpace(anyToString(object[key])); value != "" {
				return value
			}
		}
		if len(object) > 0 {
			for _, value := range object {
				if text := strings.TrimSpace(anyToString(value)); text != "" {
					return text
				}
			}
		}
	}
	return compactJSONString(raw)
}

func summarizeResourceJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var list []any
	if json.Unmarshal([]byte(raw), &list) == nil {
		result := make([]string, 0, len(list))
		for _, item := range list {
			switch value := item.(type) {
			case string:
				value = strings.TrimSpace(value)
				if value != "" {
					result = append(result, value)
				}
			case map[string]any:
				for _, key := range []string{"path", "resource", "name"} {
					if text := strings.TrimSpace(anyToString(value[key])); text != "" {
						result = append(result, text)
						break
					}
				}
			}
		}
		return result
	}

	var object map[string]any
	if json.Unmarshal([]byte(raw), &object) == nil {
		for _, key := range []string{"resources", "paths"} {
			if nested := summarizeResourceJSON(anyToString(object[key])); len(nested) > 0 {
				return nested
			}
		}
	}

	compact := compactJSONString(raw)
	if compact == "" {
		return nil
	}
	return []string{compact}
}

func summarizeDiffItems(compiled *entity.WorkflowCompiledPlans, review *entity.WorkflowPlanReviewResults) []planv1.DiffSummaryItem {
	items := make([]planv1.DiffSummaryItem, 0)

	if compiled != nil && strings.TrimSpace(compiled.CompileDiffJson) != "" {
		items = append(items, parseDiffItemsFromJSON(compiled.CompileDiffJson)...)
	}
	if len(items) == 0 && review != nil {
		items = append(items, parseSuggestionItems(review.SplitSuggestionsJson, "split")...)
		items = append(items, parseSuggestionItems(review.OverrideSuggestionsJson, "override")...)
	}
	if len(items) == 0 {
		return nil
	}
	return items
}

func parseDiffItemsFromJSON(raw string) []planv1.DiffSummaryItem {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var object map[string]any
	if json.Unmarshal([]byte(raw), &object) == nil {
		for _, key := range []string{"items", "diffs", "changes"} {
			if items := parseDiffItemsValue(object[key], "change"); len(items) > 0 {
				return items
			}
		}
	}
	return parseDiffItemsValue(json.RawMessage(raw), "change")
}

func parseDiffItemsValue(value any, fallbackKind string) []planv1.DiffSummaryItem {
	switch typed := value.(type) {
	case nil:
		return nil
	case json.RawMessage:
		var list []map[string]any
		if json.Unmarshal(typed, &list) == nil {
			return convertDiffMaps(list, fallbackKind)
		}
		var object map[string]any
		if json.Unmarshal(typed, &object) == nil {
			return convertDiffMaps([]map[string]any{object}, fallbackKind)
		}
	case []any:
		list := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if object, ok := item.(map[string]any); ok {
				list = append(list, object)
			}
		}
		return convertDiffMaps(list, fallbackKind)
	case map[string]any:
		return convertDiffMaps([]map[string]any{typed}, fallbackKind)
	}
	return nil
}

func convertDiffMaps(items []map[string]any, fallbackKind string) []planv1.DiffSummaryItem {
	result := make([]planv1.DiffSummaryItem, 0, len(items))
	for _, item := range items {
		diffItem := planv1.DiffSummaryItem{
			DiffKind:            planFirstNonEmpty(anyToString(item["diff_kind"]), anyToString(item["kind"]), fallbackKind),
			BeforeLabel:         planFirstNonEmpty(anyToString(item["before_label"]), anyToString(item["before"]), anyToString(item["source_task"])),
			AfterLabel:          planFirstNonEmpty(anyToString(item["after_label"]), anyToString(item["after"]), anyToString(item["target_task"])),
			Reason:              planFirstNonEmpty(anyToString(item["reason"]), anyToString(item["summary"]), anyToString(item["message"])),
			SourceReviewIssueID: planFirstNonEmpty(anyToString(item["source_review_issue_id"]), anyToString(item["issue_id"])),
		}
		if diffItem.DiffKind == "" && diffItem.Reason == "" && diffItem.BeforeLabel == "" && diffItem.AfterLabel == "" {
			continue
		}
		result = append(result, diffItem)
	}
	return result
}

func parseSuggestionItems(raw string, diffKind string) []planv1.DiffSummaryItem {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var list []map[string]any
	if json.Unmarshal([]byte(raw), &list) != nil {
		return nil
	}

	result := make([]planv1.DiffSummaryItem, 0, len(list))
	for _, item := range list {
		result = append(result, planv1.DiffSummaryItem{
			DiffKind:            diffKind,
			BeforeLabel:         planFirstNonEmpty(anyToString(item["task"]), anyToString(item["before"]), anyToString(item["source_task"])),
			AfterLabel:          planFirstNonEmpty(anyToString(item["proposal"]), anyToString(item["after"]), anyToString(item["target_task"])),
			Reason:              planFirstNonEmpty(anyToString(item["reason"]), anyToString(item["summary"])),
			SourceReviewIssueID: anyToString(item["issue_id"]),
		})
	}
	return result
}

func buildDiffSummaryText(aggregate *planAggregate, splitCount, overrideCount, dropCount, unchangedCount int) string {
	switch {
	case aggregate.Compiled != nil:
		return fmt.Sprintf("compiled_version=%d, split=%d, override=%d, drop=%d, unchanged=%d", aggregate.Compiled.CompiledVersion, splitCount, overrideCount, dropCount, unchangedCount)
	case aggregate.Draft != nil:
		return fmt.Sprintf("draft_version=%d, review_issues=%d, compiled=missing", aggregate.Draft.Version, reviewIssueCount(aggregate.Review))
	default:
		return "project_has_not_generated_plan_objects"
	}
}

func reviewIssueCount(review *entity.WorkflowPlanReviewResults) int {
	if review == nil {
		return 0
	}
	return review.BlockingIssueCount + review.AdvisoryIssueCount
}

func countJSONArrayItems(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	var list []any
	if json.Unmarshal([]byte(raw), &list) != nil {
		return 0
	}
	return len(list)
}

func normalizePlanState(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func deriveDraftFallbackStatus(projectStatus string) string {
	switch strings.ToLower(strings.TrimSpace(projectStatus)) {
	case "created", "draft":
		return "pending"
	default:
		return "missing"
	}
}

func deriveReviewFallbackDecision(aggregate *planAggregate) string {
	if aggregate.Compiled != nil {
		return "accepted"
	}
	if aggregate.Draft != nil {
		return "pending"
	}
	return "not_started"
}

func deriveCompiledFallbackStatus(aggregate *planAggregate) string {
	if len(aggregate.DomainTasks) > 0 {
		return "projected"
	}
	if aggregate.Review != nil {
		return "pending"
	}
	return "missing"
}

func compactJSONString(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var payload any
	if json.Unmarshal([]byte(raw), &payload) == nil {
		switch typed := payload.(type) {
		case string:
			return typed
		case []any:
			if len(typed) > 0 {
				return anyToString(typed[0])
			}
		case map[string]any:
			for _, value := range typed {
				if text := anyToString(value); strings.TrimSpace(text) != "" {
					return text
				}
			}
		}
	}
	return raw
}

func anyToString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.RawMessage:
		return string(typed)
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%d", int64(typed))
		}
		return fmt.Sprintf("%v", typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return string(data)
	}
}

func planFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func isSchemaMissingError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "no such table") || strings.Contains(text, "no such column")
}
