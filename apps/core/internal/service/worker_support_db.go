package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/repo"
)

func listPendingRunBindingIDs(ctx context.Context, limit int) ([]string, error) {
	return repo.ListPendingRunBindingIDs(ctx, limit)
}

func listProjectsForWorkspaceRefresh(ctx context.Context, limit int) ([]string, error) {
	return repo.ListProjectsForWorkspaceRefresh(ctx, limit)
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
	return repo.InsertWorkerAuditLog(ctx, workerName, projectID, summary, detail)
}
