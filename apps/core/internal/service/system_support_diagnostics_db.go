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
	categoryCounts := make(map[string]int)
	for _, row := range rows {
		item := buildProjectDiagnosticItem(row)
		if !diagnosticMatchesProject(item, projectID) {
			continue
		}
		if item.Category != "" {
			categoryCounts[item.Category] += 1
		}
		items = append(items, item)
		if len(items) >= limit {
			break
		}
	}

	return &systemv1.ListProjectDiagnosticsRes{
		Items:          items,
		CategoryCounts: categoryCounts,
		RefreshHint:    "refresh_project_diagnostics",
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

	item.Category, item.Component, item.Field, item.RecommendedAction, item.RelatedPage = classifyDiagnosticRecord(row, nil)

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
	item.Category, item.Component, item.Field, item.RecommendedAction, item.RelatedPage = classifyDiagnosticRecord(row, detail)
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

func classifyDiagnosticRecord(row entity.DiagnosticRecords, detail map[string]any) (category string, component string, field string, recommendedAction string, relatedPage string) {
	code := strings.ToLower(strings.TrimSpace(row.ErrorCode))
	scope := strings.ToLower(strings.TrimSpace(row.Scope))
	summary := strings.ToLower(strings.TrimSpace(row.Summary))
	raw := strings.ToLower(strings.TrimSpace(row.DetailJson))
	component = strings.TrimSpace(extractDiagnosticDetailString(detail, "component"))
	field = strings.TrimSpace(extractDiagnosticDetailString(detail, "field"))

	if matchDiagnosticSignal(code, scope, summary, raw, "migration") {
		return "migration_failure", fallbackString(component, "startup"), fallbackString(field, "migration_path"), "inspect_migration_and_relaunch", "recovery"
	}
	if matchDiagnosticSignal(code, scope, summary, raw, "directory is not writable", "not a directory", "data_root", "db_path", "permission denied", "read-only file system") {
		return "data_directory_unwritable", fallbackString(component, "startup"), fallbackString(field, "data_root"), "open_data_folder_and_fix_permissions", "recovery"
	}
	if matchDiagnosticSignal(code, scope, summary, raw, "runtime unavailable", "brain serve", "brain runtime health", "econnrefused", "connection refused") {
		return "core_unavailable", fallbackString(component, "runtime"), fallbackString(field, "brain_serve_base_url"), "verify_runtime_endpoint_or_restart_core", "diagnostics"
	}
	if matchDiagnosticSignal(code, scope, summary, raw, "policy_denied", "permission_denied", "tool_denied", "forbidden", "run_denied") {
		return "policy_denied", fallbackString(component, "runtime"), fallbackString(field, "policy"), "open_workspace_and_review_policy", "workspace"
	}
	if matchDiagnosticSignal(code, scope, summary, raw, "verification_conflict", "failed_checks", "missing_evidence", "verification_contract") {
		return "verification_conflict", fallbackString(component, "acceptance"), fallbackString(field, "verification_contract"), "open_acceptance_and_review_contract", "acceptance"
	}
	if matchDiagnosticSignal(code, scope, summary, raw, "fault_loop_detected", "review_fault_loop", "fault loop") {
		return "fault_loop_detected", fallbackString(component, "acceptance"), fallbackString(field, "fault_summary"), "open_repair_flow_before_retry", "workspace"
	}
	if matchDiagnosticSignal(code, scope, summary, raw, "artifact", "replay", "log segment", "raw_target", "file missing") {
		return "artifact_index_gap", fallbackString(component, "replay"), fallbackString(field, "artifact"), "open_replay_and_reconcile_artifacts", "replay"
	}
	if matchDiagnosticSignal(code, scope, summary, raw, "audit") {
		return "audit_attention", fallbackString(component, "audit"), field, "open_audit_and_review_event", "audit"
	}
	return "runtime_attention", fallbackString(component, inferDiagnosticComponent(scope)), field, inferDiagnosticAction(scope), inferDiagnosticPage(scope)
}

func matchDiagnosticSignal(values ...string) bool {
	if len(values) < 5 {
		return false
	}
	haystack := values[:4]
	signals := values[4:]
	for _, source := range haystack {
		for _, signal := range signals {
			if signal != "" && strings.Contains(source, signal) {
				return true
			}
		}
	}
	return false
}

func inferDiagnosticComponent(scope string) string {
	switch {
	case strings.Contains(scope, "runtime"):
		return "runtime"
	case strings.Contains(scope, "audit"):
		return "audit"
	case strings.Contains(scope, "replay"):
		return "replay"
	case strings.Contains(scope, "acceptance"):
		return "acceptance"
	case strings.Contains(scope, "worker"):
		return "worker"
	default:
		return "system"
	}
}

func inferDiagnosticAction(scope string) string {
	switch inferDiagnosticComponent(scope) {
	case "audit":
		return "open_audit_and_review_event"
	case "replay":
		return "open_replay_and_review_artifacts"
	case "acceptance":
		return "open_acceptance_and_review_blockers"
	case "runtime":
		return "open_execution_and_retry_sync"
	default:
		return "open_diagnostics_and_export_snapshot"
	}
}

func inferDiagnosticPage(scope string) string {
	switch inferDiagnosticComponent(scope) {
	case "audit":
		return "audit"
	case "replay":
		return "replay"
	case "acceptance":
		return "acceptance"
	case "runtime":
		return "execution"
	default:
		return "diagnostics"
	}
}

func fallbackString(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(fallback)
}
