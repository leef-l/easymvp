package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func redesignPlanDraft(ctx context.Context, projectID string, review *entity.WorkflowPlanReviewResults, feedback string) (*entity.WorkflowPlanDrafts, error) {
	if review == nil {
		return nil, gerror.New("review result is required for redesign")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	draft, err := getPlanDraftForProject(ctx, *project)
	if err != nil {
		return nil, err
	}
	if draft == nil {
		return nil, gerror.New("plan draft is required for redesign")
	}

	rewriteHints := parseStringArrayJSON(review.SplitSuggestionsJson)

	envelope, result, err := EasyMVPBrain().CallPlanRedesign(ctx, braincontracts.PlanRedesignInput{
		PlanDraftID:      draft.Id,
		PlanDraftJSON:    mustMarshalRawJSON(buildPlanDraftContractPayload(project, draft)),
		ReviewResultID:   review.Id,
		ReviewResultJSON: mustMarshalRawJSON(buildPlanReviewContractPayload(review)),
		RewriteHints:     rewriteHints,
		Feedback:         strings.TrimSpace(feedback),
	})
	if err != nil {
		return nil, err
	}

	goalSummary, inputRequirementsJSON, draftTasksJSON, projectCategory := extractRedesignedDraftFields(result.RedesignedPlanJSON, draft)

	now := nowText()
	newDraftID := newResourceID("plan_draft")
	newDraft := entity.WorkflowPlanDrafts{
		Id:                    newDraftID,
		ProjectId:             projectID,
		Version:               draft.Version + 1,
		SourceKind:            "plan_redesign",
		SourceRunId:           strings.TrimSpace(envelope.TraceID),
		ProjectCategory:       projectCategory,
		GoalSummary:           goalSummary,
		InputRequirementsJson: inputRequirementsJSON,
		DraftTasksJson:        draftTasksJSON,
		Status:                "ready",
		CreatedBy:             "system",
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, gerror.Wrap(err, "begin plan redesign transaction failed")
	}
	if err = insertPlanDraftRow(ctx, tx, newDraft); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err = updateProjectCurrentPlanDraft(ctx, tx, projectID, newDraftID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err = insertAuditLogSqlTx(ctx, tx, projectID, "plan.draft.redesigned", "system", "Plan draft redesigned based on review feedback", map[string]any{
		"previous_draft_id":   draft.Id,
		"previous_version":    draft.Version,
		"new_draft_id":        newDraftID,
		"new_version":         newDraft.Version,
		"review_result_id":    review.Id,
		"changes_summary":     result.ChangesSummary,
		"feedback":            strings.TrimSpace(feedback),
		"rewrite_hints_count": len(rewriteHints),
	}); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, gerror.Wrap(err, "commit plan redesign transaction failed")
	}
	return &newDraft, nil
}

func extractRedesignedDraftFields(raw json.RawMessage, original *entity.WorkflowPlanDrafts) (goalSummary, inputRequirementsJSON, draftTasksJSON, projectCategory string) {
	if original == nil {
		original = &entity.WorkflowPlanDrafts{}
	}
	goalSummary = original.GoalSummary
	inputRequirementsJSON = original.InputRequirementsJson
	draftTasksJSON = original.DraftTasksJson
	projectCategory = original.ProjectCategory

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}

	if v, ok := payload["goal_summary"].(string); ok && strings.TrimSpace(v) != "" {
		goalSummary = strings.TrimSpace(v)
	}
	if v, ok := payload["project_category"].(string); ok && strings.TrimSpace(v) != "" {
		projectCategory = strings.TrimSpace(v)
	}
	if v, ok := payload["input_requirements_json"].(map[string]any); ok && v != nil {
		inputRequirementsJSON = mustMarshalJSONString(v, original.InputRequirementsJson)
	} else if v, ok := payload["input_requirements_json"].([]any); ok && len(v) > 0 {
		inputRequirementsJSON = mustMarshalJSONString(v, original.InputRequirementsJson)
	} else if v, ok := payload["input_requirements_json"].(string); ok && strings.TrimSpace(v) != "" {
		inputRequirementsJSON = strings.TrimSpace(v)
	}
	if v, ok := payload["draft_tasks_json"].([]any); ok && len(v) > 0 {
		draftTasksJSON = mustMarshalJSONString(v, original.DraftTasksJson)
	} else if v, ok := payload["draft_tasks_json"].(string); ok && strings.TrimSpace(v) != "" {
		draftTasksJSON = strings.TrimSpace(v)
	}
	return
}
