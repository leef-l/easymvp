package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

var planCompileTaskKeySanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func createInitialDraftIfNeeded(ctx context.Context, projectID string) error {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return err
	}
	if project.CurrentPlanDraftId != "" {
		draft, err := getPlanDraftForProject(ctx, db, *project)
		if err == nil && draft != nil {
			return nil
		}
		if err != nil && err != sql.ErrNoRows && !isSchemaMissingError(err) {
			return err
		}
	}

	profile, err := getProjectProfileByProjectID(ctx, db, projectID)
	if err != nil {
		return err
	}

	var (
		now               = nowText()
		draftID           = newResourceID("plan_draft")
		inputRequirements = mustMarshalJSONString(map[string]any{
			"goal_summary":               project.GoalSummary,
			"workspace_root":             project.WorkspaceRoot,
			"repo_root":                  project.RepoRoot,
			"category_profile_version":   profile.CategoryProfileVersion,
			"acceptance_profile_version": profile.AcceptanceProfileVersion,
			"role_profile_version":       profile.RoleProfileVersion,
		}, "{}")
		draftTasks = mustMarshalJSONString([]map[string]any{
			{
				"task_key":   "project_delivery",
				"name":       project.GoalSummary,
				"phase":      "design",
				"task_kind":  "planning",
				"summary":    "Auto-generated initial draft from project goal",
				"brain_kind": "easymvp-brain",
			},
		}, "[]")
	)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return gerror.Wrap(err, "begin create initial draft transaction failed")
	}
	if err = insertPlanDraftRow(ctx, tx, entity.WorkflowPlanDrafts{
		Id:                    draftID,
		ProjectId:             projectID,
		Version:               1,
		SourceKind:            "project_create",
		ProjectCategory:       project.ProjectCategory,
		GoalSummary:           project.GoalSummary,
		InputRequirementsJson: inputRequirements,
		DraftTasksJson:        draftTasks,
		Status:                "ready",
		CreatedBy:             "system",
		CreatedAt:             now,
		UpdatedAt:             now,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = updateProjectCurrentPlanDraft(ctx, tx, projectID, draftID); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		return gerror.Wrap(err, "commit create initial draft transaction failed")
	}
	return nil
}

func compilePlanForProject(ctx context.Context, req CompilePlanCommand) (string, error) {
	if err := createInitialDraftIfNeeded(ctx, req.ProjectID); err != nil {
		return "", err
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return "", err
	}
	defer closeFn()

	project, err := getProjectByID(ctx, db, req.ProjectID)
	if err != nil {
		return "", err
	}
	profile, err := getProjectProfileByProjectID(ctx, db, req.ProjectID)
	if err != nil {
		return "", err
	}
	aggregate, err := loadPlanAggregate(ctx, db, req.ProjectID)
	if err != nil {
		return "", err
	}
	if aggregate.Draft == nil {
		return "", gerror.New("plan draft is required")
	}
	if req.PlanDraftID != "" && req.PlanDraftID != aggregate.Draft.Id {
		return "", gerror.New("plan draft id does not match current draft")
	}

	review := aggregate.Review
	reviewCompileAllowed := false
	if review != nil {
		reviewCompileAllowed = planReviewCompileAllowed(review)
	}
	if review == nil || !reviewCompileAllowed || req.ForceRecompile {
		review, err = runPlanReview(ctx, project, profile, aggregate.Draft)
		if err != nil {
			return "", err
		}
	}
	if !planReviewCompileAllowed(review) {
		return "", gerror.New("plan review does not allow compile")
	}

	if aggregate.Compiled != nil && aggregate.Compiled.PlanReviewResultId == review.Id && !req.ForceRecompile {
		return aggregate.Compiled.Id, nil
	}

	compiledPlanID, err := runPlanCompile(ctx, project, profile, aggregate.Draft, review)
	if err != nil {
		return "", err
	}
	if _, _, err = refreshAcceptanceProfilesAfterCompile(ctx, req.ProjectID, profile.AcceptanceProfileVersion); err != nil {
		return compiledPlanID, gerror.Wrapf(err, "compiled plan %s created but acceptance profiles refresh failed", compiledPlanID)
	}
	return compiledPlanID, nil
}

