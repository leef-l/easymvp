package service

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func TestQueryPlanBaselineUsesExpectedIndexes(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		seedAcceptanceFlowProject(t, ctx, db, acceptanceFlowSeed{
			projectID:             "proj_query_plan",
			taskID:                "task_query_plan",
			runID:                 "run_query_plan",
			projectStatus:         "acceptance",
			projectProduction:     "pending",
			runStatus:             "running",
			functionalStatus:      "pending",
			productionStatus:      "pending",
			manualReleaseRequired: 0,
		})

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin seed tx failed: %v", err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO audit_logs (id, project_id, event_type, actor_kind, summary, payload_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			"audit_query_plan", "proj_query_plan", "runtime.synced", "system", "synced", `{"ok":true}`, "2026-04-20T10:01:00Z"); err != nil {
			_ = tx.Rollback()
			t.Fatalf("insert audit log failed: %v", err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO evidence_items (id, project_id, run_id, surface, journey, evidence_type, file_path, content_hash, file_size, captured_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"evidence_query_plan", "proj_query_plan", "run_query_plan", "desktop", "smoke", "screenshot", "/tmp/evidence.png", "hash", 12, "2026-04-20T10:01:10Z", "2026-04-20T10:01:10Z"); err != nil {
			_ = tx.Rollback()
			t.Fatalf("insert evidence item failed: %v", err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO workflow_replay_index (
id, replay_id, project_id, run_id, domain_task_id, compiled_task_id, event_id, trace_id, span_id, replay_kind, seq_no, title, summary, file_path, file_ext, mime_type, file_size, sha256, source_object_kind, source_object_id, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"replay_row_query_plan", "replay_query_plan", "proj_query_plan", "run_query_plan", "task_query_plan", "", "event_1", "trace_1", "span_1", "step_snapshot", 1, "step", "summary", "/tmp/replay.json", ".json", "application/json", 32, "sha", "domain_task", "task_query_plan", "available", "2026-04-20T10:01:20Z", "2026-04-20T10:01:20Z"); err != nil {
			_ = tx.Rollback()
			t.Fatalf("insert replay index failed: %v", err)
		}
		if err = tx.Commit(); err != nil {
			t.Fatalf("commit seed tx failed: %v", err)
		}

		assertQueryPlanUsesIndex(t, db,
			`SELECT id FROM projects ORDER BY updated_at DESC, created_at DESC LIMIT ?`,
			[]any{12},
			"idx_projects_updated_created",
		)
		assertQueryPlanUsesIndex(t, db,
			`SELECT id FROM acceptance_runs WHERE project_id = ? ORDER BY created_at DESC LIMIT 1`,
			[]any{"proj_query_plan"},
			"idx_acceptance_runs_project_created",
		)
		assertQueryPlanUsesIndex(t, db,
			`SELECT id FROM audit_logs WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`,
			[]any{"proj_query_plan", 20},
			"idx_audit_project_time",
		)
		assertQueryPlanUsesIndex(t, db,
			`SELECT id FROM evidence_items WHERE project_id = ? ORDER BY COALESCE(captured_at, created_at) DESC, created_at DESC LIMIT ?`,
			[]any{"proj_query_plan", 12},
			"idx_evidence_project_effective_captured",
		)
		assertQueryPlanUsesIndex(t, db,
			`SELECT id FROM workflow_replay_index WHERE project_id = ? AND run_id = ? ORDER BY seq_no ASC, created_at ASC, id ASC LIMIT ?`,
			[]any{"proj_query_plan", "run_query_plan", 10},
			"idx_workflow_replay_index_run_seq",
		)
	})
}

func assertQueryPlanUsesIndex(t *testing.T, db *sql.DB, query string, args []any, indexName string) {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), "EXPLAIN QUERY PLAN "+query, args...)
	if err != nil {
		t.Fatalf("explain query plan failed for %s: %v", indexName, err)
	}
	defer rows.Close()

	details := make([]string, 0, 4)
	for rows.Next() {
		var (
			id      int
			parent  int
			notused int
			detail  string
		)
		if err = rows.Scan(&id, &parent, &notused, &detail); err != nil {
			t.Fatalf("scan query plan failed for %s: %v", indexName, err)
		}
		details = append(details, detail)
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("iterate query plan failed for %s: %v", indexName, err)
	}

	joined := strings.Join(details, "\n")
	if !strings.Contains(joined, indexName) {
		t.Fatalf("expected query plan to use %s, got:\n%s", indexName, joined)
	}
}
