package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	systemv1 "github.com/leef-l/easymvp/apps/core/api/system/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func listProjectDiagnosticsView(ctx context.Context, projectID string, limit int) (*systemv1.ListProjectDiagnosticsRes, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	rows, err := listRecentDiagnosticRecords(ctx, db, diagnosticCandidateLimit(limit))
	if err != nil {
		return nil, err
	}

	items := make([]systemv1.ProjectDiagnosticItem, 0, limit)
	for _, row := range rows {
		item := buildProjectDiagnosticItem(row)
		if !diagnosticMatchesProject(item, projectID) {
			continue
		}
		items = append(items, item)
		if len(items) >= limit {
			break
		}
	}

	return &systemv1.ListProjectDiagnosticsRes{
		Items:       items,
		RefreshHint: "refresh_project_diagnostics",
	}, nil
}

func diagnosticCandidateLimit(limit int) int {
	switch {
	case limit <= 0:
		return 120
	case limit < 20:
		return 120
	case limit > 100:
		return 400
	default:
		return limit * 6
	}
}

func listRecentDiagnosticRecords(ctx context.Context, db *sql.DB, limit int) ([]entity.DiagnosticRecords, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT id, scope, severity, error_code, summary, COALESCE(detail_json, ''), created_at
FROM `+dao.DiagnosticRecords.Table()+`
ORDER BY created_at DESC
LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "query diagnostic records failed")
	}
	defer rows.Close()

	items := make([]entity.DiagnosticRecords, 0, limit)
	for rows.Next() {
		var item entity.DiagnosticRecords
		if err = rows.Scan(
			&item.Id,
			&item.Scope,
			&item.Severity,
			&item.ErrorCode,
			&item.Summary,
			&item.DetailJson,
			&item.CreatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan diagnostic record failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate diagnostic records failed")
	}
	return items, nil
}

func buildProjectDiagnosticItem(row entity.DiagnosticRecords) systemv1.ProjectDiagnosticItem {
	item := systemv1.ProjectDiagnosticItem{
		ID:         row.Id,
		Scope:      row.Scope,
		Severity:   row.Severity,
		ErrorCode:  row.ErrorCode,
		Summary:    row.Summary,
		DetailJSON: row.DetailJson,
		CreatedAt:  row.CreatedAt,
	}

	if strings.TrimSpace(row.DetailJson) == "" {
		return item
	}

	var detail map[string]any
	if err := json.Unmarshal([]byte(row.DetailJson), &detail); err != nil {
		return item
	}
	item.ProjectID = extractDiagnosticDetailString(detail, "project_id")
	item.TaskID = extractDiagnosticDetailString(detail, "task_id")
	item.RunID = extractDiagnosticDetailString(detail, "run_id")
	item.BindingID = extractDiagnosticDetailString(detail, "binding_id")
	return item
}

func diagnosticMatchesProject(item systemv1.ProjectDiagnosticItem, projectID string) bool {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(item.ProjectID), projectID) {
		return true
	}
	return strings.Contains(strings.ToLower(item.DetailJSON), strings.ToLower(projectID))
}

func extractDiagnosticDetailString(detail map[string]any, key string) string {
	value, ok := detail[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}
