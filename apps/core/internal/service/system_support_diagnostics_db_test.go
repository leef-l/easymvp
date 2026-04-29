package service

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"

	replayv1 "github.com/leef-l/easymvp/apps/core/api/replay/v1"
	systemv1 "github.com/leef-l/easymvp/apps/core/api/system/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func TestBuildProjectDiagnosticItemExtractsIDs(t *testing.T) {
	t.Parallel()

	item := buildProjectDiagnosticItem(entity.DiagnosticRecords{
		Id:         "diag_1",
		Scope:      "runtime.sync_run",
		Severity:   "warning",
		ErrorCode:  "RUN_002",
		Summary:    "sync failed",
		DetailJson: `{"project_id":"proj_1","task_id":"task_1","run_id":"run_1","binding_id":"bind_1"}`,
		CreatedAt:  "2026-04-19T10:00:00Z",
	})

	if item.ProjectID != "proj_1" || item.TaskID != "task_1" || item.RunID != "run_1" || item.BindingID != "bind_1" {
		t.Fatalf("unexpected extracted diagnostic ids: %#v", item)
	}
}

func TestDiagnosticMatchesProjectUsesStructuredAndRawFallback(t *testing.T) {
	t.Parallel()

	if !diagnosticMatchesProject(systemv1.ProjectDiagnosticItem{ProjectID: "proj_1"}, "proj_1") {
		t.Fatalf("expected structured project id to match")
	}
	if !diagnosticMatchesProject(systemv1.ProjectDiagnosticItem{DetailJSON: `{"project_id":"proj_2"}`}, "proj_2") {
		t.Fatalf("expected raw detail json fallback to match")
	}
	if diagnosticMatchesProject(systemv1.ProjectDiagnosticItem{ProjectID: "proj_3"}, "proj_9") {
		t.Fatalf("unexpected project match")
	}
}

func TestBuildProjectDiagnosticItemClassifiesStructuredDiagnostics(t *testing.T) {
	t.Parallel()

	item := buildProjectDiagnosticItem(entity.DiagnosticRecords{
		Id:         "diag_2",
		Scope:      "runtime.sync_run",
		Severity:   "warning",
		ErrorCode:  "verification_conflict",
		Summary:    "verification contract mismatch",
		DetailJson: `{"project_id":"proj_1","component":"acceptance","field":"verification_contract","missing_evidence":["screen-recording"]}`,
		CreatedAt:  "2026-04-20T10:00:00Z",
	})

	if item.Category != "verification_conflict" {
		t.Fatalf("unexpected category: got %s", item.Category)
	}
	if item.Component != "acceptance" {
		t.Fatalf("unexpected component: got %s", item.Component)
	}
	if item.Field != "verification_contract" {
		t.Fatalf("unexpected field: got %s", item.Field)
	}
	if item.RelatedPage != "acceptance" {
		t.Fatalf("unexpected related page: got %s", item.RelatedPage)
	}
}

func TestMapProjectArtifactContextIncludesRunAndAction(t *testing.T) {
	t.Parallel()

	got := mapProjectArtifactContext("run_1", replayv1.ReplayArtifactIssue{
		Kind:              "stale_index",
		Source:            "replay",
		ID:                "replay_1",
		Status:            "available",
		FilePath:          "/tmp/replay.json",
		Summary:           "step",
		RecommendedAction: "refresh_replay_artifact_index",
	})

	if got.RunID != "run_1" || got.Kind != "stale_index" || got.RecommendedAction != "refresh_replay_artifact_index" {
		t.Fatalf("unexpected artifact context: %#v", got)
	}
}

func TestBuildProjectEvidenceOverviewCountsAndFlagsMissingFiles(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	defer db.Close()
	if _, err = db.Exec(`CREATE TABLE evidence_items (id TEXT PRIMARY KEY, project_id TEXT NOT NULL)`); err != nil {
		t.Fatalf("create evidence table failed: %v", err)
	}
	if _, err = db.Exec(`INSERT INTO evidence_items (id, project_id) VALUES ('ev_1', 'proj_1'), ('ev_2', 'proj_1')`); err != nil {
		t.Fatalf("insert evidence rows failed: %v", err)
	}

	existingPath := filepath.Join(t.TempDir(), "screen.png")
	if err = os.WriteFile(existingPath, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write evidence file failed: %v", err)
	}
	missingPath := filepath.Join(t.TempDir(), "missing.png")
	overview, err := buildProjectEvidenceOverview(context.Background(), db, &acceptanceAggregate{
		Project: entity.Projects{Id: "proj_1"},
		EvidenceItems: []entity.EvidenceItems{
			{Id: "ev_1", ProjectId: "proj_1", EvidenceType: "screenshot", FilePath: existingPath, CapturedAt: "2026-04-20T10:00:00Z"},
			{Id: "ev_2", ProjectId: "proj_1", EvidenceType: "screenshot", FilePath: missingPath, CapturedAt: "2026-04-20T10:01:00Z"},
		},
	})
	if err != nil {
		t.Fatalf("build evidence overview failed: %v", err)
	}
	if overview.TotalCount != 2 {
		t.Fatalf("unexpected total count: got %d want 2", overview.TotalCount)
	}
	if len(overview.MissingFiles) != 1 || overview.MissingFiles[0].ID != "ev_2" {
		t.Fatalf("unexpected missing evidence files: %#v", overview.MissingFiles)
	}
}
