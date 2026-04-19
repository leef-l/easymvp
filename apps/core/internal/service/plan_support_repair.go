package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
)

const repairPlanDraftsTable = "repair_plan_drafts"

func createRepairDraftForProject(ctx context.Context, req CreateRepairDraftCommand) (string, error) {
	normalized, err := normalizeCreateRepairDraftCommand(req)
	if err != nil {
		return "", err
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return "", err
	}
	defer closeFn()

	project, err := getProjectByID(ctx, db, normalized.ProjectID)
	if err != nil {
		return "", err
	}

	_, result, err := EasyMVPBrain().CallRepairDesign(ctx, braincontracts.RepairDesignInput{
		FailedTaskContextJSON: json.RawMessage(normalized.FailedTaskContextJSON),
		FailureReasonJSON:     json.RawMessage(normalized.FailureReasonJSON),
		OriginalContractsJSON: json.RawMessage(normalized.OriginalContractsJSON),
		RuntimeSummaryJSON:    json.RawMessage(normalized.RuntimeSummaryJSON),
		ArtifactRefs:          normalized.ArtifactRefs,
	})
	if err != nil {
		return "", err
	}

	now := nowText()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", gerror.Wrap(err, "begin repair draft transaction failed")
	}

	if err = insertRepairPlanDraftRow(ctx, tx, repairPlanDraftRow{
		ID:                     result.RepairPlanDraftID,
		ProjectID:              project.Id,
		FailedTaskContextJSON:  normalized.FailedTaskContextJSON,
		FailureReasonJSON:      normalized.FailureReasonJSON,
		OriginalContractsJSON:  normalized.OriginalContractsJSON,
		RuntimeSummaryJSON:     normalized.RuntimeSummaryJSON,
		ArtifactRefsJSON:       mustMarshalJSONString(normalized.ArtifactRefs, "[]"),
		RepairPlanJSON:         strings.TrimSpace(string(result.RepairPlanJSON)),
		RepairReasoningSummary: strings.TrimSpace(result.RepairReasoningSummary),
		ReplacedConstraintsJSON: mustMarshalJSONString(
			result.ReplacedConstraints,
			"[]",
		),
		Status:    "ready",
		CreatedBy: firstNonEmpty(normalized.CreatedBy, "system"),
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", gerror.Wrap(err, "commit repair draft transaction failed")
	}
	return result.RepairPlanDraftID, nil
}

type normalizedCreateRepairDraftCommand struct {
	ProjectID             string
	FailedTaskContextJSON string
	FailureReasonJSON     string
	OriginalContractsJSON string
	RuntimeSummaryJSON    string
	ArtifactRefs          []braincontracts.ArtifactRef
	CreatedBy             string
}

type repairPlanDraftRow struct {
	ID                      string
	ProjectID               string
	FailedTaskContextJSON   string
	FailureReasonJSON       string
	OriginalContractsJSON   string
	RuntimeSummaryJSON      string
	ArtifactRefsJSON        string
	RepairPlanJSON          string
	RepairReasoningSummary  string
	ReplacedConstraintsJSON string
	Status                  string
	CreatedBy               string
	CreatedAt               string
	UpdatedAt               string
}

func normalizeCreateRepairDraftCommand(req CreateRepairDraftCommand) (*normalizedCreateRepairDraftCommand, error) {
	normalized := &normalizedCreateRepairDraftCommand{
		ProjectID:             strings.TrimSpace(req.ProjectID),
		FailedTaskContextJSON: strings.TrimSpace(req.FailedTaskContextJSON),
		FailureReasonJSON:     strings.TrimSpace(req.FailureReasonJSON),
		OriginalContractsJSON: strings.TrimSpace(req.OriginalContractsJSON),
		RuntimeSummaryJSON:    strings.TrimSpace(req.RuntimeSummaryJSON),
		ArtifactRefs:          req.ArtifactRefs,
		CreatedBy:             strings.TrimSpace(req.CreatedBy),
	}

	if normalized.ProjectID == "" {
		return nil, gerror.New("project id is required")
	}
	if err := requireJSONObject("failed task context", normalized.FailedTaskContextJSON); err != nil {
		return nil, err
	}
	if err := requireJSONObject("failure reason", normalized.FailureReasonJSON); err != nil {
		return nil, err
	}
	if err := requireJSONObject("original contracts", normalized.OriginalContractsJSON); err != nil {
		return nil, err
	}
	if err := requireJSONObject("runtime summary", normalized.RuntimeSummaryJSON); err != nil {
		return nil, err
	}
	return normalized, nil
}

func requireJSONObject(label string, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return gerror.New(label + " json is required")
	}
	if !json.Valid([]byte(raw)) {
		return gerror.New(label + " json is invalid")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return gerror.Wrap(err, label+" json is invalid")
	}
	return nil
}

func insertRepairPlanDraftRow(ctx context.Context, tx *sql.Tx, row repairPlanDraftRow) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+repairPlanDraftsTable+` (
id, project_id, failed_task_context_json, failure_reason_json, original_contracts_json, runtime_summary_json, artifact_refs_json, repair_plan_json, repair_reasoning_summary, replaced_constraints_json, status, created_by, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID,
		row.ProjectID,
		row.FailedTaskContextJSON,
		row.FailureReasonJSON,
		row.OriginalContractsJSON,
		row.RuntimeSummaryJSON,
		nullIfEmpty(row.ArtifactRefsJSON),
		row.RepairPlanJSON,
		row.RepairReasoningSummary,
		nullIfEmpty(row.ReplacedConstraintsJSON),
		row.Status,
		row.CreatedBy,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		if isSchemaMissingError(err) {
			return gerror.Wrap(err, "insert repair draft failed: migration may be missing")
		}
		return gerror.Wrap(err, "insert repair draft failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert repair draft affected unexpected rows")
	}
	return nil
}
