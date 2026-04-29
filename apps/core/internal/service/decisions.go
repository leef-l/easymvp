package service

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	decisionsv1 "github.com/leef-l/easymvp/apps/core/api/decisions/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

type ApplyManualDecisionCommand struct {
	ProjectID  string
	TargetKind string
	TargetID   string
	Decision   string
	Reason     string
	Comment    string
}

type IDecisions interface {
	ApplyManualDecision(ctx context.Context, req ApplyManualDecisionCommand) (res *decisionsv1.ApplyManualDecisionRes, err error)
}

var localDecisions IDecisions = (*sDecisions)(nil)

type sDecisions struct{}

func Decisions() IDecisions {
	if localDecisions == nil {
		localDecisions = &sDecisions{}
	}
	return localDecisions
}

func (s *sDecisions) ApplyManualDecision(ctx context.Context, req ApplyManualDecisionCommand) (res *decisionsv1.ApplyManualDecisionRes, err error) {
	if req.ProjectID == "" {
		return nil, gerror.New("project_id is required")
	}
	if req.TargetKind == "" {
		return nil, gerror.New("target_kind is required")
	}
	if req.TargetID == "" {
		return nil, gerror.New("target_id is required")
	}
	if req.Decision == "" {
		return nil, gerror.New("decision is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	now := nowText()
	commandID := newResourceID("cmd")

	// Validate target exists based on kind
	switch strings.ToLower(req.TargetKind) {
	case "plan":
		if _, err := getPlanDraftForProject(ctx, entity.Projects{Id: req.ProjectID}); err != nil {
			return nil, gerror.Wrap(err, "target plan not found")
		}
	case "task":
		if _, err := getDomainTaskByID(ctx, db, req.TargetID); err != nil {
			return nil, gerror.Wrap(err, "target task not found")
		}
	case "acceptance_run":
		if _, err := getAcceptanceRunByID(ctx, req.TargetID); err != nil {
			return nil, gerror.Wrap(err, "target acceptance run not found")
		}
	default:
		// generic — no validation for unknown kinds
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, gerror.Wrap(err, "begin manual decision transaction failed")
	}

	// Insert manual decision record into task_manual_gates as a generic gate record
	gateID := newResourceID("gate")
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO `+dao.TaskManualGates.Table()+` (id, project_id, task_id, gate_kind, gate_status, comment, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		gateID,
		req.ProjectID,
		req.TargetID,
		req.TargetKind+":"+req.Decision,
		req.Decision,
		firstNonEmpty(req.Comment, req.Reason),
		now,
		now,
	); err != nil {
		_ = tx.Rollback()
		return nil, gerror.Wrap(err, "insert manual decision record failed")
	}

	// Write audit log
	if err := insertAuditLogSqlTx(ctx, tx, req.ProjectID, "manual_decision.applied", "user:local_operator",
		"Manual decision applied: "+req.Decision+" on "+req.TargetKind,
		map[string]any{
			"target_kind": req.TargetKind,
			"target_id":   req.TargetID,
			"decision":    req.Decision,
			"reason":      req.Reason,
			"comment":     req.Comment,
		}); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, gerror.Wrap(err, "commit manual decision failed")
	}

	return &decisionsv1.ApplyManualDecisionRes{
		CommandID:  commandID,
		Accepted:   true,
		ResourceID: gateID,
		NextAction: "refresh_workspace_view",
	}, nil
}
