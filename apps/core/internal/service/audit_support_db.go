package service

import (
	"context"

	auditv1 "github.com/leef-l/easymvp/apps/core/api/audit/v1"
)

func listProjectAuditLogsView(ctx context.Context, projectID string, limit int) (*auditv1.ListProjectAuditLogsRes, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	rows, err := listProjectAuditLogs(ctx, db, projectID, limit)
	if err != nil {
		return nil, err
	}

	items := make([]auditv1.AuditLogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, auditv1.AuditLogItem{
			ID:          row.Id,
			ProjectID:   row.ProjectId,
			EventType:   row.EventType,
			ActorKind:   row.ActorKind,
			Summary:     row.Summary,
			PayloadJSON: row.PayloadJson,
			CreatedAt:   row.CreatedAt,
		})
	}
	return &auditv1.ListProjectAuditLogsRes{
		Items:       items,
		RefreshHint: "refresh_audit_logs",
	}, nil
}