func refreshAcceptanceProfilesAfterCompile(
	ctx context.Context,
	projectID string,
	profileVersion string,
) (*acceptanceProfileRecord, *productionAcceptanceProfileRecord, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer closeFn()

	return ensureAcceptanceProfilesCurrent(ctx, db, projectID, profileVersion)
}

func runPlanReview(
	ctx context.Context,
	project *entity.Projects,
	profile *entity.ProjectProfiles,
	draft *entity.WorkflowPlanDrafts,
) (*entity.WorkflowPlanReviewResults, error) {
	if project == nil || profile == nil || draft == nil {
		return nil, gerror.New("project, profile and draft are required")
	}

	envelope, result, err := EasyMVPBrain().CallPlanReview(ctx, braincontracts.PlanReviewInput{
		PlanDraftID:            draft.Id,
		PlanDraftVersion:       draft.Version,
		PlanDraftJSON:          mustMarshalRawJSON(buildPlanDraftContractPayload(project, draft)),
		ProjectCategory:        project.ProjectCategory,
		CategoryProfileVersion: 1,
		CategoryProfileJSON:    mustMarshalRawJSON(buildCategoryProfilePayload(project, profile)),
		ProjectContextJSON:     mustMarshalRawJSON(buildProjectContextPayload(project, profile)),
	})
	if err != nil {
		return nil, err
	}

	var (
		now        = nowText()
		issuesJSON = mustMarshalJSONString(map[string]any{
			"compile_allowed": result.CompileAllowed,
			"blocking_issues": result.BlockingIssues,
			"advisory_issues": result.AdvisoryIssues,
		}, "{}")
		splitSuggestions = mustMarshalJSONString(result.RewriteHints, "[]")
		row              = entity.WorkflowPlanReviewResults{
			Id:                      result.ReviewResultID,
			ProjectId:               project.Id,
			PlanDraftId:             draft.Id,
			ReviewVersion:           result.ReviewVersion,
			ReviewRunId:             strings.TrimSpace(envelope.TraceID),
			Decision:                result.Decision,
			BlockingIssueCount:      len(result.BlockingIssues),
			AdvisoryIssueCount:      len(result.AdvisoryIssues),
			IssuesJson:              issuesJSON,
			SplitSuggestionsJson:    splitSuggestions,
			OverrideSuggestionsJson: "[]",
			Status:                  planReviewStatusFromDecision(result.Decision, result.CompileAllowed),
			ReviewedAt:              now,
		}
	)

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, gerror.Wrap(err, "begin plan review transaction failed")
	}
	if err = insertPlanReviewRow(ctx, tx, row); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, gerror.Wrap(err, "commit plan review transaction failed")
	}
	return &row, nil
}

func runPlanCompile(
	ctx context.Context,
	project *entity.Projects,
	profile *entity.ProjectProfiles,
	draft *entity.WorkflowPlanDrafts,
	review *entity.WorkflowPlanReviewResults,
) (string, error) {
	if project == nil || profile == nil || draft == nil || review == nil {
		return "", gerror.New("project, profile, draft and review are required")
	}

	envelope, result, err := EasyMVPBrain().CallPlanCompile(ctx, braincontracts.PlanCompileInput{
		PlanDraftJSON:        mustMarshalRawJSON(buildPlanDraftContractPayload(project, draft)),
		PlanReviewResultJSON: mustMarshalRawJSON(buildPlanReviewContractPayload(review)),
		CategoryProfileJSON:  mustMarshalRawJSON(buildCategoryProfilePayload(project, profile)),
		RoleContextJSON:      mustMarshalRawJSON(buildRoleContextPayload(profile)),
	})
	if err != nil {
		return "", err
	}

	var (
		now         = nowText()
		compileDiff = mustMarshalJSONString(map[string]any{
			"source_review_id": review.Id,
			"decision_summary": envelope.DecisionSummary,
			"items":            mustUnmarshalJSONArray(review.SplitSuggestionsJson),
		}, "{}")
		compiledRow = entity.WorkflowCompiledPlans{
			Id:                 result.CompiledPlanID,
			ProjectId:          project.Id,
			PlanDraftId:        draft.Id,
			PlanReviewResultId: review.Id,
			CompiledVersion:    result.CompiledVersion,
			CompileRunId:       strings.TrimSpace(envelope.TraceID),
			ProjectCategory:    project.ProjectCategory,
			Status:             "compiled",
			RiskSummaryJson:    string(result.RiskSummary),
			CompileDiffJson:    compileDiff,
			GeneratedAt:        now,
			ActivatedAt:        now,
		}
		taskRows = buildCompiledTaskRows(result.CompiledPlanID, result.CompiledTasks, now)
	)

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return "", err
	}
	defer closeFn()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", gerror.Wrap(err, "begin plan compile transaction failed")
	}
	if err = deleteCompiledTasksByPlanID(ctx, tx, result.CompiledPlanID); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if err = insertCompiledPlanRow(ctx, tx, compiledRow); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	for _, taskRow := range taskRows {
		if err = insertCompiledTaskRow(ctx, tx, taskRow); err != nil {
			_ = tx.Rollback()
			return "", err
		}
	}
	if err = syncDomainTasksForCompiledPlan(ctx, tx, project.Id, compiledRow, taskRows, now); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if err = updateProjectCurrentCompiledPlan(ctx, tx, project.Id, result.CompiledPlanID); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", gerror.Wrap(err, "commit plan compile transaction failed")
	}
	return result.CompiledPlanID, nil
}

