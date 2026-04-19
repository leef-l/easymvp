package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
)

func listPendingRunBindingIDs(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	rows, err := db.QueryContext(
		ctx,
		`SELECT id
FROM `+dao.BrainRunBindings.Table()+`
WHERE run_status IN (?, ?)
ORDER BY updated_at ASC
LIMIT ?`,
		"run_pending",
		"run_active",
		limit,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "query pending run bindings failed")
	}
	defer rows.Close()

	ids := make([]string, 0, limit)
	for rows.Next() {
		var bindingID string
		if err = rows.Scan(&bindingID); err != nil {
			return nil, gerror.Wrap(err, "scan pending run binding failed")
		}
		if bindingID != "" {
			ids = append(ids, bindingID)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate pending run bindings failed")
	}
	return ids, nil
}

func listProjectsForWorkspaceRefresh(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 12
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	rows, err := db.QueryContext(
		ctx,
		`SELECT id
FROM `+dao.Projects.Table()+`
WHERE status <> ?
ORDER BY updated_at DESC
LIMIT ?`,
		"completed",
		limit,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "query workspace refresh projects failed")
	}
	defer rows.Close()

	projectIDs := make([]string, 0, limit)
	for rows.Next() {
		var projectID string
		if err = rows.Scan(&projectID); err != nil {
			return nil, gerror.Wrap(err, "scan workspace refresh project failed")
		}
		if projectID != "" {
			projectIDs = append(projectIDs, projectID)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate workspace refresh projects failed")
	}
	return projectIDs, nil
}

func handleWorkerFailure(
	ctx context.Context,
	workerName string,
	projectID string,
	errorCode string,
	summary string,
	detail map[string]any,
) {
	workerName = strings.TrimSpace(workerName)
	projectID = strings.TrimSpace(projectID)
	summary = strings.TrimSpace(summary)
	if workerName == "" {
		workerName = "unknown_worker"
	}
	if summary == "" {
		summary = "worker error"
	}
	if errorCode == "" {
		errorCode = "WORKER_001"
	}

	g.Log().Errorf(ctx, "[worker:%s] %s", workerName, summary)

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "[worker:%s] open database failed: %v", workerName, err)
		return
	}
	defer closeFn()

	if err = insertWorkerDiagnostic(ctx, db, workerName, errorCode, summary, detail); err != nil {
		g.Log().Errorf(ctx, "[worker:%s] write diagnostic failed: %v", workerName, err)
	}
	if projectID != "" {
		if err = insertWorkerAuditLog(ctx, db, workerName, projectID, summary, detail); err != nil {
			g.Log().Errorf(ctx, "[worker:%s] write audit log failed: %v", workerName, err)
		}
	}
}

func insertWorkerDiagnostic(ctx context.Context, db *sql.DB, workerName, errorCode, summary string, detail map[string]any) error {
	return insertDiagnosticRecord(ctx, db, "worker:"+workerName, "warning", errorCode, summary, detail)
}

func insertWorkerAuditLog(ctx context.Context, db *sql.DB, workerName, projectID, summary string, detail map[string]any) error {
	result, err := db.ExecContext(
		ctx,
		`INSERT INTO `+dao.AuditLogs.Table()+` (
id, project_id, event_type, actor_kind, summary, payload_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		newResourceID("audit"),
		projectID,
		"worker.error",
		"worker:"+workerName,
		summary,
		marshalWorkerDetail(detail),
		nowText(),
	)
	if err != nil {
		return gerror.Wrap(err, "insert worker audit log failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert worker audit log affected unexpected rows")
	}
	return nil
}

func marshalWorkerDetail(detail map[string]any) any {
	return marshalDiagnosticDetail(detail)
}
