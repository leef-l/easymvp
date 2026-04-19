package service

import (
	"testing"

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