func getProjectProfileByProjectID(ctx context.Context, db *sql.DB, projectID string) (*entity.ProjectProfiles, error) {
	query := `
SELECT
  id,
  project_id,
  category_profile_version,
  acceptance_profile_version,
  COALESCE(role_profile_version, ''),
  created_at,
  updated_at
FROM ` + dao.ProjectProfiles.Table() + `
WHERE project_id = ?
LIMIT 1`

	row := db.QueryRowContext(ctx, query, projectID)
	var item entity.ProjectProfiles
	if err := row.Scan(
		&item.Id,
		&item.ProjectId,
		&item.CategoryProfileVersion,
		&item.AcceptanceProfileVersion,
		&item.RoleProfileVersion,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.Newf("project profile not found: %s", projectID)
		}
		return nil, gerror.Wrap(err, "query project profile failed")
	}
	return &item, nil
}

func insertPlanDraftRow(ctx context.Context, tx *sql.Tx, row entity.WorkflowPlanDrafts) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.WorkflowPlanDrafts.Table()+` (
id, project_id, version, source_kind, source_run_id, project_category, goal_summary, input_requirements_json, draft_tasks_json, status, created_by, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.Version,
		row.SourceKind,
		nullIfEmpty(row.SourceRunId),
		row.ProjectCategory,
		row.GoalSummary,
		row.InputRequirementsJson,
		row.DraftTasksJson,
		row.Status,
		row.CreatedBy,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert plan draft failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert plan draft affected unexpected rows")
	}
	return nil
}

func insertPlanReviewRow(ctx context.Context, tx *sql.Tx, row entity.WorkflowPlanReviewResults) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.WorkflowPlanReviewResults.Table()+` (
id, project_id, plan_draft_id, review_version, review_run_id, decision, blocking_issue_count, advisory_issue_count, issues_json, split_suggestions_json, override_suggestions_json, status, reviewed_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.PlanDraftId,
		row.ReviewVersion,
		nullIfEmpty(row.ReviewRunId),
		row.Decision,
		row.BlockingIssueCount,
		row.AdvisoryIssueCount,
		row.IssuesJson,
		nullIfEmpty(row.SplitSuggestionsJson),
		nullIfEmpty(row.OverrideSuggestionsJson),
		row.Status,
		row.ReviewedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert plan review failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert plan review affected unexpected rows")
	}
	return nil
}

