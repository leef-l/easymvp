package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// RequirementAnalysisResult is the response returned by AnalyzeRequirement.
type RequirementAnalysisResult struct {
	RequirementID  string `json:"requirement_id"`
	Status         string `json:"status"`
	Summary        string `json:"summary"`
	RequirementDoc string `json:"requirement_doc_json"`
}

// IRequirement defines the service interface for requirement management.
type IRequirement interface {
	// AnalyzeRequirement submits raw natural-language input to the Central Brain
	// and produces a structured requirement document.
	AnalyzeRequirement(ctx context.Context, projectID, rawInput string) (*RequirementAnalysisResult, error)
	// ConfirmRequirement marks a requirement as confirmed by the user.
	ConfirmRequirement(ctx context.Context, requirementID string) error
	// GetRequirement retrieves a single requirement by ID.
	GetRequirement(ctx context.Context, requirementID string) (*entity.Requirements, error)
	// GetProjectRequirement retrieves the latest requirement for a project.
	GetProjectRequirement(ctx context.Context, projectID string) (*entity.Requirements, error)
}

var localRequirement IRequirement = (*sRequirement)(nil)

type sRequirement struct{}

// Requirement returns the singleton IRequirement implementation.
func Requirement() IRequirement {
	if localRequirement == nil {
		localRequirement = &sRequirement{}
	}
	return localRequirement
}

// RegisterRequirement allows replacing the default IRequirement implementation.
func RegisterRequirement(impl IRequirement) {
	localRequirement = impl
}

// ---------------------------------------------------------------------------
// AnalyzeRequirement
// ---------------------------------------------------------------------------

func (s *sRequirement) AnalyzeRequirement(ctx context.Context, projectID, rawInput string) (*RequirementAnalysisResult, error) {
	projectID = strings.TrimSpace(projectID)
	rawInput = strings.TrimSpace(rawInput)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	if rawInput == "" {
		return nil, gerror.New("raw input is required")
	}

	// Open database and load project to get goal_summary.
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	// Insert initial requirement record with status=analyzing.
	reqID := newResourceID("req")
	now := nowText()
	_, err = db.ExecContext(ctx,
		`INSERT INTO `+dao.Requirements.Table()+` (id, project_id, raw_input, status, requirement_doc_json, user_confirmed, confirmed_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		reqID, projectID, rawInput, "analyzing", "{}", 0, "", now, now,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "insert requirement record failed")
	}

	// Call brain to analyze the requirement.
	brainInput := braincontracts.RequirementAnalysisInput{
		ProjectID:   project.Id,
		GoalSummary: project.GoalSummary,
		RawInput:    rawInput,
		Instruction: "请根据用户的自然语言需求描述，生成结构化的需求文档。需求文档应包含：标题、概述、功能需求、非功能需求、用户故事、验收标准、约束条件和假设。请以 JSON 格式返回 requirement_doc 字段。",
	}

	_, brainResult, brainErr := EasyMVPBrain().CallRequirementAnalysis(ctx, brainInput)
	if brainErr != nil {
		g.Log().Warningf(ctx, "requirement analysis brain call failed: %v", brainErr)
		// Fallback: generate a minimal requirement doc from the raw input.
		brainResult = s.fallbackRequirementResult(rawInput, project)
	}

	// Serialize the requirement doc JSON.
	docJSON, err := json.Marshal(brainResult.RequirementDoc)
	if err != nil {
		docJSON = []byte("{}")
	}

	// Update the requirement record: set doc and status=draft.
	now = nowText()
	_, err = db.ExecContext(ctx,
		`UPDATE `+dao.Requirements.Table()+` SET requirement_doc_json = ?, status = ?, updated_at = ? WHERE id = ?`,
		string(docJSON), "draft", now, reqID,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "update requirement with analysis result failed")
	}

	// Write audit log.
	if auditErr := insertAuditLog(ctx, projectID, "requirement.analyzed", "user:local_operator", "Requirement analyzed", map[string]any{
		"requirement_id": reqID,
		"status":         "draft",
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}

	return &RequirementAnalysisResult{
		RequirementID:  reqID,
		Status:         "draft",
		Summary:        brainResult.Summary,
		RequirementDoc: string(docJSON),
	}, nil
}

// ---------------------------------------------------------------------------
// ConfirmRequirement
// ---------------------------------------------------------------------------

func (s *sRequirement) ConfirmRequirement(ctx context.Context, requirementID string) error {
	requirementID = strings.TrimSpace(requirementID)
	if requirementID == "" {
		return gerror.New("requirement id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	// Load existing requirement to validate it exists and is in draft status.
	req, err := s.getRequirementByID(ctx, db, requirementID)
	if err != nil {
		return err
	}
	if req.Status != "draft" {
		return gerror.Newf("requirement status '%s' does not allow confirmation; expected 'draft'", req.Status)
	}

	now := nowText()
	_, err = db.ExecContext(ctx,
		`UPDATE `+dao.Requirements.Table()+` SET status = ?, user_confirmed = ?, confirmed_at = ?, updated_at = ? WHERE id = ?`,
		"confirmed", 1, now, now, requirementID,
	)
	if err != nil {
		return gerror.Wrap(err, "update requirement confirmation failed")
	}

	// Write audit log.
	if auditErr := insertAuditLog(ctx, req.ProjectId, "requirement.confirmed", "user:local_operator", "Requirement confirmed", map[string]any{
		"requirement_id": requirementID,
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}

	return nil
}

// ---------------------------------------------------------------------------
// GetRequirement
// ---------------------------------------------------------------------------

func (s *sRequirement) GetRequirement(ctx context.Context, requirementID string) (*entity.Requirements, error) {
	requirementID = strings.TrimSpace(requirementID)
	if requirementID == "" {
		return nil, gerror.New("requirement id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	return s.getRequirementByID(ctx, db, requirementID)
}

// ---------------------------------------------------------------------------
// GetProjectRequirement
// ---------------------------------------------------------------------------

func (s *sRequirement) GetProjectRequirement(ctx context.Context, projectID string) (*entity.Requirements, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	return s.getLatestRequirementByProjectID(ctx, db, projectID)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (s *sRequirement) getRequirementByID(ctx context.Context, db *sql.DB, requirementID string) (*entity.Requirements, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, project_id, raw_input, status, requirement_doc_json, user_confirmed, confirmed_at, created_at, updated_at FROM `+dao.Requirements.Table()+` WHERE id = ? LIMIT 1`,
		requirementID,
	)
	var req entity.Requirements
	if err := row.Scan(
		&req.Id, &req.ProjectId, &req.RawInput, &req.Status,
		&req.RequirementDocJson, &req.UserConfirmed, &req.ConfirmedAt,
		&req.CreatedAt, &req.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.Newf("requirement not found: %s", requirementID)
		}
		return nil, gerror.Wrap(err, "query requirement by id failed")
	}
	return &req, nil
}

