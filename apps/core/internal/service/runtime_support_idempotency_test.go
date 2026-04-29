package service

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/leef-l/easymvp/apps/core/internal/model/do"
)

func TestFindReusableRunBindingForTaskReturnsActiveBinding(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		seedRuntimeTask(t, ctx, db, "proj_runtime_reuse", "task_runtime_reuse")
		seedRuntimeBinding(t, ctx, "proj_runtime_reuse", "task_runtime_reuse", "bind_runtime_reuse", "run_runtime_reuse", "run_active")

		binding, err := findReusableRunBindingForTask(ctx, "proj_runtime_reuse", "task_runtime_reuse", "coder")
		if err != nil {
			t.Fatalf("findReusableRunBindingForTask failed: %v", err)
		}
		if binding == nil {
			t.Fatal("expected reusable binding")
		}
		if binding.Id != "bind_runtime_reuse" || binding.BrainRunId != "run_runtime_reuse" {
			t.Fatalf("unexpected reusable binding: %#v", binding)
		}
	})
}

func TestAppendRunEventIndexDeduplicatesRepeatedPayload(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		seedRuntimeTask(t, ctx, db, "proj_runtime_event_dedupe", "task_runtime_event_dedupe")
		seedRuntimeBinding(t, ctx, "proj_runtime_event_dedupe", "task_runtime_event_dedupe", "bind_runtime_event_dedupe", "run_runtime_event_dedupe", "run_active")

		payload := map[string]any{"run_id": "run_runtime_event_dedupe", "mapped_status": "run_active"}
		for i := 0; i < 2; i++ {
			if err := appendRunEventIndex(ctx, "proj_runtime_event_dedupe", "bind_runtime_event_dedupe", "run.active", "info", "run is active", payload); err != nil {
				t.Fatalf("appendRunEventIndex failed: %v", err)
			}
		}

		var count int
		mustQueryRow(t, db, `SELECT COUNT(*) FROM run_event_index WHERE run_binding_id = ?`, "bind_runtime_event_dedupe").Scan(&count)
		if count != 1 {
			t.Fatalf("duplicate run event count = %d, want 1", count)
		}
	})
}

func TestAppendRunCheckpointForStateDeduplicatesUnchangedState(t *testing.T) {
	withAcceptanceFlowDB(t, func(ctx context.Context, db *sql.DB) {
		seedRuntimeTask(t, ctx, db, "proj_runtime_checkpoint", "task_runtime_checkpoint")
		seedRuntimeBinding(t, ctx, "proj_runtime_checkpoint", "task_runtime_checkpoint", "bind_runtime_checkpoint", "run_runtime_checkpoint", "run_active")

		binding, err := getBrainRunBindingByID(ctx, "bind_runtime_checkpoint")
		if err != nil {
			t.Fatalf("getBrainRunBindingByID failed: %v", err)
		}
		state := &BrainRunState{
			RunID:       "run_runtime_checkpoint",
			ExecutionID: "run_runtime_checkpoint",
			Status:      "running",
			Brain:       "coder",
		}
		for i := 0; i < 2; i++ {
			if err = appendRunCheckpointForState(ctx, binding, state, "run_active"); err != nil {
				t.Fatalf("appendRunCheckpointForState failed: %v", err)
			}
		}

		var count int
		mustQueryRow(t, db, `SELECT COUNT(*) FROM run_checkpoints WHERE run_binding_id = ?`, "bind_runtime_checkpoint").Scan(&count)
		if count != 1 {
			t.Fatalf("checkpoint count = %d, want 1", count)
		}

		state.Status = "succeeded"
		if err = appendRunCheckpointForState(ctx, binding, state, "run_succeeded"); err != nil {
			t.Fatalf("append changed checkpoint failed: %v", err)
		}
		mustQueryRow(t, db, `SELECT COUNT(*) FROM run_checkpoints WHERE run_binding_id = ?`, "bind_runtime_checkpoint").Scan(&count)
		if count != 2 {
			t.Fatalf("checkpoint count after changed state = %d, want 2", count)
		}
	})
}

func seedRuntimeTask(t *testing.T, ctx context.Context, db *sql.DB, projectID string, taskID string) {
	t.Helper()

	now := "2026-04-24T10:00:00Z"
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		t.Fatalf("begin runtime seed transaction failed: %v", err)
	}
	if err = insertProjectRow(ctx, tx, &do.Projects{
		Id:               projectID,
		Name:             "Runtime Project",
		ProjectCategory:  "web",
		GoalSummary:      "Runtime seed state",
		Status:           "running",
		ProductionStatus: "pending",
		WorkspaceRoot:    filepath.Join("/tmp", projectID),
		CreatedAt:        now,
		UpdatedAt:        now,
	}); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert runtime project failed: %v", err)
	}
	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO domain_tasks (
id, project_id, source_compiled_plan_id, source_compiled_task_id, source_task_key, compiled_version, name, phase, task_kind, role_type, brain_kind, risk_level, status, manual_review_required, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskID,
		projectID,
		"",
		"",
		"",
		1,
		"Runtime Task",
		"implementation",
		"code",
		"executor",
		"coder",
		"medium",
		"running",
		0,
		now,
		now,
	); err != nil {
		_ = tx.Rollback()
		t.Fatalf("insert runtime task failed: %v", err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatalf("commit runtime seed transaction failed: %v", err)
	}
}

func seedRuntimeBinding(t *testing.T, ctx context.Context, projectID string, taskID string, bindingID string, runID string, status string) {
	t.Helper()

	now := "2026-04-24T10:01:00Z"
	if err := insertBrainRunBinding(ctx, &do.BrainRunBindings{
		Id:         bindingID,
		ProjectId:  projectID,
		TaskId:     taskID,
		BrainKind:  "coder",
		BrainRunId: runID,
		RunStatus:  status,
		StartedAt:  now,
		LastSyncAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("insert runtime binding failed: %v", err)
	}
}