func insertCompiledPlanRow(ctx context.Context, tx *sql.Tx, row entity.WorkflowCompiledPlans) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.WorkflowCompiledPlans.Table()+` (
id, project_id, plan_draft_id, plan_review_result_id, compiled_version, compile_run_id, project_category, status, risk_summary_json, compile_diff_json, generated_at, activated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.PlanDraftId,
		row.PlanReviewResultId,
		row.CompiledVersion,
		nullIfEmpty(row.CompileRunId),
		row.ProjectCategory,
		row.Status,
		nullIfEmpty(row.RiskSummaryJson),
		nullIfEmpty(row.CompileDiffJson),
		row.GeneratedAt,
		nullIfEmpty(row.ActivatedAt),
	)
	if err != nil {
		return gerror.Wrap(err, "insert compiled plan failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert compiled plan affected unexpected rows")
	}
	return nil
}

func insertCompiledTaskRow(ctx context.Context, tx *sql.Tx, row entity.WorkflowCompiledTasks) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.WorkflowCompiledTasks.Table()+` (
id, compiled_plan_id, task_key, name, description, phase, task_kind, role_type, brain_kind, risk_level, affected_resources_json, delivery_contract_json, verification_contract_json, manual_review_required, depends_on_task_keys_json, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.CompiledPlanId,
		row.TaskKey,
		row.Name,
		nullIfEmpty(row.Description),
		row.Phase,
		row.TaskKind,
		row.RoleType,
		row.BrainKind,
		row.RiskLevel,
		row.AffectedResourcesJson,
		row.DeliveryContractJson,
		row.VerificationContractJson,
		row.ManualReviewRequired,
		nullIfEmpty(row.DependsOnTaskKeysJson),
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert compiled task failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert compiled task affected unexpected rows")
	}
	return nil
}

func deleteCompiledTasksByPlanID(ctx context.Context, tx *sql.Tx, compiledPlanID string) error {
	if strings.TrimSpace(compiledPlanID) == "" {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM `+dao.WorkflowCompiledTasks.Table()+` WHERE compiled_plan_id = ?`, compiledPlanID); err != nil {
		return gerror.Wrap(err, "delete compiled tasks by plan failed")
	}
	return nil
}

func syncDomainTasksForCompiledPlan(
	ctx context.Context,
	tx *sql.Tx,
	projectID string,
	compiledPlan entity.WorkflowCompiledPlans,
	compiledTasks []entity.WorkflowCompiledTasks,
	now string,
) error {
	if err := deleteDomainTasksByCompiledPlanID(ctx, tx, compiledPlan.Id); err != nil {
		return err
	}

	domainTaskRows := buildDomainTaskRows(projectID, compiledPlan, compiledTasks, now)
	taskIDByKey := make(map[string]string, len(domainTaskRows))
	for _, row := range domainTaskRows {
		if err := insertDomainTaskRow(ctx, tx, row); err != nil {
			return err
		}
		taskIDByKey[row.SourceTaskKey] = row.Id
	}
	for _, row := range buildTaskDependencyRows(domainTaskRows, compiledTasks, taskIDByKey, now) {
		if err := insertTaskDependencyRow(ctx, tx, row); err != nil {
			return err
		}
	}
	return nil
}

func buildDomainTaskRows(
	projectID string,
	compiledPlan entity.WorkflowCompiledPlans,
	compiledTasks []entity.WorkflowCompiledTasks,
	now string,
) []entity.DomainTasks {
	rows := make([]entity.DomainTasks, 0, len(compiledTasks))
	for _, item := range compiledTasks {
		rows = append(rows, entity.DomainTasks{
			Id:                   newResourceID("task"),
			ProjectId:            projectID,
			SourceCompiledPlanId: compiledPlan.Id,
			SourceCompiledTaskId: item.Id,
			SourceTaskKey:        item.TaskKey,
			CompiledVersion:      compiledPlan.CompiledVersion,
			Name:                 item.Name,
			Phase:                item.Phase,
			TaskKind:             item.TaskKind,
			RoleType:             item.RoleType,
			BrainKind:            item.BrainKind,
			RiskLevel:            item.RiskLevel,
			Status:               domainTaskStatusForCompiledTask(item),
			ManualReviewRequired: item.ManualReviewRequired,
			CreatedAt:            now,
			UpdatedAt:            now,
		})
	}
	return rows
}

func buildTaskDependencyRows(
	domainTasks []entity.DomainTasks,
	compiledTasks []entity.WorkflowCompiledTasks,
	taskIDByKey map[string]string,
	now string,
) []entity.TaskDependencies {
	compiledByKey := make(map[string]entity.WorkflowCompiledTasks, len(compiledTasks))
	for _, item := range compiledTasks {
		compiledByKey[item.TaskKey] = item
	}

	rows := make([]entity.TaskDependencies, 0)
	for _, task := range domainTasks {
		compiledTask, ok := compiledByKey[task.SourceTaskKey]
		if !ok {
			continue
		}
		for _, dependsKey := range parseStringArrayJSON(compiledTask.DependsOnTaskKeysJson) {
			dependsKey = strings.TrimSpace(dependsKey)
			if dependsKey == "" {
				continue
			}
			dependsOnTaskID := strings.TrimSpace(taskIDByKey[dependsKey])
			if dependsOnTaskID == "" {
				continue
			}
			rows = append(rows, entity.TaskDependencies{
				TaskId:          task.Id,
				DependsOnTaskId: dependsOnTaskID,
				CreatedAt:       now,
			})
		}
	}
	return rows
}

func domainTaskStatusForCompiledTask(item entity.WorkflowCompiledTasks) string {
	switch strings.ToLower(strings.TrimSpace(item.Status)) {
	case "completed", "done":
		return "verify_pending"
	case "running", "active":
		return "running"
	case "failed":
		return "failed"
	case "cancelled", "canceled":
		return "cancelled"
	default:
		return "queued"
	}
}

func deleteDomainTasksByCompiledPlanID(ctx context.Context, tx *sql.Tx, compiledPlanID string) error {
	if strings.TrimSpace(compiledPlanID) == "" {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM `+dao.DomainTasks.Table()+` WHERE source_compiled_plan_id = ?`, compiledPlanID); err != nil {
		return gerror.Wrap(err, "delete domain tasks by compiled plan failed")
	}
	return nil
}

