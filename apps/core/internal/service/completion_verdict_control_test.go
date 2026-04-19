package service

import (
	"testing"

	projectsv1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func TestAcceptanceCompletionVerdictStopsAtManualCheckpoint(t *testing.T) {
	t.Parallel()

	data := &acceptanceAggregate{
		Project: entity.Projects{Id: "proj_1"},
		LatestAcceptanceRun: &entity.AcceptanceRuns{
			Id:                    "acc_1",
			ProductionStatus:      "production_passed",
			FunctionalStatus:      "production_passed",
			CreatedAt:             "2026-04-20 10:00:00",
			ManualReleaseRequired: 0,
		},
		Issues: []entity.AcceptanceIssues{
			{
				Id:       "issue_1",
				Blocking: 1,
				Summary:  "human review is still required",
			},
		},
	}

	got := buildCompletionVerdictView(data)
	if got.Completed {
		t.Fatalf("completion verdict should not complete when manual review is required: %#v", got)
	}
	if got.Decision != "manual_checkpoint" {
		t.Fatalf("unexpected decision: got %s want %s", got.Decision, "manual_checkpoint")
	}
	if got.FinalStatus != "accepting" {
		t.Fatalf("unexpected final status: got %s want %s", got.FinalStatus, "accepting")
	}
}

func TestWorkspaceCompletionVerdictStopsAtRuntimeEscalation(t *testing.T) {
	t.Parallel()

	data := &projectWorkspaceAggregate{
		Project: entity.Projects{
			Id:        "proj_1",
			UpdatedAt: "2026-04-20 10:00:00",
		},
		LatestAcceptanceRun: &entity.AcceptanceRuns{
			Id:               "acc_1",
			ProductionStatus: "production_passed",
			FunctionalStatus: "production_passed",
			CreatedAt:        "2026-04-20 10:00:00",
			FinishedAt:       "2026-04-20 10:05:00",
		},
		RunBindings: []entity.BrainRunBindings{
			{
				Id:         "bind_1",
				TaskId:     "task_1",
				BrainRunId: "run_1",
				RunStatus:  "run_failed",
				UpdatedAt:  "2026-04-20 10:04:00",
			},
		},
	}

	got := buildWorkspaceCompletionVerdict(
		data,
		projectsv1.AcceptanceCoverage{
			EvidenceReady:    3,
			EvidenceRequired: 3,
			ProductionPassed: true,
		},
	)
	if got.Completed {
		t.Fatalf("workspace completion verdict should not complete when runtime escalation exists: %#v", got)
	}
	if got.Decision != "manual_checkpoint" {
		t.Fatalf("unexpected decision: got %s want %s", got.Decision, "manual_checkpoint")
	}
	if got.NextAction != "resolve_runtime_escalation" {
		t.Fatalf("unexpected next action: got %s want %s", got.NextAction, "resolve_runtime_escalation")
	}
}
