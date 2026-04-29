package service

import (
	"context"
	"testing"

	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func TestBuildVerificationResultViewMarksIncompleteWhenEvidenceMissing(t *testing.T) {
	t.Parallel()

	view := buildVerificationResultView(context.Background(), &acceptanceAggregate{
		Project: entity.Projects{
			Id:              "proj_acceptance_incomplete",
			ProjectCategory: "web",
			UpdatedAt:       "2026-04-20T00:00:00Z",
		},
		ProductionProfile: &productionAcceptanceProfileRecord{
			ProfileVersion:          "v3",
			RequiredEvidenceJSON:    `["ui_capture","api_trace"]`,
			ReleaseRequirementsJSON: `["acceptance_profile","coverage_matrix","evidence_artifacts"]`,
		},
		LatestAcceptanceRun: &entity.AcceptanceRuns{
			Id:         "run_incomplete",
			Status:     "running",
			FinishedAt: "2026-04-20T00:01:00Z",
		},
		EvidenceItems: []entity.EvidenceItems{
			{Id: "evidence_1"},
		},
		ChannelAvailable:     true,
		EnvironmentAvailable: true,
	})

	if view.Status != "incomplete" {
		t.Fatalf("verification status = %q, want incomplete", view.Status)
	}
	if view.Decision != "collect_evidence" {
		t.Fatalf("verification decision = %q, want collect_evidence", view.Decision)
	}
	if len(view.MissingEvidence) != 1 || view.MissingEvidence[0] != "api_trace" {
		t.Fatalf("missing evidence = %#v, want [\"api_trace\"]", view.MissingEvidence)
	}
	if view.Completed {
		t.Fatal("verification should not be completed when evidence is missing")
	}
}

func TestBuildCompletionVerdictViewRequiresManualCheckpointOnRuntimeEscalation(t *testing.T) {
	t.Parallel()

	verdict := buildCompletionVerdictView(&acceptanceAggregate{
		Project: entity.Projects{
			Id:              "proj_acceptance_runtime",
			ProjectCategory: "web",
			UpdatedAt:       "2026-04-20T00:00:00Z",
		},
		LatestAcceptanceRun: &entity.AcceptanceRuns{
			Id:               "run_passed",
			Status:           "completed",
			FunctionalStatus: "functional_passed",
			ProductionStatus: "production_passed",
			FinishedAt:       "2026-04-20T00:02:00Z",
		},
		RunBindings: []entity.BrainRunBindings{
			{
				Id:         "binding_denied",
				TaskId:     "task_secure_repo",
				BrainKind:  "coder",
				BrainRunId: "run_brain_1",
				RunStatus:  "run_denied",
				LastSyncAt: "2026-04-20T00:03:00Z",
			},
		},
		ChannelAvailable:     true,
		EnvironmentAvailable: true,
	})

	if verdict.Decision != "manual_checkpoint" {
		t.Fatalf("completion decision = %q, want manual_checkpoint", verdict.Decision)
	}
	if verdict.NextAction != "resolve_runtime_escalation" {
		t.Fatalf("completion next action = %q, want resolve_runtime_escalation", verdict.NextAction)
	}
	if verdict.Completed {
		t.Fatal("completion verdict should not be completed while runtime escalation remains")
	}
}

func TestBuildCompletionVerdictViewCompletesWhenVerificationIsClean(t *testing.T) {
	t.Parallel()

	verdict := buildCompletionVerdictView(&acceptanceAggregate{
		Project: entity.Projects{
			Id:              "proj_acceptance_complete",
			ProjectCategory: "web",
			UpdatedAt:       "2026-04-20T00:00:00Z",
		},
		LatestAcceptanceRun: &entity.AcceptanceRuns{
			Id:               "run_clean",
			Status:           "completed",
			FunctionalStatus: "functional_passed",
			ProductionStatus: "production_passed",
			FinishedAt:       "2026-04-20T00:04:00Z",
		},
		ChannelAvailable:     true,
		EnvironmentAvailable: true,
	})

	if verdict.Decision != "complete" {
		t.Fatalf("completion decision = %q, want complete", verdict.Decision)
	}
	if verdict.FinalStatus != "completed" {
		t.Fatalf("completion final status = %q, want completed", verdict.FinalStatus)
	}
	if !verdict.Completed || !verdict.ReleaseReady {
		t.Fatalf("expected completion verdict to be release-ready: %#v", verdict)
	}
}