func insertDomainTaskRow(ctx context.Context, tx *sql.Tx, row entity.DomainTasks) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.DomainTasks.Table()+` (
id, project_id, source_compiled_plan_id, source_compiled_task_id, source_task_key, compiled_version, name, phase, task_kind, role_type, brain_kind, risk_level, status, manual_review_required, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.SourceCompiledPlanId,
		row.SourceCompiledTaskId,
		row.SourceTaskKey,
		row.CompiledVersion,
		row.Name,
		row.Phase,
		row.TaskKind,
		row.RoleType,
		row.BrainKind,
		row.RiskLevel,
		row.Status,
		row.ManualReviewRequired,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert domain task failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert domain task affected unexpected rows")
	}
	return nil
}

func insertTaskDependencyRow(ctx context.Context, tx *sql.Tx, row entity.TaskDependencies) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.TaskDependencies.Table()+` (
task_id, depends_on_task_id, created_at
) VALUES (?, ?, ?)`,
		row.TaskId,
		row.DependsOnTaskId,
		row.CreatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert task dependency failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert task dependency affected unexpected rows")
	}
	return nil
}

func updateProjectCurrentPlanDraft(ctx context.Context, tx *sql.Tx, projectID string, draftID string) error {
	if _, err := tx.ExecContext(ctx, `UPDATE `+dao.Projects.Table()+` SET current_plan_draft_id = ?, updated_at = ? WHERE id = ?`, draftID, nowText(), projectID); err != nil {
		return gerror.Wrap(err, "update project current plan draft failed")
	}
	return nil
}

func updateProjectCurrentCompiledPlan(ctx context.Context, tx *sql.Tx, projectID string, compiledPlanID string) error {
	if _, err := tx.ExecContext(ctx, `UPDATE `+dao.Projects.Table()+` SET current_compiled_plan_id = ?, updated_at = ? WHERE id = ?`, compiledPlanID, nowText(), projectID); err != nil {
		return gerror.Wrap(err, "update project current compiled plan failed")
	}
	return nil
}

func buildPlanDraftContractPayload(project *entity.Projects, draft *entity.WorkflowPlanDrafts) map[string]any {
	return map[string]any{
		"id":                      draft.Id,
		"project_id":              draft.ProjectId,
		"version":                 draft.Version,
		"source_kind":             draft.SourceKind,
		"project_category":        draft.ProjectCategory,
		"goal_summary":            draft.GoalSummary,
		"input_requirements_json": mustUnmarshalJSONObject(draft.InputRequirementsJson),
		"draft_tasks_json":        mustUnmarshalJSONArray(draft.DraftTasksJson),
		"project_name":            project.Name,
	}
}

func buildPlanReviewContractPayload(review *entity.WorkflowPlanReviewResults) map[string]any {
	payload := mustUnmarshalJSONObject(review.IssuesJson)
	payload["review_result_id"] = review.Id
	payload["review_version"] = review.ReviewVersion
	payload["decision"] = review.Decision
	payload["split_suggestions"] = mustUnmarshalJSONArray(review.SplitSuggestionsJson)
	payload["override_suggestions"] = mustUnmarshalJSONArray(review.OverrideSuggestionsJson)
	return payload
}

func buildCategoryProfilePayload(project *entity.Projects, profile *entity.ProjectProfiles) map[string]any {
	return map[string]any{
		"project_category":           project.ProjectCategory,
		"category_profile_version":   profile.CategoryProfileVersion,
		"acceptance_profile_version": profile.AcceptanceProfileVersion,
	}
}

func buildProjectContextPayload(project *entity.Projects, profile *entity.ProjectProfiles) map[string]any {
	return map[string]any{
		"project_id":                 project.Id,
		"name":                       project.Name,
		"goal_summary":               project.GoalSummary,
		"workspace_root":             project.WorkspaceRoot,
		"repo_root":                  project.RepoRoot,
		"project_category":           project.ProjectCategory,
		"category_profile_version":   profile.CategoryProfileVersion,
		"acceptance_profile_version": profile.AcceptanceProfileVersion,
		"role_profile_version":       profile.RoleProfileVersion,
	}
}

func buildRoleContextPayload(profile *entity.ProjectProfiles) map[string]any {
	return map[string]any{
		"role_profile_version": profile.RoleProfileVersion,
	}
}

func buildCompiledTaskRows(compiledPlanID string, items []braincontracts.CompiledTaskItem, now string) []entity.WorkflowCompiledTasks {
	rows := make([]entity.WorkflowCompiledTasks, 0, len(items))
	for index, item := range items {
		taskKey := buildCompiledTaskKey(index+1, item.Name)
		rows = append(rows, entity.WorkflowCompiledTasks{
			Id:                       item.CompiledTaskID,
			CompiledPlanId:           compiledPlanID,
			TaskKey:                  taskKey,
			Name:                     item.Name,
			Description:              summarizeContractJSON(string(item.DeliveryContract)),
			Phase:                    "execute",
			TaskKind:                 "implementation",
			RoleType:                 item.RoleType,
			BrainKind:                item.BrainKind,
			RiskLevel:                item.RiskLevel,
			AffectedResourcesJson:    "[]",
			DeliveryContractJson:     string(item.DeliveryContract),
			VerificationContractJson: string(item.VerificationContract),
			ManualReviewRequired:     boolToInt(planManualReviewRequired(item.RiskLevel)),
			DependsOnTaskKeysJson:    "[]",
			Status:                   "planned",
			CreatedAt:                now,
			UpdatedAt:                now,
		})
	}
	return rows
}

func buildCompiledTaskKey(index int, name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = planCompileTaskKeySanitizer.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		normalized = "task"
	}
	return fmt.Sprintf("%s_%03d", normalized, index)
}

func planReviewStatusFromDecision(decision string, compileAllowed bool) string {
	if strings.TrimSpace(decision) == "rejected" || !compileAllowed {
		return "rejected"
	}
	return "approved"
}

func planReviewCompileAllowed(review *entity.WorkflowPlanReviewResults) bool {
	if review == nil {
		return false
	}
	if strings.TrimSpace(review.Decision) == "rejected" {
		return false
	}
	if strings.TrimSpace(review.IssuesJson) == "" {
		return true
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(review.IssuesJson), &payload); err != nil {
		return true
	}
	value, ok := payload["compile_allowed"]
	if !ok {
		return true
	}
	switch typed := value.(type) {
	case bool:
		return typed
	default:
		return strings.EqualFold(anyToString(typed), "true")
	}
}

func planManualReviewRequired(riskLevel string) bool {
	switch strings.ToLower(strings.TrimSpace(riskLevel)) {
	case "high", "critical":
		return true
	default:
		return false
	}
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func mustMarshalRawJSON(value any) json.RawMessage {
	return json.RawMessage(mustMarshalJSONString(value, "{}"))
}

func mustMarshalJSONString(value any, fallback string) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fallback
	}
	return string(data)
}

func mustUnmarshalJSONObject(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return map[string]any{}
	}
	return payload
}

func mustUnmarshalJSONArray(raw string) []any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var payload []any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil
	}
	return payload
}
