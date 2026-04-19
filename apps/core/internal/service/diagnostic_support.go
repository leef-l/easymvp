package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
)

func recordDiagnostic(ctx context.Context, scope string, severity string, errorCode string, summary string, detail map[string]any) {
	scope = strings.TrimSpace(scope)
	severity = strings.TrimSpace(severity)
	errorCode = strings.TrimSpace(errorCode)
	summary = strings.TrimSpace(summary)
	if scope == "" {
		scope = "unknown"
	}
	if severity == "" {
		severity = "warning"
	}
	if errorCode == "" {
		errorCode = "DIAG_001"
	}
	if summary == "" {
		summary = "diagnostic event"
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return
	}
	defer closeFn()

	_ = insertDiagnosticRecord(ctx, db, scope, severity, errorCode, summary, detail)
}

func insertDiagnosticRecord(ctx context.Context, db *sql.DB, scope string, severity string, errorCode string, summary string, detail map[string]any) error {
	result, err := db.ExecContext(
		ctx,
		`INSERT INTO `+dao.DiagnosticRecords.Table()+` (
id, scope, severity, error_code, summary, detail_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		newResourceID("diag"),
		scope,
		severity,
		errorCode,
		summary,
		marshalDiagnosticDetail(detail),
		nowText(),
	)
	if err != nil {
		return gerror.Wrap(err, "insert diagnostic record failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert diagnostic record affected unexpected rows")
	}
	return nil
}

func marshalDiagnosticDetail(detail map[string]any) any {
	if len(detail) == 0 {
		return nil
	}
	encoded, err := json.Marshal(detail)
	if err != nil {
		return `{"marshal_error":"diagnostic detail encode failed"}`
	}
	return string(encoded)
}