func (s *sRequirement) getLatestRequirementByProjectID(ctx context.Context, db *sql.DB, projectID string) (*entity.Requirements, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, project_id, raw_input, status, requirement_doc_json, user_confirmed, confirmed_at, created_at, updated_at FROM `+dao.Requirements.Table()+` WHERE project_id = ? ORDER BY created_at DESC LIMIT 1`,
		projectID,
	)
	var req entity.Requirements
	if err := row.Scan(
		&req.Id, &req.ProjectId, &req.RawInput, &req.Status,
		&req.RequirementDocJson, &req.UserConfirmed, &req.ConfirmedAt,
		&req.CreatedAt, &req.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.Newf("no requirements found for project: %s", projectID)
		}
		return nil, gerror.Wrap(err, "query latest requirement by project id failed")
	}
	return &req, nil
}

func (s *sRequirement) fallbackRequirementResult(rawInput string, project *entity.Projects) *braincontracts.RequirementAnalysisResult {
	summary := rawInput
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}
	return &braincontracts.RequirementAnalysisResult{
		RequirementDoc: braincontracts.RequirementDoc{
			Title:    project.GoalSummary,
			Overview: rawInput,
			FunctionalReqs: []braincontracts.RequirementItem{
				{
					ID:          "FR-001",
					Description: rawInput,
					Priority:    "must",
				},
			},
			NonFunctionalReqs:  []braincontracts.RequirementItem{},
			UserStories:        []braincontracts.UserStory{},
			AcceptanceCriteria: []braincontracts.AcceptanceCriterion{},
			Constraints:        []string{},
			Assumptions:        []string{},
		},
		Summary:             summary,
		SuggestedNextAction: "confirm_requirement",
	}
}
