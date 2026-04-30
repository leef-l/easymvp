package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// ---------------------------------------------------------------------------
// Result types
// ---------------------------------------------------------------------------

// DeliveryResult is the result returned after preparing a project delivery.
type DeliveryResult struct {
	DeliveryID string
	Status     string
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

// IDelivery defines the service interface for project delivery operations.
type IDelivery interface {
	PrepareDelivery(ctx context.Context, projectID string) (*DeliveryResult, error)
	AcceptDelivery(ctx context.Context, deliveryID string) error
	RejectDelivery(ctx context.Context, deliveryID string, reason string) error
	GetDelivery(ctx context.Context, deliveryID string) (*entity.ProjectDeliveries, error)
}

// ---------------------------------------------------------------------------
// Singleton registration (GoFrame pattern)
// ---------------------------------------------------------------------------

var localDelivery IDelivery = (*sDelivery)(nil)

type sDelivery struct{}

func Delivery() IDelivery {
	if localDelivery == nil {
		localDelivery = &sDelivery{}
	}
	return localDelivery
}

// ---------------------------------------------------------------------------
// PrepareDelivery
// ---------------------------------------------------------------------------

func (s *sDelivery) PrepareDelivery(ctx context.Context, projectID string) (*DeliveryResult, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
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

	// 2. Gather statistics from domain_tasks.
	stats, err := gatherDeliveryStatistics(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	statsJSON, _ := json.Marshal(stats)

	// 3. Generate test report summary.
	testReport := map[string]any{
		"total_tasks":     stats["total_tasks"],
		"completed_tasks": stats["completed_tasks"],
		"failed_tasks":    stats["failed_tasks"],
	}
	testReportJSON, _ := json.Marshal(testReport)

	// 4. Insert a new delivery record.
	deliveryID := newResourceID("delivery")
	now := nowText()
	_, err = db.ExecContext(ctx,
		`INSERT INTO project_deliveries (id, project_id, status, workspace_path, readme, architecture_doc, api_docs, deployment_doc, test_report_json, statistics_json, user_accepted, accepted_at, delivered_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		deliveryID, projectID, "prepared",
		project.WorkspaceRoot,
		"", "", "", "",
		string(testReportJSON), string(statsJSON),
		0, "", now, now, now,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "insert project delivery failed")
	}

	// 5. Write audit log.
	if auditErr := insertAuditLog(ctx, projectID, "delivery.prepared", "system", "Project delivery prepared", map[string]any{
		"delivery_id": deliveryID,
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}

	return &DeliveryResult{
		DeliveryID: deliveryID,
		Status:     "prepared",
	}, nil
}

// ---------------------------------------------------------------------------
// AcceptDelivery
// ---------------------------------------------------------------------------

func (s *sDelivery) AcceptDelivery(ctx context.Context, deliveryID string) error {
	deliveryID = strings.TrimSpace(deliveryID)
	if deliveryID == "" {
		return gerror.New("delivery id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	delivery, err := getDeliveryByID(ctx, db, deliveryID)
	if err != nil {
		return err
	}
	if delivery.Status != "prepared" {
		return gerror.Newf("cannot accept delivery in status '%s'; expected 'prepared'", delivery.Status)
	}

	now := nowText()
	_, err = db.ExecContext(ctx,
		`UPDATE project_deliveries SET status = ?, user_accepted = ?, accepted_at = ?, updated_at = ? WHERE id = ?`,
		"accepted", 1, now, now, deliveryID,
	)
	if err != nil {
		return gerror.Wrap(err, "accept project delivery failed")
	}

	if auditErr := insertAuditLog(ctx, delivery.ProjectId, "delivery.accepted", "user:local_operator", "Project delivery accepted", map[string]any{
		"delivery_id": deliveryID,
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	return nil
}

// ---------------------------------------------------------------------------
// RejectDelivery
// ---------------------------------------------------------------------------

func (s *sDelivery) RejectDelivery(ctx context.Context, deliveryID string, reason string) error {
	deliveryID = strings.TrimSpace(deliveryID)
	if deliveryID == "" {
		return gerror.New("delivery id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	delivery, err := getDeliveryByID(ctx, db, deliveryID)
	if err != nil {
		return err
	}
	if delivery.Status != "prepared" {
		return gerror.Newf("cannot reject delivery in status '%s'; expected 'prepared'", delivery.Status)
	}

	now := nowText()
	_, err = db.ExecContext(ctx,
		`UPDATE project_deliveries SET status = ?, updated_at = ? WHERE id = ?`,
		"rejected", now, deliveryID,
	)
	if err != nil {
		return gerror.Wrap(err, "reject project delivery failed")
	}

	if auditErr := insertAuditLog(ctx, delivery.ProjectId, "delivery.rejected", "user:local_operator", "Project delivery rejected", map[string]any{
		"delivery_id": deliveryID,
		"reason":      strings.TrimSpace(reason),
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}
	return nil
}

// ---------------------------------------------------------------------------
// GetDelivery
// ---------------------------------------------------------------------------

func (s *sDelivery) GetDelivery(ctx context.Context, deliveryID string) (*entity.ProjectDeliveries, error) {
	deliveryID = strings.TrimSpace(deliveryID)
	if deliveryID == "" {
		return nil, gerror.New("delivery id is required")
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	return getDeliveryByID(ctx, db, deliveryID)
}

// ---------------------------------------------------------------------------
// DB helpers (package-private)
// ---------------------------------------------------------------------------

func getDeliveryByID(ctx context.Context, db *sql.DB, deliveryID string) (*entity.ProjectDeliveries, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, project_id, status, workspace_path, readme, architecture_doc, api_docs, deployment_doc, test_report_json, statistics_json, user_accepted, accepted_at, delivered_at, created_at, updated_at FROM project_deliveries WHERE id = ? LIMIT 1`,
		deliveryID,
	)
	var d entity.ProjectDeliveries
	if err := row.Scan(&d.Id, &d.ProjectId, &d.Status, &d.WorkspacePath, &d.Readme, &d.ArchitectureDoc, &d.ApiDocs, &d.DeploymentDoc, &d.TestReportJson, &d.StatisticsJson, &d.UserAccepted, &d.AcceptedAt, &d.DeliveredAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.Newf("project delivery not found: %s", deliveryID)
		}
		return nil, gerror.Wrap(err, "query project delivery failed")
	}
	return &d, nil
}

func gatherDeliveryStatistics(ctx context.Context, db *sql.DB, projectID string) (map[string]any, error) {
	stats := map[string]any{
		"total_tasks":     0,
		"completed_tasks": 0,
		"failed_tasks":    0,
	}

	// Count total tasks.
	row := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM domain_tasks WHERE project_id = ?`, projectID)
	var total int
	if err := row.Scan(&total); err != nil && err != sql.ErrNoRows {
		g.Log().Warningf(ctx, "count domain tasks failed: %v", err)
	}
	stats["total_tasks"] = total

	// Count completed tasks.
	row = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM domain_tasks WHERE project_id = ? AND status = 'done'`, projectID)
	var completed int
	if err := row.Scan(&completed); err != nil && err != sql.ErrNoRows {
		g.Log().Warningf(ctx, "count completed domain tasks failed: %v", err)
	}
	stats["completed_tasks"] = completed

	// Count failed tasks.
	row = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM domain_tasks WHERE project_id = ? AND status = 'failed'`, projectID)
	var failed int
	if err := row.Scan(&failed); err != nil && err != sql.ErrNoRows {
		g.Log().Warningf(ctx, "count failed domain tasks failed: %v", err)
	}
	stats["failed_tasks"] = failed

	return stats, nil
}
