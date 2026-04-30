package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// ---------------------------------------------------------------------------
// Result types
// ---------------------------------------------------------------------------

// DesignGenerationResult is the result returned after generating a solution design.
type DesignGenerationResult struct {
	DesignID string
	Status   string
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

// IDesign defines the service interface for solution design operations.
type IDesign interface {
	GenerateDesign(ctx context.Context, projectID, requirementID string) (*DesignGenerationResult, error)
	ConfirmDesign(ctx context.Context, designID string) error
	RejectDesign(ctx context.Context, designID string, reason string) error
	GetDesign(ctx context.Context, designID string) (*entity.SolutionDesigns, error)
	GetProjectLatestDesign(ctx context.Context, projectID string) (*entity.SolutionDesigns, error)
}

// ---------------------------------------------------------------------------
// Singleton registration (GoFrame pattern)
// ---------------------------------------------------------------------------

var localDesign IDesign = (*sDesign)(nil)

type sDesign struct{}

func Design() IDesign {
	if localDesign == nil {
		localDesign = &sDesign{}
	}
	return localDesign
}

// ---------------------------------------------------------------------------
// GenerateDesign
// ---------------------------------------------------------------------------

func (s *sDesign) GenerateDesign(ctx context.Context, projectID, requirementID string) (*DesignGenerationResult, error) {
	projectID = strings.TrimSpace(projectID)
	requirementID = strings.TrimSpace(requirementID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	if requirementID == "" {
		return nil, gerror.New("requirement id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	// 1. Verify the project exists.
	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	// 2. Verify the requirement exists and is confirmed.
	requirement, err := getRequirementByID(ctx, db, requirementID)
	if err != nil {
		return nil, err
	}
	if requirement.ProjectId != projectID {
		return nil, gerror.New("requirement does not belong to the specified project")
	}
	if requirement.UserConfirmed != 1 {
		return nil, gerror.New("requirement must be confirmed before generating a design")
	}

	// 3. Determine the next version number.
	nextVersion, err := getNextDesignVersion(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	// 4. Insert a new solution_designs record with status=designing.
	designID := newResourceID("design")
	now := nowText()
	_, err = db.ExecContext(ctx,
		`INSERT INTO solution_designs (id, project_id, requirement_id, version, status, architecture, modules_json, data_models_json, pages_json, task_drafts_json, user_confirmed, confirmed_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		designID, projectID, requirementID, nextVersion, "designing",
		"", "[]", "[]", "[]", "[]",
		0, "", now, now,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "insert solution design failed")
	}

	// 5. Call Central Brain to generate the solution design.
	designInput := braincontracts.SolutionDesignInput{
		ProjectID:          project.Id,
		GoalSummary:        project.GoalSummary,
		RequirementID:      requirement.Id,
		RequirementDocJSON: requirement.RequirementDocJson,
		Instruction:        "根据需求文档生成完整技术方案，包括架构设计、模块划分、数据模型、页面设计和任务草案。",
	}

	_, result, err := EasyMVPBrain().CallSolutionDesign(ctx, designInput)
	if err != nil {
		g.Log().Warningf(ctx, "solution design brain call failed: %v", err)
		// Update status to draft with empty content on brain failure (still persisted for retry).
		_, updateErr := db.ExecContext(ctx,
			`UPDATE solution_designs SET status = ?, updated_at = ? WHERE id = ?`,
			"draft", nowText(), designID,
		)
		if updateErr != nil {
			g.Log().Errorf(ctx, "update solution design status after brain failure: %v", updateErr)
		}
		if auditErr := insertAuditLog(ctx, projectID, "design.generate.brain_failed", "system:brain", "Solution design brain call failed", map[string]any{
			"design_id": designID,
			"error":     err.Error(),
		}); auditErr != nil {
			g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
		}
		return &DesignGenerationResult{
			DesignID: designID,
			Status:   "draft",
		}, nil
	}

	// 6. Store the brain result into the solution design record.
	architecture := strings.TrimSpace(result.Architecture)
	modulesJSON := normalizeJSONField(result.ModulesJSON, "[]")
	dataModelsJSON := normalizeJSONField(result.DataModelsJSON, "[]")
	pagesJSON := normalizeJSONField(result.PagesJSON, "[]")
	taskDraftsJSON := normalizeJSONField(result.TaskDraftsJSON, "[]")

	_, err = db.ExecContext(ctx,
		`UPDATE solution_designs SET status = ?, architecture = ?, modules_json = ?, data_models_json = ?, pages_json = ?, task_drafts_json = ?, updated_at = ? WHERE id = ?`,
		"draft", architecture, modulesJSON, dataModelsJSON, pagesJSON, taskDraftsJSON, nowText(), designID,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "update solution design with brain result failed")
	}

	// 7. Write audit log.
	if auditErr := insertAuditLog(ctx, projectID, "design.generated", "system:brain", "Solution design generated", map[string]any{
		"design_id": designID,
		"version":   nextVersion,
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}

	return &DesignGenerationResult{
		DesignID: designID,
		Status:   "draft",
	}, nil
}

// ---------------------------------------------------------------------------
// ConfirmDesign
// ---------------------------------------------------------------------------

func (s *sDesign) ConfirmDesign(ctx context.Context, designID string) error {
	designID = strings.TrimSpace(designID)
	if designID == "" {
		return gerror.New("design id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	design, err := getDesignByID(ctx, db, designID)
	if err != nil {
		return err
	}
	if design.Status != "draft" {
		return gerror.Newf("cannot confirm design in status '%s'; expected 'draft'", design.Status)
	}

	now := nowText()
	_, err = db.ExecContext(ctx,
		`UPDATE solution_designs SET status = ?, user_confirmed = ?, confirmed_at = ?, updated_at = ? WHERE id = ?`,
		"approved", 1, now, now, designID,
	)
	if err != nil {
		return gerror.Wrap(err, "confirm solution design failed")
	}

	if auditErr := insertAuditLog(ctx, design.ProjectId, "design.confirmed", "user:local_operator", "Solution design confirmed", map[string]any{
		"design_id": designID,
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	return nil
}

// ---------------------------------------------------------------------------
// RejectDesign
// ---------------------------------------------------------------------------

func (s *sDesign) RejectDesign(ctx context.Context, designID string, reason string) error {
	designID = strings.TrimSpace(designID)
	if designID == "" {
		return gerror.New("design id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	design, err := getDesignByID(ctx, db, designID)
	if err != nil {
		return err
	}
	if design.Status != "draft" {
		return gerror.Newf("cannot reject design in status '%s'; expected 'draft'", design.Status)
	}

	now := nowText()
	_, err = db.ExecContext(ctx,
		`UPDATE solution_designs SET status = ?, updated_at = ? WHERE id = ?`,
		"rejected", now, designID,
	)
	if err != nil {
		return gerror.Wrap(err, "reject solution design failed")
	}

	if auditErr := insertAuditLog(ctx, design.ProjectId, "design.rejected", "user:local_operator", "Solution design rejected", map[string]any{
		"design_id": designID,
		"reason":    strings.TrimSpace(reason),
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	return nil
}

// ---------------------------------------------------------------------------
// GetDesign
// ---------------------------------------------------------------------------

func (s *sDesign) GetDesign(ctx context.Context, designID string) (*entity.SolutionDesigns, error) {
	designID = strings.TrimSpace(designID)
	if designID == "" {
		return nil, gerror.New("design id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	return getDesignByID(ctx, db, designID)
}

// ---------------------------------------------------------------------------
// GetProjectLatestDesign
// ---------------------------------------------------------------------------

func (s *sDesign) GetProjectLatestDesign(ctx context.Context, projectID string) (*entity.SolutionDesigns, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	return getLatestDesignByProjectID(ctx, db, projectID)
}

// ---------------------------------------------------------------------------
// DB helpers (package-private)
// ---------------------------------------------------------------------------

func getRequirementByID(ctx context.Context, db *sql.DB, requirementID string) (*entity.Requirements, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, project_id, raw_input, status, requirement_doc_json, user_confirmed, confirmed_at, created_at, updated_at FROM requirements WHERE id = ? LIMIT 1`,
		requirementID,
	)
	var r entity.Requirements
	if err := row.Scan(&r.Id, &r.ProjectId, &r.RawInput, &r.Status, &r.RequirementDocJson, &r.UserConfirmed, &r.ConfirmedAt, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.Newf("requirement not found: %s", requirementID)
		}
		return nil, gerror.Wrap(err, "query requirement failed")
	}
	return &r, nil
}

func getDesignByID(ctx context.Context, db *sql.DB, designID string) (*entity.SolutionDesigns, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, project_id, requirement_id, version, status, architecture, modules_json, data_models_json, pages_json, task_drafts_json, user_confirmed, confirmed_at, created_at, updated_at FROM solution_designs WHERE id = ? LIMIT 1`,
		designID,
	)
	var d entity.SolutionDesigns
	if err := row.Scan(&d.Id, &d.ProjectId, &d.RequirementId, &d.Version, &d.Status, &d.Architecture, &d.ModulesJson, &d.DataModelsJson, &d.PagesJson, &d.TaskDraftsJson, &d.UserConfirmed, &d.ConfirmedAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.Newf("solution design not found: %s", designID)
		}
		return nil, gerror.Wrap(err, "query solution design failed")
	}
	return &d, nil
}

func getLatestDesignByProjectID(ctx context.Context, db *sql.DB, projectID string) (*entity.SolutionDesigns, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, project_id, requirement_id, version, status, architecture, modules_json, data_models_json, pages_json, task_drafts_json, user_confirmed, confirmed_at, created_at, updated_at FROM solution_designs WHERE project_id = ? ORDER BY version DESC LIMIT 1`,
		projectID,
	)
	var d entity.SolutionDesigns
	if err := row.Scan(&d.Id, &d.ProjectId, &d.RequirementId, &d.Version, &d.Status, &d.Architecture, &d.ModulesJson, &d.DataModelsJson, &d.PagesJson, &d.TaskDraftsJson, &d.UserConfirmed, &d.ConfirmedAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, gerror.Wrap(err, "query latest solution design failed")
	}
	return &d, nil
}

func getNextDesignVersion(ctx context.Context, db *sql.DB, projectID string) (int, error) {
	row := db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(version), 0) FROM solution_designs WHERE project_id = ?`,
		projectID,
	)
	var maxVersion int
	if err := row.Scan(&maxVersion); err != nil {
		return 0, gerror.Wrap(err, "query max design version failed")
	}
	return maxVersion + 1, nil
}

// normalizeJSONField returns the value if non-empty, otherwise the fallback.
func normalizeJSONField(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
